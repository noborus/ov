package oviewer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v3"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/jwalton/gchalk"
	"github.com/noborus/guesswidth"
	"github.com/noborus/ov/biomap"
)

// document type.
const (
	DocNormal = iota
	DocHelp
	DocLog
	DocFilter
)

// documentType represents the type of document (e.g., normal, help, log, filter).
type documentType int

// String returns the string representation of the document type.
func (d documentType) String() string {
	switch d {
	case DocNormal:
		return "normal"
	case DocHelp:
		return "help"
	case DocLog:
		return "log"
	case DocFilter:
		return "filter"
	}
	return "unknown"
}

// Document represents a document with various properties and methods for handling
// file operations, caching, synchronization, and display settings.
type Document struct {
	// documentType is the type of document.
	documentType documentType
	// File is the os.File.
	file *os.File

	// cache is an LRU cache for storing lines.
	cache *lru.Cache[int, LineC]

	// parent is the parent document.
	parent *Document
	// lineNumMap maps line numbers.
	lineNumMap *biomap.Map[int, int]

	// ticker is used for periodic updates.
	ticker *time.Ticker
	// tickerDone is a channel to signal ticker termination.
	tickerDone chan struct{}
	// ctlCh is the channel for controlling the reader goroutine.
	ctlCh chan controlSpecifier

	// multiColorRegexps holds multicolor regular expressions in slices.
	multiColorRegexps []*regexp.Regexp
	// store represents store management.
	store *store
	// followStore represents follow store management.
	followStore *store

	// conv is an interface that converts escape sequences, etc.
	conv Converter
	// alignConv is an interface that converts alignment.
	alignConv *align

	// cond is a condition variable for synchronization.
	cond *sync.Cond

	// FileName is the file name to display.
	FileName string
	// Caption is an additional caption to display after the file name.
	Caption string
	// filepath stores the absolute pathname for file watching.
	filepath string

	// marked is a list of marked line numbers.
	marked []int
	// columnWidths is a slice of column widths.
	columnWidths []int

	// RunTimeSettings contains RunTimeSettings settings.
	RunTimeSettings

	// memoryLimit is the maximum chunk size.
	memoryLimit int

	// currentChunk represents the current chunk number.
	currentChunk int

	// headerHeight is the height of the header.
	headerHeight int
	// sectionHeaderHeight is the height of the section header.
	sectionHeaderHeight int

	// statusPos is the position of the status line.
	statusPos int

	// width is the width of the screen.
	width int
	// height is the height of the screen.
	height int
	// markedPoint is the position of the marked line.
	markedPoint int

	// lastSectionPosNum is the last moved section position.
	lastSectionPosNum int
	// latestNum is the endNum read at the end of the screen update.
	latestNum int
	// topLN is the starting position of the current y.
	topLN int
	// topLX represents the x position of the top line.
	topLX int
	// bottomLN is the last line number displayed.
	bottomLN int
	// bottomLX is the leftmost X position on the last line.
	bottomLX int

	// scrollX is the starting position of the current scrollX.
	scrollX int
	// columnCursor is the number of columns.
	columnCursor int
	// columnStart is the starting position of the column.
	columnStart int

	// lastSearchLN is the last search line number.
	lastSearchLN int
	// showGotoF displays the specified line if it is true.
	showGotoF bool

	// jumpTargetHeight is the display position of search results.
	jumpTargetHeight int
	// jumpTargetSection is the display position of search results.
	jumpTargetSection bool

	// CFormat is a compressed format.
	CFormat Compressed

	// watchRestart indicates the number of times the watch has restarted.
	watchRestart int32
	// tickerState indicates the state of the ticker.
	tickerState int32
	// closed indicates if the document is closed (1 if closed).
	closed int32

	// tmpFollow indicates if there is a temporary follow mode (1 if true).
	tmpFollow int32
	// tmpLN is a temporary line number when the number of lines is undetermined.
	tmpLN int32

	// WatchMode indicates if watch mode is enabled.
	WatchMode bool
	// preventReload is true to prevent reload.
	preventReload bool
	// seekable indicates if seeking is possible.
	seekable bool
	// reopenable indicates if reopening is possible.
	reopenable bool
	// nonMatch indicates if non-matching lines are searched.
	nonMatch bool
	// pauseFollow indicates if follow mode is paused.
	pauseFollow bool
	// pauseLastNum is the line number where follow mode was paused.
	pauseLastNum int
	// General is the General settings.
	General General
}

