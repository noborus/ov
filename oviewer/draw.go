package oviewer

import (
	"fmt"

	"github.com/gdamore/tcell"
)

// draw is the main routine that draws the screen.
func (root *Root) draw() {
	if root.skipDraw {
		root.skipDraw = false
		return
	}

	m := root.Doc
	screen := root.Screen
	if m.BufEndNum() == 0 || root.vHight == 0 {
		root.Doc.lineNum = 0
		root.statusDraw()
		root.Show()
		return
	}

	l, b := root.bottomLineNum(root.Doc.endNum)
	if root.Doc.lineNum > l || (root.Doc.lineNum == l && root.Doc.branch > b) {
		if root.Doc.BufEOF() {
			root.message = "EOF"
		}
		root.Doc.lineNum = l
		root.Doc.branch = b
	}
	if root.Doc.lineNum < 0 {
		root.Doc.lineNum = 0
	}

	root.lnumber = make([]lineNumber, root.vHight+1)

	lY := 0
	lX := 0
	branch := 0
	// Header
	for hy := 0; lY < root.Doc.Header; hy++ {
		lc, err := m.lineToContents(lY, root.Doc.TabWidth)
		if err != nil {
			// EOF
			continue
		}
		if hy > root.vHight {
			break
		}
		root.headerStyle(lc)

		// column highlight
		if root.input.mode == Normal && root.Doc.ColumnMode {
			str, byteMap := contentsToStr(lc)
			start, end := rangePosition(str, root.Doc.ColumnDelimiter, root.Doc.columnNum)
			reverseContents(lc, byteMap[start], byteMap[end])
		}

		root.lnumber[hy] = lineNumber{
			line:   lY,
			branch: branch,
		}

		for x := 0; x < root.startX; x++ {
			root.Screen.SetContent(x, hy, 0, nil, tcell.StyleDefault.Normal())
		}

		if root.Doc.WrapMode {
			lX, lY = root.wrapContents(hy, lX, lY, lc)
			if lX > 0 {
				branch++
			} else {
				branch = 0
			}
		} else {
			lX, lY = root.noWrapContents(hy, root.Doc.x, lY, lc)
		}
	}

	// Body
	lX = root.Doc.branch * root.vWidth
	for y := root.headerLen(); y < root.vHight; y++ {
		lc, err := m.lineToContents(root.Doc.lineNum+lY, root.Doc.TabWidth)
		if err != nil {
			// EOF
			screen.SetContent(0, y, '~', nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
			root.drawEOL(1, y)

			root.lnumber[y] = lineNumber{
				line:   -1,
				branch: 0,
			}
			continue
		}

		for n := range lc {
			lc[n].style = lc[n].style.Reverse(false)
		}

		if root.input.reg != nil || (root.input.mode == Normal && root.Doc.ColumnMode) {
			lineStr, byteMap := contentsToStr(lc)

			// search highlight
			if root.input.reg != nil {
				poss := searchPosition(lineStr, root.input.reg)
				for _, r := range poss {
					reverseContents(lc, byteMap[r[0]], byteMap[r[1]])
				}
			}

			// column highlight
			if root.input.mode == Normal && root.Doc.ColumnMode {
				start, end := rangePosition(lineStr, root.Doc.ColumnDelimiter, root.Doc.columnNum)
				reverseContents(lc, byteMap[start], byteMap[end])
			}
		}

		// line number mode
		if root.Doc.LineNumMode {
			lineNum := strToContents(fmt.Sprintf("%*d", root.startX-1, root.Doc.lineNum+lY-root.Doc.Header+1), root.Doc.TabWidth)
			for i := 0; i < len(lineNum); i++ {
				lineNum[i].style = tcell.StyleDefault.Bold(true)
			}
			root.setContentString(0, y, lineNum)
		}

		root.lnumber[y] = lineNumber{
			line:   root.Doc.lineNum + lY,
			branch: branch,
		}

		var nextY int
		if root.Doc.WrapMode {
			lX, nextY = root.wrapContents(y, lX, lY, lc)
			if lX > 0 {
				branch++
			} else {
				branch = 0
			}
		} else {
			lX, nextY = root.noWrapContents(y, root.Doc.x, lY, lc)
		}

		// alternate background color
		if root.Doc.AlternateRows {
			bgColor := root.ColorNormalBg
			reverse := false
			if (root.Doc.lineNum+lY)%2 == 1 {
				if ColorAlternate == tcell.ColorDefault {
					// default to reversed colors
					reverse = true
				} else {
					bgColor = ColorAlternate
				}
			}
			for x := 0; x < root.vWidth; x++ {
				r, c, style, _ := root.GetContent(x, y)
				root.SetContent(x, y, r, c, style.Background(bgColor).Reverse(reverse))
			}
		}
		lY = nextY
	}

	root.bottomPos = root.Doc.lineNum + max(lY, 0) - 1

	if root.mouseSelect {
		root.drawSelect(root.x1, root.y1, root.x2, root.y2, true)
	}

	root.statusDraw()
	root.Show()
}

// drawEOL fills with blanks from the end of the line to the screen width.
func (root *Root) drawEOL(eol int, y int) {
	space := content{
		mainc: 0,
		combc: nil,
		width: 1,
		style: tcell.StyleDefault.Normal(),
	}
	for x := eol; x < root.vWidth; x++ {
		root.Screen.SetContent(x, y, space.mainc, space.combc, space.style)
	}
}

// reverses the specified range.
func reverseContents(lc lineContents, start int, end int) {
	for n := start; n < end; n++ {
		lc[n].style = lc[n].style.Reverse(true)
	}
}

// wrapContents wraps and draws the contents and returns the next drawing position.
func (root *Root) wrapContents(y int, lX int, lY int, lc lineContents) (int, int) {
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
		if lX+x < 0 {
			root.Screen.SetContent(x, y, 0, nil, tcell.StyleDefault.Normal())
			continue
		}
		if lX+x >= len(lc) {
			// EOL
			root.drawEOL(root.startX+x, y)
			break
		}
		content := lc[lX+x]
		root.Screen.SetContent(root.startX+x, y, content.mainc, content.combc, content.style)
	}
	lY++
	return lX, lY
}

// headerStyle applies the style of the header.
func (root *Root) headerStyle(lc lineContents) {
	for i := 0; i < len(lc); i++ {
		lc[i].style = HeaderStyle
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

	input := root.input
	caseSensitive := ""
	if root.CaseSensitive && (input.mode == Search || input.mode == Backsearch) {
		caseSensitive = "(Aa)"
	}

	switch input.mode {
	case Normal, Help, LogDoc:
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
	rightStatus := fmt.Sprintf("(%d/%d%s)", root.Doc.lineNum, root.Doc.BufEndNum(), next)
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
