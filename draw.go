package oviewer

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

func (root *root) Draw() {
	m := root.Model
	screen := root.Screen
	if len(m.buffer) == 0 || m.vHight == 0 {
		root.statusDraw()
		root.Show()
		return
	}

	space := content{
		mainc: ' ',
		combc: nil,
		width: 1,
	}
	for y := 0; y < m.vHight; y++ {
		for x := 0; x < m.vWidth; x++ {
			screen.SetContent(x, y, space.mainc, space.combc, space.style)
		}
	}

	bottom := root.bottomLineNum(m.endY) - m.HeaderLen
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
	for lY < m.HeaderLen {
		contents := m.getContents(lY)
		if len(contents) == 0 {
			lY++
			continue
		}
		for n := range contents {
			contents[n].style = contents[n].style.Bold(true)
		}
		if m.WrapMode {
			lX, lY = root.wrapContents(hy, lX, lY, contents)
		} else {
			lX, lY = root.noWrapContents(hy, m.x, lY, contents)
		}
		hy++
	}

	// Body
	lX = m.yy * m.vWidth
	for y := root.HeaderLen(); y < m.vHight; y++ {
		contents := m.getContents(m.lineNum + lY)
		if contents == nil {
			// EOF
			contents = strToContents("~", 0)
			contents[0].style = contents[0].style.Foreground(tcell.ColorGray)
		}
		if len(contents) == 0 {
			// blank line
			lY++
			continue
		}
		doReverse(&contents, searchWord)
		if m.WrapMode {
			lX, lY = root.wrapContents(y, lX, lY, contents)
		} else {
			lX, lY = root.noWrapContents(y, m.x, lY, contents)
		}
	}
	if lY > 0 {
		root.bottomPos = m.lineNum + lY - 1
	} else {
		root.bottomPos = m.lineNum + 1
	}

	root.statusDraw()
	if Debug {
		debug := fmt.Sprintf("m.y:%v header:%v bpts:%v bottom:%v\n", m.lineNum, m.HeaderLen, root.bottomPos, bottom)
		root.setContentString(30, root.statusPos, strToContents(debug, 0), tcell.StyleDefault)
	}
	root.Show()
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
		root.Screen.SetContent(wX, y, content.mainc, content.combc, content.style)
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
		root.Screen.SetContent(x, y, content.mainc, content.combc, content.style)
	}
	lY++
	return lX, lY
}

func doReverse(cp *[]content, searchWord string) {
	contents := *cp
	for n := range contents {
		contents[n].style = contents[n].style.Normal()
	}
	if searchWord == "" {
		return
	}
	s, cIndex := contentsToStr(contents)
	for i := strings.Index(s, searchWord); i >= 0; {
		start := cIndex[i]
		end := cIndex[i+len(searchWord)]
		for ci := start; ci < end; ci++ {
			contents[ci].style = contents[ci].style.Reverse(true)
		}
		j := strings.Index(s[i+1:], searchWord)
		if j >= 0 {
			i += j + 1
		} else {
			break
		}
	}
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
	leftStyle := style
	switch root.mode {
	case search:
		leftStatus = "/" + root.input
	case previous:
		leftStatus = "?" + root.input
	case goline:
		leftStatus = "Goto line:" + root.input
	case header:
		leftStatus = "Header length:" + root.input
	default:
		leftStatus = fmt.Sprintf("%s:", root.fileName)
		leftStyle = style.Reverse(true)
	}
	leftContents := strToContents(leftStatus, root.Model.TabWidth)
	root.setContentString(0, root.statusPos, leftContents, leftStyle)

	screen.ShowCursor(runewidth.StringWidth(leftStatus), root.statusPos)

	rightStatus := fmt.Sprintf("(%d/%d%s)", root.Model.lineNum, root.Model.endY, next)
	rightContents := strToContents(rightStatus, root.Model.TabWidth)
	root.setContentString(root.Model.vWidth-len(rightStatus), root.statusPos, rightContents, style)
}
