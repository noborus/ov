package oviewer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cbind"
)

// main is manages and executes events in the main routine.
func (root *Root) main(quitChan chan<- struct{}) {
	go root.followTimer()
	ctx := context.Background()

	for {
		if root.General.FollowAll {
			root.followAll()
		}

		if root.Doc.FollowMode {
			root.follow()
		}

		if !root.skipDraw {
			root.draw()
		}
		root.skipDraw = false

		ev := root.Screen.PollEvent()
		switch ev := ev.(type) {
		case *eventAppQuit:
			if root.input.mode == Help || root.input.mode == LogDoc {
				root.toNormal()
				continue
			}
			close(quitChan)
			return
		case *eventUpdateEndNum:
			root.updateEndNum()
		case *eventDocument:
			root.CurrentDoc = ev.docNum
			m := root.DocList[root.CurrentDoc]
			root.setDocument(m)
			root.debugMessage(fmt.Sprintf("switch document %s", m.FileName))
		case *eventAddDocument:
			root.addDocument(ev.m)
		case *eventCloseDocument:
			root.closeDocument()
		case *eventCopySelect:
			root.putClipboard(ctx)
		case *eventPaste:
			root.getClipboard(ctx)
		case *eventSearch:
			root.search(ctx, root.Doc.topLN+root.Doc.Header+1, root.searchLine)
		case *eventBackSearch:
			root.search(ctx, root.Doc.topLN+root.Doc.Header-1, root.backSearchLine)
		case *searchInput:
			root.forwardSearch(ctx, ev.value)
		case *backSearchInput:
			root.backSearch(ctx, ev.value)
		case *gotoInput:
			root.goLine(ev.value)
		case *headerInput:
			root.setHeader(ev.value)
		case *delimiterInput:
			root.setDelimiter(ev.value)
		case *tabWidthInput:
			root.setTabWidth(ev.value)
		case *tcell.EventResize:
			root.resize()
		case *tcell.EventMouse:
			root.mouseEvent(ev)
		case *tcell.EventKey:
			root.setMessage("")
			switch root.input.mode {
			case Normal, Help, LogDoc:
				root.keyCapture(ev)
			default:
				root.inputEvent(ev)
			}
		case nil:
			close(quitChan)
			return
		}
	}
}

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
	go func() {
		root.Screen.PostEventWait(ev)
	}()
}

// Cancel follow mode and follow all mode.
func (root *Root) Cancel() {
	root.General.FollowAll = false
	root.Doc.FollowMode = false
}

// WriteQuit sets the write flag and executes a quit event.
func (root *Root) WriteQuit() {
	root.AfterWrite = true
	root.Quit()
}

// eventUpdateEndNum represents a timer event.
type eventUpdateEndNum struct {
	tcell.EventTime
}

// runOnTime runs at time.
func (root *Root) runOnTime(ev tcell.Event) {
	if !root.checkScreen() {
		return
	}
	go func() {
		root.Screen.PostEventWait(ev)
	}()
}

func (root *Root) followAll() {
	if root.input.mode != Normal {
		return
	}

	current := root.CurrentDoc
	for n, doc := range root.DocList {
		go root.followModeOpen(doc)
		if (atomic.LoadInt32(&doc.changed) == 1) && (doc.latestNum != doc.BufEndNum()) {
			current = n
		}
		atomic.StoreInt32(&doc.changed, 0)
	}

	if root.CurrentDoc != current {
		root.CurrentDoc = current
		root.SetDocument(root.CurrentDoc)
	}
}

func (root *Root) follow() {
	go root.followModeOpen(root.Doc)
	num := root.Doc.BufEndNum()
	if root.Doc.latestNum != num {
		root.TailSync()
		root.Doc.latestNum = num
	}
}

// followTimer fires events.
func (root *Root) followTimer() {
	timer := time.NewTicker(time.Millisecond * 100)
	for {
		<-timer.C
		eventFlag := false
		for _, doc := range root.DocList {
			if atomic.LoadInt32(&doc.changed) == 1 {
				eventFlag = true
			}
		}
		if !eventFlag {
			continue
		}

		root.debugMessage(fmt.Sprintf("eventUpdateEndNum %s", root.Doc.FileName))
		root.UpdateEndNum()
	}
}

func (root *Root) UpdateEndNum() {
	ev := &eventUpdateEndNum{}
	ev.SetEventNow()
	root.runOnTime(ev)
}

// MoveLine fires an event that moves to the specified line.
func (root *Root) MoveLine(num int) {
	if !root.checkScreen() {
		return
	}
	ev := &gotoInput{}
	ev.value = strconv.Itoa(num)
	ev.SetEventNow()
	go func() {
		err := root.Screen.PostEvent(ev)
		if err != nil {
			log.Println(err)
		}
	}()
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
	tcell.EventTime
}

func (root *Root) eventNextSearch() {
	ev := &eventSearch{}
	ev.SetEventNow()
	go func() {
		err := root.Screen.PostEvent(ev)
		if err != nil {
			log.Println(err)
		}
	}()
}

// eventBackSearch represents backward search event.
type eventBackSearch struct {
	tcell.EventTime
}

func (root *Root) eventNextBackSearch() {
	ev := &eventBackSearch{}
	ev.SetEventNow()
	go func() {
		err := root.Screen.PostEvent(ev)
		if err != nil {
			log.Println(err)
		}
	}()
}

// Search fires a forward search event.
// This is for calling Search from the outside.
// Normally, the event is executed from Confirm.
func (root *Root) Search(str string) {
	if !root.checkScreen() {
		return
	}
	if str == "" {
		root.input.reg = nil
		return
	}
	root.input.value = str
	root.input.reg = regexpComple(str, root.CaseSensitive)
	ev := &eventSearch{}
	ev.SetEventNow()
	go func() {
		root.Screen.PostEventWait(ev)
	}()
}

// BackSearch fires a backward search event.
// This is for calling Search from the outside.
// Normally, the event is executed from Confirm.
func (root *Root) BackSearch(str string) {
	if !root.checkScreen() {
		return
	}
	if str == "" {
		root.input.reg = nil
		return
	}
	root.input.value = str
	root.input.reg = regexpComple(str, root.CaseSensitive)
	ev := &eventBackSearch{}
	ev.SetEventNow()
	go func() {
		root.Screen.PostEventWait(ev)
	}()
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
	if docNum >= 0 && docNum < len(root.DocList) {
		ev.docNum = docNum
	}
	ev.SetEventNow()
	go func() {
		root.Screen.PostEventWait(ev)
	}()
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
	go func() {
		root.Screen.PostEventWait(ev)
	}()
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
	go func() {
		root.Screen.PostEventWait(ev)
	}()
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
	go func() {
		root.Screen.PostEventWait(ev)
	}()
}

func (root *Root) cancelWait(cancel context.CancelFunc) error {
	cancelApp := func(ev *tcell.EventKey) *tcell.EventKey {
		cancel()
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
