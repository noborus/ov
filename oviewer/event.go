package oviewer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"code.rocketnine.space/tslocum/cbind"
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
		case *eventSearch:
			root.forwardSearch(ctx, ev.str)
		case *eventBackSearch:
			root.backwardSearch(ctx, ev.str)
		case *viewModeEvent:
			root.setViewMode(ev.value)
		case *inputSearch:
			root.inputSearch(ctx)
		case *inputBackSearch:
			root.inputBackSearch(ctx)
		case *gotoEvent:
			root.goLine(ev.value)
		case *headerEvent:
			root.setHeader(ev.value)
		case *skipLinesEvent:
			root.setSkipLines(ev.value)
		case *delimiterEvent:
			root.setDelimiter(ev.value)
		case *tabWidthEvent:
			root.setTabWidth(ev.value)
		case *watchIntervalEvent:
			root.setWatchInterval(ev.value)
		case *writeBAEvent:
			root.setWriteBA(ev.value)
		case *sectionDelimiterEvent:
			root.setSectionDelimiter(ev.value)
		case *sectionStartEvent:
			root.setSectionStart(ev.value)
		case *multiColorEvent:
			root.setMultiColor(ev.value)
		case *jumpTargetEvent:
			root.setJumpTarget(ev.value)
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

func (root *Root) inputSearch(ctx context.Context) {
	searcher := root.setSearcher(root.input.value, root.CaseSensitive)
	l := root.lineInfo(root.headerLen + root.Doc.JumpTarget)
	if l.number-root.Doc.topLN > root.Doc.topLN {
		l.number = 0
	}
	root.searchMove(ctx, true, l.number, searcher)
}

func (root *Root) inputBackSearch(ctx context.Context) {
	searcher := root.setSearcher(root.input.value, root.CaseSensitive)
	l := root.lineInfo(root.headerLen)
	root.searchMove(ctx, false, l.number, searcher)
}

func (root *Root) forwardSearch(ctx context.Context, str string) {
	searcher := root.setSearcher(str, root.CaseSensitive)
	l := root.lineInfo(root.headerLen + root.Doc.JumpTarget)
	root.searchMove(ctx, true, l.number+1, searcher)
}

func (root *Root) backwardSearch(ctx context.Context, str string) {
	searcher := root.setSearcher(str, root.CaseSensitive)
	l := root.lineInfo(root.headerLen + root.Doc.JumpTarget)
	root.searchMove(ctx, false, l.number-1, searcher)
}

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

// Quit executes a quit event.
func (root *Root) Quit() {
	if !root.checkScreen() {
		return
	}
	ev := &eventAppQuit{}
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// eventAppSuspend represents a suspend event.
type eventAppSuspend struct {
	tcell.EventTime
}

// Quit executes a quit event.
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

// eventUpdateEndNum represents a timer event.
type eventUpdateEndNum struct {
	tcell.EventTime
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
			root.eventUpdate()
		case <-ctx.Done():
			return
		}
	}
}

// eventUpdate fires the event if it needs to be updated.
func (root *Root) eventUpdate() {
	if !root.checkScreen() {
		return
	}

	if !root.hasDocChanged() {
		return
	}

	ev := &eventUpdateEndNum{}
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// MoveLine fires an event that moves to the specified line.
func (root *Root) MoveLine(num int) {
	if !root.checkScreen() {
		return
	}
	ev := &gotoEvent{}
	ev.value = strconv.Itoa(num)
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// MoveTop fires the event of moving to top.
func (root *Root) MoveTop() {
	root.MoveLine(0)
}

// MoveBottom fires the event of moving to bottom.
func (root *Root) MoveBottom() {
	root.MoveLine(root.Doc.BufEndNum())
}

// eventSearch represents search event.
type eventSearch struct {
	str string
	tcell.EventTime
}

func (root *Root) eventNextSearch() {
	ev := &eventSearch{}
	ev.str = root.searchWord
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// eventBackSearch represents backward search event.
type eventBackSearch struct {
	str string
	tcell.EventTime
}

func (root *Root) eventNextBackSearch() {
	ev := &eventBackSearch{}
	ev.str = root.searchWord
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// Search fires a forward search event.
// This is for calling Search from the outside.
// Normally, the event is executed from Confirm.
func (root *Root) Search(str string) {
	if !root.checkScreen() {
		return
	}
	ev := &eventSearch{}
	ev.str = str
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// BackSearch fires a backward search event.
// This is for calling Search from the outside.
// Normally, the event is executed from Confirm.
func (root *Root) BackSearch(str string) {
	if !root.checkScreen() {
		return
	}
	ev := &eventBackSearch{}
	ev.str = str
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// eventDocument represents a set document event.
type eventDocument struct {
	docNum int
	tcell.EventTime
}

// SetDocument fires a set document event.
func (root *Root) SetDocument(docNum int) {
	if !root.checkScreen() {
		return
	}
	ev := &eventDocument{}
	if docNum >= 0 && docNum < root.DocumentLen() {
		ev.docNum = docNum
	}
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// eventAddDocument represents a set document event.
type eventAddDocument struct {
	m *Document
	tcell.EventTime
}

// AddDocument fires a add document event.
func (root *Root) AddDocument(m *Document) {
	if !root.checkScreen() {
		return
	}
	ev := &eventAddDocument{}
	ev.m = m
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// eventCloseDocument represents a close document event.
type eventCloseDocument struct {
	tcell.EventTime
}

// CloseDocument fires a del document event.
func (root *Root) CloseDocument(m *Document) {
	if !root.checkScreen() {
		return
	}
	ev := &eventCloseDocument{}
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// eventSearchQuit represents a search quit event.
type eventSearchQuit struct {
	tcell.EventTime
}

// searchQuit executes a quit event.
func (root *Root) searchQuit() {
	if !root.checkScreen() {
		return
	}
	ev := &eventSearchQuit{}
	ev.SetEventNow()
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

func (root *Root) cancelWait() error {
	cancelApp := func(ev *tcell.EventKey) *tcell.EventKey {
		if root.cancelFunc != nil {
			root.cancelFunc()
			root.setMessage("cancel")
			root.cancelFunc = nil
		}
		return nil
	}

	c := cbind.NewConfiguration()

	for _, k := range root.cancelKeys {
		mod, key, ch, err := cbind.Decode(k)
		if err != nil {
			return fmt.Errorf("%w [%s] for cancel: %s", ErrFailedKeyBind, k, err)
		}
		if key == tcell.KeyRune {
			c.SetRune(mod, ch, cancelApp)
		} else {
			c.SetKey(mod, key, cancelApp)
		}
	}

	for {
		ev := root.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			c.Capture(ev)
		case *eventSearchQuit:
			return nil
		}
	}
}

// eventReload represents a reload event.
type eventReload struct {
	m *Document
	tcell.EventTime
}

// Reload executes a reload event.
func (root *Root) Reload() {
	if !root.checkScreen() {
		return
	}
	root.setMessagef("reload %s", root.Doc.FileName)
	log.Printf("reload %s", root.Doc.FileName)
	ev := &eventReload{}
	ev.SetEventNow()
	ev.m = root.Doc
	err := root.Screen.PostEvent(ev)
	if err != nil {
		log.Println(err)
	}
}

// releaseEventBuffer will release all event buffers.
func (root *Root) releaseEventBuffer() {
	for root.HasPendingEvent() {
		_ = root.Screen.PollEvent()
	}
}
