package oviewer

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
)

func (root *root) inputEvent(ev tcell.Event, fn func()) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape:
			root.mode = normal
			return true
		case tcell.KeyEnter:
			fn()
			root.mode = normal
			return true
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if len(root.input) > 0 {
				r := []rune(root.input)
				root.input = string(r[:len(r)-1])
			}
			return true
		case tcell.KeyRune:
			root.input += string(ev.Rune())
		}
	}
	return true
}

func (root *root) goLine() {
	line, err := strconv.Atoi(root.input)
	if err != nil {
		return
	}
	root.input = ""
	root.moveNum(line)
}

func (root *root) headerLen() {
	line, _ := strconv.Atoi(root.input)
	if line >= 0 && line <= root.Model.vHight-1 {
		root.Model.HeaderLen = line
		root.setWrapHeaderLen()
	}
	root.input = ""
}

func (root *root) _search(startY int) {
	for y := startY; y < root.Model.endY; y++ {
		if strings.Contains(root.Model.buffer[y], root.input) {
			root.moveNum(y)
			break
		}
	}
}

func (root *root) search() {
	root._search(root.Model.lineNum)
}

func (root *root) nextSearch() {
	root._search(root.Model.lineNum + root.Model.HeaderLen + 1)
}

func (root *root) _previous(startY int) {
	for y := startY; y >= 0; y-- {
		if strings.Contains(root.Model.buffer[y], root.input) {
			root.moveNum(y)
			return
		}
	}
}

func (root *root) previous() {
	root._previous(root.Model.lineNum)
}

func (root *root) previousSearch() {
	root._previous(root.Model.lineNum + root.Model.HeaderLen - 1)
}

func (root *root) HandleEvent(ev tcell.Event) bool {
	switch root.mode {
	case search:
		return root.inputEvent(ev, root.search)
	case previous:
		return root.inputEvent(ev, root.previous)
	case goline:
		return root.inputEvent(ev, root.goLine)
	case header:
		return root.inputEvent(ev, root.headerLen)
	}
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape:
			root.Quit()
			return true
		case tcell.KeyEnter:
			root.moveDown()
			return true
		case tcell.KeyHome:
			root.moveTop()
			return true
		case tcell.KeyEnd:
			root.moveEnd()
			return true
		case tcell.KeyLeft:
			if ev.Modifiers()&tcell.ModCtrl > 0 {
				root.moveHfLeft()
				return true
			}
			root.moveLeft()
			return true
		case tcell.KeyRight:
			if ev.Modifiers()&tcell.ModCtrl > 0 {
				root.moveHfRight()
			} else {
				root.moveRight()
				return true
			}
		case tcell.KeyDown, tcell.KeyCtrlN:
			root.moveDown()
			return true
		case tcell.KeyUp, tcell.KeyCtrlP:
			root.moveUp()
			return true
		case tcell.KeyPgDn, tcell.KeyCtrlV:
			root.movePgDn()
			return true
		case tcell.KeyPgUp, tcell.KeyCtrlB:
			root.movePgUp()
			return true
		case tcell.KeyCtrlU:
			root.moveHfUp()
			return true
		case tcell.KeyCtrlD:
			root.moveHfDn()
			return true
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'q':
				root.Quit()
				return true
			case 'Q':
				root.Model.PostWrite = true
				root.Quit()
				return true
			case 'W', 'w':
				root.keyWrap()
				return true
			case '?':
				root.input = ""
				root.keyPrevious()
				return true
			case '/':
				root.input = ""
				root.keySearch()
				return true
			case 'n':
				root.nextSearch()
				return true
			case 'N':
				root.previousSearch()
				return true
			case 'g':
				root.input = ""
				root.keyGoLine()
				return true
			case 'H':
				root.input = ""
				root.keyHeader()
				return true
			}
		}
	}
	return true
}

func (root *root) keyWrap() {
	if root.Model.WrapMode {
		root.Model.WrapMode = false
	} else {
		root.Model.WrapMode = true
		root.Model.x = 0
	}
	root.setWrapHeaderLen()
}

func (root *root) keySearch() {
	root.mode = search
}

func (root *root) keyPrevious() {
	root.mode = previous
}

func (root *root) keyGoLine() {
	root.mode = goline
}

func (root *root) keyHeader() {
	root.mode = header
}
