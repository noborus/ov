package oviewer

import (
	"fmt"

	"github.com/gdamore/tcell"
)

// Draw is the main routine that draws the screen.
func (root *Root) Draw() {
	m := root.Model
	screen := root.Screen
	if m.BufLen() == 0 || m.vHight == 0 {
		m.lineNum = 0
		root.statusDraw()
		root.Show()
		return
	}

	root.ResetScreen()

	bottom := root.bottomLineNum(m.BufEndNum()) - root.Header
	if m.lineNum > bottom+1 {
		m.lineNum = bottom + 1
	}
	if m.lineNum < 0 {
		m.lineNum = 0
	}

	_, normalBgColor, _ := tcell.StyleDefault.Decompose()
	searchWord := ""
	if root.mode == normal {
		searchWord = root.input
	}

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
			r := rangePosition(line, root.ColumnDelimiter, root.columnNum)
			reverseContents(lc, r)
		}

		if root.WrapMode {
			lX, lY = root.wrapContents(hy, lX, lY, contents)
		} else {
			lX, lY = root.noWrapContents(hy, m.x, lY, contents)
		}
	}

	// Body
	lX = m.yy * m.vWidth
	for y := root.HeaderLen(); y < m.vHight; y++ {
		lc, err := m.lineToContents(m.lineNum+lY, root.TabWidth)
		if err != nil {
			// EOF
			screen.SetContent(0, y, '~', nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
			continue
		}

		line := m.GetLine(m.lineNum + lY)

		for n := range lc.contents {
			lc.contents[n].style = lc.contents[n].style.Reverse(false)
		}

		// search highlight
		if searchWord != "" {
			poss := searchPosition(line, root.inputRegexp)
			for _, r := range poss {
				reverseContents(lc, r)
			}
		}

		// column highlight
		if root.ColumnMode {
			r := rangePosition(line, root.ColumnDelimiter, root.columnNum)
			reverseContents(lc, r)
		}

		// line number mode
		if root.LineNumMode {
			lineNum := strToContents(fmt.Sprintf("%*d", root.startPos-1, root.Model.lineNum+lY-root.Header+1), root.TabWidth)
			for i := 0; i < len(lineNum); i++ {
				lineNum[i].style = tcell.StyleDefault.Bold(true)
			}
			root.setContentString(0, y, lineNum)
		}

		var nextY int
		if root.WrapMode {
			lX, nextY = root.wrapContents(y, lX, lY, lc.contents)
		} else {
			lX, nextY = root.noWrapContents(y, m.x, lY, lc.contents)
		}

		// alternate background color
		if root.AlternateRows {
			bgColor := normalBgColor
			if (m.lineNum+lY)%2 == 1 {
				bgColor = root.ColorAlternate
			}
			for x := 0; x < m.vWidth; x++ {
				r, c, style, _ := root.GetContent(x, y)
				root.SetContent(x, y, r, c, style.Background(bgColor))
			}
		}
		lY = nextY
	}

	if lY > 0 {
		root.bottomPos = m.lineNum + lY - 1
	} else {
		root.bottomPos = m.lineNum + 1
	}

	root.statusDraw()
	root.Show()
}

// ResetScreen initializes the screen with a blank
func (root *Root) ResetScreen() {
	space := Content{
		mainc: ' ',
		combc: nil,
		width: 1,
		style: tcell.StyleDefault.Normal(),
	}
	for y := 0; y < root.Model.vHight; y++ {
		for x := 0; x < root.Model.vWidth; x++ {
			root.Screen.SetContent(x, y, space.mainc, space.combc, space.style)
		}
	}
}

// reverses the specified range.
func reverseContents(lc lineContents, r rangePos) {
	for n := lc.cMap[r.start]; n < lc.cMap[r.end]; n++ {
		lc.contents[n].style = lc.contents[n].style.Reverse(true)
	}
}

