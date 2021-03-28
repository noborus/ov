package oviewer

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

// draw is the main routine that draws the screen.
func (root *Root) draw() {
	m := root.Doc

	if m.BufEndNum() == 0 || root.vHight == 0 {
		m.topLN = 0
		root.statusDraw()
		root.Show()
		return
	}

	// Header
	lY := root.drawHeader()

	lX := 0
	if m.WrapMode {
		lX = m.topLX
	}

	if m.topLN < 0 {
		m.topLN = 0
	}

	// Body
	lX, lY = root.drawBody(lX, lY)

	root.bottomLN = m.topLN + max(lY, 0)
	root.bottomLX = lX

	if root.mouseSelect {
		root.drawSelect(root.x1, root.y1, root.x2, root.y2, true)
	}

	root.statusDraw()
	root.Show()
}

func (root *Root) drawHeader() int {
	m := root.Doc

	lY := 0
	lX := 0
	wrap := 0
	for hy := 0; lY < m.Header; hy++ {
		if hy > root.vHight {
			break
		}

		lc := root.getLineContents(lY, m.TabWidth)
		root.headerStyle(lc)

		// column highlight
		if m.ColumnMode && root.input.mode == Normal {
			str, byteMap := contentsToStr(lc)
			start, end := rangePosition(str, m.ColumnDelimiter, m.columnNum)
			reverseContents(lc, byteMap[start], byteMap[end])
		}

		root.lnumber[hy] = lineNumber{
			line: lY,
			wrap: wrap,
		}

		for x := 0; x < root.startX; x++ {
			root.Screen.SetContent(x, hy, 0, nil, tcell.StyleDefault.Normal())
		}

		if m.WrapMode {
			lX, lY = root.wrapContents(hy, lX, lY, lc)
			if lX > 0 {
				wrap++
			} else {
				wrap = 0
			}
		} else {
			lX, lY = root.noWrapContents(hy, m.x, lY, lc)
		}
	}

	return lY
}

func (root *Root) drawBody(lX int, lY int) (int, int) {
	m := root.Doc

	listX, err := root.leftMostX(m.topLN + lY)
	if err != nil {
		log.Println(err, "drawBody", m.topLN+lY)
	}
	wrap := numOfSlice(listX, lX)

	lastLY := -1
	var lc lineContents
	var lineStr string
	var byteMap map[int]int

	for y := root.headerLen(); y < root.vHight-1; y++ {
		if lastLY != lY {
			lc = root.getLineContents(m.topLN+lY, m.TabWidth)
			root.lnumber[y] = lineNumber{
				line: -1,
				wrap: 0,
			}
			lineStr, byteMap = root.getContentsStr(m.topLN+lY, lc)
			lastLY = lY
		}

		// search highlight
		if root.input.reg != nil {
			poss := searchPosition(lineStr, root.input.reg)
			for _, r := range poss {
				reverseContents(lc, byteMap[r[0]], byteMap[r[1]])
			}
		}

		// column highlight
		if root.input.mode == Normal && root.Doc.ColumnMode {
			start, end := rangePosition(lineStr, m.ColumnDelimiter, m.columnNum)
			reverseContents(lc, byteMap[start], byteMap[end])
		}

		// line number mode
		if m.LineNumMode {
			lc := strToContents(fmt.Sprintf("%*d", root.startX-1, m.topLN+lY-m.Header+1), m.TabWidth)
			for i := 0; i < len(lc); i++ {
				lc[i].style = applyStyle(tcell.StyleDefault, root.StyleLineNumber)
			}
			root.setContentString(0, y, lc)
		}

		root.lnumber[y] = lineNumber{
			line: m.topLN + lY,
			wrap: wrap,
		}

		var nextY int
		if m.WrapMode {
			lX, nextY = root.wrapContents(y, lX, lY, lc)
			if lX > 0 {
				wrap++
			} else {
				wrap = 0
			}
		} else {
			lX, nextY = root.noWrapContents(y, m.x, lY, lc)
		}

		// alternate style
		if m.AlternateRows {
			if (m.topLN+lY)%2 == 1 {
				for x := 0; x < root.vWidth; x++ {
					r, c, style, _ := root.GetContent(x, y)
					root.SetContent(x, y, r, c, applyStyle(style, root.StyleAlternate))
				}
			}
		}
		lY = nextY
	}

	return lX, lY
}

func (root *Root) getContentsStr(lN int, lc lineContents) (string, map[int]int) {
	if root.Doc.lastContentsNum != lN {
		root.Doc.lastContentsStr, root.Doc.lastContentsMap = contentsToStr(lc)
		root.Doc.lastContentsNum = lN
	}
	return root.Doc.lastContentsStr, root.Doc.lastContentsMap
}

