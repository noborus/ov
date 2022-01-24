package oviewer

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/dgraph-io/ristretto"
)

// The Document structure contains the values
// for the logical screen.
type Document struct {
	// fileName is the file name to display.
	FileName string

	// File is the os.File.
	file *os.File
	// offset
	offset int64
	// CFormat is a compressed format.
	CFormat Compressed

	// Is it possible to seek.
	seekable bool

	// lines stores the contents of the file in slices of strings.
	// lines,endNum and eof is updated by reader goroutine.
	lines []string
	// endNum is the number of the last line read.
	endNum int

	// 1 if EOF is reached.
	eof int32
	// notify when eof is reached.
	eofCh chan struct{}
	// notify when reopening.
	followCh chan struct{}
	// onceFollow represents the open followMode file.
	onceFollow sync.Once

	// 1 if there is a changed.
	changed int32
	// notify when a file changes.
	changCh chan struct{}

	// 1 if there is a closed.
	closed int32

	// cache represents a cache of contents.
	cache *ristretto.Cache

	lastContentsNum int
	lastContentsStr string
	lastContentsMap map[int]int

	// status is the display status of the document.
	general

	// latestNum is the endNum read at the end of the screen update.
	latestNum int
	// topLN is the starting position of the current y.
	topLN int
	// topLX represents the x position of the top line.
	topLX int
	// x is the starting position of the current x.
	x int
	// columnNum is the number of columns.
	columnNum int

	// marked is a list of marked line numbers.
	marked      []int
	markedPoint int

	// mu controls the mutex.
	mu sync.Mutex
}

// NewDocument returns Document.
func NewDocument() (*Document, error) {
	m := &Document{
		lines:    make([]string, 0),
		eofCh:    make(chan struct{}),
		followCh: make(chan struct{}),
		changCh:  make(chan struct{}),
		general: general{
			ColumnDelimiter: "",
			TabWidth:        8,
			MarkStyleWidth:  1,
		},
		lastContentsNum: -1,
		seekable:        true,
	}

	if err := m.NewCache(); err != nil {
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
	return m.lines[n]
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

// NewCache creates a new cache.
func (m *Document) NewCache() error {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10000, // number of keys to track frequency of.
		MaxCost:     1000,  // maximum cost of cache.
		BufferItems: 64,    // number of keys per Get buffer.
	})
	if err != nil {
		return err
	}
	m.cache = cache
	return nil
}

// ClearCache clears the cache.
func (m *Document) ClearCache() {
	m.cache.Clear()
}

// contentsLN returns contents from line number and tabwidth.
func (m *Document) contentsLN(lN int, tabWidth int) (lineContents, error) {
	if lN < 0 || lN >= m.BufEndNum() {
		return nil, ErrOutOfRange
	}

	key := fmt.Sprintf("contents:%d", lN)
	if value, found := m.cache.Get(key); found {
		// It was cached.
		lc, ok := value.(lineContents)
		if !ok {
			return lc, ErrFatalCache
		}
		return lc, nil
	}

	// It wasn't cached.
	str := m.GetLine(lN)
	lc := parseString(str, tabWidth)
	m.cache.Set(key, lc, 1)
	return lc, nil
}

// getContents returns lineContents from line number and tabwidth.
// If the line number does not exist, EOF content is returned.
func (m *Document) getContents(lN int, tabWidth int) lineContents {
	org, err := m.contentsLN(lN, tabWidth)
	if err != nil {
		// EOF
		lc := make(lineContents, 1)
		lc[0] = EOFContent
		return lc
	}
	lc := make(lineContents, len(org))
	copy(lc, org)
	return lc
}

// getContentsStr is returns a converted string
// and byte length and contents length conversion table.
// getContentsStr saves the last result
// and reduces the number of executions of contentsToStr.
// Because it takes time to analyze a line with a very long line.
func (m *Document) getContentsStr(lN int, lc lineContents) (string, map[int]int) {
	if m.lastContentsNum != lN {
		m.lastContentsStr, m.lastContentsMap = ContentsToStr(lc)
		m.lastContentsNum = lN
	}
	return m.lastContentsStr, m.lastContentsMap
}

// fistLine is the first line that excludes the SkipLines and Header.
func (m *Document) firstLine() int {
	return m.SkipLines + m.Header
}
