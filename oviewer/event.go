package oviewer

import (
	"context"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v3"
)

// UpdateInterval is the update interval that calls eventUpdate().
var UpdateInterval = 50 * time.Millisecond

// eventLoop is manages and executes events in the eventLoop routine.
func (root *Root) eventLoop(ctx context.Context, quitChan chan<- struct{}) {
	if root.Doc.WatchMode {
		atomic.StoreInt32(&root.Doc.watchRestart, 1)
	}
	go root.updateInterval(ctx)
	defer root.debugNumOfChunk()

	for {
		root.everyUpdate(ctx)
		ev := <-root.Screen.EventQ()
		if quit := root.event(ctx, ev); quit {
			close(quitChan)
			return
		}
	}
}

// event represents the event processing.
func (root *Root) event(ctx context.Context, ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *eventAppQuit:
		if root.Doc.documentType != DocHelp && root.Doc.documentType != DocLog {
			return true
		}
		// Help and logDoc return to Doc display.
		root.toNormal(ctx)
	case *eventReload:
		root.reload(ev.m)
	case *eventAppSuspend:
		root.suspend(ctx)
	case *eventUpdateEndNum:
		root.updateEndNum()
	case *eventDocument:
		root.switchDocument(ctx, ev.docNum)
	case *eventAddDocument:
		root.addDocument(ctx, ev.m)
	case *eventCloseDocument:
		root.closeDocument(ctx)
	case *eventCloseAllFilter:
		root.closeAllFilter(ctx)
	case *eventCopySelect:
		root.copyToClipboard(ctx)
	case *eventPaste:
		root.pasteFromClipboard(ctx)
	case *eventSearch:
		root.forwardSearch(ctx, ev.str, 0)
	case *eventNextSearch:
		root.forwardSearch(ctx, ev.str, 1)
	case *eventNextBackSearch:
		root.backSearch(ctx, ev.str, -1)
	case *eventSearchMove:
		root.searchGo(ctx, ev.ln, ev.searcher)
	case *eventReachEOF:
		// Quit if small doc and config allows
		if root.quitCheck() {
			return true
		}
		root.Config.QuitSmall = false
		root.notifyEOFReached(ev.m)

	// Input confirmation action event.
	case *eventConverter:
		root.setConverter(ctx, ev.value)
	case *eventDelimiter:
		root.setDelimiter(ev.value)
	case *eventGoto:
		root.goLine(ev.value)
	case *eventHeaderColumn:
		root.setHeaderColumn(ev.value)
	case *eventHeader:
		root.setHeader(ev.value)
	case *eventJumpTarget:
		root.setJumpTarget(ev.value)
	case *eventMultiColor:
		root.setMultiColor(ev.value)
	case *eventSaveBuffer:
		root.saveBuffer(ev.value)
	case *eventPipeBuffer:
		root.pipeBuffer(ev.value)
	case *eventInputSearch:
		root.firstSearch(ctx, ev.searchType)
	case *eventSkipLines:
		root.setSkipLines(ev.value)
	case *eventTabWidth:
		root.setTabWidth(ev.value)
	case *eventSectionDelimiter:
		root.setSectionDelimiter(ev.value)
	case *eventSectionStart:
		root.setSectionStart(ev.value)
	case *eventSectionNum:
		root.setSectionNum(ev.value)
	case *eventVerticalHeader:
		root.setVerticalHeader(ev.value)
	case *eventViewMode:
		root.setViewMode(ctx, ev.value)
	case *eventWatchInterval:
		root.setWatchInterval(ev.value)
	case *eventWriteBA:
		root.setWriteBA(ctx, ev.value)

	// tcell events.
	case *tcell.EventResize:
		root.resize(ctx)
	case *tcell.EventMouse:
		root.mouseEvent(ctx, ev)
	case *tcell.EventKey:
		root.keyEvent(ctx, ev)
	case nil:
		return true
	}
	return false
}

// quitCheck checks if it should quit when the document is small.
func (root *Root) quitCheck() bool {
	if !root.Config.QuitSmall {
		return false
	}
	if !root.docSmall() {
		return false
	}
	if root.DocumentLen() == 1 && !root.Config.QuitSmallFilter {
		root.Config.IsWriteOnExit = true
		return true
	}
	return false
}

// notifyEOFReached notifies that EOF has been reached.
func (root *Root) notifyEOFReached(m *Document) {
	if root.Config.NotifyEOF == 0 {
		return
	}
	msg := "EOF reached: " + m.FileName
	root.execNotify(msg, root.Config.NotifyEOF)
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
func (root *Root) everyUpdate(ctx context.Context) {
	// If tmpLN is set, set top position to position from bottom.
	// This process is executed when temporary read is switched to normal read.
	if n := atomic.SwapInt32(&root.Doc.tmpLN, 0); n > 0 {
		tmpN := int(n) - root.Doc.topLN
		if tmpN > 0 {
			root.Doc.topLN = root.Doc.BufEndNum() - tmpN
		}
	}

	switch {
	case root.FollowAll:
		root.followAll(ctx)
	case root.Doc.FollowMode:
		root.follow(ctx)
	case root.Doc.FollowSection:
		root.followSection(ctx)
	}

	if !root.skipDraw && root.Doc.height > 0 {
		root.draw(ctx)
	}
	root.skipDraw = false

	if atomic.SwapInt32(&root.Doc.watchRestart, 0) == 1 {
		root.watchControl()
	}

	if root.quitSmallCountDown > 0 {
		root.quitSmallCountDown--
		if root.DocumentLen() == 2 && root.Doc.documentType == DocFilter && root.docSmall() {
			root.draw(ctx)
			root.WriteQuit(ctx)
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

// eventAppQuit represents a quit event.
type eventAppQuit struct {
	tcell.EventTime
}

// Quit fires the eventAppQuit event.
func (root *Root) Quit(context.Context) {
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
	if !root.Doc.Normal.ProcessOfCount && !root.Doc.BufEOF() {
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
func (root *Root) SetDocument(docNum int) error {
	if docNum < 0 || docNum >= root.DocumentLen() {
		return ErrInvalidDocumentNum
	}
	root.sendDocument(docNum)
	return nil
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
func (root *Root) CloseDocument(_ *Document) {
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

// CloseAllFilter fires the eventCloseAllFilter event.
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
func (root *Root) Reload(context.Context) {
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
	for {
		select {
		case <-root.Screen.EventQ():
		default:
			return
		}
	}
}

// postEvent is a wrapper for tcell.Event.
func (root *Root) postEvent(ev tcell.Event) {
	if root == nil {
		return
	}
	if root.Screen == nil {
		return
	}
	select {
	case root.Screen.EventQ() <- ev:
		return
	default:
		log.Printf("postEvent: %v\n", tcell.ErrEventQFull)
		root.releaseEventBuffer()
	}
}

// eventReachEOF represents an EOF event.
type eventReachEOF struct {
	m *Document
	tcell.EventTime
}

// sendReachEOF fires the eventReachEOF event.
func (root *Root) sendReachEOF(m *Document) {
	ev := &eventReachEOF{}
	ev.m = m
	ev.SetEventNow()
	root.postEvent(ev)
}

// monitorEOF monitors the EOF of the document.
func (root *Root) monitorEOF() {
	for _, m := range root.DocList {
		go root.waitForEOF(m)
	}
}

// waitForEOF waits for a document to reach EOF and sends notification.
func (root *Root) waitForEOF(m *Document) {
	if !m.BufEOF() {
		m.cond.L.Lock()
		m.cond.Wait()
		m.cond.L.Unlock()
	}
	root.sendReachEOF(m)
}
