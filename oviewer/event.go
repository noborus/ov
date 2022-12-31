package oviewer

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
)

// UpdateInterval is the update interval that calls eventUpdate().
var UpdateInterval = 50 * time.Millisecond

// eventLoop is manages and executes events in the eventLoop routine.
func (root *Root) eventLoop(ctx context.Context, quitChan chan<- struct{}) {
	if root.Doc.WatchMode {
		root.watchStart()
	}
	go root.updateInterval(ctx)

	for {
		if root.General.FollowAll || root.Doc.FollowMode || root.Doc.FollowSection {
			root.follow()
		}

		if !root.skipDraw {
			root.draw()
		}
		root.skipDraw = false

		ev := root.Screen.PollEvent()
		switch ev := ev.(type) {
		case *eventAppQuit:
			if root.screenMode == Docs {
				close(quitChan)
				return
			}
			// Help and logDoc return to Doc display.
			root.toNormal()
		case *eventReload:
			root.reload(ev.m)
		case *eventAppSuspend:
			root.suspend()
		case *eventUpdateEndNum:
			root.updateEndNum()
		case *eventDocument:
			root.switchDocument(ev.docNum)
		case *eventAddDocument:
			root.addDocument(ev.m)
		case *eventCloseDocument:
			root.closeDocument()
		case *eventCopySelect:
			root.putClipboard(ctx)
		case *eventPaste:
			root.getClipboard(ctx)
		case *eventViewMode:
			root.setViewMode(ev.value)
		case *eventInputSearch:
			root.firstSearch(ctx)
		case *eventNextSearch:
			root.nextSearch(ctx, ev.str)
		case *eventInputBackSearch:
			root.firstBackSearch(ctx)
		case *eventNextBackSearch:
			root.nextBackSearch(ctx, ev.str)
		case *eventGoto:
			root.goLine(ev.value)
		case *eventHeader:
			root.setHeader(ev.value)
		case *eventSkipLines:
			root.setSkipLines(ev.value)
		case *eventDelimiter:
			root.setDelimiter(ev.value)
		case *eventTabWidth:
			root.setTabWidth(ev.value)
		case *eventWatchInterval:
			root.setWatchInterval(ev.value)
		case *eventWriteBA:
			root.setWriteBA(ev.value)
		case *eventSectionDelimiter:
			root.setSectionDelimiter(ev.value)
		case *eventSectionStart:
			root.setSectionStart(ev.value)
		case *eventMultiColor:
			root.setMultiColor(ev.value)
		case *eventJumpTarget:
			root.setJumpTarget(ev.value)

		// tcell events
		case *tcell.EventResize:
			root.resize()
		case *tcell.EventMouse:
			root.mouseEvent(ev)
		case *tcell.EventKey:
			root.keyEvent(ctx, ev)
		case nil:
			close(quitChan)
			return
		}
	}
}

// keyEvent processes key events.
func (root *Root) keyEvent(ctx context.Context, ev *tcell.EventKey) {
	root.setMessage("")
	switch root.input.Event.Mode() {
	case Normal:
		root.keyCapture(ev)
	default:
		root.inputEvent(ctx, ev)
	}
}

// checkScreen returns true if the screen is ready.
// checkScreen is used in case it is called directly from the outside.
// True if called from the event loop.
func (root *Root) checkScreen() bool {
	if root == nil {
		return false
	}
	return root.Screen != nil
}

// eventAppQuit represents a quit event.
type eventAppQuit struct {
	tcell.EventTime
}

