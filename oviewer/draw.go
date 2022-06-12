package oviewer

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
)

const statusLine = 1

// draw is the main routine that draws the screen.
func (root *Root) draw() {
	m := root.Doc

	if root.vHight == 0 {
		m.topLN = 0
		root.drawStatus()
		root.Show()
		return
	}

	// Header
	lY := root.drawHeader()

	lX := 0
	if m.WrapMode {
		lX = m.topLX
	}

	m.topLN = max(m.topLN, 0)
	lY = m.topLN + lY
	// Body
	lX, lY = root.drawBody(lX, lY)

	m.bottomLN = max(lY, 0)
	m.bottomLX = lX

	if root.mouseSelect {
		root.drawSelect(root.x1, root.y1, root.x2, root.y2, true)
	}

	root.drawStatus()
	root.Show()
}

// drawHeader draws header.
func (root *Root) drawHeader() int {
	m := root.Doc

	// lY is a logical line.
	lY := m.SkipLines
	// lX is the x position of Contents.
	lX := 0

	// wrapNum is the number of wrapped lines.
	wrapNum := 0
	// hy is the drawing line.
	hy := 0
	for ; lY < m.firstLine(); hy++ {
		if hy > root.vHight {
			break
		}

		lc := m.getContents(lY, m.TabWidth)
		lineStr, posCV := m.getContentsStr(lY, lc)
		root.lnumber[hy] = lineNumber{
			line: lY,
			wrap: wrapNum,
		}

		root.columnHighlight(lc, lineStr, posCV)
		root.blankLineNumber(hy)

		lX, lY = root.drawLine(hy, lX, lY, lc)

		// header style
		root.lineStyle(hy, root.StyleHeader)

		if lX > 0 {
			wrapNum++
		} else {
			wrapNum = 0
		}
	}
	root.headerLen = hy
	return lY
}

// drawBody draws body.
func (root *Root) drawBody(lX int, lY int) (int, int) {
	m := root.Doc

	wrapNum := root.numOfWrap(lX, lY)
	markStyleWidth := min(root.vWidth, root.Doc.general.MarkStyleWidth)
	root.Doc.lastContentsNum = -1

	// lc, lineStr, byteMap store the previous value.
	// Because it may be a continuation from the previous line in wrap mode.
	lastLN := -1
	var lc contents
	var lineStr string
	var posCV map[int]int
	for y := root.headerLen; y < root.vHight-statusLine; y++ {
		if lastLN != lY {
			lc = m.getContents(lY, m.TabWidth)
			lineStr, posCV = m.getContentsStr(lY, lc)
			root.bodyStyle(lc, root.StyleBody)
			lastLN = lY
		}

		root.lnumber[y] = lineNumber{
			line: lY,
			wrap: wrapNum,
		}

		root.columnHighlight(lc, lineStr, posCV)
		root.searchHighlight(lY, lc, lineStr, posCV)
		root.drawLineNumber(lY, y)

		currentY := lY
		lX, lY = root.drawLine(y, lX, lY, lc)

		root.alternateRowsStyle(currentY, y)
		root.markStyle(currentY, y, markStyleWidth)
		root.sectionLineHighlight(y, lineStr)

		if lX > 0 {
			wrapNum++
		} else {
			wrapNum = 0
		}
	}

	return lX, lY
}

// drawWrapLine wraps and draws the contents and returns the next drawing position.
func (root *Root) drawLine(y int, lX int, lY int, lc contents) (int, int) {
	if root.Doc.WrapMode {
		return root.drawWrapLine(y, lX, lY, lc)
	}

	return root.drawNoWrapLine(y, root.Doc.x, lY, lc)
}

// drawWrapLine wraps and draws the contents and returns the next drawing position.
func (root *Root) drawWrapLine(y int, lX int, lY int, lc contents) (int, int) {
	if lX < 0 {
		log.Printf("Illegal lX:%d", lX)
		return 0, 0
	}

	for x := 0; ; x++ {
		if lX+x >= len(lc) {
			// EOL
			root.clearEOL(root.startX+x, y)
			lX = 0
			lY++
			break
		}
		content := lc[lX+x]
		if x+root.startX+content.width > root.vWidth {
			// Right edge.
			root.clearEOL(root.startX+x, y)
			lX += x
			break
		}
		root.Screen.SetContent(root.startX+x, y, content.mainc, content.combc, content.style)
	}

	return lX, lY
}