// wrapContents wraps and draws the contents and returns the next drawing position.
func (root *Root) wrapContents(y int, lX int, lY int, contents []Content) (int, int) {
	for x := 0; ; x++ {
		if lX+x >= len(contents) {
			// EOL
			lX = 0
			lY++
			break
		}
		content := contents[lX+x]
		if x+content.width+root.startPos > root.Model.vWidth {
			// next line
			lX += x
			break
		}
		root.Screen.SetContent(x+root.startPos, y, content.mainc, content.combc, content.style)
	}
	return lX, lY
}

// noWrapContents draws contents without wrapping and returns the next drawing position.
func (root *Root) noWrapContents(y int, lX int, lY int, contents []Content) (int, int) {
	if lX < root.minStartPos {
		lX = root.minStartPos
	}
	for x := 0; x+root.startPos < root.Model.vWidth; x++ {
		if lX+x < 0 {
			continue
		}
		if lX+x >= len(contents) {
			break
		}
		content := contents[lX+x]
		root.Screen.SetContent(x+root.startPos, y, content.mainc, content.combc, content.style)
	}
	lY++
	return lX, lY
}

// headerStyle applies the style of the header.
func (root *Root) headerStyle(contents []Content) {
	for i := 0; i < len(contents); i++ {
		contents[i].style = root.HeaderStyle
	}
}

// statusDraw draws a status line.
func (root *Root) statusDraw() {
	screen := root.Screen
	style := tcell.StyleDefault

	for x := 0; x < root.Model.vWidth; x++ {
		screen.SetContent(x, root.statusPos, 0, nil, style)
	}

	leftStatus := fmt.Sprintf("%s:%s", root.fileName, root.message)
	leftContents := strToContents(leftStatus, -1)
	caseSensitive := ""
	if root.CaseSensitive {
		caseSensitive = "(Aa)"
	}
	switch root.mode {
	case search:
		p := caseSensitive + "/"
		leftStatus = p + root.input
		root.Screen.ShowCursor(len(p)+root.cursorX, root.statusPos)
		leftContents = strToContents(leftStatus, -1)
	case previous:
		p := caseSensitive + "?"
		leftStatus = p + root.input
		root.Screen.ShowCursor(len(p)+root.cursorX, root.statusPos)
		leftContents = strToContents(leftStatus, -1)
	case goline:
		p := "Goto line:"
		leftStatus = p + root.input
		root.Screen.ShowCursor(len(p)+root.cursorX, root.statusPos)
		leftContents = strToContents(leftStatus, -1)
	case header:
		p := "Header length:"
		leftStatus = p + root.input
		root.Screen.ShowCursor(len(p)+root.cursorX, root.statusPos)
		leftContents = strToContents(leftStatus, -1)
	case delimiter:
		p := "Delimiter:"
		leftStatus = p + root.input
		root.Screen.ShowCursor(len(p)+root.cursorX, root.statusPos)
		leftContents = strToContents(leftStatus, -1)
	case tabWidth:
		p := "TAB width:"
		leftStatus = p + root.input
		root.Screen.ShowCursor(len(p)+root.cursorX, root.statusPos)
		leftContents = strToContents(leftStatus, -1)
	default:
		for i := 0; i < len(leftContents); i++ {
			leftContents[i].style = leftContents[i].style.Reverse(true)
		}
		root.Screen.ShowCursor(len(leftContents), root.statusPos)
	}
	root.setContentString(0, root.statusPos, leftContents)

	next := ""
	if !root.Model.BufEOF() {
		next = "..."
	}
	rightStatus := fmt.Sprintf("(%d/%d%s)", root.Model.lineNum, root.Model.BufEndNum(), next)
	rightContents := strToContents(rightStatus, -1)
	root.setContentString(root.Model.vWidth-len(rightStatus), root.statusPos, rightContents)

	if Debug {
		debugMsg := fmt.Sprintf("header:%d(%d) body:%d-%d \n", root.Header, root.HeaderLen(), root.Model.lineNum, root.bottomPos)
		c := strToContents(debugMsg, 0)
		root.setContentString(30, root.statusPos, c)
	}
}

// setContentString is a helper function that draws a string with setContent.
func (root *Root) setContentString(vx int, vy int, contents []Content) {
	screen := root.Screen
	for x, content := range contents {
		screen.SetContent(vx+x, vy, content.mainc, content.combc, content.style)
	}
}
