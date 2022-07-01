package oviewer

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto"
)

// The Document structure contains the values
// for the logical screen.
type Document struct {
	// fileName is the file name to display.
	FileName string
	// Caption is an additional caption to display after the file name.
	Caption string

	// File is the os.File.
	file *os.File
	// offset
	offset int64
	// CFormat is a compressed format.
	CFormat Compressed

	// preventReload is true to prevent reload.
	preventReload bool
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
	// openFollow represents the open followMode file.
	openFollow int32

	// 1 if there is a changed.
	changed int32
	// notify when a file changes.
	changCh chan struct{}

	cancel context.CancelFunc
	// 1 if there is a closed.
	closed int32

	// cache represents a cache of contents.
	cache *ristretto.Cache

	lastContentsNum int
	lastContentsStr string
	lastContentsMap map[int]int

	// status is the display status of the document.
	general

	// WatchMode is watch mode.
	WatchMode bool
	ticker    *time.Ticker

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
	// columnNum is the number of columns.
	columnNum int

	// marked is a list of marked line numbers.
	marked      []int
	markedPoint int

	// Last moved Section position.
	lastSectionPosNum int

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
		preventReload:   false,
	}

	if err := m.NewCache(); err != nil {
		return nil, err
	}
	return m, nil
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
	return m.lines[n]
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
		fmt.Fprintln(w, m.GetLine(n))
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
func (m *Document) contentsLN(lN int, tabWidth int) (contents, error) {
	if lN < 0 || lN >= m.BufEndNum() {
		return nil, ErrOutOfRange
	}

	key := fmt.Sprintf("contents:%d", lN)
	if value, found := m.cache.Get(key); found {
		// It was cached.
		lc, ok := value.(contents)
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

// getContents returns contents from line number and tabwidth.
// If the line number does not exist, EOF content is returned.
func (m *Document) getContents(lN int, tabWidth int) contents {
	org, err := m.contentsLN(lN, tabWidth)
	if err != nil {
		// EOF
		lc := make(contents, 1)
		lc[0] = EOFContent
		return lc
	}
	lc := make(contents, len(org))
	copy(lc, org)
	return lc
}

// getContentsStr is returns a converted string
// and byte length and contents length conversion table.
// getContentsStr saves the last result
// and reduces the number of executions of contentsToStr.
// Because it takes time to analyze a line with a very long line.
func (m *Document) getContentsStr(lN int, lc contents) (string, map[int]int) {
	if m.lastContentsNum != lN {
		m.lastContentsStr, m.lastContentsMap = ContentsToStr(lc)
		m.lastContentsNum = lN
	}
	return m.lastContentsStr, m.lastContentsMap
}

// firstLine is the first line that excludes the SkipLines and Header.
func (m *Document) firstLine() int {
	return m.SkipLines + m.Header
}

// SearchLine searches the document and returns the matching line.
func (m *Document) SearchLine(ctx context.Context, searcher Searcher, num int) (int, error) {
	num = max(num, 0)

	for n := num; n < m.BufEndNum(); n++ {
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

// BackSearchLine does a backward search on the document and returns a matching line.
func (m *Document) BackSearchLine(ctx context.Context, searcher Searcher, num int) (int, error) {
	num = min(num, m.BufEndNum()-1)

	for n := num; n >= 0; n-- {
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

func (m *Document) watchMode() {
	m.WatchMode = true
	if m.SectionDelimiter == "" {
		m.setSectionDelimiter("^" + FormFeed)
	}
	m.SectionStartPosition = 1
	m.FollowSection = true
}

func (m *Document) unWatchMode() {
	m.WatchMode = false
	m.FollowSection = false
}

func (m *Document) setSectionDelimiter(delm string) {
	m.SectionDelimiter = delm
	m.SectionDelimiterReg = regexpCompile(delm, true)
}
