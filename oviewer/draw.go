package oviewer

import (
	"fmt"

	"github.com/gdamore/tcell"
)

// draw is the main routine that draws the screen.
func (root *Root) draw() {
	m := root.Doc
	screen := root.Screen
	if m.BufEndNum() == 0 || root.vHight == 0 {
		root.lineNum = 0
		root.statusDraw()
		root.Show()
		return
	}

	root.resetScreen()

	bottom := root.bottomLineNum(m.BufEndNum()) - root.Header
	if root.lineNum > bottom+1 {
		root.lineNum = bottom + 1
	}
	if root.lineNum < 0 {
		root.lineNum = 0
	}

	_, normalBgColor, _ := tcell.StyleDefault.Decompose()

	lY := 0
	lX := 0
	// Header
	for hy := 0; lY < root.Header; hy++ {
		line := m.GetLine(lY)
		lc, err := m.lineToContents(lY, root.TabWidth)
		if err != nil {
			// EOF
			continue
		}
		contents := lc.contents
		root.headerStyle(contents)

		// column highlight
		if root.ColumnMode {
			start, end := rangePosition(line, root.ColumnDelimiter, root.columnNum)
			reverseContents(lc, start, end)
		}

		if root.WrapMode {
			lX, lY = root.wrapContents(hy, lX, lY, contents)
		} else {
			lX, lY = root.noWrapContents(hy, root.x, lY, contents)
		}
	}

	// Body
	lX = root.yy * root.vWidth
	for y := root.headerLen(); y < root.vHight; y++ {
		lc, err := m.lineToContents(root.lineNum+lY, root.TabWidth)
		if err != nil {
			// EOF
			screen.SetContent(0, y, '~', nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
			continue
		}

		line := m.GetLine(root.lineNum + lY)

		for n := range lc.contents {
			lc.contents[n].style = lc.contents[n].style.Reverse(false)
		}

		// search highlight
		if root.input.reg != nil {
			poss := searchPosition(line, root.input.reg)
			for _, r := range poss {
				reverseContents(lc, r[0], r[1])
			}
		}

		// column highlight
		if root.ColumnMode {
			start, end := rangePosition(line, root.ColumnDelimiter, root.columnNum)
			reverseContents(lc, start, end)
		}

		// line number mode
		if root.LineNumMode {
			lineNum := strToContents(fmt.Sprintf("%*d", root.startX-1, root.lineNum+lY-root.Header+1), root.TabWidth)
			for i := 0; i < len(lineNum); i++ {
				lineNum[i].style = tcell.StyleDefault.Bold(true)
			}
			root.setContentString(0, y, lineNum)
		}

		var nextY int
		if root.WrapMode {
			lX, nextY = root.wrapContents(y, lX, lY, lc.contents)
		} else {
			lX, nextY = root.noWrapContents(y, root.x, lY, lc.contents)
		}

		// alternate background color
		if root.AlternateRows {
			bgColor := normalBgColor
			if (root.lineNum+lY)%2 == 1 {
				bgColor = ColorAlternate
			}
			for x := 0; x < root.vWidth; x++ {
				r, c, style, _ := root.GetContent(x, y)
				root.SetContent(x, y, r, c, style.Background(bgColor))
			}
		}
		lY = nextY
	}

	if lY > 0 {
		root.bottomPos = root.lineNum + lY - 1
	} else {
		root.bottomPos = root.lineNum + 1
	}

	root.statusDraw()
	root.Show()
}

// resetScreen initializes the screen with a blank.
func (root *Root) resetScreen() {
	space := content{
		mainc: ' ',
		combc: nil,
		width: 1,
		style: tcell.StyleDefault.Normal(),
	}
	for y := 0; y < root.vHight; y++ {
		for x := 0; x < root.vWidth; x++ {
			root.Screen.SetContent(x, y, space.mainc, space.combc, space.style)
		}
	}
}

// reverses the specified range.
func reverseContents(lc lineContents, start int, end int) {
	for n := lc.byteMap[start]; n < lc.byteMap[end]; n++ {
		lc.contents[n].style = lc.contents[n].style.Reverse(true)
	}
}

// wrapContents wraps and draws the contents and returns the next drawing position.
func (root *Root) wrapContents(y int, lX int, lY int, contents []content) (int, int) {
	for x := 0; ; x++ {
		if lX+x >= len(contents) {
			// EOL
			lX = 0
			lY++
			break
		}
		content := contents[lX+x]
		if x+content.width+root.startX > root.vWidth {
			// next line
			lX += x
			break
		}
		root.Screen.SetContent(x+root.startX, y, content.mainc, content.combc, content.style)
	}
	return lX, lY
}

// noWrapContents draws contents without wrapping and returns the next drawing position.
func (root *Root) noWrapContents(y int, lX int, lY int, contents []content) (int, int) {
	if lX < root.minStartX {
		lX = root.minStartX
	}
	for x := 0; x+root.startX < root.vWidth; x++ {
		if lX+x < 0 {
			continue
		}
		if lX+x >= len(contents) {
			break
		}
		content := contents[lX+x]
		root.Screen.SetContent(x+root.startX, y, content.mainc, content.combc, content.style)
	}
	lY++
	return lX, lY
}

// headerStyle applies the style of the header.
func (root *Root) headerStyle(contents []content) {
	for i := 0; i < len(contents); i++ {
		contents[i].style = HeaderStyle
	}
}

// statusDraw draws a status line.
func (root *Root) statusDraw() {
	screen := root.Screen
	style := tcell.StyleDefault

	for x := 0; x < root.vWidth; x++ {
		screen.SetContent(x, root.statusPos, 0, nil, style)
	}
	leftStatus := fmt.Sprintf("%s:%s", root.Doc.FileName, root.message)
	leftContents := strToContents(leftStatus, -1)
	caseSensitive := ""
	if root.CaseSensitive {
		caseSensitive = "(Aa)"
	}
	input := root.input
	switch input.mode {
	case Normal, Help:
		for i := 0; i < len(leftContents); i++ {
			leftContents[i].style = leftContents[i].style.Reverse(true)
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
	rightStatus := fmt.Sprintf("(%d/%d%s)", root.lineNum, root.Doc.BufEndNum(), next)
	rightContents := strToContents(rightStatus, -1)
	root.setContentString(root.vWidth-len(rightStatus), root.statusPos, rightContents)

	if root.Debug {
		debugMsg := fmt.Sprintf("header:%d(%d) body:%d-%d \n", root.Header, root.headerLen(), root.lineNum, root.bottomPos)
		c := strToContents(debugMsg, 0)
		root.setContentString(root.vWidth/2, root.statusPos, c)
	}
}

// setContentString is a helper function that draws a string with setContent.
func (root *Root) setContentString(vx int, vy int, contents []content) {
	screen := root.Screen
	for x, content := range contents {
		screen.SetContent(vx+x, vy, content.mainc, content.combc, content.style)
	}
}
