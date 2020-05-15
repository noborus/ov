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
		if r == '\t' {
			width += 2
		}
		if width >= cursor {
			return i
		}
		i++
	}
	return i
}

func runeWidth(str string) int {
	width := 0
	for _, r := range string(str) {
		width += runewidth.RuneWidth(r)
		if r == '\t' {
			width += 2
		}
	}
	return width
}

func (root *Root) inputEvent(ev *tcell.EventKey, fn func()) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		root.mode = normal
	case tcell.KeyEnter:
		fn()
		root.mode = normal
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if root.cursorX > 0 {
			pos := stringWidth(root.input, root.cursorX)
			runes := []rune(root.input)
			if pos >= 0 {
				root.input = string(runes[:pos])
				root.cursorX = runeWidth(root.input)
				root.input += string(runes[pos+1:])
			}
		}
	case tcell.KeyLeft:
		if root.cursorX > 0 {
			pos := stringWidth(root.input, root.cursorX)
			runes := []rune(root.input)
			if pos >= 0 {
				root.cursorX = runeWidth(string(runes[:pos]))
				if pos > 0 && runes[pos-1] == '\t' {
					root.cursorX--
				}
			}
		}
	case tcell.KeyRight:
		pos := stringWidth(root.input, root.cursorX+1)
		runes := []rune(root.input)
		root.cursorX = runeWidth(string(runes[:pos+1]))
	case tcell.KeyTAB:
		pos := stringWidth(root.input, root.cursorX+1)
		runes := []rune(root.input)
		root.input = string(runes[:pos])
		root.input += "\t"
		root.cursorX += 2
		root.input += string(runes[pos:])
	case tcell.KeyCtrlA:
		root.CaseSensitive = !root.CaseSensitive
	case tcell.KeyRune:
		pos := stringWidth(root.input, root.cursorX+1)
		runes := []rune(root.input)
		root.input = string(runes[:pos])
		r := ev.Rune()
		root.input += string(r)
		root.input += string(runes[pos:])
		root.cursorX += runewidth.RuneWidth(r)
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
