package oviewer

import (
	"context"
	"log"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// WheelScrollNum is the number of lines to scroll with the mouse wheel.
var WheelScrollNum = 2

// mouseEvent handles mouse events.
func (root *Root) mouseEvent(ctx context.Context, ev *tcell.EventMouse) {
	button := ev.Buttons()
	mod := ev.Modifiers()

	var horizontal bool
	if mod&tcell.ModShift != 0 {
		horizontal = true
	}

	if button&tcell.WheelUp != 0 {
		if horizontal {
			root.wheelLeft(ctx)
			return
		}
		root.wheelUp(ctx)
		return
	}

	if button&tcell.WheelDown != 0 {
		if horizontal {
			root.wheelRight(ctx)
			return
		}
		root.wheelDown(ctx)
		return
	}

	if button&tcell.WheelLeft != 0 {
		root.wheelLeft(ctx)
		return
	}
	if button&tcell.WheelRight != 0 {
		root.wheelRight(ctx)
		return
	}

	if button != tcell.ButtonNone || root.scr.mouseSelect {
		root.selectRange(ctx, ev)
		return
	}

	// Avoid redrawing with other mouse events.
	root.skipDraw = true
}

// wheelUp moves the mouse wheel up.
func (root *Root) wheelUp(context.Context) {
	root.setMessage("")
	root.moveUp(WheelScrollNum)
}

// wheelDown moves the mouse wheel down.
func (root *Root) wheelDown(context.Context) {
	root.setMessage("")
	root.moveDown(WheelScrollNum)
}

func (root *Root) wheelRight(ctx context.Context) {
	root.setMessage("")
	if root.Doc.ColumnMode {
		root.moveRightOne(ctx)
	} else {
		root.moveRight(root.Doc.HScrollWidthNum)
	}
}

func (root *Root) wheelLeft(ctx context.Context) {
	root.setMessage("")
	if root.Doc.ColumnMode {
		root.moveLeftOne(ctx)
	} else {
		root.moveLeft(root.Doc.HScrollWidthNum)
	}
}

// selectRange saves the position by selecting the range with the mouse.
// The selection range is represented by (x1, y1), (x2, y2).
func (root *Root) selectRange(ctx context.Context, ev *tcell.EventMouse) {
	button := ev.Buttons()
	if button == tcell.ButtonMiddle {
		root.Paste(ctx)
		return
	}

	if !root.scr.mouseSelect && button == tcell.ButtonPrimary {
		root.setMessage("")
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			root.scr.mouseRectangle = true
		} else {
			root.scr.mouseRectangle = false
		}
		root.scr.mouseSelect = true
		root.scr.mousePressed = true
		root.scr.x1, root.scr.y1 = ev.Position()
	}

	if root.scr.mousePressed {
		root.scr.x2, root.scr.y2 = ev.Position()
	}

	if root.scr.mouseSelect {
		if button == tcell.ButtonNone {
			root.scr.mousePressed = false
		} else if !root.scr.mousePressed {
			root.resetSelect()
			if button == tcell.ButtonPrimary || button == tcell.ButtonSecondary {
				root.CopySelect(ctx)
			}
		}
	}
}

// resetSelect resets the selection.
func (root *Root) resetSelect() {
	root.scr.mouseSelect = false
	root.scr.mousePressed = false
}

// eventCopySelect represents a mouse select event.
type eventCopySelect struct {
	tcell.EventTime
}

// CopySelect fires the eventCopySelect event.
func (root *Root) CopySelect(context.Context) {
	root.sendCopySelect()
}