// Quit fires the eventAppQuit event.
func (root *Root) Quit() {
	if !root.checkScreen() {
		return
	}
	ev := &eventAppQuit{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// eventAppSuspend represents a suspend event.
type eventAppSuspend struct {
	tcell.EventTime
}

// Suspend fires the eventAppsuspend event.
func (root *Root) Suspend() {
	if !root.checkScreen() {
		return
	}
	ev := &eventAppSuspend{}
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// Cancel follow mode and follow all mode.
func (root *Root) Cancel() {
	root.General.FollowAll = false
	root.Doc.FollowMode = false
}

// WriteQuit sets the write flag and executes a quit event.
func (root *Root) WriteQuit() {
	root.IsWriteOriginal = true
	root.Quit()
}

// follow updates the document in follow mode.
func (root *Root) follow() {
	if root.General.FollowAll {
		root.followAll()
	}

	root.Doc.onceFollowMode()

	num := root.Doc.BufEndNum()
	if root.Doc.latestNum == num {
		return
	}

	root.skipDraw = false
	if root.Doc.FollowSection {
		root.tailSection()
	} else {
		root.TailSync()
	}
	root.Doc.latestNum = num
}

// followAll monitors and switches all document updates
// in follow all mode.
func (root *Root) followAll() {
	if root.screenMode != Docs {
		return
	}

	current := root.CurrentDoc

	root.mu.RLock()
	for n, doc := range root.DocList {
		doc.onceFollowMode()
		if doc.latestNum != doc.BufEndNum() {
			current = n
		}
	}
	root.mu.RUnlock()

	if root.CurrentDoc != current {
		log.Printf("switch document: %d", current)
		root.switchDocument(current)
	}
}

// updateInterval calls eventUpdate at regular intervals.
func (root *Root) updateInterval(ctx context.Context) {
	timer := time.NewTicker(UpdateInterval)
	for {
		select {
		case <-timer.C:
			root.regularUpdate()
		case <-ctx.Done():
			return
		}
	}
}

// eventUpdateEndNum represents a timer event.
type eventUpdateEndNum struct {
	tcell.EventTime
}

// regularUpdate fires an eventUpdateEndNum event when an update is required.
func (root *Root) regularUpdate() {
	if !root.checkScreen() {
		return
	}

	if !root.hasDocChanged() {
		return
	}

	ev := &eventUpdateEndNum{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// MoveLine fires an eventGoto event that moves to the specified line.
func (root *Root) MoveLine(num int) {
	if !root.checkScreen() {
		return
	}
	ev := &eventGoto{}
	ev.value = strconv.Itoa(num)
	ev.SetEventNow()
	root.postEvent(ev)
}

// MoveTop fires the event of moving to top.
func (root *Root) MoveTop() {
	root.MoveLine(0)
}

// MoveBottom fires the event of moving to bottom.
func (root *Root) MoveBottom() {
	root.MoveLine(root.Doc.BufEndNum())
}

// eventDocument represents a set document event.
type eventDocument struct {
	docNum int
	tcell.EventTime
}

// SetDocument fires the eventDocument event.
func (root *Root) SetDocument(docNum int) {
	if !root.checkScreen() {
		return
	}
	ev := &eventDocument{}
	if docNum >= 0 && docNum < root.DocumentLen() {
		ev.docNum = docNum
	}
	ev.SetEventNow()
	root.postEvent(ev)
}

// eventAddDocument represents a set document event.
type eventAddDocument struct {
	m *Document
	tcell.EventTime
}

// AddDocument fires the eventAddDocument event.
func (root *Root) AddDocument(m *Document) {
	if !root.checkScreen() {
		return
	}
	ev := &eventAddDocument{}
	ev.m = m
	ev.SetEventNow()
	root.postEvent(ev)
}

// eventCloseDocument represents a close document event.
type eventCloseDocument struct {
	tcell.EventTime
}

// CloseDocument fires the eventCloseDocument event.
func (root *Root) CloseDocument(m *Document) {
	if !root.checkScreen() {
		return
	}
	ev := &eventCloseDocument{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// eventReload represents a reload event.
type eventReload struct {
	m *Document
	tcell.EventTime
}

// Reload fires the eventReload event.
func (root *Root) Reload() {
	if !root.checkScreen() {
		return
	}
	root.setMessagef("reload %s", root.Doc.FileName)
	log.Printf("reload %s", root.Doc.FileName)
	ev := &eventReload{}
	ev.SetEventNow()
	ev.m = root.Doc
	ev.SetEventNow()
	root.postEvent(ev)
}

// releaseEventBuffer will release all event buffers.
func (root *Root) releaseEventBuffer() {
	for root.HasPendingEvent() {
		_ = root.Screen.PollEvent()
	}
}

// postEvent is a wrapper for tcell.Event.
func (root *Root) postEvent(ev tcell.Event) {
	if err := root.Screen.PostEvent(ev); err != nil {
		log.Println(err)
	}
}
