package oviewer

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru"
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

	// File is the os.File.
	file *os.File

	// chunks is the content of the file to be stored in chunks.
	chunks []*chunk

	// notify when eof is reached.
	eofCh chan struct{}
	// notify when reopening.
	followCh chan struct{}

	// notify when a file changes.
	changCh chan struct{}

	cancel context.CancelFunc

	cache *lru.Cache

	ticker     *time.Ticker
	tickerDone chan struct{}

	// multiColorRegexps holds multicolor regular expressions in slices.
	multiColorRegexps []*regexp.Regexp

	// marked is a list of marked line numbers.
	marked []int

	// status is the display status of the document.
	general

	// size is the number of bytes read.
	size int64
	// offset
	offset int64

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
	// openFollow represents the open followMode file.
	openFollow int32
	// 1 if there is a changed.
	changed int32
	// 1 if there is a closed.
	closed int32

	// WatchMode is watch mode.
	WatchMode bool
	// preventReload is true to prevent reload.
	preventReload bool
	// Is it possible to seek.
	seekable bool
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
	lines []string
	// start is the first position of the number of bytes read.
	start int64
}

// NewDocument returns Document.
func NewDocument() (*Document, error) {
	cache, err := lru.New(1024)
	if err != nil {
		return nil, err
	}
	m := &Document{
		eofCh:      make(chan struct{}),
		followCh:   make(chan struct{}),
		changCh:    make(chan struct{}),
		tickerDone: make(chan struct{}),
		general: general{
			ColumnDelimiter: "",
			TabWidth:        8,
			MarkStyleWidth:  1,
			PlainMode:       false,
		},
		seekable:      true,
		preventReload: false,
		chunks: []*chunk{
			NewChunk(0),
		},
		cache: cache,
	}
	return m, nil
}

func NewChunk(start int64) *chunk {
	return &chunk{
		lines: make([]string, 0, ChunkSize),
		start: start,
	}
}

// OpenDocument opens a file and returns a Document.
func OpenDocument(fileName string) (*Document, error) {
	fi, err := os.Stat(fileName)
	if err != nil {
		return nil, err
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

	if err := m.ReadFile(fileName); err != nil {
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
	if err := m.ReadFile(""); err != nil {
		return nil, err
	}
	return m, nil
}

// GetLine returns one line from buffer.
func (m *Document) GetLine(n int) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if n < 0 || n >= m.endNum {
		return ""
	}

	chunkNum := n / ChunkSize
	chunkLine := n % ChunkSize
	if len(m.chunks)-1 < chunkNum {
		log.Println("over chunk size: ", chunkNum)
		return ""
	}
	chunk := m.chunks[chunkNum]
	if len(chunk.lines)-1 < chunkLine {
		log.Printf("over lines size: chunk[%d]:%d < %d", chunkNum, len(chunk.lines)-1, chunkLine)
		return ""
	}

	return chunk.lines[chunkLine]
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
		fmt.Fprint(w, m.GetLine(n))
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

// contents returns contents from line number and tabwidth.
func (m *Document) contents(lN int, tabWidth int) (contents, error) {
	if lN < 0 || lN >= m.BufEndNum() {
		return nil, ErrOutOfRange
	}

	str := m.GetLine(lN)
	lc := parseString(str, tabWidth)
	return lc, nil
}

// getLineC returns contents from line number and tabwidth.
// If the line number does not exist, EOF content is returned.
func (m *Document) getLineC(lN int, tabWidth int) (LineC, bool) {
	if v, ok := m.cache.Get(lN); ok {
		line := v.(LineC)
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
	m.cache.Add(lN, line)

	lc := make(contents, len(org))
	copy(lc, org)
	line.lc = lc
	return line, true
}

// firstLine is the first line that excludes the SkipLines and Header.
func (m *Document) firstLine() int {
	return m.SkipLines + m.Header
}

// SearchLine searches the document and returns the matching line number.
func (m *Document) SearchLine(ctx context.Context, searcher Searcher, lN int) (int, error) {
	lN = max(lN, 0)
	end := m.BufEndNum()
	for n := lN; n < end; n++ {
		if searcher.Match(m.GetLine(n)) {
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

// BackSearchLine does a backward search on the document and returns a matching line number.
func (m *Document) BackSearchLine(ctx context.Context, searcher Searcher, lN int) (int, error) {
	lN = min(lN, m.BufEndNum()-1)

	for n := lN; n >= 0; n-- {
		if searcher.Match(m.GetLine(n)) {
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