// drawNoWrapLine draws contents without wrapping and returns the next drawing position.
func (root *Root) drawNoWrapLine(y int, lX int, lY int, lc contents) (int, int) {
	if lX < root.minStartX {
		lX = root.minStartX
	}

	for x := 0; root.startX+x < root.vWidth; x++ {
		if lX+x >= len(lc) {
			// EOL
			root.clearEOL(root.startX+x, y)
			break
		}
		content := DefaultContent
		if lX+x >= 0 {
			content = lc[lX+x]
		}
		root.Screen.SetContent(root.startX+x, y, content.mainc, content.combc, content.style)
	}
	lY++

	return lX, lY
}

// bodyStyle applies the style from the beginning to the end of one line of the body.
// Apply style to contents.
func (root *Root) bodyStyle(lc contents, s OVStyle) {
	RangeStyle(lc, 0, len(lc), s)
}

// searchHighlight applies the style of the search highlight.
// Apply style to contents.
func (root *Root) searchHighlight(lY int, lc contents, lineStr string, posCV map[int]int) {
	if root.searchWord == "" {
		return
	}

	poss := root.searchPosition(lY, lineStr)
	for _, r := range poss {
		RangeStyle(lc, posCV[r[0]], posCV[r[1]], root.StyleSearchHighlight)
	}
}

// blankLineNumber should be blank for the line number.
func (root *Root) blankLineNumber(y int) {
	m := root.Doc
	// line number mode
	if !root.Doc.LineNumMode {
		return
	}
	numC := StrToContents(strings.Repeat(" ", root.startX-1), m.TabWidth)
	root.setContentString(0, y, numC)
}

// drawLineNumber draws the line number.
func (root *Root) drawLineNumber(lY int, y int) {
	m := root.Doc
	if !m.LineNumMode {
		return
	}
	// Line numbers start at 1 except for skip and header lines.
	numC := StrToContents(fmt.Sprintf("%*d", root.startX-1, lY-m.firstLine()+1), m.TabWidth)
	for i := 0; i < len(numC); i++ {
		numC[i].style = applyStyle(tcell.StyleDefault, root.StyleLineNumber)
	}
	root.setContentString(0, y, numC)
}

// columnHighlight applies the style of the column highlight.
func (root *Root) columnHighlight(lc contents, str string, posCV map[int]int) {
	if !root.Doc.ColumnMode {
		return
	}
	start, end := rangePosition(str, root.Doc.ColumnDelimiter, root.Doc.columnNum)
	RangeStyle(lc, posCV[start], posCV[end], root.StyleColumnHighlight)
}

// RangeStyle applies the style to the specified range.
// Apply style to contents.
func RangeStyle(lc contents, start int, end int, s OVStyle) {
	for x := start; x < end; x++ {
		lc[x].style = applyStyle(lc[x].style, s)
	}
}

// alternateRowsStyle applies from beginning to end of line.
func (root *Root) alternateRowsStyle(lY int, y int) {
	if root.Doc.AlternateRows {
		if (lY)%2 == 1 {
			root.lineStyle(y, root.StyleAlternate)
		}
	}
}

// lineStyle applies the style from the left edge to the right edge of the physical line.
// Apply styles to the screen.
func (root *Root) lineStyle(y int, s OVStyle) {
	for x := 0; x < root.vWidth; x++ {
		r, c, ts, _ := root.GetContent(x, y)
		root.Screen.SetContent(x, y, r, c, applyStyle(ts, s))
	}
}

// markStyle applies the style from the left edge to the specified width.
func (root *Root) markStyle(lY int, y int, width int) {
	m := root.Doc
	if containsInt(m.marked, lY) {
		for x := 0; x < width; x++ {
			r, c, style, _ := root.GetContent(x, y)
			root.SetContent(x, y, r, c, applyStyle(style, root.StyleMarkLine))
		}
	}
}

