package oviewer

import (
	"context"
	"errors"
	"log"
	"math"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"time"
)

// sectionTimeOut is the section header search timeout(in milliseconds) period.
const sectionTimeOut time.Duration = 1000

// prepareScreen prepares when the screen size is changed.
func (root *Root) prepareScreen() {
	root.scr.vWidth, root.scr.vHeight = root.Screen.Size()
	// Do not allow size 0.
	root.scr.vWidth = max(root.scr.vWidth, 1)
	root.scr.vHeight = max(root.scr.vHeight, 1)
	root.Doc.statusPos = root.scr.vHeight - statusLine

	num := int(math.Round(calculatePosition(root.Doc.HScrollWidth, root.scr.vWidth)))
	root.Doc.HScrollWidthNum = max(num, 1)

	if len(root.scr.numbers) < root.scr.vHeight+1 {
		root.scr.numbers = make([]LineNumber, root.scr.vHeight+1)
	}

	if root.scr.lines == nil {
		root.scr.lines = make(map[int]LineC)
	}
}

// prepareStartX prepares the start position of the x.
func (root *Root) prepareStartX() {
	root.scr.startX = 0
	m := root.Doc
	if !m.LineNumMode {
		return
	}

	if m.parent != nil {
		m = m.parent
	}
	root.scr.startX = len(strconv.Itoa(m.BufEndNum())) + 1
}

// ViewSync redraws the whole thing.
func (root *Root) ViewSync(context.Context) {
	root.resetSelect()
	root.prepareStartX()
	root.prepareScreen()
	root.Screen.Sync()
	root.Doc.jumpTargetHeight, root.Doc.jumpTargetSection = jumpPosition(root.Doc.JumpTarget, root.scr.vHeight)
}

// prepareDraw prepares the screen for drawing.
func (root *Root) prepareDraw(ctx context.Context) {
	// Header.
	root.scr.headerLN = root.Doc.SkipLines
	root.scr.headerEnd = root.Doc.firstLine()
	// Set the header height.
	root.Doc.headerHeight = root.Doc.getHeight(root.scr.headerLN, root.scr.headerEnd)

	// Section header.
	root.scr.sectionHeaderLN = -1
	root.scr.sectionHeaderEnd = 0
	root.Doc.sectionHeaderHeight = 0
	if root.Doc.SectionHeader {
		root.scr.sectionHeaderLN, root.scr.sectionHeaderEnd = root.sectionHeader(ctx, root.Doc.topLN+root.scr.headerEnd-root.Doc.SectionStartPosition, sectionTimeOut)
		// Set the section header height.
		root.Doc.sectionHeaderHeight = root.Doc.getHeight(root.scr.sectionHeaderLN, root.scr.sectionHeaderEnd)
	}

	// Shift the initial position of the body to the displayed position.
	if root.Doc.showGotoF {
		fX, fN := root.Doc.topLX, root.Doc.topLN+root.scr.headerEnd
		tX, tN := root.Doc.shiftBody(fX, fN, root.scr.sectionHeaderLN, root.scr.sectionHeaderEnd)
		root.Doc.topLX, root.Doc.topLN = tX, tN-root.scr.headerEnd
		root.Doc.showGotoF = false
	}

	if root.Doc.ColumnWidth && len(root.Doc.columnWidths) == 0 {
		root.Doc.setColumnWidths()
	}

	// Sets alignConv if the converter is align.
	if root.Doc.Converter == alignConv {
		root.setAlignConverter()
	}

	// Prepare the lines.
	root.scr.lines = root.prepareLines(root.scr.lines)
}

// setAlignConverter sets the maximum width of the column.
func (root *Root) setAlignConverter() {
	m := root.Doc

	m.alignConv.WidthF = m.ColumnWidth
	if !m.ColumnWidth {
		root.Doc.alignConv.delimiter = m.ColumnDelimiter
		root.Doc.alignConv.delimiterReg = m.ColumnDelimiterReg
	}

	maxWidths := make([]int, 0, len(m.alignConv.maxWidths))
	addRight := make([]int, 0, len(m.alignConv.maxWidths))
	for ln := root.scr.headerLN; ln < root.scr.headerEnd; ln++ {
		maxWidths, addRight = m.maxColumnWidths(maxWidths, addRight, ln)
	}
	for ln := root.scr.sectionHeaderLN; ln < root.scr.sectionHeaderEnd; ln++ {
		maxWidths, addRight = m.maxColumnWidths(maxWidths, addRight, ln)
	}
	startLN := m.topLN + m.firstLine()
	for ln := startLN; ln < startLN+root.scr.vHeight; ln++ {
		maxWidths, addRight = m.maxColumnWidths(maxWidths, addRight, ln)
	}

	if !slices.Equal(m.alignConv.maxWidths, maxWidths) {
		m.alignConv.orgWidths = m.columnWidths
		m.alignConv.maxWidths = maxWidths
		for n := len(m.alignConv.shrink); n < len(maxWidths)+1; n++ {
			m.alignConv.shrink = append(m.alignConv.shrink, false)
		}
		for n := len(m.alignConv.rightAlign); n < len(maxWidths)+1; n++ {
			m.alignConv.rightAlign = append(m.alignConv.rightAlign, false)
		}
		for i := 0; i < len(addRight); i++ {
			if !m.alignConv.rightAlign[i] {
				m.alignConv.rightAlign[i] = addRight[i] > 1
			}
		}
		m.ClearCache()
	}
}

