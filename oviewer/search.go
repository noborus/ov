package oviewer

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode"

	"code.rocketnine.space/tslocum/cbind"
	"github.com/gdamore/tcell/v2"
	"golang.org/x/sync/errgroup"
)

// Searcher interface provides a match method that determines
// if the search word matches the argument string.
type Searcher interface {
	// Match searches for bytes.
	Match(target []byte) bool
	// MatchString searches for strings.
	MatchString(target string) bool
	// FindAll searches for strings and returns the index of the match.
	FindAll(target string) [][]int
	// String returns the search word.
	String() string
}

// searchWord is a case-insensitive search.
type searchWord struct {
	word string
}

// searchWord Match is a case-insensitive search for bytes.
func (substr searchWord) Match(target []byte) bool {
	target = stripEscapeSequenceBytes(target)
	return bytes.Contains(bytes.ToLower(target), []byte(substr.word))
}

// searchWord MatchString is a case-insensitive search for string.
func (substr searchWord) MatchString(target string) bool {
	target = stripEscapeSequenceString(target)
	return strings.Contains(strings.ToLower(target), substr.word)
}

// searchWord FindAll searches for strings and returns the index of the match.
func (substr searchWord) FindAll(target string) [][]int {
	target = strings.ToLower(target)
	return allStringIndex(target, substr.word)
}

// searchWord String returns the search word.
func (substr searchWord) String() string {
	return substr.word
}

// sensitiveWord is a case-sensitive search.
type sensitiveWord struct {
	word string
}

// sensitiveWord Match is a case-sensitive search for bytes.
func (substr sensitiveWord) Match(target []byte) bool {
	target = stripEscapeSequenceBytes(target)
	return bytes.Contains(target, []byte(substr.word))
}

// sensitiveWord MatchString is a case-sensitive search for string.
func (substr sensitiveWord) MatchString(target string) bool {
	target = stripEscapeSequenceString(target)
	return strings.Contains(target, substr.word)
}

// sensitiveWord FindAll searches for strings and returns the index of the match.
func (substr sensitiveWord) FindAll(target string) [][]int {
	return allStringIndex(target, substr.word)
}

// String returns the search word.
func (substr sensitiveWord) String() string {
	return substr.word
}

// regexpWord is a regular expression search.
type regexpWord struct {
	word   string
	regexp *regexp.Regexp
}

// regexpWord Match is a regular expression search for bytes.
func (substr regexpWord) Match(target []byte) bool {
	target = stripEscapeSequenceBytes(target)
	return substr.regexp.Match(target)
}

// regexpWord MatchString is a regular expression search for string.
func (substr regexpWord) MatchString(target string) bool {
	target = stripEscapeSequenceString(target)
	return substr.regexp.MatchString(target)
}

// regexpWord FindAll searches for strings and returns the index of the match.
func (substr regexpWord) FindAll(target string) [][]int {
	return substr.regexp.FindAllStringIndex(target, -1)
}

// regexpWord String returns the search word.
func (substr regexpWord) String() string {
	return substr.word
}

// stripRegexpES is a regular expression that excludes escape sequences.
var stripRegexpES = regexp.MustCompile("(\x1b\\[[\\d;*]*m)|.\b")

// stripEscapeSequenceString strips if it contains escape sequences.
func stripEscapeSequenceString(src string) string {
	if !strings.ContainsAny(src, "\x1b\b") {
		return src
	}
	// Remove EscapeSequence.
	return stripRegexpES.ReplaceAllString(src, "")
}

// stripEscapeSequence strips if it contains escape sequences.
func stripEscapeSequenceBytes(src []byte) []byte {
	if !bytes.ContainsAny(src, "\x1b\b") {
		return src
	}
	// Remove EscapeSequence.
	return stripRegexpES.ReplaceAll(src, []byte(""))
}

