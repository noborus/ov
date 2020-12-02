package oviewer

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell"
)

// draw is the main routine that draws the screen.
func (root *Root) draw() {
	m := root.Doc
	if m.BufEndNum() == 0 || root.vHight == 0 {
		m.lineNum = 0
		root.statusDraw()
		root.Show()
		return
	}
	//l, x := root.displayLineNum(m.lineNum, m.firstStartX)
	//log.Printf("startLine:%d startX:%d, endLine:%d endX:%d", m.lineNum, m.firstStartX, l, x)
	// Calculate the bottom line when it is possible to reach EOF.
	if m.lineNum+root.vHight >= m.endNum {
		l, x := root.bottomLineNum(m.endNum)
		if m.lineNum > l || (m.lineNum == l && m.firstStartX > x) {
			if m.BufEOF() {
				root.message = "EOF"
			}
			m.lineNum = l
			m.firstStartX = x
		}
	}

	if m.lineNum < 0 {
		m.lineNum = 0
	}

	root.lnumber = make([]lineNumber, root.vHight+1)

	lY := 0
	lX := 0
	branch := 0
	// Header
	for hy := 0; lY < m.Header; hy++ {
		lc := root.getLineContents(lY, m.TabWidth)
		if hy > root.vHight {
			break
		}
		root.headerStyle(lc)

		// column highlight
		if root.input.mode == Normal && m.ColumnMode {
			str, byteMap := contentsToStr(lc)
			start, end := rangePosition(str, m.ColumnDelimiter, m.columnNum)
			reverseContents(lc, byteMap[start], byteMap[end])
		}

		root.lnumber[hy] = lineNumber{
			line:   lY,
			branch: branch,
		}

		for x := 0; x < root.startX; x++ {
			root.Screen.SetContent(x, hy, 0, nil, tcell.StyleDefault.Normal())
		}

		if m.WrapMode {
			lX, lY = root.wrapContents(hy, lX, lY, lc)
			if lX > 0 {
				branch++
			} else {
				branch = 0
			}
		} else {
			lX, lY = root.noWrapContents(hy, m.x, lY, lc)
		}
	}

	lastLY := -1
	var lc lineContents
	var lineStr string
	var byteMap map[int]int

	if m.WrapMode {
		lX = m.firstStartX
	}

	// Body
	for y := root.headerLen(); y < root.vHight-1; y++ {
		if lastLY != lY {
			lc = root.getLineContents(m.lineNum+lY, m.TabWidth)
			root.lnumber[y] = lineNumber{
				line:   -1,
				branch: 0,
			}
			lineStr, byteMap = root.getContentsStr(m.lineNum+lY, lc)
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
			lineNum := strToContents(fmt.Sprintf("%*d", root.startX-1, m.lineNum+lY-m.Header+1), m.TabWidth)
			for i := 0; i < len(lineNum); i++ {
				lineNum[i].style = tcell.StyleDefault.Bold(true)
			}
			root.setContentString(0, y, lineNum)
		}

		root.lnumber[y] = lineNumber{
			line:   m.lineNum + lY,
			branch: branch,
		}

		var nextY int
		if m.WrapMode {
			lX, nextY = root.wrapContents(y, lX, lY, lc)
			if lX > 0 {
				branch++
			} else {
				branch = 0
			}
		} else {
			lX, nextY = root.noWrapContents(y, m.x, lY, lc)
		}

		// alternate style
		if m.AlternateRows {
			if (m.lineNum+lY)%2 == 1 {
				for x := 0; x < root.vWidth; x++ {
					r, c, style, _ := root.GetContent(x, y)
					root.SetContent(x, y, r, c, applyStyle(style, root.StyleAlternate))
				}
			}
		}
		lY = nextY
	}

	root.bottomPos = m.lineNum + max(lY, 0)
	root.bottomEndX = lX

	if root.mouseSelect {
		root.drawSelect(root.x1, root.y1, root.x2, root.y2, true)
	}

	root.statusDraw()
	root.Show()
}

func (root *Root) getContentsStr(lineNum int, lc lineContents) (string, map[int]int) {
	if root.Doc.lastContentsNum != lineNum {
		root.Doc.lastContentsStr, root.Doc.lastContentsMap = contentsToStr(lc)
		root.Doc.lastContentsNum = lineNum
	}
	return root.Doc.lastContentsStr, root.Doc.lastContentsMap
}

func (root *Root) getLineContents(lineNum int, tabWidth int) lineContents {
	lc, err := root.Doc.lineToContents(lineNum, tabWidth)
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
	space := content{
		mainc: 0,
		combc: nil,
		width: 1,
		style: tcell.StyleDefault.Normal(),
	}
	for x := 1; x < width; x++ {
		lc[x] = space
	}
	return lc
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
		lc[i].style = applyStyle(lc[i].style, root.StyleHeader)
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
