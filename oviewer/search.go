package oviewer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"

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
	Match([]byte) bool
	MatchString(string) bool
}

// searchWord is a case-insensitive search.
type searchWord struct {
	word string
}

// searchWord Match is a case-insensitive search.
func (substr searchWord) Match(s []byte) bool {
	s = stripEscapeSequenceBytes(s)
	return bytes.Contains(bytes.ToLower(s), []byte(substr.word))
}

// searchWord Match is a case-insensitive search.
func (substr searchWord) MatchString(s string) bool {
	s = stripEscapeSequenceString(s)
	return strings.Contains(strings.ToLower(s), substr.word)
}

// sensitiveWord is a case-sensitive search.
type sensitiveWord struct {
	word string
}

// sensitiveWord Match is a case-sensitive search.
func (substr sensitiveWord) Match(s []byte) bool {
	s = stripEscapeSequenceBytes(s)
	return bytes.Contains(s, []byte(substr.word))
}

// sensitiveWord Match is a case-sensitive search.
func (substr sensitiveWord) MatchString(s string) bool {
	s = stripEscapeSequenceString(s)
	return strings.Contains(s, substr.word)
}

// regexpWord is a regular expression search.
type regexpWord struct {
	word *regexp.Regexp
}

// regexpWord Match is a regular expression search.
func (substr regexpWord) Match(s []byte) bool {
	s = stripEscapeSequenceBytes(s)
	return substr.word.Match(s)
}

// regexpWord Match is a regular expression search.
func (substr regexpWord) MatchString(s string) bool {
	s = stripEscapeSequenceString(s)
	return substr.word.MatchString(s)
}

// stripRegexpES is a regular expression that excludes escape sequences.
var stripRegexpES = regexp.MustCompile("(\x1b\\[[\\d;*]*m)|.\b")

// stripEscapeSequenceString strips if it contains escape sequences.
func stripEscapeSequenceString(s string) string {
	// Remove EscapeSequence.
	if strings.ContainsAny(s, "\x1b\b") {
		s = stripRegexpES.ReplaceAllString(s, "")
	}
	return s
}

// stripEscapeSequence strips if it contains escape sequences.
func stripEscapeSequenceBytes(s []byte) []byte {
	// Remove EscapeSequence.
	if bytes.ContainsAny(s, "\x1b\b") {
		s = stripRegexpES.ReplaceAll(s, []byte(""))
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
	// Use sensitiveWord when not needed.
	if strings.ToLower(word) == strings.ToUpper(word) {
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

// multiRegexpCompile compiles multi-word regular expressions.
func multiRegexpCompile(words []string) []*regexp.Regexp {
	regexps := make([]*regexp.Regexp, len(words))
	for n, w := range words {
		s, err := strconv.Unquote(w)
		if err != nil {
			s = w
		}
		regexps[n] = regexpCompile(s, true)
	}
	return regexps
}

// condRegexpCompile conditionally compiles a regular expression.
func condRegexpCompile(in string) *regexp.Regexp {
	if len(in) < 2 {
		return nil
	}

	if in[0] != '/' || in[len(in)-1] != '/' {
		return nil
	}

	re, err := regexp.Compile(in[1 : len(in)-1])
	if err != nil {
		return nil
	}
	return re
}

// searchPosition returns the position where the search in the argument line matched.
// searchPosition uses cache.
func (root *Root) searchPosition(lN int, str string) [][]int {
	var indexes [][]int
	if root.Config.RegexpSearch {
		indexes = searchPositionReg(str, root.searchReg)
	} else {
		indexes = searchPositionStr(root.Config.CaseSensitive, str, root.searchWord)
	}
	return indexes
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
	return allStringIndex(s, substr)
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
		root.searchGo(lN)
		return nil
	})

	if err := eg.Wait(); err != nil {
		root.setMessage(err.Error())
		return
	}
	root.setMessagef("search:%v", root.searchWord)
}

func (m *Document) Search(ctx context.Context, searcher Searcher, chunkNum int, line int) (int, error) {
	chunk := m.chunks[chunkNum]
	if len(chunk.lines) == 0 {
		if !m.storageSearch(ctx, searcher, chunkNum, line) {
			return 0, ErrNotFound
		}
	}
	for n := line; n < ChunkSize; n++ {
		buf, err := m.GetChunkLine(chunkNum, n)
		if err != nil {
			return n, err
		}
		if searcher.Match(buf) {
			return n, nil
		}
		select {
		case <-ctx.Done():
			return 0, ErrCancel
		default:
		}
	}
	return 0, ErrNotFound
}
func (m *Document) BackSearch(ctx context.Context, searcher Searcher, chunkNum int, line int) (int, error) {
	chunk := m.chunks[chunkNum]
	if len(chunk.lines) == 0 {
		if !m.storageSearch(ctx, searcher, chunkNum, line) {
			return 0, ErrNotFound
		}
	}
	for n := line; n >= 0; n-- {
		buf, err := m.GetChunkLine(chunkNum, n)
		if err != nil {
			return n, err
		}
		if searcher.Match(buf) {
			return n, nil
		}
		select {
		case <-ctx.Done():
			return 0, ErrCancel
		default:
		}
	}
	return 0, ErrNotFound
}

func (m *Document) storageSearch(ctx context.Context, searcher Searcher, chunkNum int, line int) bool {
	chunk := m.chunks[chunkNum]
	if len(chunk.lines) == 0 && atomic.LoadInt32(&m.closed) == 0 {
		if m.searchControl(chunkNum, searcher) {
			return true
		}
	}
	return false
}

// SearchLine searches the document and returns the matching line number.
func (m *Document) SearchLine(ctx context.Context, searcher Searcher, lN int) (int, error) {
	lN = max(lN, 0)
	end := m.BufEndNum()

	startChunk, sn := chunkLine(lN)
	endChunk, _ := chunkLine(end)

	for cn := startChunk; cn <= endChunk; cn++ {
		n, err := m.Search(ctx, searcher, cn, sn)
		if err == nil {
			return cn*ChunkSize + n, nil
		}
		select {
		case <-ctx.Done():
			return 0, ErrCancel
		default:
		}
		sn = 0
	}

	return 0, ErrNotFound
}

// BackSearchLine does a backward search on the document and returns a matching line number.
func (m *Document) BackSearchLine(ctx context.Context, searcher Searcher, lN int) (int, error) {
	lN = min(lN, m.BufEndNum()-1)
	startChunk, sn := chunkLine(lN)
	for cn := startChunk; cn >= 0; cn-- {
		n, err := m.BackSearch(ctx, searcher, cn, sn)
		if err == nil {
			return cn*ChunkSize + n, nil
		}
		select {
		case <-ctx.Done():
			return 0, ErrCancel
		default:
		}
		sn = ChunkSize - 1
	}
	return 0, ErrNotFound
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
	// Allow only some events while searching.
	for {
		ev := root.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey: // cancel key?
			c.Capture(ev)
		case *eventSearchQuit: // found
			return nil
		case *eventUpdateEndNum:
			root.updateEndNum()
		default:
			log.Printf("unexpected event %#v", ev)
			return nil
		}
	}
}

