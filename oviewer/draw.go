package oviewer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// defaultStyle is used when the style is not specified.
var defaultStyle = tcell.StyleDefault

var anchorPointStyle = OVStyle{
	Reverse: true,
}

// draw is the main routine that draws the screen.
func (root *Root) draw(ctx context.Context) {
	shouldSync := root.scr.forceDisplaySync
	root.scr.forceDisplaySync = false
	defer func() {
		if shouldSync {
			root.Screen.Sync()
		} else {
			root.Screen.Show()
		}
	}()

	root.prepareDraw(ctx)
	root.drawSidebar()
	root.drawBody()

	root.drawRuler()
	root.drawHeader()
	root.drawSectionHeader()
	root.drawSelect()
	root.drawStatus()
	root.drawTitle()
}

// drawBody draws the body.
// drawBody sets bottomLN and bottomLX of the document.
func (root *Root) drawBody() {
	m := root.Doc
	lX := m.topLX
	lN := m.topLN + root.scr.headerEnd
	if !m.WrapMode { // If WrapMode is off, topLX is always 0.
		lX = 0
	}

	markStyleWidth := min(root.Doc.width, root.Doc.MarkStyleWidth)

	wrapNum := m.numOfWrap(lX, lN)
	for y := m.headerHeight; y < root.scr.vHeight-root.scr.statusLineHeight; y++ {
		lineC, ok := root.scr.lines[lN]
		if !ok {
			log.Panicf("line is not found %d", lN)
		}
		root.scr.numbers[y] = newLineNumber(lN, wrapNum)
		root.drawLineNumber(lN, y, lineC.valid)

		nextLX, nextLN := root.drawLine(y, lX, lN, lineC)
		if root.Doc.SectionHeader {
			root.sectionLineHighlight(y, lineC)
			if root.Doc.HideOtherSection {
				root.hideOtherSection(y, lineC)
			}
		}

		root.drawVerticalHeader(y, wrapNum, lineC)
		if lineC.valid {
			root.coordinatesStyle(lN, y)
		}

		root.applyMarkStyle(lN, y, markStyleWidth)

		wrapNum++
		if nextLX == 0 {
			wrapNum = 0
		}

		lX = nextLX
		lN = nextLN
	}
	m.bottomLN = lN
	m.bottomLX = lX
}

// drawHeader draws header.
func (root *Root) drawHeader() {
	m := root.Doc
	if m.headerHeight == 0 {
		return
	}
	// wrapNum is the number of wrapped lines.
	wrapNum := 0
	lX := 0
	lN := root.scr.headerLN
	for y := root.Doc.startY; y < m.headerHeight && lN < root.scr.headerEnd; y++ {
		lineC, ok := root.scr.lines[lN]
		if !ok {
			log.Panicf("line is not found %d", lN)
		}
		root.scr.numbers[y] = newLineNumber(lN, wrapNum)
		root.blankLineNumber(y)

		lX, lN = root.drawLine(y, lX, lN, lineC)
		root.drawVerticalHeader(y, wrapNum, lineC)
		// header style.
		root.applyStyleToLine(y, m.Style.Header)

		wrapNum++
		if lX == 0 {
			wrapNum = 0
		}
	}
	if root.scr.headerEnd > 0 {
		root.applyStyleToLine(m.headerHeight-1, m.Style.HeaderBorder)
	}
}

// drawSectionHeader draws section header.
// The section header overrides the body.
func (root *Root) drawSectionHeader() {
	m := root.Doc
	if m.sectionHeaderHeight == 0 {
		return
	}
	wrapNum := 0
	lX := 0
	lN := root.scr.sectionHeaderLN
	for y := m.headerHeight; y < m.headerHeight+m.sectionHeaderHeight && lN < root.scr.sectionHeaderEnd; y++ {
		lineC, ok := root.scr.lines[lN]
		if !ok {
			log.Panicf("line is not found %d", lN)
		}
		root.scr.numbers[y] = newLineNumber(lN, wrapNum)
		root.drawLineNumber(lN, y, lineC.valid)

		nextLX, nextLN := root.drawLine(y, lX, lN, lineC)
		// section header style.
		root.applyStyleToLine(y, m.Style.SectionLine)

		root.drawVerticalHeader(y, wrapNum, lineC)
		// markStyle is displayed above the section header.
		markStyleWidth := min(root.scr.vWidth, m.MarkStyleWidth)
		root.applyMarkStyle(lN, y, markStyleWidth)
		// Underline search lines when they overlap in section headers.
		if lN == m.lastSearchLN {
			root.applyStyleToLine(y, OVStyle{Underline: true})
		}

		wrapNum++
		if nextLX == 0 {
			wrapNum = 0
		}

		lX = nextLX
		lN = nextLN
	}
	root.applyStyleToLine(m.headerHeight+m.sectionHeaderHeight-1, m.Style.SectionHeaderBorder)
}

