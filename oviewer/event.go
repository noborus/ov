package oviewer

import (
	"context"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
)

// UpdateInterval is the update interval that calls eventUpdate().
var UpdateInterval = 50 * time.Millisecond

// eventLoop is manages and executes events in the eventLoop routine.
func (root *Root) eventLoop(ctx context.Context, quitChan chan<- struct{}) {
	if root.Doc.WatchMode {
		root.Doc.watchRestart.Store(true)
	}
	go root.updateInterval(ctx)
	defer root.debugNumOfChunk()

	for {
		root.everyUpdate()
		ev := root.Screen.PollEvent()
		switch ev := ev.(type) {
		case *eventAppQuit:
			if root.Doc.documentType != DocHelp && root.Doc.documentType != DocLog {
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
		case *eventCloseAllFilter:
			root.closeAllFilter()
		case *eventCopySelect:
			root.copyToClipboard(ctx)
		case *eventPaste:
			root.pasteFromClipboard(ctx)
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
		case *eventSearchMove:
			root.searchGo(ev.value)
		case *eventInputFilter:
			root.filter(ctx)
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
		case *eventSaveBuffer:
			root.saveBuffer(ev.value)
		case *eventSectionNum:
			root.setSectionNum(ev.value)

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

// sendGoto fires an eventGoto event that moves to the specified line.
func (root *Root) sendGoto(num int) {
	ev := &eventGoto{}
	ev.value = strconv.Itoa(num)
	ev.SetEventNow()
	root.postEvent(ev)
}

// MoveLine fires an eventGoto event that moves to the specified line.
func (root *Root) MoveLine(num int) {
	root.sendGoto(num)
}

// MoveTop fires the event of moving to top.
func (root *Root) MoveTop() {
	root.MoveLine(root.Doc.BufStartNum())
}

// MoveBottom fires the event of moving to bottom.
func (root *Root) MoveBottom() {
	root.MoveLine(root.Doc.BufEndNum())
}

// everyUpdate is called every time before running the event.
func (root *Root) everyUpdate() {
	// If tmpLN is set, set top position to position from bottom.
	// This process is executed when temporary read is switched to normal read.
	if n := atomic.SwapInt32(&root.Doc.tmpLN, 0); n > 0 {
		tmpN := int(n) - root.Doc.topLN
		root.Doc.topLN = root.Doc.BufEndNum() - tmpN
	}

	root.Doc.width = root.scr.vWidth - root.scr.startX
	root.Doc.height = root.Doc.statusPos - root.Doc.headerLen

	if root.General.FollowAll || root.Doc.FollowMode || root.Doc.FollowSection {
		root.follow()
	}

	if !root.skipDraw {
		root.draw()
	}
	root.skipDraw = false

	if root.Doc.watchRestart.Swap(false) == true {
		root.watchControl()
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

// eventAppQuit represents a quit event.
type eventAppQuit struct {
	tcell.EventTime
}

// Quit fires the eventAppQuit event.
func (root *Root) Quit() {
	root.sendQuit()
}

func (root *Root) sendQuit() {
	ev := &eventAppQuit{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// eventAppSuspend represents a suspend event.
type eventAppSuspend struct {
	tcell.EventTime
}

// Suspend fires the eventAppSuspend event.
func (root *Root) Suspend() {
	root.sendSuspend()
}

func (root *Root) sendSuspend() {
	ev := &eventAppSuspend{}
	ev.SetEventNow()
	root.postEvent(ev)
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
	root.sendUpdateEndNum()
}

func (root *Root) sendUpdateEndNum() {
	if !root.hasDocChanged() {
		return
	}
	if !root.Config.Prompt.Normal.ProcessOfCount && !root.Doc.BufEOF() {
		return
	}
	ev := &eventUpdateEndNum{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// eventDocument represents a set document event.
type eventDocument struct {
	tcell.EventTime
	docNum int
}

// SetDocument fires the eventDocument event.
func (root *Root) SetDocument(docNum int) {
	if docNum < 0 || docNum < root.DocumentLen() {
		return
	}
	root.sendDocument(docNum)
}

func (root *Root) sendDocument(docNum int) {
	ev := &eventDocument{}
	ev.docNum = docNum
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
	root.sendAddDocument(m)
}

func (root *Root) sendAddDocument(m *Document) {
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
	root.sendCloseDocument()
}

func (root *Root) sendCloseDocument() {
	ev := &eventCloseDocument{}
	ev.SetEventNow()
	root.postEvent(ev)
}

type eventCloseAllFilter struct {
	tcell.EventTime
}

func (root *Root) CloseAllFilter() {
	root.sendCloseAllFilter()
}

func (root *Root) sendCloseAllFilter() {
	ev := &eventCloseAllFilter{}
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
	root.sendReload(root.Doc)
}

func (root *Root) sendReload(m *Document) {
	ev := &eventReload{}
	ev.m = m
	ev.SetEventNow()
	root.postEvent(ev)
}

// releaseEventBuffer will release all event buffers.
func (root *Root) releaseEventBuffer() {
	for root.Screen.HasPendingEvent() {
		_ = root.Screen.PollEvent()
	}
}

// postEvent is a wrapper for tcell.Event.
func (root *Root) postEvent(ev tcell.Event) {
	if !root.checkScreen() {
		return
	}

	if err := root.Screen.PostEvent(ev); err != nil {
		log.Printf("postEvent %s", err)
		root.releaseEventBuffer()
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