// drawStatus draws a status line.
func (root *Root) drawStatus() {
	root.clearLine(root.statusPos)
	leftContents, cursorPos := root.leftStatus()
	root.setContentString(0, root.statusPos, leftContents)

	rightContents := root.rightStatus()
	root.setContentString(root.vWidth-len(rightContents), root.statusPos, rightContents)

	root.Screen.ShowCursor(cursorPos, root.statusPos)
}

func (root *Root) leftStatus() (contents, int) {
	if root.input.mode == Normal {
		return root.normalLeftStatus()
	}
	return root.inputLeftStatus()
}

func (root *Root) normalLeftStatus() (contents, int) {
	number := ""
	if root.DocumentLen() > 1 && root.screenMode == Docs {
		number = fmt.Sprintf("[%d]", root.CurrentDoc)
	}

	modeStatus := ""
	if root.Doc.FollowMode {
		modeStatus = "(Follow Mode)"
	}
	if root.General.FollowAll {
		modeStatus = "(Follow All)"
	}
	// Watch mode doubles as FollowSection mode.
	if root.Doc.WatchMode {
		modeStatus += "(Watch)"
	} else if root.Doc.FollowSection {
		modeStatus = "(Follow Section)"
	}
	caption := root.Doc.FileName
	if root.Doc.Caption != "" {
		caption = root.Doc.Caption
	}

	leftStatus := fmt.Sprintf("%s%s%s:%s", number, modeStatus, caption, root.message)
	leftContents := StrToContents(leftStatus, -1)
	color := tcell.ColorWhite
	if root.CurrentDoc != 0 {
		color = tcell.Color((root.CurrentDoc + 8) % 16)
	}
	for i := 0; i < len(leftContents); i++ {
		leftContents[i].style = leftContents[i].style.Foreground(tcell.ColorValid + color).Reverse(true)
	}
	return leftContents, len(leftContents)
}

func (root *Root) inputLeftStatus() (contents, int) {
	input := root.input
	searchMode := ""
	if input.mode == Search || input.mode == Backsearch {
		if root.Config.RegexpSearch {
			searchMode += "(R)"
		}
		if root.Config.Incsearch {
			searchMode += "(I)"
		}
		if root.CaseSensitive {
			searchMode += "(Aa)"
		}
	}
	p := searchMode + input.EventInput.Prompt()
	leftStatus := p + input.value
	leftContents := StrToContents(leftStatus, -1)
	return leftContents, len(p) + input.cursorX
}

func (root *Root) rightStatus() contents {
	next := ""
	if !root.Doc.BufEOF() {
		next = "..."
	}
	str := fmt.Sprintf("(%d/%d%s)", root.Doc.topLN, root.Doc.BufEndNum(), next)
	return StrToContents(str, -1)
}

// setContentString is a helper function that draws a string with setContent.
func (root *Root) setContentString(vx int, vy int, lc contents) {
	screen := root.Screen
	for x, content := range lc {
		screen.SetContent(vx+x, vy, content.mainc, content.combc, content.style)
	}
	screen.SetContent(vx+len(lc), vy, 0, nil, tcell.StyleDefault.Normal())
}

// clearEOL clears from the specified position to the right end.
func (root *Root) clearEOL(x int, y int) {
	for ; x < root.vWidth; x++ {
		root.Screen.SetContent(x, y, ' ', nil, tcell.StyleDefault)
	}
}

// clearLine clear the specified line.
func (root *Root) clearLine(y int) {
	root.clearEOL(0, y)
}

// sectionLineHighlight applies the style of the section line highlight.
func (root *Root) sectionLineHighlight(y int, str string) {
	if root.Doc.SectionDelimiter == "" {
		return
	}

	if root.Doc.SectionDelimiterReg == nil {
		log.Printf("Regular expression is not set: %s", root.Doc.SectionDelimiter)
		return
	}

	if root.Doc.SectionDelimiterReg.MatchString(str) {
		root.lineStyle(y, root.StyleSectionLine)
	}
}