// store represents store management.
type store struct {
	// loadedChunks manages chunks loaded into memory.
	loadedChunks *lru.Cache[int, struct{}]
	// chunks is the content of the file to be stored in chunks.
	chunks []*chunk
	// mu controls the mutex.
	mu sync.RWMutex

	// startNum is the number of the first line that can be moved.
	startNum int32
	// endNum is the number of the last line read.
	endNum int32

	// 1 if there is a changed.
	changed int32
	// 1 if there is a read cancel.
	readCancel int32
	// 1 if newline at end of file.
	noNewlineEOF int32
	// 1 if EOF is reached.
	eof int32
	// size is the number of bytes read.
	size int64
	// offset is the current byte offset in the file.
	offset int64
	// formfeedTime adds time on formfeed.
	formfeedTime bool
}

// chunk stores the contents of the split file as slices of strings.
type chunk struct {
	// lines stores the contents of the file in slices of strings.
	// lines,endNum and eof is updated by reader goroutine.
	lines [][]byte
	// start is the first position of the number of bytes read.
	start int64
}

// LineC is one line of information.
// Contains content, string, location information.
type LineC struct {
	// line contents.
	lc contents
	// string representation of the line.
	str string
	// for converting the width of str and lc.
	pos widthPos
	// columnRanges is the range of columnRanges.
	columnRanges []columnRange
	// valid is true if the line is valid.
	valid bool
	// The number of the section in the screen.
	section int
	// Line number within a section.
	sectionNm int
	// eolStyle is the style of the end of the line.
	eolStyle tcell.Style
}

// columnRange represents the start and end positions of a columnRange.
type columnRange struct {
	start int
	end   int
}

// NewDocument creates and initializes a new [Document] with default settings.
// It returns a pointer to the Document and an error if the cache initialization fails.
func NewDocument() (*Document, error) {
	m := &Document{
		documentType:    DocNormal,
		tickerDone:      make(chan struct{}),
		RunTimeSettings: NewRunTimeSettings(),
		ctlCh:           make(chan controlSpecifier),
		memoryLimit:     100,
		seekable:        true,
		reopenable:      true,
		store:           NewStore(),
		lastSearchLN:    -1,
	}
	if err := m.NewCache(); err != nil {
		return nil, err
	}
	m.alignConv = newAlignConverter(m.ColumnWidth)
	m.conv = m.converterType(m.Converter)

	m.cond = sync.NewCond(&sync.Mutex{})
	return m, nil
}

// NewCache initializes a new LRU cache for the Document.
// It returns an error if the cache cannot be created.
func (m *Document) NewCache() error {
	cache, err := lru.New[int, LineC](1024)
	if err != nil {
		return fmt.Errorf("new cache %w", err)
	}
	m.cache = cache

	return nil
}

// converterType returns the appropriate Converter based on the provided name.
// Parameters:
// - name: a string representing the name of the converter type.
// Returns:
// - A Converter interface corresponding to the specified name.
// converterType returns the Converter type.
func (m *Document) converterType(name string) Converter {
	switch name {
	case convRaw:
		return newRawConverter()
	case convEscaped:
		return newESConverter()
	case convAlign:
		return m.alignConv
	}
	return defaultConverter
}