func (root *Root) sendCopySelect() {
	ev := &eventCopySelect{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// copyToClipboard writes the selection to the clipboard.
func (root *Root) copyToClipboard(_ context.Context) {
	str, err := root.rangeToString(root.scr.x1, root.scr.y1, root.scr.x2, root.scr.y2)
	if err != nil {
		root.debugMessage("copyToClipboard: " + err.Error())
		return
	}
	if len(str) == 0 {
		return
	}

	root.copyClipboard(str)
	root.setMessage("Copy")
}

// copyClipboard copies the string to the clipboard.
// The method of copying to the clipboard is determined by the configuration.
// The default method is to use the clipboard package.
// "OSC52” can be specified (available only if the terminal supports it).
func (root *Root) copyClipboard(str string) {
	switch root.Config.ClipboardMethod {
	case "OSC52":
		root.Screen.SetClipboard([]byte(str))
	default:
		if err := clipboard.WriteAll(str); err != nil {
			log.Printf("copyToClipboard: %v\n", err)
		}
	}
}

type eventPaste struct {
	tcell.EventTime
}

// Paste fires the eventPaste event.
func (root *Root) Paste(context.Context) {
	root.sendPaste()
}

func (root *Root) sendPaste() {
	ev := &eventPaste{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// pasteFromClipboard writes a string from the clipboard.
func (root *Root) pasteFromClipboard(context.Context) {
	input := root.input
	if input.Event.Mode() == Normal {
		return
	}

	str, err := clipboard.ReadAll()
	if err != nil {
		log.Printf("pasteFromClipboard: %v\n", err)
		return
	}

	pos := countToCursor(input.value, input.cursorX+1)
	runes := []rune(input.value)
	input.value = string(runes[:pos])
	input.value += str
	input.value += string(runes[pos:])
	input.cursorX += runewidth.StringWidth(str)
}

func (root *Root) rangeToString(x1, y1, x2, y2 int) (string, error) {
	x1, y1, x2, y2 = normalizeRange(x1, y1, x2, y2)
	if root.scr.mouseRectangle {
		return root.scr.rectangleToString(root.Doc, x1, y1, x2, y2)
	}
	return root.scr.lineRangeToString(root.Doc, x1, y1, x2, y2)
}

func normalizeRange(x1, y1, x2, y2 int) (int, int, int, int) {
	if y2 < y1 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}
	if y1 == y2 {
		if x2 < x1 {
			x1, x2 = x2, x1
		}
	}
	return x1, y1, x2, y2
}

// lineRangeToString returns the selection.
// The arguments must be x1 <= x2 and y1 <= y2.
// ...◼◻◻◻◻◻
// ◻◻◻◻◻◻◻◻
// ◻◻◻◼......
func (scr SCR) lineRangeToString(m *Document, x1, y1, x2, y2 int) (string, error) {
	l1 := scr.lineNumber(y1)
	line1, ok := scr.lines[l1.number]
	if !ok || !line1.valid {
		return "", ErrOutOfRange
	}
	wx1 := scr.branchWidth(line1.lc, l1.wrap)

	var l2 LineNumber
	for y := y2; ; y-- {
		l := scr.lineNumber(y)
		if l.number < m.BufEndNum() {
			l2 = l
			y2 = y
			break
		}
	}

	line2, ok := scr.lines[l2.number]
	if !ok || !line2.valid {
		return "", ErrOutOfRange
	}
	wx2 := scr.branchWidth(line2.lc, l2.wrap)

	if l1.number == l2.number {
		x1 := m.x + x1 + wx1
		x2 := m.x + x2 + wx2
		return scr.selectLine(line1, x1, x2+1), nil
	}

	var buff strings.Builder

	first := scr.selectLine(line1, m.x+x1+wx1, -1)
	buff.WriteString(first)
	buff.WriteByte('\n')

	for y := y1 + 1; y < y2; y++ {
		ln := scr.lineNumber(y)
		if ln.number == l1.number || ln.number == l2.number || ln.wrap > 0 {
			continue
		}
		line, ok := scr.lines[ln.number]
		if !ok || !line.valid {
			break
		}
		str := scr.selectLine(line, 0, -1)
		buff.WriteString(str)
		buff.WriteByte('\n')
	}

	last := scr.selectLine(line2, 0, m.x+x2+wx2+1)
	buff.WriteString(last)
	return buff.String(), nil
}

// rectangleToByte returns a rectangular range.
// The arguments must be x1 <= x2 and y1 <= y2.
// ...◼◻◻...
// ...◻◻◻...
// ...◻◻◼...
func (scr SCR) rectangleToString(m *Document, x1, y1, x2, y2 int) (string, error) {
	var buff strings.Builder
	for y := y1; y <= y2; y++ {
		ln := scr.lineNumber(y)
		lineC, ok := scr.lines[ln.number]
		if !ok || !lineC.valid {
			break
		}
		wx := scr.branchWidth(lineC.lc, ln.wrap)
		str := scr.selectLine(lineC, m.x+x1+wx, m.x+x2+wx+1)
		buff.WriteString(str)
		buff.WriteByte('\n')
	}
	return buff.String(), nil
}

// branchWidth returns the leftmost position of the number of wrapped line.
// If the wide character is at the right end, it may wrap one character forward.
func (scr SCR) branchWidth(lc contents, branch int) int {
	i := 0
	w := scr.startX
	x := 0
	for n := range lc {
		c := lc[n]
		if w+c.width > scr.vWidth {
			i++
			w = scr.startX
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
// The arguments must be x1 <= x2.
func (scr SCR) selectLine(lineC LineC, x1 int, x2 int) string {
	maxX := len(lineC.lc)
	// -1 is a special max value.
	if x2 == -1 {
		x2 = maxX
	}
	x1 -= scr.startX
	x2 -= scr.startX
	x1 = max(0, x1)
	x2 = max(0, x2)
	x1 = min(x1, maxX)
	x2 = min(x2, maxX)
	start := lineC.pos.n(x1)
	end := lineC.pos.n(x2)

	if start > len(lineC.str) || end > len(lineC.str) || start > end {
		log.Printf("selectLine:len(%d):start(%d):end(%d)\n", len(lineC.str), start, end)
		return ""
	}
	return lineC.str[start:end]
}