// eventSearchQuit represents a search quit event.
type eventSearchQuit struct {
	tcell.EventTime
}

// searchQuit fires the eventSearchQuit event.
func (root *Root) searchQuit() {
	ev := &eventSearchQuit{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// eventSearchMove represents the moveo input mode.
type eventSearchMove struct {
	tcell.EventTime
	value int
}

func (root *Root) searchGo(lN int) {
	ev := &eventSearchMove{}
	ev.SetEventNow()
	ev.value = lN
	root.postEvent(ev)
}

// incrementalSearch performs incremental search by setting and input mode.
func (root *Root) incrementalSearch(ctx context.Context) {
	if !root.Config.Incsearch {
		return
	}

	switch root.input.Event.Mode() {
	case Search:
		root.incSearch(ctx, true, root.startSearchLN())
	case Backsearch:
		root.incSearch(ctx, false, root.startSearchLN())
	}
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
				log.Printf("incSearch %s", err)
			}
			return
		}
		root.searchGo(lN)
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

// startSearchLN returns the start position of the search.
func (root *Root) startSearchLN() int {
	l := root.scr.lineNumber(root.headerLen + root.Doc.JumpTarget)
	if l.number-root.Doc.topLN > root.Doc.topLN {
		return 0
	}
	return l.number
}

// firstSearch performs the first search immediately after the input.
func (root *Root) firstSearch(ctx context.Context) {
	searcher := root.setSearcher(root.input.value, root.Config.CaseSensitive)
	root.searchMove(ctx, true, root.startSearchLN(), searcher)
}

// nextSearch performs the next search.
func (root *Root) nextSearch(ctx context.Context, str string) {
	searcher := root.setSearcher(str, root.Config.CaseSensitive)
	l := root.scr.lineNumber(root.headerLen + root.Doc.JumpTarget)
	root.searchMove(ctx, true, l.number+1, searcher)
}

// firstBackSearch performs the first back search immediately after the input.
func (root *Root) firstBackSearch(ctx context.Context) {
	searcher := root.setSearcher(root.input.value, root.Config.CaseSensitive)
	l := root.scr.lineNumber(root.headerLen)
	root.searchMove(ctx, false, l.number, searcher)
}

// nextBackSearch performs the next back search.
func (root *Root) nextBackSearch(ctx context.Context, str string) {
	searcher := root.setSearcher(str, root.Config.CaseSensitive)
	l := root.scr.lineNumber(root.headerLen + root.Doc.JumpTarget)
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
