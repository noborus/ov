package zpager

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
	root.model.y = line
	root.input = ""
}

func (root *root) headerLen() {
	line, _ := strconv.Atoi(root.input)
	if line >= 0 && line <= root.model.vHight-1 {
		root.model.HeaderLen = line
	}
	root.input = ""
}

func (root *root) search() {
	for y := root.model.y; y < root.model.endY; y++ {
		if strings.Contains(root.model.text[y], root.input) {
			root.model.y = y
			break
		}
	}
}

func (root *root) previous() {
	for y := root.model.y; y >= 0; y-- {
		if strings.Contains(root.model.text[y], root.input) {
			root.model.y = y
			break
		}
	}
}

func (root *root) HandleEvent(ev tcell.Event) bool {
	switch root.mode {
	case search:
		return root.inputEvent(ev, root.search)
	case previous:
		return root.inputEvent(ev, root.previous)
	case goline:
		return root.inputEvent(ev, root.goLine)
	case headerLen:
		return root.inputEvent(ev, root.headerLen)
	}
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape:
			root.Quit()
			return true
		case tcell.KeyEnter:
			root.keyDown()
			return true
		case tcell.KeyHome:
			root.keyHome()
			return true
		case tcell.KeyEnd:
			root.keyEnd()
			return true
		case tcell.KeyLeft:
			root.keyLeft()
			return true
		case tcell.KeyRight:
			root.keyRight()
			return true
		case tcell.KeyDown, tcell.KeyCtrlN:
			root.keyDown()
			return true
		case tcell.KeyUp, tcell.KeyCtrlP:
			root.keyUp()
			return true
		case tcell.KeyPgDn, tcell.KeyCtrlV:
			root.keyPgDn()
			return true
		case tcell.KeyPgUp, tcell.KeyCtrlB:
			root.keyPgUp()
			return true
		case tcell.KeyCtrlU:
			root.keyHfUp()
			return true
		case tcell.KeyCtrlD:
			root.keyHfDn()
			return true
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'q':
				root.Quit()
				return true
			case 'Q':
				root.model.PostWrite = true
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
				root.model.y++
				root.search()
				return true
			case 'N':
				root.model.y--
				root.previous()
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

func (root *root) keyHome() {
	root.model.y = 0
}
func (root *root) keyEnd() {
	root.model.y = root.model.endY
}
func (root *root) keyPgUp() {
	root.model.y -= root.model.vHight
}
func (root *root) keyPgDn() {
	root.model.y += root.model.vHight
}
func (root *root) keyHfUp() {
	root.model.y -= (root.model.vHight / 2)
}
func (root *root) keyHfDn() {
	root.model.y += (root.model.vHight / 2)
}

func (root *root) keyUp() {
	root.model.y--
}

func (root *root) keyDown() {
	root.model.y++
}

func (root *root) keyLeft() {
	if !root.model.WrapMode {
		root.model.x--
	}
}

func (root *root) keyRight() {
	if !root.model.WrapMode {
		root.model.x++
	}
}

func (root *root) keyWrap() {
	if root.model.WrapMode {
		root.model.WrapMode = false
	} else {
		root.model.WrapMode = true
		root.model.x = 0
	}
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
	root.mode = headerLen
}
