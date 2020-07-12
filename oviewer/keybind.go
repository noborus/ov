package oviewer

import (
	"github.com/gdamore/tcell"
)

// DefaultKeyEvent default input key events.
func (root *Root) defaultKeyEvent(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		root.Quit()
	case tcell.KeyCtrlL:
		root.viewSync()
	case tcell.KeyEnter:
		root.moveDown()
	case tcell.KeyHome:
		root.MoveTop()
	case tcell.KeyEnd:
		root.MoveBottom()
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
			root.Input.SetMode(Backsearch)
		case 'c':
			root.toggleColumnMode()
		case 'd':
			root.Input.SetMode(Delimiter)
		case '/':
			root.Input.SetMode(Search)
		case 'n':
			root.nextSearch()
		case 'N':
			root.nextBackSearch()
		case 'g':
			root.Input.SetMode(Goline)
		case 'G':
			root.toggleLineNumMode()
		case '>':
			root.goLine(newGotoInput(root.Input.GoCandidate).Up(""))
		case '<':
			root.goLine(newGotoInput(root.Input.GoCandidate).Down(""))
		case 'm':
			root.markLineNum(root.lineNum)
		case 'H':
			root.Input.SetMode(Header)
		case 'C':
			root.toggleAlternateRows()
		case 't':
			root.Input.SetMode(TabWidth)
		}
	}
	return true
}
