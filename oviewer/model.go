package oviewer

import (
	"io"
	"os"
	"sync"

	"github.com/dgraph-io/ristretto"
	"golang.org/x/crypto/ssh/terminal"
)

// The Model structure contains the values
// for the logical screen.
type Model struct {
	// fileName is the file name to display.
	fileName string
	// buffer stores the contents of the file in slices of strings.
	// buffer,endNum and eof is updated by reader goroutine.
	buffer []string
	// endNum is the number of the last line read.
	endNum int
	// true if EOF is reached.
	eof bool

	// x is the starting position of the current x.
	x int
	// lineNum is the starting position of the current y.
	lineNum int
	// yy represents the number of wrapped lines.
	yy int
	// header represents the header line.
	header []string
	// beforeSize represents the number of lines to read first.
	beforeSize int
	// vWidth represents the screen width.
	vWidth int
	// vHight represents the screen height.
	vHight int
	// cache represents a cache of contents.
	cache *ristretto.Cache
	// mu controls the mutex.
	mu sync.Mutex
}

// NewModel reads files (or stdin) and returns Model.
func NewModel(args []string) (*Model, error) {
	m := &Model{
		buffer:     make([]string, 0, 1000),
		header:     make([]string, 0),
		beforeSize: 1000,
	}
	err := m.NewCache()
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	fileName := ""
	switch len(args) {
	case 0:
		if terminal.IsTerminal(0) {
			return nil, ErrMissingFile
		}
		fileName = "(STDIN)"
		reader = uncompressedReader(os.Stdin)
	case 1:
		fileName = args[0]
		r, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}
		reader = uncompressedReader(r)
		defer r.Close()
	default:
		readers := make([]io.Reader, 0)
		for _, fileName := range args {
			r, err := os.Open(fileName)
			if err != nil {
				return nil, err
			}
			readers = append(readers, uncompressedReader(r))
			reader = io.MultiReader(readers...)
			defer r.Close()
		}
	}
	err = m.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	m.fileName = fileName

	return m, nil
}

// GetLine returns one line from buffer.
func (m *Model) GetLine(lineNum int) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if lineNum < 0 || lineNum >= len(m.buffer) {
		return ""
	}
	return m.buffer[lineNum]
}

// BufEndNum return last line number.
func (m *Model) BufEndNum() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.endNum
}

// BufEOF return true if EOF is reached.
func (m *Model) BufEOF() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.eof
}

// NewCache creates a new cache.
func (m *Model) NewCache() error {
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
func (m *Model) ClearCache() {
	m.cache.Clear()
}

// GetContents returns one line of contents from buffer.
func (m *Model) GetContents(lineNum int, tabWidth int) []content {
	lc, err := m.lineToContents(lineNum, tabWidth)
	if err != nil {
		return nil
	}
	return lc.contents
}

// lineToContents returns contents from line number.
func (m *Model) lineToContents(lineNum int, tabWidth int) (lineContents, error) {
	var lc lineContents

	if lineNum < 0 || lineNum >= m.BufEndNum() {
		return lc, ErrOutOfRange
	}

	value, found := m.cache.Get(lineNum)
	if found {
		var ok bool
		if lc, ok = value.(lineContents); !ok {
			return lc, ErrFatalCache
		}
	} else {
		lc = parseString(m.GetLine(lineNum), tabWidth)
		m.cache.Set(lineNum, lc, 1)
	}
	return lc, nil
}