func (root *Root) getLineContents(lN int, tabWidth int) lineContents {
	lc, err := root.Doc.lineToContents(lN, tabWidth)
	if err == nil {
		for n := range lc {
			lc[n].style = lc[n].style.Reverse(false)
		}
		return lc
	}

	// EOF
	width := root.vWidth - root.startX
	lc = make(lineContents, width)
	eof := content{
		mainc: '~',
		combc: nil,
		width: 1,
		style: tcell.StyleDefault.Foreground(tcell.ColorGray),
	}
	lc[0] = eof

	for x := 1; x < width; x++ {
		lc[x] = DefaultContent
	}
	return lc
}

// drawEOL fills with blanks from the end of the line to the screen width.
func (root *Root) drawEOL(eol int, y int) {
	for x := eol; x < root.vWidth; x++ {
		root.Screen.SetContent(x, y, DefaultContent.mainc, DefaultContent.combc, DefaultContent.style)
	}
}

// reverses the specified range.
func reverseContents(lc lineContents, start int, end int) {
	for x := start; x < end; x++ {
		lc[x].style = lc[x].style.Reverse(true)
	}
}

// wrapContents wraps and draws the contents and returns the next drawing position.
func (root *Root) wrapContents(y int, lX int, lY int, lc lineContents) (int, int) {
	if lX < 0 {
		log.Printf("Illegal lX:%d", lX)
		return 0, 0
	}

	for x := 0; ; x++ {
		if lX+x >= len(lc) {
			// EOL
			root.drawEOL(root.startX+x, y)
			lX = 0
			lY++
			break
		}
		content := lc[lX+x]
		if x+content.width+root.startX > root.vWidth {
			// next line
			root.drawEOL(root.startX+x, y)
			lX += x
			break
		}
		root.Screen.SetContent(root.startX+x, y, content.mainc, content.combc, content.style)
	}

	return lX, lY
}

// noWrapContents draws contents without wrapping and returns the next drawing position.
func (root *Root) noWrapContents(y int, lX int, lY int, lc lineContents) (int, int) {
	if lX < root.minStartX {
		lX = root.minStartX
	}

	for x := 0; x+root.startX < root.vWidth; x++ {
		if lX+x >= len(lc) {
			// EOL
			root.drawEOL(root.startX+x, y)
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

// headerStyle applies the style of the header.
func (root *Root) headerStyle(lc lineContents) {
	for x := 0; x < len(lc); x++ {
		lc[x].style = applyStyle(lc[x].style, root.StyleHeader)
	}
}

// statusDraw draws a status line.
func (root *Root) statusDraw() {
	screen := root.Screen
	style := tcell.StyleDefault

	for x := 0; x < root.vWidth; x++ {
		screen.SetContent(x, root.statusPos, 0, nil, style)
	}

	number := ""
	if root.DocumentLen() > 1 {
		number = fmt.Sprintf("[%d]", root.CurrentDoc)
	}
	follow := ""
	if root.Doc.FollowMode {
		follow = "(Follow Mode)"
	}
	if root.General.FollowAll {
		follow = "(Follow All)"
	}
	leftStatus := fmt.Sprintf("%s%s%s:%s", number, follow, root.Doc.FileName, root.message)
	leftContents := strToContents(leftStatus, -1)
	input := root.input
	caseSensitive := ""
	if root.CaseSensitive && (input.mode == Search || input.mode == Backsearch) {
		caseSensitive = "(Aa)"
	}

	switch input.mode {
	case Normal, Help, LogDoc:
		color := tcell.ColorWhite
		if root.CurrentDoc != 0 {
			color = tcell.Color((root.CurrentDoc + 8) % 16)
		}

		for i := 0; i < len(leftContents); i++ {
			leftContents[i].style = leftContents[i].style.Foreground(tcell.ColorValid + color).Reverse(true)
		}
		root.Screen.ShowCursor(len(leftContents), root.statusPos)
	default:
		p := caseSensitive + input.EventInput.Prompt()
		leftStatus = p + input.value
		root.Screen.ShowCursor(len(p)+input.cursorX, root.statusPos)
		leftContents = strToContents(leftStatus, -1)
	}
	root.setContentString(0, root.statusPos, leftContents)

	next := ""
	if !root.Doc.BufEOF() {
		next = "..."
	}
	rightStatus := fmt.Sprintf("(%d/%d%s)", root.Doc.topLN, root.Doc.BufEndNum(), next)
	rightContents := strToContents(rightStatus, -1)
	root.setContentString(root.vWidth-len(rightStatus), root.statusPos, rightContents)
}

// setContentString is a helper function that draws a string with setContent.
func (root *Root) setContentString(vx int, vy int, lc lineContents) {
	screen := root.Screen
	for x, content := range lc {
		screen.SetContent(vx+x, vy, content.mainc, content.combc, content.style)
	}
	screen.SetContent(vx+len(lc), vy, 0, nil, tcell.StyleDefault.Normal())
}
