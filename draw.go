package oviewer

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

func (root *root) Draw() {
	m := root.Model
	screen := root.Screen
	if len(m.buffer) == 0 || m.vHight == 0 {
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

	bottom := root.bottomLineNum(m.endNum) - root.Header
	if m.lineNum > bottom+1 {
		m.lineNum = bottom + 1
	}
	if m.lineNum < 0 {
		m.lineNum = 0
	}

	searchWord := ""
	if root.mode == normal {
		searchWord = root.input
	}

	// Header
	lY := 0
	lX := 0
	hy := 0
	for lY < root.Header {
		contents := m.getContents(lY, root.TabWidth)
		if len(contents) == 0 {
			lY++
			continue
		}
		if root.WrapMode {
			lX, lY = root.wrapContents(hy, lX, lY, contents, true)
		} else {
			lX, lY = root.noWrapContents(hy, m.x, lY, contents, true)
		}
		hy++
	}

	// Body
	lX = m.yy * m.vWidth
	for y := root.HeaderLen(); y < m.vHight; y++ {
		contents := m.getContents(m.lineNum+lY, root.TabWidth)
		if contents == nil {
			// EOF
			root.Screen.SetContent(0, y, '~', nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
			continue
		}
		searchContents(&contents, searchWord, root.CaseSensitive)
		if root.WrapMode {
			lX, lY = root.wrapContents(y, lX, lY, contents, false)
		} else {
			lX, lY = root.noWrapContents(y, m.x, lY, contents, false)
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

func (root *root) wrapContents(y int, lX int, lY int, contents []content, headerFlag bool) (rX int, rY int) {
	wX := 0
	for {
		if lX+wX >= len(contents) {
			// EOL
			if root.AlternateRows && !headerFlag && (root.Model.lineNum+lY)%2 == 1 {
				for x := wX; x < root.Model.vWidth; x++ {
					root.Screen.SetContent(x, y, ' ', nil, tcell.StyleDefault.Background(root.ColorAlternate))
				}
			}
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
		if headerFlag {
			style = root.HeaderStyle
		} else {
			if root.AlternateRows && (root.Model.lineNum+lY)%2 == 1 {
				style = style.Background(root.ColorAlternate)
			}
		}
		root.Screen.SetContent(wX, y, content.mainc, content.combc, style)
		wX++
	}
	return lX, lY
}

func (root *root) noWrapContents(y int, lX int, lY int, contents []content, headerFlag bool) (rX int, rY int) {
	if lX < root.minStartPos {
		lX = root.minStartPos
	}
	for x := 0; x < root.Model.vWidth; x++ {
		if lX+x < 0 {
			continue
		}
		if lX+x >= len(contents) {
			if root.AlternateRows && !headerFlag && (root.Model.lineNum+lY)%2 == 1 {
				for ; x < root.Model.vWidth; x++ {
					root.Screen.SetContent(x, y, ' ', nil, tcell.StyleDefault.Background(root.ColorAlternate))
				}
			}
			break
		}
		content := contents[lX+x]
		style := content.style
		if headerFlag {
			style = root.HeaderStyle
		} else {
			if root.AlternateRows && (root.Model.lineNum+lY)%2 == 1 {
				style = style.Background(root.ColorAlternate)
			}
		}
		root.Screen.SetContent(x, y, content.mainc, content.combc, style)
	}
	lY++
	return lX, lY
}

func (root *root) statusDraw() {
	screen := root.Screen
	style := tcell.StyleDefault
	next := "..."
	if root.Model.eof {
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
	default:
		leftStatus = fmt.Sprintf("%s:%s", root.fileName, root.message)
		leftStyle = style.Reverse(true)
	}
	leftContents := strToContents(leftStatus, root.TabWidth)
	root.setContentString(0, root.statusPos, leftContents, leftStyle)

	screen.ShowCursor(runewidth.StringWidth(leftStatus), root.statusPos)

	rightStatus := fmt.Sprintf("(%d/%d%s)", root.Model.lineNum, root.Model.endNum, next)
	rightContents := strToContents(rightStatus, root.TabWidth)
	root.setContentString(root.Model.vWidth-len(rightStatus), root.statusPos, rightContents, style)
	if Debug {
		debug := fmt.Sprintf("header:%d(%d) body:%d-%d \n", root.Header, root.HeaderLen(), root.Model.lineNum, root.bottomPos)
		root.setContentString(30, root.statusPos, strToContents(debug, 0), tcell.StyleDefault)
	}
}

func (root *root) setContentString(vx int, vy int, contents []content, style tcell.Style) {
	screen := root.Screen
	for x, content := range contents {
		screen.SetContent(vx+x, vy, content.mainc, content.combc, style)
	}
}
