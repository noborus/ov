package oviewer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cbind"
)

// main is manages and executes events in the main routine.
func (root *Root) main(quitChan chan<- struct{}) {
	go root.countTimer()
	ctx := context.Background()
	if root.Doc.FollowMode {
		root.followModeFire(root.Doc)
		go root.followTimer()
	}

	for {
		if !root.skipDraw {
			root.draw()
		} else {
			root.skipDraw = false
		}
		ev := root.Screen.PollEvent()
		switch ev := ev.(type) {
		case *eventAppQuit:
			if root.input.mode == Help || root.input.mode == LogDoc {
				root.toNormal()
				continue
			}
			close(quitChan)
			return
		case *eventTimer:
			root.updateEndNum()
		case *eventFollow:
			root.TailSync()
		case *eventDocument:
			root.setDocument(ev.m)
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

// Cancel usually does nothing.
func (root *Root) Cancel() {
}

// WriteQuit sets the write flag and executes a quit event.
func (root *Root) WriteQuit() {
	root.AfterWrite = true
	root.Quit()
}

// eventTimer represents a timer event.
type eventTimer struct {
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

// countTimer fires events periodically until it reaches EOF.
func (root *Root) countTimer() {
	timer := time.NewTicker(time.Millisecond * 500)
	defer timer.Stop()
loop:
	for {
		<-timer.C
		ev := &eventTimer{}
		ev.SetEventNow()
		root.runOnTime(ev)
		if root.Doc.BufEOF() {
			break loop
		}
	}
}

// eventFollow represents a follow event.
type eventFollow struct {
	tcell.EventTime
}

// followTimer fires events.
func (root *Root) followTimer() {
	log.Println("followTimer on")
	timer := time.NewTicker(time.Millisecond * 50)
	defer timer.Stop()
	for {
		<-timer.C
		ev := &eventFollow{}
		ev.SetEventNow()
		root.runOnTime(ev)
		if !root.Doc.FollowMode {
			break
		}
	}
	log.Println("followTimer off")
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
	root.MoveLine(root.Doc.endNum)
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
	m *Document
	tcell.EventTime
}

// SetDocument fires a set document event.
func (root *Root) SetDocument(m *Document) {
	if !root.checkScreen() {
		return
	}
	ev := &eventDocument{}
	ev.m = m
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
