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

		lc, _ := m.getContents(lY, m.TabWidth)
		lineStr, pos := m.getContentsStr(lY, lc)
		root.lines[hy] = line{
			number: lY,
			wrap:   wrapNum,
		}

		if root.Doc.ColumnMode {
			root.columnHighlight(lc, lineStr, pos)
		}
		if root.Doc.LineNumMode {
			root.blankLineNumber(hy)
		}

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
	root.Doc.lastContentsNum = -1

	// lc, lineStr, byteMap store the previous value.
	// Because it may be a continuation from the previous line in wrap mode.
	lastLN := -1
	var lc contents
	var valid bool
	var lineStr string
	var pos widthPos
	for y := root.headerLen; y < root.vHight-statusLine; y++ {
		if lastLN != lY {
			lc, valid = m.getContents(lY, m.TabWidth)
			if valid {
				lineStr, pos = m.getContentsStr(lY, lc)
			} else {
				lineStr, pos = string(EOFC), widthPos{0: 0, 1: 1}
			}

			root.bodyStyle(lc, root.StyleBody)
			lastLN = lY
		}

		root.lines[y] = line{
			number: lY,
			wrap:   wrapNum,
		}

		if valid {
			root.styleContent(lY, lc, lineStr, pos)
		}

		if root.Doc.LineNumMode {
			if valid {
				root.drawLineNumber(lY, y)
			} else {
				root.blankLineNumber(y)
			}
		}

		currentlY := lY
		lX, lY = root.drawLine(y, lX, lY, lc)

		if valid {
			root.styleLine(currentlY, y, lc, lineStr)
		}

		if lX > 0 {
			wrapNum++
		} else {
			wrapNum = 0
		}
	}

	return lX, lY
}

func (root *Root) styleContent(lY int, lc contents, lineStr string, pos widthPos) {
	if root.Doc.PlainMode {
		root.plainStyle(lc)
	}
	if root.Doc.ColumnMode {
		root.columnHighlight(lc, lineStr, pos)
	}
	root.multiColorHighlight(lc, lineStr, pos)
	root.searchHighlight(lY, lc, lineStr, pos)
}

func (root *Root) styleLine(currentY int, y int, lc contents, lineStr string) {
	root.alternateRowsStyle(currentY, y)
	markStyleWidth := min(root.vWidth, root.Doc.general.MarkStyleWidth)
	root.markStyle(currentY, y, markStyleWidth)
	root.sectionLineHighlight(y, lineStr)
	if root.Doc.JumpTarget != 0 && root.headerLen+root.Doc.JumpTarget == y {
		root.lineStyle(y, root.StyleJumpTargetLine)
	}
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
func (root *Root) searchHighlight(lY int, lc contents, lineStr string, pos widthPos) {
	if root.searchWord == "" {
		return
	}

	indexes := root.searchPosition(lY, lineStr)
	for _, idx := range indexes {
		RangeStyle(lc, pos.x(idx[0]), pos.x(idx[1]), root.StyleSearchHighlight)
	}
}

// plainStyle defaults to the original style.
func (root *Root) plainStyle(lc contents) {
	for x := 0; x < len(lc); x++ {
		lc[x].style = tcell.StyleDefault
	}
}

// multiColorHighlight applies styles to multiple words (regular expressions) individually.
// The style of the first specified word takes precedence.
func (root *Root) multiColorHighlight(lc contents, str string, pos widthPos) {
	numC := len(root.StyleMultiColorHighlight)
	for i := len(root.Doc.multiColorRegexps) - 1; i >= 0; i-- {
		indexes := searchPositionReg(str, root.Doc.multiColorRegexps[i])
		for _, idx := range indexes {
			RangeStyle(lc, pos.x(idx[0]), pos.x(idx[1]), root.StyleMultiColorHighlight[i%numC])
		}
	}
}

// blankLineNumber should be blank for the line number.
func (root *Root) blankLineNumber(y int) {
	m := root.Doc
	numC := StrToContents(strings.Repeat(" ", root.startX-1), m.TabWidth)
	root.setContentString(0, y, numC)
}

// drawLineNumber draws the line number.
func (root *Root) drawLineNumber(lY int, y int) {
	m := root.Doc
	// Line numbers start at 1 except for skip and header lines.
	numC := StrToContents(fmt.Sprintf("%*d", root.startX-1, lY-m.firstLine()+1), m.TabWidth)
	for i := 0; i < len(numC); i++ {
		numC[i].style = applyStyle(tcell.StyleDefault, root.StyleLineNumber)
	}
	root.setContentString(0, y, numC)
}

// columnHighlight applies the style of the column highlight.
func (root *Root) columnHighlight(lc contents, str string, pos widthPos) {
	m := root.Doc

	indexes := allIndex(str, m.ColumnDelimiter, m.ColumnDelimiterReg)
	if len(indexes) == 0 {
		return
	}
	numC := len(root.StyleColumnRainbow)

	var iStart, iEnd int
	for c := 0; c < len(indexes)+1; c++ {
		switch {
		case c == 0:
			iStart = 0
			iEnd = indexes[0][1] - 1
		case c < len(indexes):
			iStart = iEnd + 1
			iEnd = indexes[c][0]
		case c == len(indexes):
			iStart = iEnd + 1
			iEnd = len(str)
		}

		start, end := pos.x(iStart), pos.x(iEnd)
		if m.ColumnRainbow {
			RangeStyle(lc, start, end, root.StyleColumnRainbow[c%numC])
		}
		if c == m.columnCursor {
			RangeStyle(lc, start, end, root.StyleColumnHighlight)
		}
	}
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
	if contains(m.marked, lY) {
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
	if root.input.Event.Mode() == Normal {
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
	p := root.inputOpts() + input.Event.Prompt()
	leftStatus := p + input.value
	leftContents := StrToContents(leftStatus, -1)
	return leftContents, len(p) + input.cursorX
}

func (root *Root) inputOpts() string {
	opt := ""
	switch root.input.Event.Mode() {
	case Search, Backsearch:
		if root.Config.RegexpSearch {
			opt += "(R)"
		}
		if root.Config.Incsearch {
			opt += "(I)"
		}
		if root.Config.CaseSensitive {
			opt += "(Aa)"
		}
	}
	return opt
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