const (
	minusRuler = "0"
	tenRuler   = "         1         2         3         4         5         6         7         8         9         0"
	oneRuler   = "1234567890"
)

// drawRuler draws the ruler.
func (root *Root) drawRuler() {
	rulerType := root.Doc.RulerType
	if rulerType == RulerNone {
		return
	}

	style := applyStyle(defaultStyle, root.Doc.Style.Ruler)

	startX := 0
	offset := 0
	if rulerType == RulerRelative {
		scrollX := 0
		if !root.Doc.WrapMode {
			scrollX = root.Doc.scrollX
		}
		startX = root.Doc.bodyStartX - scrollX
		if startX < 0 {
			offset = -startX
			startX = 0
		}
	}
	root.Screen.PutStrStyled(0, 0, strings.Repeat(" ", startX), style)
	root.Screen.PutStrStyled(0, 1, strings.Repeat(minusRuler, startX), style)
	tensLine := strings.Repeat(tenRuler, (root.scr.vWidth+offset)/100+1)
	onesLine := strings.Repeat(oneRuler, (root.scr.vWidth+offset)/10+1)
	root.Screen.PutStrStyled(startX, 0, tensLine[offset:], style)
	root.Screen.PutStrStyled(startX, 1, onesLine[offset:], style)
}

// drawWrapLine wraps and draws the contents and returns the next drawing position.
func (root *Root) drawLine(y int, lX int, lN int, lineC LineC) (int, int) {
	if root.Doc.WrapMode {
		return root.drawWrapLine(y, lX, lN, lineC)
	}

	return root.drawNoWrapLine(y, root.Doc.scrollX, lN, lineC)
}

// drawWrapLine wraps and draws the contents and returns the next drawing position.
func (root *Root) drawWrapLine(y int, lX int, lN int, lineC LineC) (int, int) {
	if lX < 0 {
		log.Printf("Illegal lX: %d\n", lX)
		return 0, 0
	}
	for n := 0; ; n++ {
		x := root.Doc.bodyStartX + n
		if lX+n >= len(lineC.lc) {
			// EOL
			root.clearEOL(x, y, lineC.eolStyle)
			lX = 0
			lN++
			break
		}
		c := lineC.lc[lX+n]
		if x+c.width > root.Doc.bodyStartX+root.Doc.bodyWidth {
			// Right edge.
			root.clearEOL(x, y, defaultStyle)
			lX += n
			break
		}
		root.put(x, y, c.str, c.style)
		if c.width == 2 {
			n++
		}
	}
	return lX, lN
}

// drawNoWrapLine draws contents without wrapping and returns the next drawing position.
func (root *Root) drawNoWrapLine(y int, lX int, lN int, lineC LineC) (int, int) {
	lX = max(lX, root.minStartX)
	for n := 0; n < root.Doc.bodyWidth; n++ {
		x := root.Doc.bodyStartX + n
		if lX+n >= len(lineC.lc) {
			// EOL
			root.clearEOL(x, y, lineC.eolStyle)
			break
		}
		if lX+n < 0 {
			root.Screen.Put(x, y, " ", defaultStyle)
			continue
		}
		c := lineC.lc[lX+n]
		root.put(x, y, c.str, c.style)
		if c.width == 2 {
			n++
		}
	}
	lN++
	return lX, lN
}