// NewSearcher returns the Searcher interface suitable for the search term.
func NewSearcher(word string, searchReg *regexp.Regexp, caseSensitive bool, regexpSearch bool) Searcher {
	if regexpSearch && word != regexp.QuoteMeta(word) {
		return regexpWord{
			word:   word,
			regexp: searchReg,
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
	opt := ""
	if !caseSensitive {
		opt = "(?i)"
	}
	re, err := regexp.Compile(opt + r)
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
func (root *Root) searchPosition(str string) [][]int {
	return root.searcher.FindAll(str)
}

// searchXPos returns the x position of the first match.
func (root *Root) searchXPos(lineNum int) int {
	line := root.Doc.getLineC(lineNum, root.Doc.TabWidth)
	indexes := root.searcher.FindAll(line.str)
	if len(indexes) == 0 {
		return 0
	}
	return line.pos.x(indexes[0][0])
}

// searchPositionReg returns an array of the beginning and end of the string
// that matched the regular expression search.
func searchPositionReg(s string, re *regexp.Regexp) [][]int {
	if re == nil || re.String() == "" {
		return nil
	}
	return re.FindAllStringIndex(s, -1)
}

// setSearcher is a wrapper for NewSearcher and returns a Searcher interface.
// Returns nil if there is no search term.
func (root *Root) setSearcher(word string, caseSensitive bool) Searcher {
	if word == "" {
		root.searcher = nil
		return nil
	}
	root.input.value = word

	if root.Config.SmartCaseSensitive {
		for _, ch := range word {
			if unicode.IsUpper(ch) {
				caseSensitive = true
				break
			}
		}
	}
	reg := regexpCompile(word, caseSensitive)
	searcher := NewSearcher(word, reg, caseSensitive, root.Config.RegexpSearch)
	root.searcher = searcher
	return searcher
}

// searchMove searches forward/backward and moves to the nearest matching line.
func (root *Root) searchMove(ctx context.Context, forward bool, lineNum int, searcher Searcher) bool {
	if searcher == nil {
		return false
	}
	word := searcher.String()
	root.setMessagef("search:%v (%v)Cancel", word, strings.Join(root.cancelKeys, ","))
	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg.Go(func() error {
		return root.cancelWait(cancel)
	})

	eg.Go(func() error {
		n, err := root.Doc.searchLine(ctx, searcher, forward, lineNum)
		root.sendSearchQuit()
		if err != nil {
			return fmt.Errorf("search:%w:%v", err, word)
		}
		root.sendSearchMove(n)
		return nil
	})

	if err := eg.Wait(); err != nil {
		root.setMessageLog(err.Error())
		return false
	}
	root.setMessagef("search:%v", word)
	return true
}

// searchLine is a forward/backward search wrap function.
func (m *Document) searchLine(ctx context.Context, searcher Searcher, forward bool, lineNum int) (int, error) {
	if forward {
		return m.SearchLine(ctx, searcher, lineNum)
	}
	return m.BackSearchLine(ctx, searcher, lineNum)
}

// Search searches for the search term and moves to the nearest matching line.
func (m *Document) Search(ctx context.Context, searcher Searcher, chunkNum int, lineNum int) (int, error) {
	if !m.seekable {
		if chunkNum != 0 && m.store.lastChunkNum() <= chunkNum {
			m.requestLoad(chunkNum)
		}
	} else {
		if m.store.lastChunkNum() < chunkNum {
			return 0, ErrOutOfChunk
		}
		if !m.store.isLoadedChunk(chunkNum, m.seekable) && !m.storageSearch(searcher, chunkNum) {
			return 0, ErrNotFound
		}
	}

	if m.nonMatch {
		return m.SearchChunkNonMatch(ctx, searcher, chunkNum, lineNum)
	}
	return m.SearchChunk(ctx, searcher, chunkNum, lineNum)
}

// SearchChunk searches forward from the specified line.
func (m *Document) SearchChunk(ctx context.Context, searcher Searcher, chunkNum int, lineNum int) (int, error) {
	for n := lineNum; n < ChunkSize; n++ {
		buf, err := m.store.GetChunkLine(chunkNum, n)
		if err != nil {
			return n, fmt.Errorf("%w: %d:%d", err, chunkNum, n)
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

// SearchChunkNonMatch returns unmatched line number.
func (m *Document) SearchChunkNonMatch(ctx context.Context, searcher Searcher, chunkNum int, lineNum int) (int, error) {
	for n := lineNum; n < ChunkSize; n++ {
		buf, err := m.store.GetChunkLine(chunkNum, n)
		if err != nil {
			return n, fmt.Errorf("%w: %d:%d", err, chunkNum, n)
		}
		if !searcher.Match(buf) {
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

// BackSearch searches backward from the specified line.
func (m *Document) BackSearch(ctx context.Context, searcher Searcher, chunkNum int, line int) (int, error) {
	if !m.store.isLoadedChunk(chunkNum, m.seekable) && !m.storageSearch(searcher, chunkNum) {
		return 0, ErrNotFound
	}
	if m.nonMatch {
		return m.BackSearchChunkNonMatch(ctx, searcher, chunkNum, line)
	}
	return m.BackSearchChunk(ctx, searcher, chunkNum, line)
}

// BackSearchChunk searches backward from the specified line.
func (m *Document) BackSearchChunk(ctx context.Context, searcher Searcher, chunkNum int, line int) (int, error) {
	for n := line; n >= 0; n-- {
		buf, err := m.store.GetChunkLine(chunkNum, n)
		if err != nil {
			return n, fmt.Errorf("%w: %d:%d", err, chunkNum, n)
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

// BackSearchChunkNonMatch returns unmatched line number.
func (m *Document) BackSearchChunkNonMatch(ctx context.Context, searcher Searcher, chunkNum int, line int) (int, error) {
	for n := line; n >= 0; n-- {
		buf, err := m.store.GetChunkLine(chunkNum, n)
		if err != nil {
			return n, fmt.Errorf("%w: %d:%d", err, chunkNum, n)
		}
		if !searcher.Match(buf) {
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

// storageSearch searches for line not in memory(storage).
func (m *Document) storageSearch(searcher Searcher, chunkNum int) bool {
	if !m.store.isLoadedChunk(chunkNum, m.seekable) && atomic.LoadInt32(&m.closed) == 0 {
		if m.requestSearch(chunkNum, searcher) {
			return true
		}
	}
	return false
}

// SearchLine searches the document and returns the matching line number.
func (m *Document) SearchLine(ctx context.Context, searcher Searcher, lineNum int) (int, error) {
	lineNum = max(lineNum, m.BufStartNum())
	startChunk, sn := chunkLineNum(lineNum)

	for cn := startChunk; ; cn++ {
		n, err := m.Search(ctx, searcher, cn, sn)
		if err == nil {
			return cn*ChunkSize + n, nil
		}
		select {
		case <-ctx.Done():
			return 0, ErrCancel
		default:
		}

		// lastChunkNum may be updated by Search.
		if cn >= m.store.lastChunkNum() {
			lineNum = cn*ChunkSize + n
			break
		}
		sn = 0
	}

	return lineNum, ErrNotFound
}

// BackSearchLine does a backward search on the document and returns a matching line number.
func (m *Document) BackSearchLine(ctx context.Context, searcher Searcher, lineNum int) (int, error) {
	lineNum = min(lineNum, m.BufEndNum()-1)
	startChunk, sn := chunkLineNum(lineNum)
	minChunk, _ := chunkLineNum(m.BufStartNum())
	for cn := startChunk; cn >= minChunk; cn-- {
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
func (root *Root) cancelWait(cancel context.CancelFunc) error {
	cancelApp := func(_ *tcell.EventKey) *tcell.EventKey {
		cancel()
		root.setMessage("cancel")
		return nil
	}

	c := cbind.NewConfiguration()
	c, err := cancelKeys(c, root.cancelKeys, cancelApp)
	if err != nil {
		return err
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
			// ignore other events.
		}
	}
}

// cancelKeys sets the cancel key.
func cancelKeys(c *cbind.Configuration, cancelKeys []string, cancelApp func(_ *tcell.EventKey) *tcell.EventKey) (*cbind.Configuration, error) {
	for _, k := range cancelKeys {
		mod, key, ch, err := cbind.Decode(k)
		if err != nil {
			return c, fmt.Errorf("%w [%s] for cancel: %w", ErrFailedKeyBind, k, err)
		}
		if key == tcell.KeyRune {
			c.SetRune(mod, ch, cancelApp)
		} else {
			c.SetKey(mod, key, cancelApp)
		}
	}
	return c, nil
}

// eventSearchQuit represents a search quit event.
type eventSearchQuit struct {
	tcell.EventTime
}

// sendSearchQuit fires the eventSearchQuit event.
func (root *Root) sendSearchQuit() {
	ev := &eventSearchQuit{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// eventSearchMove represents the move input mode.
type eventSearchMove struct {
	tcell.EventTime
	value int
}

func (root *Root) sendSearchMove(lineNum int) {
	ev := &eventSearchMove{}
	ev.SetEventNow()
	ev.value = lineNum
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
func (root *Root) incSearch(ctx context.Context, forward bool, lineNum int) {
	root.Doc.topLN = root.returnStartPosition()

	searcher := root.setSearcher(root.input.value, root.Config.CaseSensitive)
	if searcher == nil {
		return
	}

	ctx = root.cancelRestart(ctx)
	go func() {
		n, err := root.Doc.searchLine(ctx, searcher, forward, lineNum)
		if err != nil {
			root.debugMessage(fmt.Sprintf("incSearch: %s", err))
			return
		}
		root.sendSearchMove(n)
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
	word := ""
	if root.searcher != nil {
		word = root.searcher.String()
	}
	start := root.Doc.topLN
	if !strings.Contains(word, root.OriginStr) {
		start = root.OriginPos
	}
	root.OriginStr = word
	return start
}

// startSearchLN returns the start position of the search.
func (root *Root) startSearchLN() int {
	top := root.Doc.topLN + root.Doc.firstLine()
	height := root.Doc.sectionHeaderHeight
	if root.Doc.jumpTargetHeight > 0 {
		height = root.Doc.jumpTargetHeight
	}
	current := root.scr.lineNumber(root.Doc.headerHeight + height)
	if root.Doc.topLN == 0 && root.Doc.jumpTargetHeight > 0 {
		top = top - current.number
	}

	if root.Doc.lastSearchLN < 0 {
		return top
	}
	if top > root.Doc.lastSearchLN || root.Doc.bottomLN < root.Doc.lastSearchLN {
		return top
	}
	if root.Doc.lastSearchLN < current.number-root.Doc.SectionHeaderNum {
		return top
	}
	if root.Doc.lastSearchLN > current.number+root.Doc.SectionHeaderNum {
		return top
	}
	return root.Doc.lastSearchLN
}

// firstSearch performs the first search immediately after the input.
func (root *Root) firstSearch(ctx context.Context, t searchType) {
	switch t {
	case forward:
		root.forwardSearch(ctx, root.input.value, 0)
	case backward:
		root.backSearch(ctx, root.input.value, 0)
	case filter:
		root.filter(ctx, root.input.value)
	}
}

// forwardSearch performs the forward search.
func (root *Root) forwardSearch(ctx context.Context, str string, next int) {
	searcher := root.setSearcher(str, root.Config.CaseSensitive)
	root.searchMove(ctx, true, root.startSearchLN()+next, searcher)
}

// backSearch performs the back search.
func (root *Root) backSearch(ctx context.Context, str string, next int) {
	searcher := root.setSearcher(str, root.Config.CaseSensitive)
	root.searchMove(ctx, false, root.startSearchLN()+next, searcher)
}

// eventNextSearch represents search event.
type eventNextSearch struct {
	tcell.EventTime
	str string
}

// sendNextSearch fires the eventNextSearch event.
func (root *Root) sendNextSearch(context.Context) {
	if root.searcher == nil {
		return
	}

	ev := &eventNextSearch{}
	ev.str = root.searcher.String()
	ev.SetEventNow()
	root.postEvent(ev)
}

// eventNextBackSearch represents backward search event.
type eventNextBackSearch struct {
	tcell.EventTime
	str string
}

// sendNextBackSearch fires the eventNextBackSearch event.
func (root *Root) sendNextBackSearch(context.Context) {
	if root.searcher == nil {
		return
	}

	ev := &eventNextBackSearch{}
	ev.str = root.searcher.String()
	ev.SetEventNow()
	root.postEvent(ev)
}

// Search fires a forward search event.
// This is for calling Search from the outside.
// Normally, the event is executed from Confirm.
func (root *Root) Search(str string) {
	root.sendSearch(str)
}

// eventSearch represents search event.
type eventSearch struct {
	tcell.EventTime
	str string
}

// sendSearch fires the eventSearch event.
func (root *Root) sendSearch(str string) {
	ev := &eventSearch{}
	ev.str = str
	ev.SetEventNow()
	root.postEvent(ev)
}

// BackSearch fires a backward search event.
// This is for calling Search from the outside.
// Normally, the event is executed from Confirm.
func (root *Root) BackSearch(str string) {
	root.sendBackSearch(str)
}

func (root *Root) sendBackSearch(str string) {
	ev := &eventNextBackSearch{}
	ev.str = str
	ev.SetEventNow()
	root.postEvent(ev)
}

// searchChunk searches in a Chunk without loading it into memory.
func (m *Document) searchChunk(chunkNum int, searcher Searcher) (int, error) {
	// Seek to the start of the chunk.
	chunk := m.store.chunks[chunkNum]
	if _, err := m.file.Seek(chunk.start, io.SeekStart); err != nil {
		return 0, fmt.Errorf("seek: %w", err)
	}

	// Read the chunk line by line.
	reader := bufio.NewReader(m.file)
	var line bytes.Buffer
	var isPrefix bool
	num := 0
	for num < ChunkSize {
		// Read a line.
		buf, err := reader.ReadSlice('\n')
		if errors.Is(err, bufio.ErrBufferFull) {
			isPrefix = true
			err = nil
		}
		line.Write(buf)

		// If the line is complete, check if it matches.
		if !isPrefix {
			if searcher.Match(bytes.TrimSuffix(line.Bytes(), []byte("\n"))) {
				return num, nil
			}
			num++
			line.Reset()
		}

		// If we hit the end of the file, stop.
		if err != nil {
			break
		}
	}
	return 0, ErrNotFound
}