// OpenDocument opens a file specified by fileName and returns a Document.
// If the fileName is "-", it reads from stdin. It returns an error if the file
// cannot be opened, is a directory, or if there are issues initializing the Document.
func OpenDocument(fileName string) (*Document, error) {
	if fileName == "-" {
		return STDINDocument()
	}
	fi, err := os.Stat(fileName)
	if err != nil {
		return nil, fmt.Errorf("'%s' %w", fileName, ErrNotFound)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("'%s' %w", fileName, ErrIsDirectory)
	}

	m, err := NewDocument()
	if err != nil {
		return nil, err
	}
	// Check if the file is a named pipe.
	if fi.Mode()&fs.ModeNamedPipe != 0 {
		m.reopenable = false
	}

	f, err := open(fileName)
	if err != nil {
		return nil, err
	}
	// Check if the file is seekable.
	if n, err := f.Seek(1, io.SeekStart); n != 1 || err != nil {
		m.seekable = false
	} else {
		_, _ = f.Seek(0, io.SeekStart)
	}

	m.FileName = fileName
	// Read the control file.
	if err := m.ControlFile(f); err != nil {
		return nil, err
	}
	return m, nil
}

// STDINDocument creates and returns a Document that reads from stdin.
// It returns a pointer to the Document and an error if the Document initialization fails.
func STDINDocument() (*Document, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}

	m.seekable = false
	m.reopenable = false
	m.FileName = "(STDIN)"
	f, err := open("")
	if err != nil {
		return nil, fmt.Errorf("failed to open stdin: %w", err)
	}
	if err := m.ControlFile(f); err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}
	return m, nil
}

// Line returns one line from buffer.
func (m *Document) Line(n int) ([]byte, error) {
	if atomic.LoadInt32(&m.tmpFollow) == 1 {
		return m.followStore.GetChunkLine(0, n)
	}

	s := m.store
	if n < 0 || n >= m.BufEndNum() {
		return nil, fmt.Errorf("%w %d>%d", ErrOutOfRange, n, m.BufEndNum())
	}

	chunkNum, cn := chunkLineNum(n)

	if s.lastChunkNum() < chunkNum {
		return nil, fmt.Errorf("%w %d<%d", ErrOutOfRange, s.lastChunkNum(), chunkNum)
	}
	if m.currentChunk != chunkNum {
		m.currentChunk = chunkNum
		m.requestLoad(chunkNum)
	}

	return s.GetChunkLine(chunkNum, cn)
}

// GetChunkLine returns a specific line from a specified chunk.
// Parameters:
// - chunkNum: the chunk number to retrieve the line from.
// - cn: the line number within the chunk.
// Returns:
// - A byte slice containing the line data.
// - An error if the chunk or line number is out of range.
func (s *store) GetChunkLine(chunkNum int, cn int) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.chunks) <= chunkNum {
		return nil, fmt.Errorf("over chunk(%d) %w", chunkNum, ErrOutOfRange)
	}
	chunk := s.chunks[chunkNum]

	if cn >= len(chunk.lines) {
		return nil, fmt.Errorf("over line (%d:%d) %w", chunkNum, cn, ErrOutOfRange)
	}
	return bytes.TrimSuffix(chunk.lines[cn], []byte("\n")), nil
}

// LineStr returns one line from buffer.
func (m *Document) LineStr(n int) (string, error) {
	s, err := m.Line(n)
	if err != nil {
		return gchalk.Red(err.Error()), err
	}
	return string(s), nil
}

// CurrentLN returns the currently displayed line number.
func (m *Document) CurrentLN() int {
	return m.topLN
}

// Export exports the document in the specified range.
func (m *Document) Export(w io.Writer, start int, end int) error {
	return m.exportRange(w, start, end, false)
}

// ExportPlain exports the document in the specified range without ANSI escape sequences.
func (m *Document) ExportPlain(w io.Writer, start int, end int) error {
	return m.exportRange(w, start, end, true)
}