// maxColumnWidths returns the maximum width of the column.
func (m *Document) maxColumnWidths(maxWidths []int, rightCount []int, lN int) ([]int, []int) {
	if lN < 0 {
		return maxWidths, rightCount
	}
	str, err := m.LineStr(lN)
	if err != nil {
		return maxWidths, rightCount
	}
	lc := StrToContents(str, m.TabWidth)
	if m.ColumnWidth {
		return maxWidthsWidth(lc, maxWidths, m.columnWidths, rightCount)
	}
	return maxWidthsDelm(lc, maxWidths, m.ColumnDelimiter, m.ColumnDelimiterReg), rightCount
}

// maxWidthsDelm returns the maximum width of the column.
func maxWidthsDelm(lc contents, maxWidths []int, delimiter string, delimiterReg *regexp.Regexp) []int {
	str, pos := ContentsToStr(lc)
	indexes := allIndex(str, delimiter, delimiterReg)
	if len(indexes) == 0 {
		return maxWidths
	}
	s := 1
	for i := 0; i < len(indexes); i++ {
		e := pos.x(indexes[i][0])
		e1 := pos.x(indexes[i][1])
		delWidth := e1 - e
		width := e - s
		if len(maxWidths) <= i {
			maxWidths = append(maxWidths, width)
		} else {
			maxWidths[i] = max(width, maxWidths[i])
		}
		s = e + delWidth
	}
	return maxWidths
}

// maxWidthsWidth returns the maximum width of the column.
func maxWidthsWidth(lc contents, maxWidths []int, widths []int, rightCount []int) ([]int, []int) {
	if len(widths) == 0 {
		return maxWidths, rightCount
	}
	s := 0
	for i := 0; i < len(widths); i++ {
		e := findColumnEnd(lc, widths, i) + 1
		width, addRight := trimWidth(lc[s:e])
		if len(maxWidths) <= i {
			maxWidths = append(maxWidths, width)
		} else {
			maxWidths[i] = max(width, maxWidths[i])
		}
		if len(rightCount) <= i {
			rightCount = append(rightCount, addRight)
		} else {
			rightCount[i] += addRight
		}
		s = e
	}
	return maxWidths, rightCount
}

// trimWidth returns the column width and 1 if the column is right-aligned.
func trimWidth(lc contents) (int, int) {
	l, r := trimmedIndices(lc)
	addRight := 0
	if l > len(lc)-r {
		addRight = 1
	}
	return r - l, addRight
}

// trimmedIndices returns the number of left spaces.
func trimmedIndices(lc contents) (int, int) {
	ts := 0
	for i := 0; i < len(lc); i++ {
		if lc[i].mainc != ' ' {
			ts = i
			break
		}
	}
	te := len(lc)
	for i := len(lc) - 1; i >= 0; i-- {
		if lc[i].mainc != ' ' {
			te = i + 1
			break
		}
	}
	return ts, te
}

// shiftBody shifts the section header so that it is not hidden by it.
func (m *Document) shiftBody(lX int, lN int, shStart int, shEnd int) (int, int) {
	if m.jumpTargetHeight != 0 {
		return lX, lN
	}
	if m.sectionHeaderHeight <= 0 {
		return lX, lN
	}

	// Move to the first line of the section, if within the section header.
	if lN < shEnd {
		return lX, shStart
	}

	// Go to the first line after the section header.
	if !m.WrapMode {
		return lX, lN - m.sectionHeaderHeight
	}
	return m.numUp(lX, lN, m.sectionHeaderHeight)
}

