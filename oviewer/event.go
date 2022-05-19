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

// main is manages and executes events in the main routine.
func (root *Root) main(ctx context.Context, quitChan chan<- struct{}) {
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
			if root.screenMode != Docs {
				root.toNormal()
				continue
			}
			close(quitChan)
			return
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
			searcher := root.setSearcher(ev.str, root.CaseSensitive)
			root.searchMove(ctx, true, root.Doc.topLN+root.Doc.firstLine()+1, searcher)
		case *eventBackSearch:
			searcher := root.setSearcher(ev.str, root.CaseSensitive)
			root.searchMove(ctx, false, root.Doc.topLN+root.Doc.firstLine()-1, searcher)
		case *viewModeInput:
			root.setViewMode(ev.value)
		case *searchInput:
			searcher := root.setSearcher(root.input.value, root.CaseSensitive)
			root.searchMove(ctx, true, root.Doc.topLN+root.Doc.firstLine(), searcher)
		case *backSearchInput:
			searcher := root.setSearcher(root.input.value, root.CaseSensitive)
			root.searchMove(ctx, false, root.Doc.topLN+root.Doc.firstLine(), searcher)
		case *gotoInput:
			root.goLine(ev.value)
		case *headerInput:
			root.setHeader(ev.value)
		case *skipLinesInput:
			root.setSkipLines(ev.value)
		case *delimiterInput:
			root.setDelimiter(ev.value)
		case *tabWidthInput:
			root.setTabWidth(ev.value)
		case *watchIntervalInput:
			root.setWatchInterval(ev.value)
		case *writeBAInput:
			root.setWriteBA(ev.value)
		case *sectionDelimiterInput:
			root.setSectionDelimiter(ev.value)
		case *sectionStartInput:
			root.setSectionStart(ev.value)
		case *tcell.EventResize:
			root.resize()
		case *tcell.EventMouse:
			root.mouseEvent(ev)
		case *tcell.EventKey:
			root.setMessage("")
			switch root.input.mode {
			case Normal:
				root.keyCapture(ev)
			default:
				root.inputEvent(ctx, ev)
			}
		case nil:
			close(quitChan)
			return
		}
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
	ev := &gotoInput{}
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
	ev.str = root.input.value
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
	ev.str = root.input.value
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
