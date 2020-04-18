package oviewer

import (
	"fmt"

	"github.com/gdamore/tcell"
)

func (root *root) Draw() {
	m := root.Model
	screen := root.Screen
	if m.BufLen() == 0 || m.vHight == 0 {
		m.lineNum = 0
		root.statusDraw()
		root.Show()
		return
	}

	space := content{
		mainc: ' ',
		combc: nil,
		width: 1,
		style: tcell.StyleDefault.Normal(),
	}
	for y := 0; y < m.vHight; y++ {
		for x := 0; x < m.vWidth; x++ {
			screen.SetContent(x, y, space.mainc, space.combc, space.style)
		}
	}

	if root.AlternateRows {
		root.controlBgColor = true
	}

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

	// Header
	lY := 0
	lX := 0
	hy := 0
	for lY < root.Header {
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
		hy++
	}

	// Body
	lX = m.yy * m.vWidth
	for y := root.HeaderLen(); y < m.vHight; y++ {
		lc, err := m.lineToContents(m.lineNum+lY, root.TabWidth)
		if err != nil {
			// EOF
			root.Screen.SetContent(0, y, '~', nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
			continue
		}

		line := m.GetLine(m.lineNum + lY)

		if root.controlBgColor {
			bgColor := normalBgColor
			// alternate background color
			if root.AlternateRows && (root.Model.lineNum+lY)%2 == 1 {
				bgColor = root.ColorAlternate
			}
			for n := range lc.contents {
				lc.contents[n].style = lc.contents[n].style.Background(bgColor)
			}
		}

		for n := range lc.contents {
			lc.contents[n].style = lc.contents[n].style.Reverse(false)
		}
		// search highlight
		if searchWord != "" {
			poss := searchPosition(line, searchWord, root.CaseSensitive)
			for _, r := range poss {
				reverseContents(lc, r)
			}
		}

		// column highlight
		if root.ColumnMode {
			r := rangePosition(line, root.ColumnDelimiter, root.columnNum)
			reverseContents(lc, r)
		}

		if root.WrapMode {
			lX, lY = root.wrapContents(y, lX, lY, lc.contents)
		} else {
			lX, lY = root.noWrapContents(y, m.x, lY, lc.contents)
		}
	}

	if lY > 0 {
		root.bottomPos = m.lineNum + lY - 1
	} else {
		root.bottomPos = m.lineNum + 1
	}

	root.statusDraw()
	root.Show()
}

func reverseContents(lc lineContents, r rangePos) {
	for n := lc.cMap[r.start]; n < lc.cMap[r.end]; n++ {
		lc.contents[n].style = lc.contents[n].style.Reverse(true)
	}
}

func (root *root) wrapContents(y int, lX int, lY int, contents []content) (rX int, rY int) {
	wX := 0
	for {
		if lX+wX >= len(contents) {
			// EOL
			lX = 0
			lY++
			break
		}
		content := contents[lX+wX]
		if wX+content.width > root.Model.vWidth {
			// next line
			lX += wX
			break
		}
		style := content.style
		root.Screen.SetContent(wX, y, content.mainc, content.combc, style)
		wX++
	}
	return lX, lY
}

func (root *root) noWrapContents(y int, lX int, lY int, contents []content) (rX int, rY int) {
	if lX < root.minStartPos {
		lX = root.minStartPos
	}
	for x := 0; x < root.Model.vWidth; x++ {
		if lX+x < 0 {
			continue
		}
		if lX+x >= len(contents) {
			break
		}
		content := contents[lX+x]
		style := content.style
		root.Screen.SetContent(x, y, content.mainc, content.combc, style)
	}
	lY++
	return lX, lY
}

func (root *root) headerStyle(contents []content) {
	for i := 0; i < len(contents); i++ {
		contents[i].style = root.HeaderStyle
	}
}

func (root *root) statusDraw() {
	screen := root.Screen
	style := tcell.StyleDefault
	next := "..."
	if root.Model.BufEOF() {
		next = ""
	}
	for x := 0; x < root.Model.vWidth; x++ {
		screen.SetContent(x, root.statusPos, 0, nil, style)
	}
	leftStatus := ""
	caseSensitive := ""
	if root.CaseSensitive {
		caseSensitive = "(Aa)"
	}
	leftStyle := style
	switch root.mode {
	case search:
		leftStatus = caseSensitive + "/" + root.input
	case previous:
		leftStatus = caseSensitive + "?" + root.input
	case goline:
		leftStatus = "Goto line:" + root.input
	case header:
		leftStatus = "Header length:" + root.input
	case delimiter:
		leftStatus = "Delimiter:" + root.input
	default:
		leftStatus = fmt.Sprintf("%s:%s", root.fileName, root.message)
		leftStyle = style.Reverse(true)
	}
	leftContents := strToContents(leftStatus, root.TabWidth)
	root.setContentString(0, root.statusPos, leftContents, leftStyle)

	screen.ShowCursor(len(leftContents), root.statusPos)

	rightStatus := fmt.Sprintf("(%d/%d%s)", root.Model.lineNum, root.Model.BufEndNum(), next)
	rightContents := strToContents(rightStatus, root.TabWidth)
	root.setContentString(root.Model.vWidth-len(rightStatus), root.statusPos, rightContents, style)
	if Debug {
		debugMsg := fmt.Sprintf("header:%d(%d) body:%d-%d \n", root.Header, root.HeaderLen(), root.Model.lineNum, root.bottomPos)
		c := strToContents(debugMsg, 0)
		root.setContentString(30, root.statusPos, c, tcell.StyleDefault)
	}
}

func (root *root) setContentString(vx int, vy int, contents []content, style tcell.Style) {
	screen := root.Screen
	for x, content := range contents {
		screen.SetContent(vx+x, vy, content.mainc, content.combc, style)
	}
}
