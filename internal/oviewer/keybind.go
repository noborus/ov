package oviewer

import (
	"github.com/gdamore/tcell"
)

// HandleEvent handles all events.
func (root *Root) HandleEvent(ev *tcell.EventKey) bool {
	root.message = ""
	if root.mode == normal {
		return root.defaultEvent(ev)
	}
	return root.HandleInputEvent(ev)
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
			root.setMode(previous)
			return true
		case 'c':
			root.toggleColumnMode()
			return true
		case 'd':
			root.setMode(delimiter)
			return true
		case '/':
			root.setMode(search)
			return true
		case 'n':
			root.NextSearch()
			return true
		case 'N':
			root.NextBackSearch()
			return true
		case 'g':
			root.setMode(goline)
			return true
		case 'G':
			root.toggleLineNumMode()
		case 'H':
			root.setMode(header)
			return true
		case 'C':
			root.toggleAlternateRows()
			return true
		case 't':
			root.input = ""
			root.setMode(tabWidth)
			return true
		}
	}
	return true
}
