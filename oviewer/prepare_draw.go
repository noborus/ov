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

const (
	// sectionTimeOut is the section header search timeout(in milliseconds) period.
	sectionTimeOut time.Duration = 1000
	// rulerHeight is the height of the ruler.
	rulerHeight = 2
)

// prepareScreen prepares when the screen size is changed.
func (root *Root) prepareScreen() {
	root.scr.vWidth, root.scr.vHeight = root.Screen.Size()
	// Do not allow very small screens.
	root.scr.vWidth = max(root.scr.vWidth, 2)
	root.scr.vHeight = max(root.scr.vHeight, 2)
	root.updateDocumentSize()

	root.scr.rulerHeight = 0
	if root.Doc.RulerType != RulerNone {
		root.scr.rulerHeight = rulerHeight
	}
	root.Doc.startY = root.scr.rulerHeight

	n, err := calculatePosition(root.Doc.HScrollWidth, root.scr.vWidth)
	if err != nil {
		root.setMessageLogf("Invalid HScrollWidth: %s", root.Doc.HScrollWidth)
		n = 0
	}
	num := int(math.Round(n))
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
	m := root.Doc
	m.bodyStartX = 0
	m.leftMargin = root.sidebarWidth
	m.rightMargin = 0
	m.lineNumberWidth = 0
	if m.LineNumMode {
		if m.parent != nil {
			m = m.parent
		}
		m.lineNumberWidth = len(strconv.Itoa(m.BufEndNum())) + 1
	}
	m.bodyStartX = m.leftMargin + m.lineNumberWidth
}

// updateDocumentSize updates the document size.
func (root *Root) updateDocumentSize() {
	m := root.Doc
	m.width = root.scr.vWidth - m.bodyStartX
	m.bodyWidth = root.scr.vWidth - (m.bodyStartX + m.rightMargin)
	m.height = root.scr.vHeight - root.scr.statusLineHeight
	m.statusPos = m.height
}

// ViewSync redraws the whole thing.
func (root *Root) ViewSync(context.Context) {
	root.resetSelect()
	root.prepareStartX()
	root.prepareScreen()
	root.Screen.Sync()
	root.Doc.jumpTargetHeight, root.Doc.jumpTargetSection = jumpPosition(root.Doc.JumpTarget, root.scr.vHeight)
}

// determineStatusLine determines the height of the status line.
// If the status line is enabled, it returns DefaultStatusLine.
func (root *Root) determineStatusLine() int {
	if root.Doc.StatusLine {
		return DefaultStatusLine
	}
	if len(root.message) > 0 {
		return DefaultStatusLine
	}
	if root.input.Event.Mode() == Normal {
		return 0
	}
	return DefaultStatusLine
}