// exportRange exports the document in the specified range.
// If plain is true, it writes plain text (no ANSI escapes), otherwise
// it preserves ANSI escape sequences.
func (m *Document) exportRange(w io.Writer, start int, end int, plain bool) error {
	end = min(end, m.BufEndNum()-1)
	startChunk, startCn := chunkLineNum(start)
	endChunk, endCn := chunkLineNum(end)

	scn := startCn
	ecn := ChunkSize
	for chunkNum := startChunk; chunkNum <= endChunk; chunkNum++ {
		if chunkNum == endChunk {
			ecn = endCn + 1
		}
		chunk := m.store.chunks[chunkNum]
		var err error
		if plain {
			err = m.store.exportPlain(w, chunk, scn, ecn)
		} else {
			err = m.store.export(w, chunk, scn, ecn)
		}
		if err != nil {
			return err
		}
		scn = 0
	}
	return nil
}

// BufStartNum returns the starting line number of the buffer.
func (m *Document) BufStartNum() int {
	return int(atomic.LoadInt32(&m.store.startNum))
}

// BufEndNum return last line number.
func (m *Document) BufEndNum() int {
	if atomic.LoadInt32(&m.tmpFollow) == 1 {
		return int(atomic.LoadInt32(&m.followStore.endNum))
	}
	return int(atomic.LoadInt32(&m.store.endNum))
}

// storeEndNum returns the last line number from the main store, ignoring follow mode.
func (m *Document) storeEndNum() int {
	return int(atomic.LoadInt32(&m.store.endNum))
}

// WaitEOF waits for EOF.
func (m *Document) WaitEOF() {
	m.cond.L.Lock()
	defer m.cond.L.Unlock()
	if m.BufEOF() {
		return
	}
	m.cond.Wait()
}

// WaitEOFWithTimeout waits for EOF with a timeout.
// Parameters:
// - timeout: the duration to wait before timing out.
func (m *Document) WaitEOFWithTimeout(timeout time.Duration) {
	if m.BufEOF() {
		return
	}

	doneCh := make(chan struct{})
	go func() {
		m.cond.L.Lock()
		defer m.cond.L.Unlock()
		m.cond.Wait()
		doneCh <- struct{}{}
	}()

	select {
	case <-doneCh:
	case <-time.After(timeout):
		log.Println("WaitEOFWithTimeout timed out")
	}
}

// BufEOF checks if the end of the file (EOF) has been reached in the document buffer.
// Returns:
// - A boolean value: true if EOF is reached, false otherwise.
func (m *Document) BufEOF() bool {
	return atomic.LoadInt32(&m.store.eof) == 1
}

// ClearCache clears the LRU cache of the document.
// It does not take any parameters and does not return any values.
func (m *Document) ClearCache() {
	m.cache.Purge()
}

// contents returns the contents of a specific line number in the document buffer.
// Parameters:
// - lN: the line number to retrieve the contents for.
// Returns:
// - A contents type representing the line's contents.
// - An error if the line number is out of range or if there is an issue retrieving the line.
func (m *Document) contents(lN int) (contents, error) {
	if lN < 0 || lN >= m.BufEndNum() {
		return nil, ErrOutOfRange
	}

	str, err := m.LineStr(lN)
	return parseString(m.conv, str, m.TabWidth), err
}

func (m *Document) contentsLine(lN int) (contents, tcell.Style, error) {
	if lN < 0 || lN >= m.BufEndNum() {
		return nil, tcell.StyleDefault, ErrOutOfRange
	}

	str, err := m.LineStr(lN)
	lc, style := parseLine(m.conv, str, m.TabWidth)
	return lc, style, err
}