// drawVerticalHeader draws the vertical header.
func (root *Root) drawVerticalHeader(y int, wrapNum int, lineC LineC) {
	if root.Doc.WrapMode && wrapNum > 0 {
		return
	}
	if !lineC.valid {
		return
	}
	widthVH := root.Doc.widthVerticalHeader(lineC)
	if widthVH == 0 {
		return
	}
	if lineC.lc.IsFullWidth(widthVH - 1) {
		widthVH--
	}

	x := root.Doc.bodyStartX
	for n := 0; n < widthVH; n++ {
		c := DefaultContent
		if n < len(lineC.lc) {
			c = lineC.lc[n]
		}
		var style tcell.Style
		if n == widthVH-1 || (n == widthVH-2 && c.width == 2) {
			style = applyStyle(defaultStyle, root.Doc.Style.VerticalHeaderBorder)
		} else {
			style = applyStyle(c.style, root.Doc.Style.VerticalHeader)
		}
		root.put(x, y, c.str, style)
		x++
		if c.width == 2 {
			x++
			n++
		}
	}
}

// widthVerticalHeader calculates the vertical header value.
// If VerticalHeader is specified, it returns that as the width.
// If HeaderColumn is specified, it returns the width of the specified column.
func (m *Document) widthVerticalHeader(lineC LineC) int {
	if m.VerticalHeader > 0 {
		return m.VerticalHeader
	}

	if m.HeaderColumn <= 0 {
		return 0
	}
	vhc := m.HeaderColumn
	columns := lineC.columnRanges
	if len(columns) == 0 {
		return 0
	}
	vhc = min(vhc+m.columnStart, len(columns))
	return columns[vhc-1].end + 1
}

// blankLineNumber should be blank for the line number.
func (root *Root) blankLineNumber(y int) {
	if !root.Doc.LineNumMode {
		return
	}
	if root.Doc.lineNumberWidth <= 0 {
		return
	}
	for x := root.Doc.leftMargin; x < root.Doc.bodyStartX; x++ {
		root.Screen.PutStr(x, y, " ")
	}
}

// drawLineNumber draws the line number.
func (root *Root) drawLineNumber(lN int, y int, valid bool) {
	m := root.Doc
	if !m.LineNumMode {
		return
	}
	if !valid {
		root.blankLineNumber(y)
		return
	}
	if root.Doc.lineNumberWidth <= 0 {
		return
	}

	number := lN
	if m.lineNumMap != nil {
		n, ok := m.lineNumMap.LoadForward(number)
		if ok {
			number = n
		}
	}
	// Line numbers start at 1 except for skip and header lines.
	number = number - m.firstLine() + 1

	style := applyStyle(defaultStyle, m.Style.LineNumber)
	numC := fmt.Sprintf("%*d ", root.Doc.lineNumberWidth-1, number)
	root.Screen.PutStrStyled(root.Doc.leftMargin, y, numC, style)
}

// drawTitle sets the terminal title if TerminalTitle is enabled.
func (root *Root) drawTitle() {
	if root.Config.SetTerminalTitle {
		root.Screen.SetTitle(root.displayTitle())
	}
}

// displayTitle returns the caption string of the document.
func (root *Root) displayTitle() string {
	if root.Doc.Caption != "" {
		return root.Doc.Caption
	}
	return root.Doc.FileName
}

// clearEOL clears from the specified position to the right end.
func (root *Root) clearEOL(x int, y int, style tcell.Style) {
	if x >= root.scr.vWidth {
		return
	}
	root.Screen.PutStrStyled(x, y, strings.Repeat(" ", root.scr.vWidth-x), style)
}

// clearY clear the specified line.
func (root *Root) clearY(y int) {
	root.clearEOL(0, y, defaultStyle)
}

// coordinatesStyle applies the style of the coordinates.
func (root *Root) coordinatesStyle(lN int, y int) {
	if root.Doc.AlternateRows {
		root.applyStyleToAlternate(lN, y)
	}
	if root.Doc.jumpTargetHeight != 0 && root.Doc.headerHeight+root.Doc.jumpTargetHeight == y {
		root.applyStyleToLine(y, root.Doc.Style.JumpTargetLine)
	}
	if root.Doc.pauseFollow && lN == root.Doc.pauseLastNum {
		root.applyStyleToLine(y, root.Doc.Style.PauseLine)
	}
}

// applyStyleToAlternate applies from beginning to end of line.
func (root *Root) applyStyleToAlternate(lN int, y int) {
	if (lN)%2 == 0 {
		return
	}
	root.applyStyleToLine(y, root.Doc.Style.Alternate)
}

