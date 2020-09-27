package oviewer

import (
	"bytes"
	"fmt"
	"log"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

func (root *Root) mouseEvent(ev *tcell.EventMouse) {
	button := ev.Buttons()

	if button&tcell.WheelUp != 0 {
		root.wheelUp()
		return
	}

	if button&tcell.WheelDown != 0 {
		root.wheelDown()
		return
	}

	if button != tcell.ButtonNone || root.mouseSelect {
		root.selectRange(ev)
		return
	}

	root.skipDraw = true
}

func (root *Root) wheelUp() {
	root.moveUp()
	root.moveUp()
}

func (root *Root) wheelDown() {
	root.moveDown()
	root.moveDown()
}

func (root *Root) selectRange(ev *tcell.EventMouse) {
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
		} else if !root.mousePressed {
			root.resetSelect()
			if button == tcell.Button1 || button == tcell.Button3 {
				root.CopySelect()
			}
		}
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

	if y2 < y1 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}

	ln1 := root.lnumber[y1]
	lc1, err := root.Doc.lineToContents(ln1.line, root.Doc.TabWidth)
	if err != nil {
		root.debugMessage(fmt.Sprintf("%s", err))
		return
	}
	wx1 := root.branchWidth(lc1, ln1.branch)

	ln2 := root.lnumber[y2]
	lc2, err := root.Doc.lineToContents(ln2.line, root.Doc.TabWidth)
	if err != nil {
		root.debugMessage(fmt.Sprintf("%s", err))
		return
	}
	wx2 := root.branchWidth(lc2, ln2.branch)

	if ln1.line == ln2.line {
		str := root.selectLine(ln1.line, root.Doc.x+x1+wx1, root.Doc.x+x2+wx2)
		if len(str) == 0 {
			return
		}
		if err := clipboard.WriteAll(str); err != nil {
			log.Printf("putClipboard: %v", err)
		}
		return
	}

	var buff bytes.Buffer

	str := root.selectLine(ln1.line, root.Doc.x+x1+wx1, -1)
	if _, err := buff.WriteString(str); err != nil {
		log.Println(err)
		return
	}
	if err := buff.WriteByte('\n'); err != nil {
		log.Println(err)
		return
	}

	lnumber := []int{}
	for y := y1 + 1; y < y2; y++ {
		l := root.lnumber[y]
		if l.line == ln1.line || l.line == ln2.line || l.branch > 0 {
			continue
		}
		lnumber = append(lnumber, l.line)
	}
	maxCopySize := 100000
	for _, ln := range lnumber {
		line := root.selectLine(ln, 0, -1)
		if _, err := buff.WriteString(line); err != nil {
			log.Println(err)
			return
		}
		if err := buff.WriteByte('\n'); err != nil {
			log.Println(err)
			return
		}
		if buff.Len() > maxCopySize {
			log.Printf("over size %d", buff.Len())
			break
		}
	}

	str = root.selectLine(ln2.line, 0, root.Doc.x+x2+wx2)
	if _, err := buff.WriteString(str); err != nil {
		log.Println(err)
		return
	}

	if buff.Len() == 0 {
		return
	}
	if err := clipboard.WriteAll(buff.String()); err != nil {
		log.Printf("putClipboard: %v", err)
	}
}

func (root *Root) branchWidth(lc lineContents, branch int) int {
	i := 0
	w := root.startX
	x := 0
	for n := 0; n < len(lc); n++ {
		c := lc[n]
		if w+c.width > root.vWidth {
			i++
			w = root.startX
		}
		if i >= branch {
			break
		}
		w += c.width
		x += c.width
	}
	return x
}

func (root *Root) selectLine(ly int, x1 int, x2 int) string {
	lc, err := root.Doc.lineToContents(ly, root.Doc.TabWidth)
	if err != nil {
		root.debugMessage(fmt.Sprintf("%s", err))
		return ""
	}

	size := len(lc)
	// -1 is a special max value.
	if x2 == -1 {
		x2 = size
	}

	x1 = x1 - root.startX
	x2 = x2 - root.startX
	x1 = max(0, x1)
	x2 = max(0, x2)
	x1 = min(x1, size)
	x2 = min(x2, size)
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	str, _ := contentsToStr(lc[x1:x2])
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
		log.Printf("getClipboard: %v", err)
		return
	}
	pos := stringWidth(input.value, input.cursorX+1)
	runes := []rune(input.value)
	input.value = string(runes[:pos])
	input.value += str
	input.value += string(runes[pos:])
	input.cursorX += runewidth.StringWidth(str)
}
