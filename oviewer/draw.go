package oviewer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

// defaultStyle is used when the style is not specified.
var defaultStyle = tcell.StyleDefault

// draw is the main routine that draws the screen.
func (root *Root) draw(ctx context.Context) {
	m := root.Doc

	// Prepare the screen for drawing.
	root.prepareDraw(ctx)

	// Body.
	lX := m.topLX
	lN := m.topLN + root.scr.headerEnd
	// If WrapMode is off, lX is always 0.
	if !m.WrapMode {
		lX = 0
	}
	lX, lN = root.drawBody(lX, lN)
	m.bottomLX = lX
	m.bottomLN = lN

	// Header.
	if m.headerHeight > 0 {
		root.drawHeader()
	}
	// Section header.
	if m.sectionHeaderHeight > 0 {
		root.drawSectionHeader()
	}

	if root.scr.mouseSelect {
		root.drawSelect(root.scr.x1, root.scr.y1, root.scr.x2, root.scr.y2, true)
	}

	root.drawStatus()
	root.Show()
}

// drawBody draws body.
func (root *Root) drawBody(lX int, lN int) (int, int) {
	m := root.Doc

	markStyleWidth := min(root.scr.vWidth, root.Doc.general.MarkStyleWidth)

	wrapNum := m.numOfWrap(lX, lN)
	for y := m.headerHeight; y < root.scr.vHeight-statusLine; y++ {
		lineC, ok := root.scr.lines[lN]
		if !ok {
			log.Fatalf("line is not found %d", lN)
		}
		root.scr.numbers[y] = newLineNumber(lN, wrapNum)
		root.drawLineNumber(lN, y, lineC.valid)

		nextLX, nextLN := root.drawLine(y, lX, lN, lineC)
		if lineC.valid {
			root.coordinatesStyle(lN, y)
		}
		if root.Doc.SectionHeader {
			root.sectionLineHighlight(y, lineC)
			if root.Doc.HideOtherSection {
				root.hideOtherSection(y, lineC)
			}
		}

		root.drawVerticalHeader(y, wrapNum, lineC)

		root.applyMarkStyle(lN, y, markStyleWidth)

		wrapNum++
		if nextLX == 0 {
			wrapNum = 0
		}

		lX = nextLX
		lN = nextLN
	}
	return lX, lN
}

// drawHeader draws header.
func (root *Root) drawHeader() {
	m := root.Doc

	// wrapNum is the number of wrapped lines.
	wrapNum := 0
	lX := 0
	lN := root.scr.headerLN
	for y := 0; y < m.headerHeight && lN < root.scr.headerEnd; y++ {
		lineC, ok := root.scr.lines[lN]
		if !ok {
			log.Fatalf("line is not found %d", lN)
		}
		root.scr.numbers[y] = newLineNumber(lN, wrapNum)
		root.blankLineNumber(y)

		lX, lN = root.drawLine(y, lX, lN, lineC)
		root.drawVerticalHeader(y, wrapNum, lineC)
		// header style.
		root.applyStyleToLine(y, root.StyleHeader)

		wrapNum++
		if lX == 0 {
			wrapNum = 0
		}
	}
}