// sectionHeader returns the start and end of the section header.
func (root *Root) sectionHeader(ctx context.Context, topLN int, timeOut time.Duration) (int, int) {
	m := root.Doc

	// SectionHeader.
	shNum := min(m.SectionHeaderNum, root.scr.vHeight)
	shStart, err := m.searchSectionHeader(ctx, topLN+1, timeOut)
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

// searchSectionHeader searches for the section header.
func (m *Document) searchSectionHeader(ctx context.Context, lN int, timeOut time.Duration) (int, error) {
	if !m.SectionHeader || m.SectionDelimiter == "" {
		return 0, ErrNoDelimiter
	}

	ctx, cancel := context.WithTimeout(ctx, timeOut*time.Millisecond)
	defer cancel()
	sLN, err := m.prevSection(ctx, lN)
	if err != nil {
		return 0, err
	}

	sLN += m.SectionStartPosition
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

// prepareLines returns the contents of the screen.
func (root *Root) prepareLines(lines map[int]LineC) map[int]LineC {
	// clear lines.
	for k := range lines {
		delete(lines, k)
	}
	// Header.
	lines = root.setLines(lines, root.scr.headerLN, root.scr.headerEnd)
	// Section header.
	lines = root.setLines(lines, root.scr.sectionHeaderLN, root.scr.sectionHeaderEnd)
	// Body.
	startLN := root.Doc.topLN + root.Doc.firstLine()
	endLN := startLN + root.scr.vHeight // vHeight is the max line of logical lines.
	lines = root.setLines(lines, startLN, endLN)

	if root.Doc.SectionHeader {
		lines = root.sectionNum(lines)
	}
	return lines
}

// setLines sets the lines of the specified range.
func (root *Root) setLines(lines map[int]LineC, startLN int, endLN int) map[int]LineC {
	m := root.Doc
	for lN := startLN; lN < endLN; lN++ {
		line := m.getLineC(lN)
		if line.valid {
			RangeStyle(line.lc, 0, len(line.lc), root.StyleBody)
			root.styleContent(line)
		}
		lines[lN] = line
	}
	return lines
}

// sectionNum sets the section number.
func (root *Root) sectionNum(lines map[int]LineC) map[int]LineC {
	m := root.Doc
	if m.SectionDelimiter == "" {
		return lines
	}
	if m.SectionDelimiterReg == nil {
		log.Printf("Regular expression is not set: %s", m.SectionDelimiter)
		return lines
	}
	lNs := lineNumbers(lines)
	num := 1
	section := 0
	for _, lN := range lNs {
		switch {
		case lN < root.scr.sectionHeaderLN:
			// Before the first section header.
			section = 0
		case lN < root.scr.sectionHeaderEnd:
			// The first section header.
			section = 1
			if lN == root.scr.sectionHeaderLN {
				num = 1
			}
		default:
			// After the first section header.
			sp, ok := lines[lN-m.SectionStartPosition]
			if !ok {
				// section starts off screen.
				sp = m.getLineC(lN - m.SectionStartPosition)
			}
			if m.SectionDelimiterReg.MatchString(sp.str) {
				num = 1
				section++
			}
		}

		line := lines[lN]
		line.section = section
		line.sectionNm = num
		lines[lN] = line
		num++
	}

	return lines
}

// lineNumbers returns the line numbers.
func lineNumbers(s map[int]LineC) []int {
	result := make([]int, 0, len(s))
	for k := range s {
		result = append(result, k)
	}
	sort.Ints(result)
	return result
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

	numC := len(root.StyleColumnRainbow)

	lStart := 0
	if indexes[0][0] == 0 {
		if len(indexes) == 1 {
			return
		}
		lStart = indexes[0][1]
		indexes = indexes[1:]
	}

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

// columnWidthHighlight applies the style of the column width highlight.
func (root *Root) columnWidthHighlight(line LineC) {
	m := root.Doc
	indexes := m.columnWidths
	if len(indexes) == 0 {
		return
	}

	numC := len(root.StyleColumnRainbow)

	start := 0
	for c := 0; c < len(indexes)+1; c++ {
		end := 0
		if m.Converter == alignConv {
			end = alignColumnEnd(line.lc, m.alignConv.maxWidths, c, start)
		} else {
			end = findColumnEnd(line.lc, indexes, c)
		}

		if m.ColumnRainbow {
			RangeStyle(line.lc, start, end, root.StyleColumnRainbow[c%numC])
		}
		if c == m.columnCursor {
			RangeStyle(line.lc, start, end, root.StyleColumnHighlight)
		}
		start = end + 1
	}
}

// findColumnEnd returns the position of the end of a column.
func findColumnEnd(lc contents, indexes []int, n int) int {
	if len(indexes) <= n {
		return len(lc)
	}
	width := indexes[n]
	if width >= len(lc) {
		return len(lc)
	}

	if lc[width].mainc == ' ' {
		return width
	}

	f := width
	for ; f < len(lc) && lc[f].mainc != ' '; f++ {
	}
	b := width
	for ; b > 0 && lc[b].mainc != ' '; b-- {
	}

	if b == indexes[n] {
		return f
	}
	if n < len(indexes)-1 {
		if f == indexes[n+1] {
			return b
		}
		if b == indexes[n] {
			return f
		}
		if b > indexes[n] && b < indexes[n+1] {
			return b
		}
	}
	return f
}

// alignColumnEnd returns the position of the end of a column.
func alignColumnEnd(lc contents, widths []int, n int, start int) int {
	if len(widths) <= n {
		return len(lc)
	}
	end := start + widths[n]
	return min(end, len(lc))
}

// RangeStyle applies the style to the specified range.
// Apply style to contents.
func RangeStyle(lc contents, start int, end int, s OVStyle) {
	for x := start; x < end; x++ {
		lc[x].style = applyStyle(lc[x].style, s)
	}
}
