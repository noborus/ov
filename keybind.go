package oviewer

import (
	"strconv"

	"github.com/gdamore/tcell"
)

func (root *root) HandleEvent(ev tcell.Event) bool {
	root.message = ""
	switch root.mode {
	case search:
		return root.inputEvent(ev, root.Search)
	case previous:
		return root.inputEvent(ev, root.BackSearch)
	case goline:
		return root.inputEvent(ev, root.GoLine)
	case header:
		return root.inputEvent(ev, root.SetHeader)
	}

	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
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
				root.AfterWrite = true
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
				root.NextSearch()
				return true
			case 'N':
				root.NextBackSearch()
				return true
			case 'g':
				root.input = ""
				root.keyGoLine()
				return true
			case 'H':
				root.input = ""
				root.keyHeader()
				return true
			case 'C':
				root.keyAlternateRows()
				return true
			}
		}
	}
	return true
}

func (root *root) keyWrap() {
	if root.WrapMode {
		root.WrapMode = false
	} else {
		root.WrapMode = true
		root.Model.x = 0
	}
	root.setWrapHeaderLen()
}

func (root *root) keyAlternateRows() {
	if root.AlternateRows {
		root.AlternateRows = false
	} else {
		root.AlternateRows = true
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
	root.mode = header
}

func (root *root) inputEvent(ev tcell.Event, fn func()) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape:
			root.mode = normal
		case tcell.KeyEnter:
			fn()
			root.mode = normal
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if len(root.input) > 0 {
				r := []rune(root.input)
				root.input = string(r[:len(r)-1])
			}
		case tcell.KeyCtrlI:
			root.CaseSensitive = !root.CaseSensitive
		case tcell.KeyRune:
			root.input += string(ev.Rune())
		}
	}
	return true
}

func (root *root) GoLine() {
	lineNum, err := strconv.Atoi(root.input)
	if err != nil {
		return
	}
	root.input = ""
	root.moveNum(lineNum - root.Header)
}

func (root *root) SetHeader() {
	line, _ := strconv.Atoi(root.input)
	if line >= 0 && line <= root.Model.vHight-1 {
		root.Header = line
		root.setWrapHeaderLen()
	}
	root.input = ""
}

func (root *root) Search() {
	root.postSearch(root.search(root.Model.lineNum))
}

func (root *root) NextSearch() {
	root.postSearch(root.search(root.Model.lineNum + root.Header + 1))
}

func (root *root) BackSearch() {
	root.postSearch(root.backSearch(root.Model.lineNum))
}

func (root *root) NextBackSearch() {
	root.postSearch(root.backSearch(root.Model.lineNum + root.Header - 1))
}

func (root *root) postSearch(lineNum int, err error) {
	if err != nil {
		root.message = err.Error()
		return
	}
	root.moveNum(lineNum)
}