// drawSectionHeader draws section header.
// The section header overrides the body.
func (root *Root) drawSectionHeader() {
	m := root.Doc

	wrapNum := 0
	lX := 0
	lN := root.scr.sectionHeaderLN
	for y := m.headerHeight; y < m.headerHeight+m.sectionHeaderHeight && lN < root.scr.sectionHeaderEnd; y++ {
		lineC, ok := root.scr.lines[lN]
		if !ok {
			log.Fatalf("line is not found %d", lN)
		}
		root.scr.numbers[y] = newLineNumber(lN, wrapNum)
		root.drawLineNumber(lN, y, lineC.valid)

		nextLX, nextLN := root.drawLine(y, lX, lN, lineC)
		// section header style.
		root.applyStyleToLine(y, root.StyleSectionLine)

		root.drawVerticalHeader(y, wrapNum, lineC)
		// markstyle is displayed above the section header.
		markStyleWidth := min(root.scr.vWidth, m.general.MarkStyleWidth)
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
}

// drawWrapLine wraps and draws the contents and returns the next drawing position.
func (root *Root) drawLine(y int, lX int, lN int, lineC LineC) (int, int) {
	if root.Doc.WrapMode {
		return root.drawWrapLine(y, lX, lN, lineC)
	}

	return root.drawNoWrapLine(y, root.Doc.x, lN, lineC)
}

// drawWrapLine wraps and draws the contents and returns the next drawing position.
func (root *Root) drawWrapLine(y int, lX int, lN int, lineC LineC) (int, int) {
	if lX < 0 {
		log.Printf("Illegal lX:%d", lX)
		return 0, 0
	}
	screen := root.Screen
	for n := 0; ; n++ {
		x := root.scr.startX + n
		if lX+n >= len(lineC.lc) {
			// EOL
			root.clearEOL(x, y, lineC.eolStyle)
			lX = 0
			lN++
			break
		}
		c := lineC.lc[lX+n]
		if x+c.width > root.scr.vWidth {
			// Right edge.
			root.clearEOL(x, y, defaultStyle)
			lX += n
			break
		}
		screen.SetContent(x, y, c.mainc, c.combc, c.style)
	}
	return lX, lN
}

// drawNoWrapLine draws contents without wrapping and returns the next drawing position.
func (root *Root) drawNoWrapLine(y int, lX int, lN int, lineC LineC) (int, int) {
	lX = max(lX, root.minStartX)
	screen := root.Screen
	for n := 0; root.scr.startX+n < root.scr.vWidth; n++ {
		x := root.scr.startX + n
		if lX+n >= len(lineC.lc) {
			// EOL
			root.clearEOL(x, y, lineC.eolStyle)
			break
		}
		if lX+n < 0 {
			c := DefaultContent
			screen.SetContent(x, y, c.mainc, c.combc, c.style)
			continue
		}
		c := lineC.lc[lX+n]
		screen.SetContent(x, y, c.mainc, c.combc, c.style)
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

	screen := root.Screen
	for n := 0; n < widthVH; n++ {
		c := DefaultContent
		if n < len(lineC.lc) {
			c = lineC.lc[n]
		}
		if n == widthVH-1 && c.width == 2 {
			// If the width is 2, it is the last character.
			return
		}
		style := applyStyle(c.style, root.StyleVerticalHeader)
		screen.SetContent(root.scr.startX+n, y, c.mainc, c.combc, style)
	}
}

// widthVerticalHeader calculates the vertical header value.
// If VerticalHeader is specified, it returns that as the width.
// If HeaderColumn is specified, it returns the width of the specified column.
func (m *Document) widthVerticalHeader(lineC LineC) int {
	if m.VerticalHeader > 0 {
		return m.VerticalHeader
	}

	vhc := m.HeaderColumn
	if vhc <= 0 {
		return 0
	}
	columns := lineC.columnRanges
	if len(columns) == 0 {
		return 0
	}
	vhc = min(vhc, len(columns))
	return columns[vhc-1].end + 1
}

// blankLineNumber should be blank for the line number.
func (root *Root) blankLineNumber(y int) {
	if !root.Doc.LineNumMode {
		return
	}
	if root.scr.startX <= 0 {
		return
	}
	numC := StrToContents(strings.Repeat(" ", root.scr.startX-1), root.Doc.TabWidth)
	root.setContentString(0, y, numC)
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
	if root.scr.startX <= 0 {
		return
	}

	number := lN
	if m.lineNumMap != nil {
		n, ok := m.lineNumMap.LoadForward(number)
		if ok {
			number = n
		}
	}
	number = number - m.firstLine() + 1

	// Line numbers start at 1 except for skip and header lines.
	numC := StrToContents(fmt.Sprintf("%*d", root.scr.startX-1, number), m.TabWidth)
	for i := 0; i < len(numC); i++ {
		numC[i].style = applyStyle(defaultStyle, root.StyleLineNumber)
	}
	root.setContentString(0, y, numC)
}

// setContentString is a helper function that draws a string with setContent.
func (root *Root) setContentString(vx int, vy int, lc contents) {
	screen := root.Screen
	for x, content := range lc {
		screen.SetContent(vx+x, vy, content.mainc, content.combc, content.style)
	}
	screen.SetContent(vx+len(lc), vy, 0, nil, defaultStyle.Normal())
}

// clearEOL clears from the specified position to the right end.
func (root *Root) clearEOL(x int, y int, style tcell.Style) {
	for ; x < root.scr.vWidth; x++ {
		root.Screen.SetContent(x, y, ' ', nil, style)
	}
}

// clearY clear the specified line.
func (root *Root) clearY(y int) {
	root.clearEOL(0, y, defaultStyle)
}

// coordinatesStyle applies the style of the coordinates.
func (root *Root) coordinatesStyle(lN int, y int) {
	root.applyStyleToAlternate(lN, y)
	if root.Doc.jumpTargetHeight != 0 && root.Doc.headerHeight+root.Doc.jumpTargetHeight == y {
		root.applyStyleToLine(y, root.StyleJumpTargetLine)
	}
}

// applyStyleToAlternate applies from beginning to end of line.
func (root *Root) applyStyleToAlternate(lN int, y int) {
	if !root.Doc.AlternateRows {
		return
	}
	if (lN)%2 == 0 {
		return
	}
	root.applyStyleToLine(y, root.StyleAlternate)
}

// applyStyleToLine applies the style from the left edge to the right edge of the physical line.
// Apply styles to the screen.
func (root *Root) applyStyleToLine(y int, s OVStyle) {
	root.applyStyleToRange(y, s, 0, root.scr.vWidth)
}

// applyMarkStyle applies the style from the left edge to the specified width.
func (root *Root) applyMarkStyle(lN int, y int, width int) {
	m := root.Doc
	if !contains(m.marked, lN) {
		return
	}
	root.applyStyleToRange(y, root.StyleMarkLine, 0, width)
}

// applyStyleToRange applies the style from the start to the end of the physical line.
func (root *Root) applyStyleToRange(y int, s OVStyle, start int, end int) {
	for x := start; x < end; x++ {
		mainc, combc, style, width := root.GetContent(x, y)
		newStyle := applyStyle(style, s)
		if style != newStyle {
			root.Screen.SetContent(x, y, mainc, combc, newStyle)
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
	root.applyStyleToLine(y, root.StyleSectionLine)
}

// hideOtherSection hides other sections.
func (root *Root) hideOtherSection(y int, line LineC) {
	if line.section <= 1 { // 1 is the first section.
		return
	}
	root.clearY(y)
}

// drawSelect highlights the selection.
// Multi-line selection is included until the end of the line,
// but if the rectangle flag is true, the rectangle will be the range.
func (root *Root) drawSelect(x1, y1, x2, y2 int, sel bool) {
	if y2 < y1 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}
	if y1 == y2 {
		if x2 < x1 {
			x1, x2 = x2, x1
		}
		root.reverseRange(y1, x1, x2+1, sel)
		return
	}
	if root.scr.mouseRectangle {
		for y := y1; y <= y2; y++ {
			root.reverseRange(y, x1, x2+1, sel)
		}
		return
	}

	root.reverseRange(y1, x1, root.scr.vWidth, sel)
	for y := y1 + 1; y < y2; y++ {
		root.reverseRange(y, 0, root.scr.vWidth, sel)
	}
	root.reverseRange(y2, 0, x2+1, sel)
}

// reverseRange reverses the specified range.
func (root *Root) reverseRange(y int, start int, end int, sel bool) {
	if start >= end {
		return
	}
	for x := start; x < end; x++ {
		mainc, combc, style, width := root.Screen.GetContent(x, y)
		style = style.Reverse(sel)
		root.Screen.SetContent(x, y, mainc, combc, style)
		if width == 2 {
			x++
		}
	}
}

// notifyEOFReached notifies that EOF has been reached.
func (root *Root) notifyEOFReached(m *Document) {
	root.setMessagef("EOF reached %s", m.FileName)
	root.notify(root.Config.NotifyEOF)
}

// notify notifies by beeping and flashing the screen the specified number of times.
func (root *Root) notify(count int) {
	for i := 0; i < count; i++ {
		root.Screen.Beep()
		root.flash()
	}
}

// flash flashes the screen.
func (root *Root) flash() {
	root.Screen.Fill(' ', tcell.StyleDefault.Background(tcell.ColorWhite))
	root.Screen.Sync()
	time.Sleep(50 * time.Millisecond)
	root.Screen.Fill(' ', tcell.StyleDefault.Background(tcell.ColorBlack))
	root.Screen.Sync()
	time.Sleep(50 * time.Millisecond)
	root.draw(context.Background())
	time.Sleep(100 * time.Millisecond)
}
