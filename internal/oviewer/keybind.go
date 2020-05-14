package oviewer

import (
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
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
	case tabWidth:
		b := root.inputEvent(ev, root.SetTabWidth)
		root.Model.ClearCache()
		return b
	default:
		return root.defaultEvent(ev)
	}
}

func stringWidth(str string, cursor int) int {
	width := 0
	i := 0
	for _, r := range string(str) {
		width += runewidth.RuneWidth(r)
		if width >= cursor {
			return i
		}
		i++
	}
	return i
}

func (root *Root) inputEvent(ev *tcell.EventKey, fn func()) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		root.mode = normal
	case tcell.KeyEnter:
		fn()
		root.mode = normal
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if root.cursorX > 1 {
			pos := stringWidth(root.input, root.cursorX)
			runes := []rune(root.input)
			if pos > 0 {
				root.input = string(runes[:pos-1])
				root.cursorX = runewidth.StringWidth(root.input) + 1
				root.input += string(runes[pos:])
			}
		}
	case tcell.KeyLeft:
		if root.cursorX > 1 {
			_, _, _, cw := root.GetContent(root.cursorX, root.statusPos)
			_, _, _, w := root.GetContent(root.cursorX-cw, root.statusPos)
			root.cursorX -= w
		}
	case tcell.KeyRight:
		if root.cursorX <= runewidth.StringWidth(root.input) {
			_, _, _, w := root.GetContent(root.cursorX, root.statusPos)
			root.cursorX += w
		}
	case tcell.KeyTAB:
		root.input += "\t"
		root.cursorX += root.TabWidth
	case tcell.KeyCtrlA:
		root.CaseSensitive = !root.CaseSensitive
	case tcell.KeyRune:
		pos := stringWidth(root.input, root.cursorX)
		runes := []rune(root.input)
		root.input = string(runes[:pos])
		r := ev.Rune()
		root.input += string(r)
		root.input += string(runes[pos:])
		root.cursorX += runewidth.RuneWidth(r)
	}
	root.Screen.ShowCursor(root.cursorX, root.statusPos)
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
