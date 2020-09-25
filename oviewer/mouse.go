package oviewer

import (
	"bytes"
	"log"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

func (root *Root) mouseEvent(ev *tcell.EventMouse) {
	button := ev.Buttons()

	if button == tcell.Button2 {
		root.PasteSelect()
		return
	}

	if !root.mouseSelect && button == tcell.Button1 {
		root.mouseSelect = true
		root.mousePressed = true
		root.x1, root.y1 = ev.Position()
	}

	if root.mousePressed {
		root.x2, root.y2 = ev.Position()
	}

	if root.mouseSelect {
		if button == tcell.ButtonNone {
			root.mousePressed = false
		} else if !root.mousePressed && button != tcell.ButtonNone {
			root.resetSelect()
			if button == tcell.Button1 {
				root.CopySelect()
			}
		}
	}

	if button&tcell.WheelUp != 0 {
		root.moveUp()
		root.moveUp()
		return
	}

	if button&tcell.WheelDown != 0 {
		root.moveDown()
		root.moveDown()
		return
	}
}

func (root *Root) resetSelect() {
	root.mouseSelect = false
	root.mousePressed = false
}

// eventCopySelect represents a mouse select event.
type eventCopySelect struct {
	tcell.EventTime
}

// CopySelect executes a copy select event.
func (root *Root) CopySelect() {
	if !root.checkScreen() {
		return
	}
	ev := &eventCopySelect{}
	ev.SetEventNow()
	go func() {
		root.Screen.PostEventWait(ev)
	}()
}

func (root *Root) drawSelect(x1, y1, x2, y2 int, sel bool) {
	if y1 == y2 {
		if x2 < x1 {
			x1, x2 = x2, x1
		}
		root.reverseLine(y1, x1, x2, sel)
		return
	}

	if y2 < y1 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}
	root.reverseLine(y1, x1, root.vWidth, sel)
	for y := y1 + 1; y < y2; y++ {
		root.reverseLine(y, 0, root.vWidth, sel)
	}
	root.reverseLine(y2, 0, x2, sel)
}

func (root *Root) reverseLine(y int, start int, end int, sel bool) {
	for x := start; x < end; x++ {
		mainc, combc, style, width := root.Screen.GetContent(x, y)
		style = style.Reverse(sel)
		root.Screen.SetContent(x, y, mainc, combc, style)
		x += width - 1
	}
}

func (root *Root) putClipboard() {
	y1 := root.y1
	y2 := root.y2
	x1 := root.x1
	x2 := root.x2

	if y1 == y2 {
		ln := root.lnumber[y1]
		str := root.selectLine(ln.line, x1, x2)
		if len(str) == 0 {
			return
		}
		if err := clipboard.WriteAll(str); err != nil {
			log.Printf("clipboard: %v", err)
		}
		return
	}

	var buff bytes.Buffer

	if y2 < y1 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}

	ln := root.lnumber[y1]
	str := root.selectLine(ln.line, x1, -1)
	if _, err := buff.WriteString(str); err != nil {
		log.Println(err)
		return
	}
	if err := buff.WriteByte('\n'); err != nil {
		log.Println(err)
		return
	}

	maxCopySize := 5000
	for y := y1 + 1; y < y2; y++ {
		ln := root.lnumber[y]
		if ln.branch > 0 {
			continue
		}
		line := root.selectLine(ln.line, 0, -1)
		if _, err := buff.WriteString(line); err != nil {
			log.Println(err)
			return
		}
		if err := buff.WriteByte('\n'); err != nil {
			log.Println(err)
			return
		}
		if buff.Len() > maxCopySize {
			break
		}
	}

	ln = root.lnumber[y2]
	if ln.branch == 0 {
		str = root.selectLine(ln.line, 0, x2)
		if _, err := buff.WriteString(str); err != nil {
			log.Println(err)
		}
	}

	if buff.Len() == 0 {
		return
	}
	if err := clipboard.WriteAll(buff.String()); err != nil {
		log.Println(err)
	}
}

func (root *Root) selectLine(ly int, x1 int, x2 int) string {
	lc, err := root.Doc.lineToContents(ly, root.Doc.TabWidth)
	if err != nil {
		log.Println(err)
		return ""
	}

	size := len(lc.contents)
	if x2 < 0 {
		x2 = size
	}
	if x1 > size {
		x1 = size
	}
	if x2 > size {
		x2 = size
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	str, _ := contentsToStr(lc.contents[x1:x2])
	return str
}

type eventPasteSelect struct {
	tcell.EventTime
}

func (root *Root) PasteSelect() {
	if !root.checkScreen() {
		return
	}
	ev := &eventPasteSelect{}
	ev.SetEventNow()
	go func() {
		root.Screen.PostEventWait(ev)
	}()
}

func (root *Root) getClipboard() {
	input := root.input
	switch input.mode {
	case Normal, Help, LogDoc:
		return
	}

	str, err := clipboard.ReadAll()
	if err != nil {
		log.Printf("%v", err)
		return
	}
	pos := stringWidth(input.value, input.cursorX+1)
	runes := []rune(input.value)
	input.value = string(runes[:pos])
	input.value += str
	input.value += string(runes[pos:])
	input.cursorX += runewidth.StringWidth(str)
}