// prepareDraw prepares the screen for drawing.
func (root *Root) prepareDraw(ctx context.Context) {
	root.scr.statusLineHeight = root.determineStatusLine()
	root.updateDocumentSize()
	root.prepareSidebarItems()
	// Set the columnCursor at the first run.
	if len(root.scr.lines) == 0 {
		defer func() {
			root.Doc.columnCursor = root.Doc.optimalCursor(root.scr, root.Doc.columnCursor)
		}()
	}

	// Header.
	root.scr.headerLN = root.Doc.SkipLines
	root.scr.headerEnd = root.Doc.firstLine()
	// Set the header height.
	root.Doc.headerHeight = min(root.scr.vHeight, root.Doc.getHeight(root.scr.headerLN, root.scr.headerEnd)+root.scr.rulerHeight)

	// Section header.
	root.scr.sectionHeaderLN = -1
	root.scr.sectionHeaderEnd = 0
	root.Doc.sectionHeaderHeight = 0
	if root.Doc.SectionHeader {
		root.scr.sectionHeaderLN, root.scr.sectionHeaderEnd = root.sectionHeader(ctx, root.Doc.topLN+root.scr.headerEnd-root.Doc.SectionStartPosition, sectionTimeOut)
		// Set the section header height.
		root.Doc.sectionHeaderHeight = min(root.scr.vHeight, root.Doc.getHeight(root.scr.sectionHeaderLN, root.scr.sectionHeaderEnd))
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
	if root.Doc.Converter == convAlign {
		root.setAlignConverter()
	}
	root.scr.bodyLN = root.Doc.topLN + root.Doc.firstLine()
	root.scr.bodyEnd = root.scr.bodyLN + root.scr.vHeight // vHeight is the max line of logical lines.

	// Prepare the lines.
	root.scr.lines = root.prepareLines(root.scr.lines)

	root.Doc.columnStart = determineColumnStart(root.scr.lines)
}

// determineColumnStart determines the start index of the column.
// If the column starts with a delimiter, it returns 1. Otherwise, it returns 0.
// This function checks all lines to ensure that a CSV file starting with a comma
// is correctly interpreted as having an empty first value.
func determineColumnStart(lines map[int]LineC) int {
	start := 0
	for _, lineC := range lines {
		if !lineC.valid {
			continue
		}
		columns := lineC.columnRanges
		if len(columns) > 0 {
			if columns[0].end > 0 {
				return 0
			}
			start = 1
		}
	}
	return start
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

	if slices.Equal(m.alignConv.maxWidths, maxWidths) {
		return
	}
	m.alignConv.orgWidths = m.columnWidths
	m.alignConv.maxWidths = maxWidths
	// column attributes are inherited, so only the required columns are added.
	for n := len(m.alignConv.columnAttrs); n < len(maxWidths)+1; n++ {
		m.alignConv.columnAttrs = append(m.alignConv.columnAttrs, columnAttribute{})
	}
	for i := range addRight {
		if !m.alignConv.columnAttrs[i].rightAlign {
			m.alignConv.columnAttrs[i].rightAlign = addRight[i] > 1
		}
	}
	m.ClearCache()
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
		return maxWidthsWidth(lc, maxWidths, rightCount, m.columnWidths)
	}
	return maxWidthsDelm(lc, maxWidths, rightCount, m.ColumnDelimiter, m.ColumnDelimiterReg)
}

// maxWidthsDelm returns the maximum width of the column.
func maxWidthsDelm(lc contents, maxWidths []int, rightCount []int, delimiter string, delimiterReg *regexp.Regexp) ([]int, []int) {
	str, pos := ContentsToStr(lc)
	indexes := allIndex(str, delimiter, delimiterReg)
	if len(indexes) == 0 {
		return maxWidths, rightCount
	}

	start := pos.x(0)
	for i := range indexes {
		end := pos.x(indexes[i][0])
		delmEnd := pos.x(indexes[i][1])
		width := end - start
		maxWidths, rightCount = updateMaxWidths(maxWidths, rightCount, i, width, 0)
		start = delmEnd
	}
	// The last column.
	if start < len(lc) {
		width := len(lc) - start
		maxWidths, rightCount = updateMaxWidths(maxWidths, rightCount, len(indexes), width, 0)
	}
	return maxWidths, rightCount
}

// maxWidthsWidth returns the maximum width of the column.
func maxWidthsWidth(lc contents, maxWidths []int, rightCount []int, widths []int) ([]int, []int) {
	if len(widths) == 0 {
		return maxWidths, rightCount
	}
	start := 0
	for i := range widths {
		end := min(findColumnEnd(lc, widths, i, start)+1, len(lc))
		if start > end {
			break
		}
		width, addRight := trimWidth(lc[start:end])
		maxWidths, rightCount = updateMaxWidths(maxWidths, rightCount, i, width, addRight)
		start = end
	}
	// The last column.
	if start < len(lc) {
		width, addRight := trimWidth(lc[start:])
		maxWidths, rightCount = updateMaxWidths(maxWidths, rightCount, len(widths), width, addRight)
	}
	return maxWidths, rightCount
}

