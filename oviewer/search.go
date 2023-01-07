package oviewer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"code.rocketnine.space/tslocum/cbind"
	"github.com/gdamore/tcell/v2"
	"golang.org/x/sync/errgroup"
)

//  1. serch key event(/)		back search key event(?)
//  2. actionSearch				setBackSearchMode			(key event)
//  3. root.setSearchMode()		root.setBackSearchMode()
//  4. input...confirm			input...confirm
//  5. *eventInputSearch		*eventInputBackSearch		(event)
//  6. root.firstSearch()		root.firstBackSearch()
//
//  7. next search key event(n)	next backsearch key event(N)
//  8. actionNextSearch			actionNextBackSearch		(key event)
//  9. root.setNextSearch()		root.setNextBackSearch()
// 10. *eventNextSearch			*eventNextBackSearch		(event)
// 11. root.nextSearch()		root.nextBackSearch()

// Searcher interface provides a match method that determines
// if the search word matches the argument string.
type Searcher interface {
	Match(string) bool
}

// searchWord is a case-insensitive search.
type searchWord struct {
	word string
}

// searchWord Match is a case-insensitive search.
func (substr searchWord) Match(s string) bool {
	s = stripEscapeSequence(s)
	return strings.Contains(strings.ToLower(s), substr.word)
}

// sensitiveWord is a case-sensitive search.
type sensitiveWord struct {
	word string
}

// sensitiveWord Match is a case-sensitive search.
func (substr sensitiveWord) Match(s string) bool {
	s = stripEscapeSequence(s)
	return strings.Contains(s, substr.word)
}

// regexpWord is a regular expression search.
type regexpWord struct {
	word *regexp.Regexp
}

// regexpWord Match is a regular expression search.
func (substr regexpWord) Match(s string) bool {
	s = stripEscapeSequence(s)
	return substr.word.MatchString(s)
}

// stripRegexpES is a regular expression that excludes escape sequences.
var stripRegexpES = regexp.MustCompile("(\x1b\\[[\\d;*]*m)|.\b")

// stripEscapeSequence strips if it contains escape sequences.
func stripEscapeSequence(s string) string {
	// Remove EscapeSequence.
	if strings.ContainsAny(s, "\x1b\b") {
		s = stripRegexpES.ReplaceAllString(s, "")
	}
	return s
}

// NewSearcher returns the Searcher interface suitable for the search term.
func NewSearcher(word string, searchReg *regexp.Regexp, caseSensitive bool, regexpSearch bool) Searcher {
	if regexpSearch && word != regexp.QuoteMeta(word) {
		return regexpWord{
			word: searchReg,
		}
	}
	if caseSensitive {
		return sensitiveWord{
			word: word,
		}
	}
	return searchWord{
		word: strings.ToLower(word),
	}
}

// regexpCompile is regexp.Compile the search string.
func regexpCompile(r string, caseSensitive bool) *regexp.Regexp {
	if !caseSensitive {
		r = "(?i)" + r
	}
	re, err := regexp.Compile(r)
	if err == nil {
		return re
	}

	r = regexp.QuoteMeta(r)
	re, err = regexp.Compile(r)
	if err != nil {
		log.Printf("regexpCompile failed %s", r)
		return nil
	}
	return re
}

// searchPosition returns the position where the search in the argument line matched.
// searchPosition uses cache.
func (root *Root) searchPosition(lN int, lineStr string) [][]int {
	m := root.Doc
	key := root.searchKey(lN)

	if value, found := m.cache.Get(key); found {
		poss, ok := value.([][]int)
		if !ok {
			return nil
		}
		return poss
	}

	var poss [][]int
	if root.Config.RegexpSearch {
		poss = searchPositionReg(lineStr, root.searchReg)
	} else {
		poss = searchPositionStr(root.Config.CaseSensitive, lineStr, root.searchWord)
	}

	m.cache.Set(key, poss, 3)
	return poss
}

// searcKey returns a search key for the cache.
func (root *Root) searchKey(lN int) string {
	s := "s"
	if root.Config.RegexpSearch {
		s += "r"
	}
	if root.Config.CaseSensitive {
		s += "i"
	}
	return fmt.Sprintf("search:%d:%s:%s", lN, s, root.searchWord)
}

// searchPositionReg returns an array of the beginning and end of the search string.
func searchPositionReg(s string, re *regexp.Regexp) [][]int {
	if re == nil || re.String() == "" {
		return nil
	}
	return re.FindAllIndex([]byte(s), -1)
}

// searchPosition returns an array of the beginning and end of the search string.
func searchPositionStr(caseSensitive bool, s string, substr string) [][]int {
	if substr == "" {
		return nil
	}

	if !caseSensitive {
		s = strings.ToLower(s)
		substr = strings.ToLower(substr)
	}
	return allIndex(s, substr)
}

// setSearcher is a wrapper for NewSearcher and returns a Searcher interface.
// Returns nil if there is no search term.
func (root *Root) setSearcher(word string, caseSensitive bool) Searcher {
	if word == "" {
		root.searchWord = ""
		root.searchReg = nil
		return nil
	}
	root.input.value = word
	root.searchWord = word
	root.searchReg = regexpCompile(root.searchWord, caseSensitive)

	return NewSearcher(root.searchWord, root.searchReg, caseSensitive, root.Config.RegexpSearch)
}

// eventSearchQuit represents a search quit event.
type eventSearchQuit struct {
	tcell.EventTime
}

