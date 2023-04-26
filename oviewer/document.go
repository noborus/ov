package oviewer

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/noborus/guesswidth"
)

// ChunkSize is the unit of number of lines to split the file.
var ChunkSize = 10000

// The Document structure contains the values
// for the logical screen.
type Document struct {
	// fileName is the file name to display.
	FileName string
	// Caption is an additional caption to display after the file name.
	Caption string
	// filepath stores the absolute pathname for file watching.
	filepath string
	// File is the os.File.
	file *os.File

	// chunks is the content of the file to be stored in chunks.
	chunks []*chunk

	// Specifies the chunk to read. -1 reads the new last line.
	ctlCh chan controlSpecifier

	cache *lru.Cache[int, LineC]

	ticker     *time.Ticker
	tickerDone chan struct{}

	// multiColorRegexps holds multicolor regular expressions in slices.
	multiColorRegexps []*regexp.Regexp

	// marked is a list of marked line numbers.
	marked []int
	// columnWidths is a slice of column widths.
	columnWidths []int
	// status is the display status of the document.
	general

	// size is the number of bytes read.
	size int64
	// offset
	offset int64

	startNum int
	// endNum is the number of the last line read.
	endNum int

	markedPoint int

	// Last moved Section position.
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

	// x is the starting position of the current x.
	x int
	// columnCursor is the number of columns.
	columnCursor int

	// CFormat is a compressed format.
	CFormat Compressed

	// mu controls the mutex.
	mu sync.Mutex

	watchRestart int32
	tickerState  int32

	// 1 if EOF is reached.
	eof int32
	// 1 if there is a changed.
	changed int32
	// 1 if there is a closed.
	closed int32

	// 1 if there is a read cancel.
	readCancel int32

	// WatchMode is watch mode.
	WatchMode bool
	// preventReload is true to prevent reload.
	preventReload bool
	// Is it possible to seek.
	seekable bool
	// formfeedTime adds time on formfeed.
	formfeedTime bool
}

// LineC is one line of information.
// Contains content, string, location information.
type LineC struct {
	lc  contents
	str string
	pos widthPos
}

// chunk stores the contents of the split file as slices of strings.
type chunk struct {
	// lines stores the contents of the file in slices of strings.
	// lines,endNum and eof is updated by reader goroutine.
	lines [][]byte
	// start is the first position of the number of bytes read.
	start int64
}

// NewDocument returns Document.
func NewDocument() (*Document, error) {
	m := &Document{
		tickerDone: make(chan struct{}),
		general: general{
			ColumnDelimiter: "",
			TabWidth:        8,
			MarkStyleWidth:  1,
			PlainMode:       false,
		},
		ctlCh:         make(chan controlSpecifier),
		seekable:      true,
		preventReload: false,
		chunks: []*chunk{
			NewChunk(0),
		},
	}
	if err := m.NewCache(); err != nil {
		return nil, err
	}
	return m, nil
}

func NewChunk(start int64) *chunk {
	return &chunk{
		lines: make([][]byte, 0, ChunkSize),
		start: start,
	}
}

// NewCache creates a new cache.
func (m *Document) NewCache() error {
	cache, err := lru.New[int, LineC](1024)
	if err != nil {
		return fmt.Errorf("new cache %w", err)
	}
	m.cache = cache
	return nil
}

// OpenDocument opens a file and returns a Document.
func OpenDocument(fileName string) (*Document, error) {
	fi, err := os.Stat(fileName)
	if err != nil {
		return nil, fmt.Errorf("%s %w", fileName, err)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("%s %w", fileName, ErrIsDirectory)
	}

	m, err := NewDocument()
	if err != nil {
		return nil, err
	}

	// named pipe.
	if fi.Mode()&fs.ModeNamedPipe != 0 {
		m.seekable = false
	}

	f, err := open(fileName)
	if err != nil {
		return nil, err
	}
	m.FileName = fileName
	if err := m.ControlFile(f); err != nil {
		return nil, err
	}
	return m, nil
}

// STDINDocument returns a Document that reads stdin.
func STDINDocument() (*Document, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}

	m.seekable = false
	m.Caption = "(STDIN)"
	f, err := open("")
	if err != nil {
		return nil, err
	}
	if err := m.ControlFile(f); err != nil {
		return nil, err
	}
	return m, nil
}