// applyStyleToLine applies the style from the left edge to the right edge of the physical line.
// Apply styles to the screen.
func (root *Root) applyStyleToLine(y int, s OVStyle) {
	root.applyStyleToRange(y, s, root.Doc.bodyStartX, root.Doc.bodyStartX+root.Doc.width)
}

// applyMarkStyle applies the style from the left edge to the specified width.
func (root *Root) applyMarkStyle(lN int, y int, width int) {
	m := root.Doc
	if !m.marked.contains(lN) {
		return
	}
	root.applyStyleToRange(y, m.Style.MarkLine, root.Doc.bodyStartX, root.Doc.bodyStartX+width)
}

// applyStyleToRange applies the style from the start to the end of the physical line.
func (root *Root) applyStyleToRange(y int, s OVStyle, start int, end int) {
	for x := start; x < end; x++ {
		str, style, width := root.Screen.Get(x, y)
		newStyle := applyStyle(style, s)
		if style != newStyle {
			root.Screen.Put(x, y, str, newStyle)
		}
		if width == 2 {
			x++
		}
	}
}

// sectionLineHighlight applies the style of the section line highlight.
func (root *Root) sectionLineHighlight(y int, line LineC) {
	if !line.valid {
		return
	}
	if line.section <= 0 {
		return
	}
	if line.sectionNm <= 0 || line.sectionNm > root.Doc.SectionHeaderNum {
		return
	}
	root.applyStyleToLine(y, root.Doc.Style.SectionLine)
}

// hideOtherSection hides other sections.
func (root *Root) hideOtherSection(y int, line LineC) {
	if line.section <= 1 { // 1 is the first section.
		return
	}
	root.clearY(y)
}

// drawSelect highlights the mouse selection.
// Multi-line selection is included until the end of the line,
// but if the rectangle flag is true, the rectangle will be the range.
func (root *Root) drawSelect() {
	sel := root.scr.mouseSelect
	x1 := max(root.scr.x1, root.Doc.bodyStartX)
	y1 := root.scr.y1
	x2 := min(root.scr.x2, root.Doc.bodyStartX+root.Doc.bodyWidth-1)
	y2 := root.scr.y2

	if root.scr.hasAnchorPoint {
		// Highlight the anchor point if the mouse was clicked but no selection is active.
		if x1 == x2 && y1 == y2 {
			root.applyStyleToRange(y1, anchorPointStyle, x1, x2+1)
			return
		}
	}

	if sel == SelectNone {
		return
	}

	if y2 < y1 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}
	if y1 == y2 {
		if x2 < x1 {
			x1, x2 = x2, x1
		}
		root.applySelectionRange(y1, x1, x2+1, sel)
		return
	}
	if root.scr.mouseRectangle {
		for y := y1; y <= y2; y++ {
			root.applySelectionRange(y, x1, x2+1, sel)
		}
		return
	}

	root.applySelectionRange(y1, x1, root.Doc.bodyStartX+root.Doc.bodyWidth, sel)
	for y := y1 + 1; y < y2; y++ {
		root.applySelectionRange(y, root.Doc.bodyStartX, root.Doc.bodyStartX+root.Doc.bodyWidth, sel)
	}
	root.applySelectionRange(y2, root.Doc.bodyStartX, x2+1, sel)
}

// applySelectionRange applies selection style to the specified range.
func (root *Root) applySelectionRange(y int, start int, end int, selectState MouseSelectState) {
	if start >= end {
		return
	}

	var selectStyle OVStyle
	switch selectState {
	case SelectActive:
		selectStyle = root.Doc.Style.SelectActive
	case SelectCopied:
		selectStyle = root.Doc.Style.SelectCopied
	default:
		return // No selection style to apply
	}

	for x := start; x < end; x++ {
		str, style, width := root.Screen.Get(x, y)
		newStyle := applyStyle(style, selectStyle)
		root.Screen.Put(x, y, str, newStyle)
		if width == 2 {
			x++
		}
	}
}

// execNotify notifies by beeping and flashing the screen the specified number of times.
func (root *Root) execNotify(msg string, count int) {
	root.setMessage(msg)
	root.Screen.ShowNotification("ov notify", msg)
	for range count {
		_ = root.Screen.Beep()
		root.flash()
	}
}

