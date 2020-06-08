package oviewer

import (
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

// HandleInputEvent handles input events.
func (root *Root) HandleInputEvent(ev *tcell.EventKey) bool {
	// Input not confirmed or canceled.
	ok := root.inputEvent(ev)
	if !ok {
		// Input is not confirmed.
		return true
	}

	// Input is confirmed.
	nev := root.EventInput.Confirm(root.input)
	go func() { root.Screen.PostEventWait(nev) }()

	root.mode = normal
	return true
}

func (root *Root) inputEvent(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		root.mode = normal
		return true
	case tcell.KeyEnter:
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if root.cursorX == 0 {
			return false
		}
		pos := stringWidth(root.input, root.cursorX)
		runes := []rune(root.input)
		root.input = string(runes[:pos])
		root.cursorX = runeWidth(root.input)
		root.input += string(runes[pos+1:])
	case tcell.KeyDelete:
		pos := stringWidth(root.input, root.cursorX)
		runes := []rune(root.input)
		dp := 1
		if root.cursorX == 0 {
			dp = 0
		}
		root.input = string(runes[:pos+dp])
		if len(runes) > pos+1 {
			root.input += string(runes[pos+dp+1:])
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
	case tcell.KeyUp:
		root.input = root.EventInput.Up(root.input)
		runes := []rune(root.input)
		root.cursorX = runeWidth(string(runes))
	case tcell.KeyDown:
		root.input = root.EventInput.Down(root.input)
		runes := []rune(root.input)
		root.cursorX = runeWidth(string(runes))
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
	return false
}

func stringWidth(str string, cursor int) int {
	width := 0
	i := 0
	for _, r := range str {
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
	for _, r := range str {
		width += runewidth.RuneWidth(r)
		if r == '\t' {
			width += 2
		}
	}
	return width
}
