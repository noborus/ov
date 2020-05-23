package oviewer

import (
	"fmt"
	"sync"

	"github.com/dgraph-io/ristretto"
	"github.com/gdamore/tcell"
)

// The Model structure contains the values
// for the logical screen.
type Model struct {
	// updated by reader goroutine
	buffer []string
	endNum int
	eof    bool

	x          int
	lineNum    int
	yy         int
	header     []string
	beforeSize int
	vWidth     int
	vHight     int
	cache      *ristretto.Cache

	mu sync.Mutex
}

//NewModel returns Model.
func NewModel() *Model {
	return &Model{
		buffer:     make([]string, 0, 1000),
		header:     make([]string, 0),
		beforeSize: 1000,
	}
}

// Content represents one character on the terminal.
type Content struct {
	width int
	style tcell.Style
	mainc rune
	combc []rune
}

type lineContents struct {
	contents []Content
	cMap     map[int]int
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

// BufLen return buffer length.
func (m *Model) BufLen() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.buffer)
}

// BufEndNum return last line number.
func (m *Model) BufEndNum() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.endNum
}

// BufEOF return reaching EOF.
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
func (m *Model) GetContents(lineNum int, tabWidth int) []Content {
	lc, err := m.lineToContents(lineNum, tabWidth)
	if err != nil {
		return nil
	}
	return lc.contents
}

func (m *Model) lineToContents(lineNum int, tabWidth int) (lineContents, error) {
	var lc lineContents

	if lineNum < 0 || lineNum >= m.BufLen() {
		return lc, fmt.Errorf("out of range")
	}

	value, found := m.cache.Get(lineNum)
	if found {
		var ok bool
		if lc, ok = value.(lineContents); !ok {
			return lc, fmt.Errorf("fatal error not lineContents")
		}
	} else {
		lc.contents, lc.cMap = parseString(m.GetLine(lineNum), tabWidth)
		m.cache.Set(lineNum, lc, 1)
	}
	return lc, nil
}
