package oviewer

import (
	"strconv"
	"time"

	"github.com/gdamore/tcell"
)

// main is manages and executes events in the main routine.
func (root *Root) main() {
	go root.countTimer()

loop:
	for {
		root.draw()
		ev := root.Screen.PollEvent()
		switch ev := ev.(type) {
		case *eventAppQuit:
			break loop
		case *eventTimer:
			root.updateEndNum()
		case *eventModel:
			root.setModel(ev.m)
		case *searchInput:
			root.search(ev.input)
		case *backSearchInput:
			root.backSearch(ev.input)
		case *gotoInput:
			root.goLine(ev.input)
		case *headerInput:
			root.setHeader(ev.input)
		case *delimiterInput:
			root.setDelimiter(ev.input)
		case *tabWidthInput:
			root.setTabWidth(ev.input)
		case *tcell.EventResize:
			root.resize()
		case *tcell.EventKey:
			root.message = ""
			if root.input.mode == Normal {
				root.keyCapture(ev)
			} else {
				root.inputEvent(ev)
			}
		}
	}
}

// eventAppQuit represents a quit event.
type eventAppQuit struct {
	tcell.EventTime
}

// Quit executes a quit event.
func (root *Root) Quit() {
	ev := &eventAppQuit{}
	ev.SetEventNow()
	go func() { root.Screen.PostEventWait(ev) }()
}

func (root *Root) WriteQuit() {
	root.AfterWrite = true
	ev := &eventAppQuit{}
	ev.SetEventNow()
	go func() { root.Screen.PostEventWait(ev) }()
}

// eventTimer represents a timer event.
type eventTimer struct {
	tcell.EventTime
}

// runOnTime runs at time.
func (root *Root) runOnTime() {
	ev := &eventTimer{}
	ev.SetEventNow()
	go func() { root.Screen.PostEventWait(ev) }()
}

// countTimer fires events periodically until it reaches EOF.
func (root *Root) countTimer() {
	timer := time.NewTicker(time.Millisecond * 500)
loop:
	for {
		<-timer.C
		root.runOnTime()
		if root.Model.BufEOF() {
			break loop
		}
	}
	timer.Stop()
}

// MoveLine fires an event that moves to the specified line.
func (root *Root) MoveLine(num int) {
	ev := &gotoInput{}
	ev.input = strconv.Itoa(num)
	ev.SetEventNow()
	go func() {
		root.Screen.PostEventWait(ev)
	}()
}

// MoveTop fires the event of moving to top.
func (root *Root) MoveTop() {
	root.MoveLine(0)
}

// MoveBottom fires the event of moving to bottom.
func (root *Root) MoveBottom() {
	root.MoveLine(root.Model.endNum)
}

// Search fires a forward search event.
func (root *Root) Search(input string) {
	ev := &searchInput{}
	ev.input = input
	ev.SetEventNow()
	go func() {
		root.Screen.PostEventWait(ev)
	}()
}

// BackSearch fires a backward search event.
func (root *Root) BackSearch(input string) {
	ev := &backSearchInput{}
	ev.input = input
	ev.SetEventNow()
	go func() {
		root.Screen.PostEventWait(ev)
	}()
}

// eventModel represents a set model event.
type eventModel struct {
	m *Model
	tcell.EventTime
}

// SetModel fires a set model event.
func (root *Root) SetModel(m *Model) {
	ev := &eventModel{}
	ev.m = m
	ev.SetEventNow()
	go func() {
		root.Screen.PostEventWait(ev)
	}()
}
