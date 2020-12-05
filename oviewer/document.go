package oviewer

import (
	"io"
	"os"
	"sync"

	"github.com/dgraph-io/ristretto"
	"golang.org/x/term"
)

// The Document structure contains the values
// for the logical screen.
type Document struct {
	// fileName is the file name to display.
	FileName string
	// lines stores the contents of the file in slices of strings.
	// lines,endNum and eof is updated by reader goroutine.
	lines []string
	// endNum is the number of the last line read.
	endNum int
	// true if EOF is reached.
	eof bool
	// beforeSize represents the number of lines to read first.
	beforeSize int
	// cache represents a cache of contents.
	cache *ristretto.Cache

	lastContentsNum int
	lastContentsStr string
	lastContentsMap map[int]int

	// status is the display status of the document.
	status
	// topLN is the starting position of the current y.
	topLN int
	// topLX represents the x position of the top line.
	topLX int
	// x is the starting position of the current x.
	x int
	// columnNum is the number of columns.
	columnNum int

	// mu controls the mutex.
	mu sync.Mutex
}

// NewDocument returns Document.
func NewDocument() (*Document, error) {
	m := &Document{
		lines:      make([]string, 0, 1000),
		beforeSize: 1000,
		status: status{
			ColumnDelimiter: "",
			TabWidth:        8,
		},
	}

	if err := m.NewCache(); err != nil {
		return nil, err
	}
	return m, nil
}

// ReadFile reads file.
func (m *Document) ReadFile(fileName string) error {
	var reader io.ReadCloser
	if fileName == "" {
		if term.IsTerminal(0) {
			return ErrMissingFile
		}
		fileName = "(STDIN)"
		reader = uncompressedReader(os.Stdin)
	} else {
		r, err := os.Open(fileName)
		if err != nil {
			return err
		}
		reader = uncompressedReader(r)
	}

	if err := m.ReadAll(reader); err != nil {
		return err
	}

	m.FileName = fileName

	return nil
}

// GetLine returns one line from buffer.
func (m *Document) GetLine(lineNum int) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if lineNum < 0 || lineNum >= len(m.lines) {
		return ""
	}
	return m.lines[lineNum]
}

// BufEndNum return last line number.
func (m *Document) BufEndNum() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.endNum
}

// BufEOF return true if EOF is reached.
func (m *Document) BufEOF() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.eof
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

// lineToContents returns contents from line number.
func (m *Document) lineToContents(lineNum int, tabWidth int) (lineContents, error) {
	if lineNum < 0 || lineNum >= m.BufEndNum() {
		return nil, ErrOutOfRange
	}

	value, found := m.cache.Get(lineNum)
	if found {
		lc, ok := value.(lineContents)
		if !ok {
			return lc, ErrFatalCache
		}
		return lc, nil
	}

	lc := parseString(m.GetLine(lineNum), tabWidth)

	m.cache.Set(lineNum, lc, 1)
	return lc, nil
}