// updateMaxWidths updates the maximum width and right count.
func updateMaxWidths(maxWidths []int, rightCount []int, idx int, width int, addRight int) ([]int, []int) {
	if len(maxWidths) <= idx {
		maxWidths = append(maxWidths, width)
	} else {
		maxWidths[idx] = max(width, maxWidths[idx])
	}
	if len(rightCount) <= idx {
		rightCount = append(rightCount, addRight)
	} else {
		rightCount[idx] += addRight
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

// trimmedIndices returns the start and end of the trimmed contents.
func trimmedIndices(lc contents) (int, int) {
	ts := 0
	for i := range lc {
		if !lc.IsSpace(i) {
			ts = i
			break
		}
	}
	te := len(lc)
	for i := len(lc) - 1; i >= 0; i-- {
		if !lc.IsSpace(i) {
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
	lines = root.setLines(lines, root.scr.bodyLN, root.scr.bodyEnd)

	if root.Doc.SectionHeader {
		lines = root.sectionNum(lines)
	}
	return lines
}

// setLines sets the lines of the specified range.
func (root *Root) setLines(lines map[int]LineC, startLN int, endLN int) map[int]LineC {
	for lN := startLN; lN < endLN; lN++ {
		lines[lN] = root.lineContent(lN)
	}
	return lines
}

// lineContent returns the contents of the specified line.
func (root *Root) lineContent(lN int) LineC {
	m := root.Doc
	lineC := m.getLineC(lN)
	if !lineC.valid {
		return lineC
	}
	if m.ColumnMode {
		lineC = m.columnRanges(lineC)
	}
	RangeStyle(lineC.lc, 0, len(lineC.lc), m.Style.Body)
	root.styleContent(lineC)
	return lineC
}

// sectionNum sets the section number.
func (root *Root) sectionNum(lines map[int]LineC) map[int]LineC {
	m := root.Doc
	if m.SectionDelimiter == "" {
		return lines
	}
	if m.SectionDelimiterReg == nil {
		log.Printf("Regular expression is not set: %s\n", m.SectionDelimiter)
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

		lineC := lines[lN]
		lineC.section = section
		lineC.sectionNm = num
		lines[lN] = lineC
		num++
	}

	return lines
}

// lineNumbers returns the line numbers.
func lineNumbers(lines map[int]LineC) []int {
	result := make([]int, 0, len(lines))
	for k := range lines {
		result = append(result, k)
	}
	sort.Ints(result)
	return result
}

// styleContent applies the style of the content.
func (root *Root) styleContent(lineC LineC) {
	if root.Doc.PlainMode {
		root.plainStyle(lineC.lc)
	}
	if root.Doc.ColumnMode {
		root.columnHighlight(lineC)
	}
	root.multiColorHighlight(lineC)
	root.searchHighlight(lineC)
}

// plainStyle defaults to the original style.
func (*Root) plainStyle(lc contents) {
	for x := range lc {
		lc[x].style = defaultStyle
	}
}

// columnHighlight applies the style of the column highlight.
func (root *Root) columnHighlight(lineC LineC) {
	root.applyColumnStyles(lineC)
}

// multiColorHighlight applies styles to multiple words (regular expressions) individually.
// The style of the first specified word takes precedence.
func (root *Root) multiColorHighlight(lineC LineC) {
	numC := len(root.Doc.Style.MultiColorHighlight)
	for i := len(root.Doc.multiColorRegexps) - 1; i >= 0; i-- {
		indexes := searchPositionReg(lineC.str, root.Doc.multiColorRegexps[i])
		for _, idx := range indexes {
			RangeStyle(lineC.lc, lineC.pos.x(idx[0]), lineC.pos.x(idx[1]), root.Doc.Style.MultiColorHighlight[i%numC])
		}
	}
}

// searchHighlight applies the style of the search highlight.
// Apply style to contents.
func (root *Root) searchHighlight(lineC LineC) {
	if root.searcher == nil || root.searcher.String() == "" {
		return
	}

	indexes := root.searchPosition(lineC.str)
	for _, idx := range indexes {
		RangeStyle(lineC.lc, lineC.pos.x(idx[0]), lineC.pos.x(idx[1]), root.Doc.Style.SearchHighlight)
	}
}

// columnRanges sets the column ranges.
func (m *Document) columnRanges(lineC LineC) LineC {
	if m.ColumnWidth {
		lineC.columnRanges = m.columnWidthRanges(lineC)
	} else {
		lineC.columnRanges = m.columnDelimiterRange(lineC)
	}
	return lineC
}

// columnDelimiterRange returns the ranges of the columns.
func (m *Document) columnDelimiterRange(lineC LineC) []columnRange {
	indexes := allIndex(lineC.str, m.ColumnDelimiter, m.ColumnDelimiterReg)
	if len(indexes) == 0 {
		return nil
	}

	var columnRanges []columnRange
	start := lineC.pos.x(0)
	for columnNum := range indexes {
		end := lineC.pos.x(indexes[columnNum][0])
		delmEnd := lineC.pos.x(indexes[columnNum][1])
		columnRanges = append(columnRanges, columnRange{start: start, end: end})
		start = delmEnd
	}
	// The last column.
	end := lineC.pos.x(len(lineC.str))
	if start < end {
		columnRanges = append(columnRanges, columnRange{start: start, end: end})
	}
	return columnRanges
}

// columnWidthRanges returns the ranges of the columns.
func (m *Document) columnWidthRanges(lineC LineC) []columnRange {
	indexes := m.columnWidths
	if len(indexes) == 0 {
		return nil
	}

	var columnRanges []columnRange
	start, end := 0, 0
	for c := range len(indexes) + 1 {
		if m.Converter == convAlign {
			end = alignColumnEnd(lineC.lc, m.alignConv.maxWidths, c, start)
		} else {
			end = findColumnEnd(lineC.lc, indexes, c, start)
		}
		if start > end {
			break
		}
		columnRanges = append(columnRanges, columnRange{start: start, end: end})
		start = end + 1
	}
	return columnRanges
}

// applyColumnStyles applies the styles to the columns based on the calculated ranges.
func (root *Root) applyColumnStyles(lineC LineC) {
	m := root.Doc
	numC := len(m.Style.ColumnRainbow)

	for c, colRange := range lineC.columnRanges {
		if m.ColumnRainbow {
			RangeStyle(lineC.lc, colRange.start, colRange.end, m.Style.ColumnRainbow[c%numC])
		}
		if c == m.columnCursor {
			RangeStyle(lineC.lc, colRange.start, colRange.end, m.Style.ColumnHighlight)
		}
	}
}

// findColumnEnd returns the position of the end of a column.
func findColumnEnd(lc contents, indexes []int, n int, start int) int {
	// If the column index is out of bounds, return the length of the contents.
	if len(indexes) <= n {
		return len(lc)
	}

	columnEnd := indexes[n]
	// The end of the right end returns the length of the contents.
	if columnEnd >= len(lc) {
		return len(lc)
	}

	// If the character at the end of the column is a space, return the end of the column.
	if lc.IsSpace(columnEnd) {
		return columnEnd
	}

	// Column overflow.
	lPos := findPrevSpace(lc, columnEnd)
	rPos := findNextSpace(lc, columnEnd)

	// Is the column omitted?
	if lPos == columnEnd {
		return rPos
	}

	// Check the position of the next column.
	if n < len(indexes)-1 {
		// If the next column is reached, return the position shifted to the left.
		if rPos == indexes[n+1] {
			return lPos
		}
		// If the position is between the end of the column and the start of the next column,
		// return the position shifted to the left.
		if lPos > columnEnd && lPos < indexes[n+1] {
			return lPos
		}
	}

	// It overflows on the left side.
	lCount := columnEnd - lPos
	rCount := rPos - columnEnd
	if lPos > start && rCount > lCount {
		return lPos
	}

	// It overflows on the right side.
	return rPos
}

// findPrevSpace returns the position just before the previous space.
func findPrevSpace(lc contents, start int) int {
	for p := start; p > 0; p-- {
		if lc.IsSpace(p) {
			return p
		}
	}
	return 0
}

// findNextSpace returns the position of the next space or the end of contents.
func findNextSpace(lc contents, start int) int {
	for p := start; p < len(lc); p++ {
		if lc.IsSpace(p) {
			return p
		}
	}
	return len(lc)
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
