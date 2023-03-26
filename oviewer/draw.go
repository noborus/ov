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

	if root.scr.vHeight == 0 {
		m.topLN = 0
		root.drawStatus()
		root.Show()
		return
	}

	if m.ColumnWidth && m.columnWidths == nil {
		m.setColumnWidths()
	}

	// Header
	lN := root.drawHeader()

	lX := 0
	if m.WrapMode {
		lX = m.topLX
	}

	lN = m.topLN + lN
	// Body
	lX, lN = root.drawBody(lX, lN)

	m.bottomLN = max(lN, 0)
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

	// lN is a logical line number.
	lN := m.SkipLines
	// lX is the logical x position of Contents.
	lX := 0

	// wrapNum is the number of wrapped lines.
	wrapNum := 0
	line, _ := m.getLineC(lN, m.TabWidth)
	// y is the y-coordinate.
	y := 0
	for ; lN < m.firstLine(); y++ {
		if y > root.scr.vHeight {
			break
		}

		root.scr.numbers[y] = LineNumber{
			number: lN,
			wrap:   wrapNum,
		}

		if root.Doc.ColumnMode {
			root.columnHighlight(line)
		}
		if root.Doc.LineNumMode {
			root.blankLineNumber(y)
		}

		nextX, nextY := root.drawLine(y, lX, lN, line.lc)
		// header style
		root.yStyle(y, root.StyleHeader)

		lX = nextX
		if nextY != lN {
			lN = nextY
			line, _ = m.getLineC(lN, m.TabWidth)
			root.styleContent(lN, line)

		}

		if lX > 0 {
			wrapNum++
		} else {
			wrapNum = 0
		}
	}
	root.headerLen = y
	return lN
}

// drawBody draws body.
func (root *Root) drawBody(lX int, lN int) (int, int) {
	m := root.Doc

	wrapNum := root.numOfWrap(lX, lN)
	line, valid := m.getLineC(lN, m.TabWidth)
	root.bodyStyle(line.lc, root.StyleBody)
	if valid {
		root.styleContent(lN, line)
	}
	for y := root.headerLen; y < root.scr.vHeight-statusLine; y++ {
		root.scr.numbers[y] = LineNumber{
			number: lN,
			wrap:   wrapNum,
		}

		if root.Doc.LineNumMode {
			if valid {
				root.drawLineNumber(lN, y)
			} else {
				root.blankLineNumber(y)
			}
		}

		nextX, nextY := root.drawLine(y, lX, lN, line.lc)

		if valid {
			root.coordinatesStyle(lN, y, line.str)
		}

		lX = nextX
		if nextY != lN {
			lN = nextY
			line, valid = m.getLineC(lN, m.TabWidth)
			root.bodyStyle(line.lc, root.StyleBody)
			if valid {
				root.styleContent(lN, line)
			}
		}

		if lX > 0 {
			wrapNum++
		} else {
			wrapNum = 0
		}
	}
	return lX, lN
}

func (root *Root) styleContent(lN int, line LineC) {
	if root.Doc.PlainMode {
		root.plainStyle(line.lc)
	}
	if root.Doc.ColumnMode {
		root.columnHighlight(line)
	}
	root.multiColorHighlight(line)
	root.searchHighlight(lN, line)
}

func (root *Root) coordinatesStyle(lN int, y int, str string) {
	root.alternateRowsStyle(lN, y)
	markStyleWidth := min(root.scr.vWidth, root.Doc.general.MarkStyleWidth)
	root.markStyle(lN, y, markStyleWidth)
	root.sectionLineHighlight(y, str)
	if root.Doc.JumpTarget != 0 && root.headerLen+root.Doc.JumpTarget == y {
		root.yStyle(y, root.StyleJumpTargetLine)
	}
}

// drawWrapLine wraps and draws the contents and returns the next drawing position.
func (root *Root) drawLine(y int, lX int, lN int, lc contents) (int, int) {
	if root.Doc.WrapMode {
		return root.drawWrapLine(y, lX, lN, lc)
	}

	return root.drawNoWrapLine(y, root.Doc.x, lN, lc)
}

// drawWrapLine wraps and draws the contents and returns the next drawing position.
func (root *Root) drawWrapLine(y int, lX int, lN int, lc contents) (int, int) {
	if lX < 0 {
		log.Printf("Illegal lX:%d", lX)
		return 0, 0
	}

	//log.Println("len", root.scr.vHeight*root.scr.vWidth, len(lc))
	for x := 0; ; x++ {
		if lX+x >= len(lc) {
			// EOL
			root.clearEOL(root.scr.startX+x, y)
			lX = 0
			lN++
			break
		}
		content := lc[lX+x]
		if x+root.scr.startX+content.width > root.scr.vWidth {
			// Right edge.
			root.clearEOL(root.scr.startX+x, y)
			lX += x
			break
		}
		root.Screen.SetContent(root.scr.startX+x, y, content.mainc, content.combc, content.style)
	}

	return lX, lN
}

