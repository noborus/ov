package oviewer

import (
	"context"
	"errors"
	"time"
)

// sectionTimeOut is the section header search timeout(in milliseconds) period.
const sectionTimeOut = 1000

// prepareContents prepares the contents.
func (root *Root) prepareContents(ctx context.Context) {
	// Header.
	root.scr.headerLN = root.Doc.SkipLines
	root.scr.headerEnd = root.Doc.firstLine()
	// Section header.
	root.scr.sectionHeaderLN, root.scr.sectionHeaderEnd = root.sectionHeader(ctx)
	// Header and Section header height.
	root.Doc.setHeaderHeight(root.scr)

	if root.Doc.showGotoF {
		root.Doc.shiftSectionHeader(root.scr.sectionHeaderLN)
	}

	root.scr.contents = root.screenContents(root.scr.contents)
}

// sectionHeader sets the header line number.
func (root *Root) sectionHeader(ctx context.Context) (int, int) {
	m := root.Doc
	topLN := m.topLN + m.firstLine()
	// SectionHeader.
	shNum := min(m.SectionHeaderNum, root.scr.vHeight)
	shStart, err := m.searchSectionHeader(ctx, topLN+1)
	if err != nil {
		if errors.Is(err, ErrCancel) {
			root.setMessageLogf("Section header search timed out")
			m.SectionDelimiter = ""
		}
		// If the section header is not found, the section header is not displayed.
		shNum = 0
	}
	return shStart, shStart + shNum
}

// setHeaderHeight sets the header height.
func (m *Document) setHeaderHeight(scr SCR) {
	m.headerHeight = m.getHeight(scr.headerLN, scr.headerEnd)
	m.sectionHeaderHeight = m.getHeight(scr.sectionHeaderLN, scr.sectionHeaderEnd)
}

// shiftSectionHeader shifts the section header so that it is not hidden by it.
// sectionHeaderHeight should be updated before calling this function.
func (m *Document) shiftSectionHeader(lN int) {
	m.showGotoF = false

	if m.jumpTargetHeight != 0 {
		return
	}
	if m.sectionHeaderHeight <= 0 {
		return
	}

	// Move to the first line of the section, if within the section header.
	headerEndLN := (lN - m.firstLine()) + m.SectionHeaderNum
	if m.topLN < headerEndLN {
		m.topLN = lN - m.firstLine()
		return
	}

	// Go to the first line after the section header.
	m.moveYUp(m.sectionHeaderHeight)
}

// searchSectionHeader searches for the section header.
func (m *Document) searchSectionHeader(ctx context.Context, lN int) (int, error) {
	if !m.SectionHeader || m.SectionDelimiter == "" {
		return 0, ErrNoDelimiter
	}

	ctx, cancel := context.WithTimeout(ctx, sectionTimeOut*time.Millisecond)
	defer cancel()
	sLN, err := m.prevSection(ctx, lN)
	if err != nil {
		return 0, err
	}

	if m.firstLine() > sLN {
		return 0, ErrNoMoreSection
	}

	return sLN, nil
}

// getHeight returns the height of the specified range.
func (m *Document) getHeight(startLN int, endLN int) int {
	if !m.WrapMode {
		return max(endLN-startLN, 0)
	}
	height := 0
	for lN := startLN; lN < endLN; lN++ {
		height += len(m.leftMostX(lN))
	}
	return height
}

// screenContents returns the contents of the screen.
func (root *Root) screenContents(contents map[int]LineC) map[int]LineC {
	// clear contents.
	for k := range contents {
		delete(contents, k)
	}
	// Header.
	contents = root.contentsRange(contents, root.scr.headerLN, root.scr.headerEnd)
	// Section header.
	contents = root.contentsRange(contents, root.scr.sectionHeaderLN, root.scr.sectionHeaderEnd)
	// Body.
	startLN := root.Doc.topLN + root.Doc.firstLine()
	endLN := startLN + root.scr.vHeight // vHeight is the max line of logical lines.
	contents = root.contentsRange(contents, startLN, endLN)

	return contents
}

// contentsRange sets the contents of the specified range.
func (root *Root) contentsRange(contents map[int]LineC, startLN int, endLN int) map[int]LineC {
	m := root.Doc
	for lN := startLN; lN < endLN; lN++ {
		line := m.getLineC(lN, m.TabWidth)
		if line.valid {
			RangeStyle(line.lc, 0, len(line.lc), root.StyleBody)
			root.styleContent(line)
		}
		contents[lN] = line
	}
	return contents
}

// styleContent applies the style of the content.
func (root *Root) styleContent(line LineC) {
	if root.Doc.PlainMode {
		root.plainStyle(line.lc)
	}
	if root.Doc.ColumnMode {
		root.columnHighlight(line)
	}
	root.multiColorHighlight(line)
	root.searchHighlight(line)
}