// flash flashes the screen.
func (root *Root) flash() {
	root.Screen.Fill(' ', tcell.StyleDefault.Background(color.White))
	root.Screen.Sync()
	time.Sleep(50 * time.Millisecond)
	root.Screen.Fill(' ', tcell.StyleDefault.Background(color.Black))
	root.Screen.Sync()
	time.Sleep(50 * time.Millisecond)
	root.draw(context.Background())
	time.Sleep(100 * time.Millisecond)
}

// put calls Screen.Put and checks if display sync is needed.
func (root *Root) put(x, y int, str string, style tcell.Style) {
	root.Screen.Put(x, y, str, style)
	if !root.scr.forceDisplaySync && needsDisplaySync(str) {
		root.scr.forceDisplaySync = true
	}
}

// needsDisplaySync checks if display sync handling is needed.
// This is required when there's a mismatch between the character width
// recognized by tcell and the actual display width in the terminal.
// When this mismatch occurs, the display can become corrupted, so we
// need to use Screen.Sync() instead of Screen.Show() to ensure proper rendering.
//
// Common cases that require sync:
// - Combining characters (ZWJ sequences, variation selectors, skin tone modifiers)
// - Ambiguous width characters (box drawing, mathematical symbols, etc.)
// - Regional indicator symbols used in flag emojis (🇺🇸, 🇯🇵, etc.)
// - Complex character sequences that may be rendered differently across terminals
func needsDisplaySync(str string) bool {
	gr := uniseg.NewGraphemes(str)
	for gr.Next() {
		runes := gr.Runes()
		if len(runes) == 0 {
			continue
		}
		// Any combining characters can cause display issues
		if len(runes) > 1 {
			return true
		}

		mainc := runes[0]
		// Check for ambiguous width characters using runewidth library
		if runewidth.IsAmbiguousWidth(mainc) {
			return true
		}
		// Regional indicator symbols and emoji range that commonly cause issues
		// Check for regional indicator symbols (used in flag emojis).
		if mainc >= 0x1F1E0 && mainc <= 0x1F9FF {
			return true
		}
	}
	return false
}

// drawSidebar draws the sidebar.
func (root *Root) drawSidebar() {
	if !root.sidebarVisible {
		return
	}

	sidebarWidth := root.sidebarWidth
	sidebarStyle := tcell.StyleDefault
	borderStyle := tcell.StyleDefault.Background(color.Gray)
	height := root.scr.vHeight
	// Sidebar background
	for y := range height {
		for x := 0; x < sidebarWidth-1; x++ {
			root.Screen.Put(x, y, " ", sidebarStyle)
		}
		root.Screen.Put(sidebarWidth-1, y, " ", borderStyle)
	}
	// SidebarItems is prepared in prepareDraw.
	root.drawSidebarList(root.SidebarItems)
}

// drawSidebarList displays a list of SidebarItem in the sidebar.
func (root *Root) drawSidebarList(items []SidebarItem) {
	sidebarStyle := tcell.StyleDefault
	root.Screen.PutStrStyled(3, 0, root.sidebarMode.String(), sidebarStyle.Bold(true))
	currentStyle := tcell.StyleDefault.Bold(true).Reverse(true)
	height := root.scr.vHeight
	scrollY := 0
	scrollX := 0
	if root.sidebarScrolls != nil {
		scroll := root.sidebarScrolls[root.sidebarMode]
		scrollY = scroll.y
		scrollX = scroll.x
		scrollY = max(scrollY, 0)
		scrollX = min(scrollX, maxSidebarWidth+(root.sidebarWidth-1))
		scrollX = max(scrollX, 0)
		// Update scroll info.
		root.sidebarScrolls[root.sidebarMode] = sidebarScroll{
			x:        scrollX,
			y:        scrollY,
			currentY: root.sidebarScrolls[root.sidebarMode].currentY,
		}
	}
	maxList := min(len(items), height-2)
	for i := range maxList {
		item := items[i]
		style := sidebarStyle
		if item.IsCurrent {
			style = currentStyle
		}
		label := item.Label
		labelLen := uniseg.StringWidth(label)
		left := min(scrollX, len(item.Contents))
		width := max(min(root.sidebarWidth-(labelLen+2), len(item.Contents)-left), 0)
		right := min(left+width, len(item.Contents))
		out := item.Contents[left:right].String()
		root.Screen.PutStrStyled(labelLen, i+1, out, style)
		root.Screen.PutStrStyled(0, i+1, label, style)
	}
}
