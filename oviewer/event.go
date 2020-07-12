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
		case *eventTimer:
			root.updateEndNum()
		case *eventAppQuit:
			break loop
		case *searchInput:
			root.search(root.Input.value)
		case *backSearchInput:
			root.backSearch(root.Input.value)
		case *gotoInput:
			root.goLine(root.Input.value)
		case *headerInput:
			root.setHeader(root.Input.value)
		case *delimiterInput:
			root.setDelimiter(root.Input.value)
		case *tabWidthInput:
			root.setTabWidth(root.Input.value)
		case *tcell.EventKey:
			root.message = ""
			if root.Input.mode == Normal {
				root.defaultKeyEvent(ev)
			} else {
				root.inputEvent(ev)
			}
		case *tcell.EventResize:
			root.resize()
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
	root.Input.value = strconv.Itoa(num)
	ev := &gotoInput{}
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