// searchMove searches forward/backward and moves to the nearest matching line.
func (root *Root) searchMove(ctx context.Context, forward bool, lN int, searcher Searcher) {
	if searcher == nil {
		return
	}
	root.setMessagef("search:%v (%v)Cancel", root.searchWord, strings.Join(root.cancelKeys, ","))
	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	root.cancelFunc = cancel

	eg.Go(func() error {
		return root.cancelWait()
	})

	eg.Go(func() error {
		var err error
		if forward {
			lN, err = root.Doc.SearchLine(ctx, searcher, lN)
		} else {
			lN, err = root.Doc.BackSearchLine(ctx, searcher, lN)
		}
		root.searchQuit()
		if err != nil {
			return err
		}
		root.Doc.topLN = lN - root.Doc.firstLine()
		root.Doc.topLX = 0
		root.moveNumUp(root.Doc.JumpTarget)
		return nil
	})

	if err := eg.Wait(); err != nil {
		root.setMessage(err.Error())
		return
	}
	root.setMessagef("search:%v", root.searchWord)
}

// cancelWait waits for key to cancel.
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

// searchQuit fires the eventSearchQuit event.
func (root *Root) searchQuit() {
	if !root.checkScreen() {
		return
	}
	ev := &eventSearchQuit{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// incSearch implements incremental forward/back search.
func (root *Root) incSearch(ctx context.Context, forward bool, lN int) {
	root.Doc.topLN = root.returnStartPosition()

	searcher := root.setSearcher(root.input.value, root.Config.CaseSensitive)
	if searcher == nil {
		return
	}

	ctx = root.cancelRestart(ctx)
	go func() {
		var err error
		if forward {
			lN, err = root.Doc.SearchLine(ctx, searcher, lN)
		} else {
			lN, err = root.Doc.BackSearchLine(ctx, searcher, lN)
		}
		root.searchQuit()
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Println(err)
			}
			return
		}
		root.Doc.topLN = lN - root.Doc.firstLine()
		root.Doc.topLX = 0
		root.moveNumUp(root.Doc.JumpTarget)
	}()
}

// cancelRestart calls the cancel function and sets the cancel function again.
func (root *Root) cancelRestart(ctx context.Context) context.Context {
	if root.cancelFunc != nil {
		root.cancelFunc()
	}
	ctx, cancel := context.WithCancel(ctx)
	root.cancelFunc = cancel
	return ctx
}

// returnStartPosition checks the input value and returns the start position.
func (root *Root) returnStartPosition() int {
	start := root.Doc.topLN
	if !strings.Contains(root.searchWord, root.OriginStr) {
		start = root.OriginPos
	}
	root.OriginStr = root.searchWord
	return start
}

// firstSearch performs the first search immediately after the input.
func (root *Root) firstSearch(ctx context.Context) {
	searcher := root.setSearcher(root.input.value, root.Config.CaseSensitive)
	l := root.lineInfo(root.headerLen + root.Doc.JumpTarget)
	if l.number-root.Doc.topLN > root.Doc.topLN {
		l.number = 0
	}
	root.searchMove(ctx, true, l.number, searcher)
}

// nextSearch performs the next search.
func (root *Root) nextSearch(ctx context.Context, str string) {
	searcher := root.setSearcher(str, root.Config.CaseSensitive)
	l := root.lineInfo(root.headerLen + root.Doc.JumpTarget)
	root.searchMove(ctx, true, l.number+1, searcher)
}

// firstBackSearch performs the first back search immediately after the input.
func (root *Root) firstBackSearch(ctx context.Context) {
	searcher := root.setSearcher(root.input.value, root.Config.CaseSensitive)
	l := root.lineInfo(root.headerLen)
	root.searchMove(ctx, false, l.number, searcher)
}

// nextBackSearch performs the next back search.
func (root *Root) nextBackSearch(ctx context.Context, str string) {
	searcher := root.setSearcher(str, root.Config.CaseSensitive)
	l := root.lineInfo(root.headerLen + root.Doc.JumpTarget)
	root.searchMove(ctx, false, l.number-1, searcher)
}

// eventNextSearch represents search event.
type eventNextSearch struct {
	tcell.EventTime
	str string
}

// setNextSearch fires the exntNextSearch event.
func (root *Root) setNextSearch() {
	ev := &eventNextSearch{}
	ev.str = root.searchWord
	ev.SetEventNow()
	root.postEvent(ev)
}

// eventNextBackSearch represents backward search event.
type eventNextBackSearch struct {
	tcell.EventTime
	str string
}

// setNextBackSearch fires the exntNextBackSearch event.
func (root *Root) setNextBackSearch() {
	ev := &eventNextBackSearch{}
	ev.str = root.searchWord
	ev.SetEventNow()
	root.postEvent(ev)
}

// Search fires a forward search event.
// This is for calling Search from the outside.
// Normally, the event is executed from Confirm.
func (root *Root) Search(str string) {
	if !root.checkScreen() {
		return
	}
	ev := &eventNextSearch{}
	ev.str = str
	ev.SetEventNow()
	root.postEvent(ev)
}

// BackSearch fires a backward search event.
// This is for calling Search from the outside.
// Normally, the event is executed from Confirm.
func (root *Root) BackSearch(str string) {
	if !root.checkScreen() {
		return
	}
	ev := &eventNextBackSearch{}
	ev.str = str
	ev.SetEventNow()
	root.postEvent(ev)
}
