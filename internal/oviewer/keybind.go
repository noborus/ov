package oviewer

import (
	"github.com/gdamore/tcell"
)

// HandleEvent handles all events.
func (root *Root) HandleEvent(ev *tcell.EventKey) bool {
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
	case delimiter:
		return root.inputEvent(ev, root.SetDelimiter)
	default:
		return root.defaultEvent(ev)
	}
}

func (root *Root) inputEvent(ev *tcell.EventKey, fn func()) bool {
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
	case tcell.KeyTAB:
		root.input += "\t"
	case tcell.KeyCtrlA:
		root.CaseSensitive = !root.CaseSensitive
	case tcell.KeyRune:
		root.input += string(ev.Rune())
	}
	return true
}

func (root *Root) defaultEvent(ev *tcell.EventKey) bool {
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
			return true
		}
		root.moveRight()
		return true
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
			root.toggleWrap()
			return true
		case '?':
			root.input = ""
			root.setMode(previous)
			return true
		case 'c':
			root.toggleColumnMode()
			return true
		case 'd':
			root.input = ""
			root.setMode(delimiter)
			return true
		case '/':
			root.input = ""
			root.setMode(search)
			return true
		case 'n':
			root.NextSearch()
			return true
		case 'N':
			root.NextBackSearch()
			return true
		case 'g':
			root.input = ""
			root.setMode(goline)
			return true
		case 'H':
			root.input = ""
			root.setMode(header)
			return true
		case 'C':
			root.toggleAlternateRows()
			return true
		}
	}
	return true
}