// Line returns one line from buffer.
func (m *Document) Line(n int) []byte {
	if n < m.startNum || n >= m.endNum {
		return nil
	}
	chunkNum := n / ChunkSize
	if len(m.chunks)-1 < chunkNum {
		log.Println("over chunk size: ", chunkNum)
		return nil
	}
	chunk := m.chunks[chunkNum]

	if len(chunk.lines) == 0 && atomic.LoadInt32(&m.closed) == 0 {
		m.loadControl(chunkNum)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	cn := n % ChunkSize
	if cn < len(chunk.lines) {
		return chunk.lines[cn]
	}
	log.Println("not load", n, m.endNum)
	return nil
}

// Deprecated: GetLine returns one line from buffer.
func (m *Document) GetLine(n int) string {
	return string(m.Line(n))
}

// LineString returns one line from buffer.
func (m *Document) LineString(n int) string {
	return string(m.Line(n))
}

// loadControl sends instructions to load chunks into memory.
func (m *Document) loadControl(chunkNum int) {
	sc := controlSpecifier{
		control:  loadControl,
		chunkNum: chunkNum,
		done:     make(chan bool),
	}
	m.ctlCh <- sc
	<-sc.done
}

// searchControl sends instructions to load chunks into memory.
func (m *Document) searchControl(chunkNum int, searcher Searcher) bool {
	sc := controlSpecifier{
		control:  searchControl,
		searcher: searcher,
		chunkNum: chunkNum,
		done:     make(chan bool),
	}
	m.ctlCh <- sc
	return <-sc.done
}

func (m *Document) closeControl() {
	atomic.StoreInt32(&m.readCancel, 1)
	sc := controlSpecifier{
		control: closeControl,
		done:    make(chan bool),
	}

	log.Println("close send")
	m.ctlCh <- sc
	<-sc.done
	log.Println("receive done")
	atomic.StoreInt32(&m.readCancel, 0)
}

// CurrentLN returns the currently displayed line number.
func (m *Document) CurrentLN() int {
	return m.topLN
}

// Export exports the document in the specified range.
func (m *Document) Export(w io.Writer, start int, end int) {
	for n := start; n <= end; n++ {
		if n >= m.BufEndNum() {
			break
		}
		fmt.Fprint(w, m.LineString(n))
	}
}

// BufEndNum return last line number.
func (m *Document) BufEndNum() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.endNum
}

// BufEOF return true if EOF is reached.
func (m *Document) BufEOF() bool {
	return atomic.LoadInt32(&m.eof) == 1
}

// ClearCache clears the cache.
func (m *Document) ClearCache() {
	m.cache.Purge()
}

// contents returns contents from line number and tabWidth.
func (m *Document) contents(lN int, tabWidth int) (contents, error) {
	if lN < 0 || lN >= m.BufEndNum() {
		return nil, ErrOutOfRange
	}

	str := m.LineString(lN)
	return parseString(str, tabWidth), nil
}

// getLineC returns contents from line number and tabWidth.
// If the line number does not exist, EOF content is returned.
func (m *Document) getLineC(lN int, tabWidth int) (LineC, bool) {
	if line, ok := m.cache.Get(lN); ok {
		lc := make(contents, len(line.lc))
		copy(lc, line.lc)
		line.lc = lc
		return line, true
	}

	org, err := m.contents(lN, tabWidth)
	if err != nil {
		// EOF
		lc := make(contents, 1)
		lc[0] = EOFContent
		return LineC{
			lc:  lc,
			str: string(EOFC),
			pos: widthPos{0: 0, 1: 1},
		}, false
	}
	str, pos := ContentsToStr(org)
	line := LineC{
		lc:  org,
		str: str,
		pos: pos,
	}
	if len(org) != 0 {
		m.cache.Add(lN, line)
	}

	lc := make(contents, len(org))
	copy(lc, org)
	line.lc = lc
	return line, true
}

// firstLine is the first line that excludes the SkipLines and Header.
func (m *Document) firstLine() int {
	return m.SkipLines + m.Header
}

// GetChunkLine returns one line from buffer.
func (m *Document) GetChunkLine(chunkNum int, cn int) ([]byte, error) {
	if len(m.chunks) <= chunkNum {
		return nil, ErrOutOfChunk
	}
	chunk := m.chunks[chunkNum]

	m.mu.Lock()
	defer m.mu.Unlock()

	if cn >= len(chunk.lines) {
		return nil, ErrOutOfRange
	}
	return chunk.lines[cn], nil
}

// chunkLine returns chunkNum and chunk line number from line number.
func chunkLine(n int) (int, int) {
	chunkNum := n / ChunkSize
	cn := n % ChunkSize
	return chunkNum, cn
}

// watchMode sets the document to watch mode.
func (m *Document) watchMode() {
	m.WatchMode = true
	if m.SectionDelimiter == "" {
		m.setSectionDelimiter("^" + FormFeed)
	}
	m.SectionStartPosition = 1
	m.FollowSection = true
}

// unwatchMode unwatch mode for the document.
func (m *Document) unWatchMode() {
	m.WatchMode = false
	m.FollowSection = false
}

// regexpCompile compiles the new document's regular expressions.
func (m *Document) regexpCompile() {
	m.ColumnDelimiterReg = condRegexpCompile(m.ColumnDelimiter)
	m.setSectionDelimiter(m.SectionDelimiter)
	if m.MultiColorWords != nil {
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
	tl := min(1000, len(m.chunks[0].lines))
	lines := m.chunks[0].lines[m.SkipLines:tl]
	buf := make([]string, len(lines))
	for n, line := range lines {
		buf[n] = string(line)
	}
	m.columnWidths = guesswidth.Positions(buf, header, 2)
}