// getLineC returns the content of the specified line number.
// If the line number does not exist, EOF content is returned.
func (m *Document) getLineC(lN int) LineC {
	if lineC, ok := m.cache.Get(lN); ok {
		lc := make(contents, len(lineC.lc))
		copy(lc, lineC.lc)
		lineC.lc = lc
		lineC.valid = true
		return lineC
	}

	org, style, err := m.contentsLine(lN)
	if err != nil && errors.Is(err, ErrOutOfRange) {
		lc := make(contents, 1)
		lc[0] = EOFContent
		return LineC{
			lc:    lc,
			str:   string(EOFC),
			pos:   widthPos{0: 0, 1: 1},
			valid: false,
		}
	}
	str, pos := ContentsToStr(org)
	lineC := LineC{
		lc:       org,
		str:      str,
		pos:      pos,
		eolStyle: style,
	}
	if err == nil {
		m.cache.Add(lN, lineC)
	}

	lc := make(contents, len(org))
	copy(lc, org)
	lineC.lc = lc
	lineC.valid = true
	return lineC
}

// firstLine is the first line that excludes the SkipLines and Header.
func (m *Document) firstLine() int {
	return m.SkipLines + m.Header
}

// watchMode sets the document to watch mode and configures section settings.
func (m *Document) watchMode() {
	m.WatchMode = true
	if m.SectionDelimiter == "" {
		m.setSectionDelimiter("^" + FormFeed)
	}
	m.SectionHeader = false
	m.SectionStartPosition = 1
	m.FollowSection = true
}

// unwatchMode unwatch mode for the document.
func (m *Document) unwatchMode() {
	m.WatchMode = false
	m.FollowSection = false
}

// regexpCompile compiles the new document's regular expressions.
func (m *Document) regexpCompile() {
	m.ColumnDelimiterReg = condRegexpCompile(m.ColumnDelimiter)
	m.setSectionDelimiter(m.SectionDelimiter)
	if len(m.MultiColorWords) > 0 {
		m.setMultiColorWords(m.MultiColorWords)
	}
}

// setDelimiter sets the delimiter string.
func (m *Document) setDelimiter(delm string) {
	m.ColumnDelimiter = delm
	m.ColumnDelimiterReg = condRegexpCompile(delm)
}

// setSectionDelimiter sets the document section delimiter.
func (m *Document) setSectionDelimiter(delm string) {
	m.SectionDelimiter = delm
	m.SectionDelimiterReg = regexpCompile(delm, true)
}

// setMultiColorWords set multiple strings to highlight with multiple colors.
func (m *Document) setMultiColorWords(words []string) {
	m.MultiColorWords = words
	m.multiColorRegexps = multiRegexpCompile(words)
}

// setColumnWidths sets the column widths.
// Guess the width of the columns using the first 1000 lines (maximum) and the headers.
func (m *Document) setColumnWidths() {
	if m.BufEndNum() == 0 {
		return
	}

	header := m.Header - 1
	header = max(header, 0)
	tl := min(1000, len(m.store.chunks[0].lines))
	lines := m.store.chunks[0].lines[m.SkipLines:tl]
	buf := make([]string, len(lines))
	for n, line := range lines {
		buf[n] = string(line)
	}
	// Stop guessing if valid row count is not reached.
	if !m.BufEOF() && len(buf) < 20 {
		return
	}
	m.columnWidths = guesswidth.Positions(buf, header, 2)
}

// vHeaderWidth returns the width of the header.
func (m *Document) vHeaderWidth(lineC LineC) int {
	if m.VerticalHeader > 0 {
		return m.VerticalHeader
	}
	if m.HeaderColumn > 0 && m.HeaderColumn < len(lineC.columnRanges) {
		return lineC.columnRanges[(m.HeaderColumn+m.columnStart)-1].end
	}
	return 0
}

// GetLine returns one line from buffer.
//
// Deprecated: Use [Document.LineStr] instead.
func (m *Document) GetLine(n int) string {
	s, err := m.Line(n)
	if err != nil {
		return ""
	}
	return string(s)
}

// LineString returns one line from the buffer as a string.
//
// Deprecated: This method is deprecated because it does not return an error and may silently ignore failures.
// Please use [Document.LineStr] instead, which returns both the string and an error for better error handling.
func (m *Document) LineString(n int) string {
	str, _ := m.LineStr(n)
	return str
}