// drawNoWrapLine draws contents without wrapping and returns the next drawing position.
func (root *Root) drawNoWrapLine(y int, lX int, lN int, lc contents) (int, int) {
	if lX < root.minStartX {
		lX = root.minStartX
	}

	for x := 0; root.scr.startX+x < root.scr.vWidth; x++ {
		if lX+x >= len(lc) {
			// EOL
			root.clearEOL(root.scr.startX+x, y)
			break
		}
		content := DefaultContent
		if lX+x >= 0 {
			content = lc[lX+x]
		}
		root.Screen.SetContent(root.scr.startX+x, y, content.mainc, content.combc, content.style)
	}
	lN++

	return lX, lN
}

// bodyStyle applies the style from the beginning to the end of one line of the body.
// Apply style to contents.
func (root *Root) bodyStyle(lc contents, s OVStyle) {
	RangeStyle(lc, 0, len(lc), s)
}

// searchHighlight applies the style of the search highlight.
// Apply style to contents.
func (root *Root) searchHighlight(lN int, line LineC) {
	if root.searchWord == "" {
		return
	}

	indexes := root.searchPosition(lN, line.str)
	for _, idx := range indexes {
		RangeStyle(line.lc, line.pos.x(idx[0]), line.pos.x(idx[1]), root.StyleSearchHighlight)
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
func (root *Root) multiColorHighlight(line LineC) {
	numC := len(root.StyleMultiColorHighlight)
	for i := len(root.Doc.multiColorRegexps) - 1; i >= 0; i-- {
		indexes := searchPositionReg(line.str, root.Doc.multiColorRegexps[i])
		for _, idx := range indexes {
			RangeStyle(line.lc, line.pos.x(idx[0]), line.pos.x(idx[1]), root.StyleMultiColorHighlight[i%numC])
		}
	}
}

// blankLineNumber should be blank for the line number.
func (root *Root) blankLineNumber(y int) {
	m := root.Doc
	numC := StrToContents(strings.Repeat(" ", root.scr.startX-1), m.TabWidth)
	root.setContentString(0, y, numC)
}

// drawLineNumber draws the line number.
func (root *Root) drawLineNumber(lN int, y int) {
	m := root.Doc
	// Line numbers start at 1 except for skip and header lines.
	numC := StrToContents(fmt.Sprintf("%*d", root.scr.startX-1, lN-m.firstLine()+1), m.TabWidth)
	for i := 0; i < len(numC); i++ {
		numC[i].style = applyStyle(tcell.StyleDefault, root.StyleLineNumber)
	}
	root.setContentString(0, y, numC)
}

// columnHighlight applies the style of the column highlight.
func (root *Root) columnHighlight(line LineC) {
	if root.Doc.ColumnWidth {
		root.columnWidthHighlight(line)
		return
	}
	root.columnDelimiterHighlight(line)
}

// columnHighlight applies the style of the column highlight.
func (root *Root) columnDelimiterHighlight(line LineC) {
	m := root.Doc
	indexes := allIndex(line.str, m.ColumnDelimiter, m.ColumnDelimiterReg)
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

		log.Println(c, indexes)
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

// alternateRowsStyle applies from beginning to end of line.
func (root *Root) alternateRowsStyle(lN int, y int) {
	if root.Doc.AlternateRows {
		if (lN)%2 == 1 {
			root.yStyle(y, root.StyleAlternate)
		}
	}
}

// yStyle applies the style from the left edge to the right edge of the physical line.
// Apply styles to the screen.
func (root *Root) yStyle(y int, s OVStyle) {
	for x := 0; x < root.scr.vWidth; x++ {
		r, c, ts, _ := root.GetContent(x, y)
		root.Screen.SetContent(x, y, r, c, applyStyle(ts, s))
	}
}

// markStyle applies the style from the left edge to the specified width.
func (root *Root) markStyle(lN int, y int, width int) {
	m := root.Doc
	if contains(m.marked, lN) {
		for x := 0; x < width; x++ {
			r, c, style, _ := root.GetContent(x, y)
			root.SetContent(x, y, r, c, applyStyle(style, root.StyleMarkLine))
		}
	}
}

// drawStatus draws a status line.
func (root *Root) drawStatus() {
	root.clearY(root.statusPos)
	leftContents, cursorPos := root.leftStatus()
	root.setContentString(0, root.statusPos, leftContents)

	rightContents := root.rightStatus()
	root.setContentString(root.scr.vWidth-len(rightContents), root.statusPos, rightContents)

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
	if root.Doc.FollowMode && root.Doc.FollowName {
		modeStatus = "(Follow Name)"
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
	for ; x < root.scr.vWidth; x++ {
		root.Screen.SetContent(x, y, ' ', nil, tcell.StyleDefault)
	}
}

// clearY clear the specified line.
func (root *Root) clearY(y int) {
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
		root.yStyle(y, root.StyleSectionLine)
	}
}
