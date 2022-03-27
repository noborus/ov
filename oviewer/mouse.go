package oviewer

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// mouseEvent handles mouse events.
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

	// Avoid redrawing with other mouse events.
	root.skipDraw = true
}

// wheelUp moves the mouse wheel up.
func (root *Root) wheelUp() {
	root.setMessage("")
	root.moveUp()
	root.moveUp()
}

// wheelDown moves the mouse wheel down.
func (root *Root) wheelDown() {
	root.setMessage("")
	root.moveDown()
	root.moveDown()
}

// selectRange saves the position by selecting the range with the mouse.
// The selection range is represented by (x1, y1), (x2, y2).
func (root *Root) selectRange(ev *tcell.EventMouse) {
	button := ev.Buttons()
	if button == tcell.ButtonMiddle {
		root.Paste()
		return
	}

	if !root.mouseSelect && button == tcell.ButtonPrimary {
		root.setMessage("")
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			root.mouseRectangle = true
		} else {
			root.mouseRectangle = false
		}
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
			if button == tcell.ButtonPrimary || button == tcell.ButtonSecondary {
				root.CopySelect()
			}
		}
	}
}

// resetSelect resets the selection.
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
		err := root.Screen.PostEvent(ev)
		if err != nil {
			log.Println(err)
		}
	}()
}

// drawSelect highlights the selection.
// Multi-line selection is included until the end of the line,
// but if the rectangle flag is true, the rectangle will be the range.
func (root *Root) drawSelect(x1, y1, x2, y2 int, sel bool) {
	if y1 == y2 {
		if x2 < x1 {
			x1, x2 = x2, x1
		}
		root.reverseLine(y1, x1, x2+1, sel)
		return
	}

	if y2 < y1 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}

	if root.mouseRectangle {
		root.drawRectangle(x1, y1, x2, y2, sel)
		return
	}

	root.reverseLine(y1, x1, root.vWidth, sel)
	for y := y1 + 1; y < y2; y++ {
		root.reverseLine(y, 0, root.vWidth, sel)
	}
	root.reverseLine(y2, 0, x2+1, sel)
}

// drawRectangle highlights the rectangular area.
func (root *Root) drawRectangle(x1, y1, x2, y2 int, sel bool) {
	for y := y1; y <= y2; y++ {
		root.reverseLine(y, x1, x2+1, sel)
	}
}

// reverseLine reverses one line.
func (root *Root) reverseLine(y int, start int, end int, sel bool) {
	for x := start; x < end; x++ {
		mainc, combc, style, width := root.Screen.GetContent(x, y)
		style = style.Reverse(sel)
		root.Screen.SetContent(x, y, mainc, combc, style)
		x += width - 1
	}
}

// putClipboard writes the selection to the clipboard.
func (root *Root) putClipboard(_ context.Context) {
	y1 := root.y1
	y2 := root.y2
	x1 := root.x1
	x2 := root.x2

	if y2 < y1 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}

	buff, err := root.rangeToString(x1, y1, x2, y2)
	if err != nil {
		root.debugMessage(fmt.Sprintf("putClipboard: %s", err.Error()))
		return
	}

	if len(buff) == 0 {
		return
	}
	if err := clipboard.WriteAll(buff); err != nil {
		log.Printf("putClipboard: %v", err)
	}
	root.setMessage("Copy")
}

// rangeToString returns the selection.
func (root *Root) rangeToString(x1, y1, x2, y2 int) (string, error) {
	if root.mouseRectangle {
		return root.rectangleToString(x1, y1, x2, y2)
	}

	var buff strings.Builder

	ln1 := root.lnumber[y1]
	lc1, err := root.Doc.contentsLN(ln1.line, root.Doc.TabWidth)
	if err != nil {
		return "", err
	}
	wx1 := root.branchWidth(lc1, ln1.wrap)

	ln2 := root.lnumber[y2]
	lc2, err := root.Doc.contentsLN(ln2.line, root.Doc.TabWidth)
	if err != nil {
		return "", err
	}
	wx2 := root.branchWidth(lc2, ln2.wrap)

	if ln1.line == ln2.line {
		str := root.selectLine(ln1.line, root.Doc.x+x1+wx1, root.Doc.x+x2+wx2+1)
		if len(str) == 0 {
			return buff.String(), nil
		}
		if _, err := buff.WriteString(str); err != nil {
			return "", err
		}
		return buff.String(), nil
	}

	str := root.selectLine(ln1.line, root.Doc.x+x1+wx1, -1)
	if _, err := buff.WriteString(str); err != nil {
		return "", err
	}
	if err := buff.WriteByte('\n'); err != nil {
		return "", err
	}

	lnumber := []int{}
	for y := y1 + 1; y < y2; y++ {
		l := root.lnumber[y]
		if l.line == ln1.line || l.line == ln2.line || l.wrap > 0 {
			continue
		}
		lnumber = append(lnumber, l.line)
	}
	for _, ln := range lnumber {
		line := root.selectLine(ln, 0, -1)
		if _, err := buff.WriteString(line); err != nil {
			return "", err
		}
		if err := buff.WriteByte('\n'); err != nil {
			return "", err
		}
	}

	str = root.selectLine(ln2.line, 0, root.Doc.x+x2+wx2+1)
	if _, err := buff.WriteString(str); err != nil {
		return "", err
	}

	return buff.String(), nil
}

// rectangleToByte returns a rectangular range.
func (root *Root) rectangleToString(x1, y1, x2, y2 int) (string, error) {
	var buff strings.Builder

	for y := y1; y <= y2; y++ {
		ln := root.lnumber[y]
		lc, err := root.Doc.contentsLN(ln.line, root.Doc.TabWidth)
		if err != nil {
			return "", err
		}
		wx := root.branchWidth(lc, ln.wrap)
		line := root.selectLine(ln.line, root.Doc.x+x1+wx, root.Doc.x+x2+wx+1)

		if _, err := buff.WriteString(line); err != nil {
			return "", err
		}
		if err := buff.WriteByte('\n'); err != nil {
			return "", err
		}
	}
	return buff.String(), nil
}

// branchWidth returns the leftmost position of the number of wrapped line.
func (root *Root) branchWidth(lc contents, branch int) int {
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

// selectLine returns a string in the specified range on one line.
func (root *Root) selectLine(ly int, x1 int, x2 int) string {
	lc, err := root.Doc.contentsLN(ly, root.Doc.TabWidth)
	if err != nil {
		root.debugMessage(fmt.Sprintf("%s", err))
		return ""
	}

	size := len(lc)
	// -1 is a special max value.
	if x2 == -1 {
		x2 = size
	}

	x1 -= root.startX
	x2 -= root.startX
	x1 = max(0, x1)
	x2 = max(0, x2)
	x1 = min(x1, size)
	x2 = min(x2, size)
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	str, _ := ContentsToStr(lc[x1:x2])
	return str
}

type eventPaste struct {
	tcell.EventTime
}

// Paste executes the mouse paste event.
func (root *Root) Paste() {
	if !root.checkScreen() {
		return
	}
	ev := &eventPaste{}
	ev.SetEventNow()
	go func() {
		err := root.Screen.PostEvent(ev)
		if err != nil {
			log.Println(err)
		}
	}()
}

// getClipboard writes a string from the clipboard.
func (root *Root) getClipboard(_ context.Context) {
	input := root.input
	switch input.mode {
	case Normal:
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
