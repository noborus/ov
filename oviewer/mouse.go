package oviewer

import (
	"log"

	"github.com/gdamore/tcell"
)

func (root *Root) mouseEvent(ev *tcell.EventMouse) {
	button := ev.Buttons()

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

	root.mouseX, root.mouseY = ev.Position()
	if !root.mouseSelect && button == tcell.Button1 {
		root.mouseSelect = true
		root.oX, root.oY = ev.Position()
	} else if root.mouseSelect && (button == tcell.ButtonNone) {
		root.mouseSelect = false
		log.Printf("x:%d x:%d\n", root.oX, root.mouseX)
		log.Printf("y:%d:%s\n", root.Doc.lineNum+root.oY, root.Doc.GetLine(root.Doc.lineNum+root.oY))
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

func (root *Root) setCopySelect() {
	log.Printf("L:%s", root.Doc.GetLine(root.Doc.lineNum+root.mouseY))
}

func drawSelect(s tcell.Screen, x1, y1, x2, y2 int, sel bool) {
	w, _ := s.Size()

	if y1 == y2 {
		for col := x1; col < w; col++ {
			mainc, combc, style, width := s.GetContent(col, y1)
			style = style.Reverse(sel)
			s.SetContent(col, y1, mainc, combc, style)
			col += width - 1
		}
	}

	if y2 < y1 {
		y1, y2 = y2, y1
	}

	for row := y1 + 1; row < y2; row++ {
		for col := 0; col < w; col++ {
			mainc, combc, style, width := s.GetContent(col, row)
			style = style.Reverse(sel)
			s.SetContent(col, row, mainc, combc, style)
			col += width - 1
		}
	}

	if y1 == y2 {
		return
	}
	for col := 0; col < x2; col++ {
		mainc, combc, style, width := s.GetContent(col, y2)
		style = style.Reverse(sel)
		s.SetContent(col, y2, mainc, combc, style)
		col += width - 1
	}
}
