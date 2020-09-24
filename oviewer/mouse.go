package oviewer

import (
	"bytes"
	"log"
	"regexp"

	"github.com/d-tsuji/clipboard"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

func (root *Root) mouseEvent(ev *tcell.EventMouse) {
	button := ev.Buttons()

	if button == tcell.Button2 {
		root.mousePaste()
		return
	}

	if !root.mouseSelect && button == tcell.Button1 {
		root.mouseSelect = true
		root.mousePressed = true
		x, y := ev.Position()
		root.x1, root.y1 = root.logicalPos(x, y)
	}

	if root.mousePressed {
		x, y := ev.Position()
		root.x2, root.y2 = root.logicalPos(x, y)
	}

	if root.mouseSelect {
		if button == tcell.ButtonNone {
			root.mousePressed = false
		} else if !root.mousePressed && button != tcell.ButtonNone {
			root.mouseSelect = false
			root.mousePressed = false
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

func (root *Root) logicalPos(x int, y int) (int, int) {
	logical := root.lnumber[y]
	lx := 0
	for n := 0; n < logical.branch; n++ {
		lx += root.vWidth
	}
	lx = lx + x
	return lx, logical.line
}

func (root *Root) displayPos(lx int, ly int) (int, int) {
	for n := 0; n < len(root.lnumber); n++ {
		logical := root.lnumber[n]
		if logical.line >= ly {
			branch := lx / root.vWidth
			x := lx - (branch * root.vWidth)
			return x, n + branch
		}
	}
	return 0, 0
}

func (root *Root) drawSelect(lx1, ly1, lx2, ly2 int, sel bool) {
	x1, y1 := root.displayPos(lx1, ly1)
	x2, y2 := root.displayPos(lx2, ly2)
	if y1 == y2 {
		if x2 < x1 {
			x1, x2 = x2, x1
		}
		for col := x1; col < x2; col++ {
			mainc, combc, style, width := root.Screen.GetContent(col, y1)
			style = style.Reverse(sel)
			root.Screen.SetContent(col, y1, mainc, combc, style)
			col += width - 1
		}
		return
	}

	if y2 < y1 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}

	for col := x1; col < root.vWidth; col++ {
		mainc, combc, style, width := root.Screen.GetContent(col, y1)
		style = style.Reverse(sel)
		root.Screen.SetContent(col, y1, mainc, combc, style)
		col += width - 1
	}

	for row := y1 + 1; row < y2; row++ {
		for col := 0; col < root.vWidth; col++ {
			mainc, combc, style, width := root.Screen.GetContent(col, row)
			style = style.Reverse(sel)
			root.Screen.SetContent(col, row, mainc, combc, style)
			col += width - 1
		}
	}

	for col := 0; col < x2; col++ {
		mainc, combc, style, width := root.Screen.GetContent(col, y2)
		style = style.Reverse(sel)
		root.Screen.SetContent(col, y2, mainc, combc, style)
		col += width - 1
	}
}

func (root *Root) mousePaste() {
	input := root.input
	switch input.mode {
	case Normal, Help, LogDoc:
		return
	}

	str, err := clipboard.Get()
	if err != nil {
		log.Printf("%v", err)
	}
	pos := stringWidth(input.value, input.cursorX+1)
	runes := []rune(input.value)
	input.value = string(runes[:pos])
	input.value += str
	input.value += string(runes[pos:])
	input.cursorX += runewidth.StringWidth(str)
}

func (root *Root) setCopySelect() {
	y1 := root.y1
	y2 := root.y2
	x1 := root.x1
	x2 := root.x2

	if y1 == y2 {
		str := root.selectLine(y1, x1, x2)
		if err := clipboard.Set(str); err != nil {
			log.Printf("clipboard: %v", err)
		}
		return
	}

	var buff bytes.Buffer

	if y2 < y1 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}

	str := root.selectLine(y1, x1, -1)
	if _, err := buff.WriteString(str); err != nil {
		log.Println(err)
		return
	}
	if err := buff.WriteByte('\n'); err != nil {
		log.Println(err)
		return
	}

	for y := y1 + 1; y < y2; y++ {
		line := root.Doc.GetLine(y)
		if _, err := buff.WriteString(line); err != nil {
			log.Println(err)
			return
		}
		if err := buff.WriteByte('\n'); err != nil {
			log.Println(err)
			return
		}
	}

	str = root.selectLine(y2, 0, x2)
	if _, err := buff.WriteString(str); err != nil {
		log.Println(err)
	}

	s := stripEscapeSequence.ReplaceAllString(buff.String(), "")
	if err := clipboard.Set(s); err != nil {
		log.Println(err)
	}
}

func (root *Root) selectLine(y int, x1 int, x2 int) string {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	line := root.Doc.GetLine(y)
	if x2 < 0 {
		x2 = len(line)
	}

	lc, err := root.Doc.lineToContents(y, root.Doc.TabWidth)
	if err != nil {
		log.Println(err)
		return ""
	}

	sx := lc.contentsByteNum(x1)
	ex := lc.contentsByteNum(x2)
	s := substring(line, sx, ex)

	return stripEscapeSequence.ReplaceAllString(s, "")
}

var stripBackSpace = regexp.MustCompile(".\b")

func substring(str string, start int, end int) string {
	var subs bytes.Buffer
	byteNum := 0
	sFlag := false
	str = stripBackSpace.ReplaceAllString(str, "")

	for _, r := range str {
		byteNum += len(string(r))
		if byteNum > start {
			sFlag = true
		}
		if sFlag {
			if byteNum > end {
				break
			}
			_, err := subs.WriteRune(r)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return subs.String()
}
