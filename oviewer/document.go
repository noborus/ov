package oviewer

import (
	"bufio"
	"errors"
	"io"
	"log"
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

	// File is the os.File.
	file *os.File
	// offset
	offset int64
	// CFormat is a compressed format.
	CFormat Compressed

	// lines stores the contents of the file in slices of strings.
	// lines,endNum and eof is updated by reader goroutine.
	lines []string
	// endNum is the number of the last line read.
	endNum int
	// true if EOF is reached.
	eof bool
	// true if the file is changed.
	changed chan bool
	// beforeSize represents the number of lines to read first.
	beforeSize int
	// cache represents a cache of contents.
	cache *ristretto.Cache

	lastContentsNum int
	lastContentsStr string
	lastContentsMap map[int]int

	// status is the display status of the document.
	general
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
		lines:      make([]string, 0),
		changed:    make(chan bool),
		beforeSize: 100,
		general: general{
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
	r, err := os.Open(fileName)
	if err != nil {
		return err
	}
	m.file = r
	m.FileName = fileName

	cFormat, reader := uncompressedReader(r)
	m.CFormat = cFormat

	if err := m.ReadAll(reader); err != nil {
		return err
	}
	return nil
}

// ReadSTDIN read STDIN.
func (m *Document) ReadSTDIN() error {
	if term.IsTerminal(0) {
		return ErrMissingFile
	}

	m.FileName = "(STDIN)"

	cFormat, reader := uncompressedReader(os.Stdin)
	m.CFormat = cFormat

	if err := m.ReadAll(reader); err != nil {
		return err
	}

	return nil
}

func (m *Document) Close() error {
	pos, err := m.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	m.offset = pos
	return m.file.Close()
}

func (m *Document) reOpen() error {
	go func() {
		if err := m.Close(); err != nil {
			return
		}

		if !m.BufEOF() {
			return
		}

		m.mu.Lock()
		m.eof = false
		m.mu.Unlock()

		r, err := os.Open(m.FileName)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = r.Seek(m.offset, io.SeekStart)
		if err != nil {
			log.Println(err)
			return
		}

		rr := compressedFormatReader(m.CFormat, r)
		reader := bufio.NewReader(rr)
		for {
			err := m.ReadFollow(reader)
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed) {
					log.Println("wait")
					<-m.changed
					log.Println("start")
					continue
				}
				log.Println(err)
				break
			}
		}
	}()
	return nil
}

// GetLine returns one line from buffer.
func (m *Document) GetLine(lN int) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if lN < 0 || lN >= len(m.lines) {
		return ""
	}
	return m.lines[lN]
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
func (m *Document) lineToContents(lN int, tabWidth int) (lineContents, error) {
	if lN < 0 || lN >= m.BufEndNum() {
		return nil, ErrOutOfRange
	}

	value, found := m.cache.Get(lN)
	if found {
		lc, ok := value.(lineContents)
		if !ok {
			return lc, ErrFatalCache
		}
		return lc, nil
	}

	lc := parseString(m.GetLine(lN), tabWidth)

	m.cache.Set(lN, lc, 1)
	return lc, nil
}