// plainStyle defaults to the original style.
func (*Root) plainStyle(lc contents) {
	for x := 0; x < len(lc); x++ {
		lc[x].style = defaultStyle
	}
}

// columnHighlight applies the style of the column highlight.
func (root *Root) columnHighlight(line LineC) {
	if root.Doc.ColumnWidth {
		root.columnWidthHighlight(line)
		return
	}
	root.columnDelimiterHighlight(line)
}

// multiColorHighlight applies styles to multiple words (regular expressions) individually.
// The style of the first specified word takes precedence.
func (root *Root) multiColorHighlight(line LineC) {
	numC := len(root.StyleMultiColorHighlight)
	for i := len(root.Doc.multiColorRegexps) - 1; i >= 0; i-- {
		indexes := searchPositionReg(line.str, root.Doc.multiColorRegexps[i])
		for _, idx := range indexes {
			RangeStyle(line.lc, line.pos.x(idx[0]), line.pos.x(idx[1]), root.StyleMultiColorHighlight[i%numC])
		}
	}
}

// searchHighlight applies the style of the search highlight.
// Apply style to contents.
func (root *Root) searchHighlight(line LineC) {
	if root.searcher == nil || root.searcher.String() == "" {
		return
	}

	indexes := root.searchPosition(line.str)
	for _, idx := range indexes {
		RangeStyle(line.lc, line.pos.x(idx[0]), line.pos.x(idx[1]), root.StyleSearchHighlight)
	}
}

// columnHighlight applies the style of the column highlight.
func (root *Root) columnDelimiterHighlight(line LineC) {
	m := root.Doc
	indexes := allIndex(line.str, m.ColumnDelimiter, m.ColumnDelimiterReg)
	if len(indexes) == 0 {
		return
	}

	lStart := 0
	if indexes[0][0] == 0 {
		if len(indexes) == 1 {
			return
		}
		lStart = indexes[0][1]
		indexes = indexes[1:]
	}

	numC := len(root.StyleColumnRainbow)

	var iStart, iEnd int
	for c := 0; c < len(indexes)+1; c++ {
		switch {
		case c == 0 && lStart == 0:
			iStart = lStart
			iEnd = indexes[0][1] - len(m.ColumnDelimiter)
			if iEnd < 0 {
				iEnd = 0
			}
		case c < len(indexes):
			iStart = iEnd + 1
			iEnd = indexes[c][0]
		case c == len(indexes):
			iStart = iEnd + 1
			iEnd = len(line.str)
		}
		if iStart < 0 || iEnd < 0 {
			return
		}
		start, end := line.pos.x(iStart), line.pos.x(iEnd)
		if m.ColumnRainbow {
			RangeStyle(line.lc, start, end, root.StyleColumnRainbow[c%numC])
		}
		if c == m.columnCursor {
			RangeStyle(line.lc, start, end, root.StyleColumnHighlight)
		}
	}
}

func (root *Root) columnWidthHighlight(line LineC) {
	m := root.Doc
	indexes := m.columnWidths
	if len(indexes) == 0 {
		return
	}
	iStart, iEnd := 0, 0
	numC := len(root.StyleColumnRainbow)
	for c := 0; c < len(indexes)+1; c++ {
		switch {
		case c == 0:
			iStart = 0
			iEnd = findBounds(line.lc, indexes[0]-1, indexes, c)
		case c < len(indexes):
			iStart = iEnd + 1
			iEnd = findBounds(line.lc, indexes[c], indexes, c)
		case c == len(indexes):
			iStart = iEnd + 1
			iEnd = len(line.str)
		}
		iEnd = min(iEnd, len(line.lc))

		if m.ColumnRainbow {
			RangeStyle(line.lc, iStart, iEnd, root.StyleColumnRainbow[c%numC])
		}

		if c == m.columnCursor {
			RangeStyle(line.lc, iStart, iEnd, root.StyleColumnHighlight)
		}
	}
}

// findBounds finds the bounds of values that extend beyond the column position.
func findBounds(lc contents, p int, pos []int, n int) int {
	if len(lc) <= p {
		return p
	}
	if lc[p].mainc == ' ' {
		return p
	}
	f := p
	fp := 0
	for ; f < len(lc) && lc[f].mainc != ' '; f++ {
		fp++
	}

	b := p
	bp := 0
	for ; b > 0 && lc[b].mainc != ' '; b-- {
		bp++
	}

	if b == pos[n] {
		return f
	}
	if n < len(pos)-1 {
		if f == pos[n+1] {
			return b
		}
		if b == pos[n] {
			return f
		}
		if b > pos[n] && b < pos[n+1] {
			return b
		}
	}
	return f
}

// RangeStyle applies the style to the specified range.
// Apply style to contents.
func RangeStyle(lc contents, start int, end int, s OVStyle) {
	for x := start; x < end; x++ {
		lc[x].style = applyStyle(lc[x].style, s)
	}
}
