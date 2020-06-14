package oviewer

import (
	"github.com/gdamore/tcell"
)

// DefaultKeyEvent default input key events.
func (root *Root) DefaultKeyEvent(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		root.Quit()
	case tcell.KeyCtrlL:
		root.Sync()
	case tcell.KeyEnter:
		root.moveDown()
	case tcell.KeyHome:
		root.moveTop()
	case tcell.KeyEnd:
		root.moveEnd()
	case tcell.KeyLeft:
		if ev.Modifiers()&tcell.ModCtrl > 0 {
			root.moveHfLeft()
		} else {
			root.moveLeft()
		}
	case tcell.KeyRight:
		if ev.Modifiers()&tcell.ModCtrl > 0 {
			root.moveHfRight()
		} else {
			root.moveRight()
		}
	case tcell.KeyDown, tcell.KeyCtrlN:
		root.moveDown()
	case tcell.KeyUp, tcell.KeyCtrlP:
		root.moveUp()
	case tcell.KeyPgDn, tcell.KeyCtrlV:
		root.movePgDn()
	case tcell.KeyPgUp, tcell.KeyCtrlB:
		root.movePgUp()
	case tcell.KeyCtrlU:
		root.moveHfUp()
	case tcell.KeyCtrlD:
		root.moveHfDn()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q':
			root.Quit()
		case 'Q':
			root.AfterWrite = true
			root.Quit()
		case 'W', 'w':
			root.toggleWrapMode()
		case '?':
			root.Input.SetMode(backsearch)
		case 'c':
			root.toggleColumnMode()
		case 'd':
			root.Input.SetMode(delimiter)
		case '/':
			root.Input.SetMode(search)
		case 'n':
			root.NextSearch()
		case 'N':
			root.NextBackSearch()
		case 'g':
			root.Input.SetMode(goline)
		case 'G':
			root.toggleLineNumMode()
		case 'H':
			root.Input.SetMode(header)
		case 'C':
			root.toggleAlternateRows()
		case 't':
			root.Input.SetMode(tabWidth)
		}
	}
	return true
}
