package oviewer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"golang.org/x/term"
)

type controlSpecifier struct {
	done     chan struct{}
	control  control
	chunkNum int
}

type control string

const (
	firstControl    = "first"
	continueControl = "read"
	followControl   = "follow"
	closeControl    = "close"
	reloadControl   = "reload"
	loadControl     = "load"
	searchControl   = "search"
)

const FormFeed = "\f"

// ControlFile controls file read and loads in chunks.
// ControlFile can be reloaded by file name.
func (m *Document) ControlFile(file *os.File) error {
	go func() error {
		atomic.StoreInt32(&m.closed, 0)
		r, err := m.fileReader(file)
		if err != nil {
			atomic.StoreInt32(&m.closed, 1)
			log.Println(err)
		}
		atomic.StoreInt32(&m.eof, 0)
		reader := bufio.NewReader(r)
		for sc := range m.ctlCh {
			reader, err = m.control(sc, reader)
			if err != nil {
				log.Println(err)
			}
			if sc.done != nil {
				close(sc.done)
			}
		}
		log.Println("close m.ctlCh")
		return nil
	}()

	m.startControl()
	return nil
}

func (m *Document) ControlLog() error {
	go func() error {
		for sc := range m.ctlCh {
			switch sc.control {
			case reloadControl:
				m.reset()
			default:
				panic(fmt.Sprintf("unexpected %s", sc.control))
			}
			if sc.done != nil {
				close(sc.done)
			}
		}
		log.Println("close m.ctlCh")
		return nil
	}()
	return nil
}

// controlreader is the controller for io.Reader.
// Assuming call from Exec. reload executes the argument function.
func (m *Document) ControlReader(r io.Reader, reload func() *bufio.Reader) error {
	reader := bufio.NewReader(r)
	go func() error {
		var err error
		for sc := range m.ctlCh {
			switch sc.control {
			case firstControl:
				reader, err = m.firstRead(reader)
				if !m.BufEOF() {
					m.continueControl()
				}
			case continueControl:
				reader, err = m.continueRead(reader)
				if !m.BufEOF() {
					m.continueControl()
				}
			case reloadControl:
				log.Println("reset")
				reader = reload()
				m.startControl()
			default:
				panic(fmt.Sprintf("unexpected %s", sc.control))
			}
			if err != nil {
				log.Println(err)
			}
			if sc.done != nil {
				close(sc.done)
			}
		}
		log.Println("close ctlCh")
		return nil
	}()

	m.startControl()
	return nil
}

func (m *Document) control(sc controlSpecifier, reader *bufio.Reader) (*bufio.Reader, error) {
	if atomic.LoadInt32(&m.closed) == 1 && sc.control != reloadControl {
		return nil, fmt.Errorf("closed %s", sc.control)
	}

	var err error
	switch sc.control {
	case firstControl:
		reader, err = m.firstRead(reader)
		if !m.BufEOF() {
			m.continueControl()
		}
	case continueControl:
		reader, err = m.continueRead(reader)
		if !m.BufEOF() {
			m.continueControl()
		}
	case followControl:
		reader, err = m.followRead(reader)
	case loadControl:
		reader, err = m.readChunk(reader, sc.chunkNum)
	case searchControl:
		err = m.searchChunk(reader, sc.chunkNum)
	case reloadControl:
		reader, err = m.reloadRead(reader)
		m.startControl()
	case closeControl:
		err = m.close()
		log.Println(err)
		err = nil
	default:
		panic(fmt.Sprintf("unexpected %s", sc.control))
	}
	return reader, nil
}

func (m *Document) startControl() {
	go func() {
		m.ctlCh <- controlSpecifier{
			control: firstControl,
		}
	}()
}

func (m *Document) continueControl() {
	go func() {
		m.ctlCh <- controlSpecifier{
			control: continueControl,
		}
	}()
}

// firstRead first reads the file.
// Packs the contents of the read file into the first chunk.
func (m *Document) firstRead(reader *bufio.Reader) (*bufio.Reader, error) {
	chunk := m.chunks[0]
	start := 0
	if err := m.packChunk(chunk, reader, start, true); err != nil {
		if !errors.Is(err, io.EOF) {
			atomic.StoreInt32(&m.eof, 1)
			return nil, err
		}
		reader = m.afterEOF(reader)
	}
	return reader, nil
}

// continueRead is executed after the second
// and only reads the file or counts the lines of the file.
func (m *Document) continueRead(reader *bufio.Reader) (*bufio.Reader, error) {
	if m.seekable {
		if _, err := m.file.Seek(m.offset, io.SeekStart); err != nil {
			atomic.StoreInt32(&m.eof, 1)
			return nil, err
		}
		reader.Reset(m.file)
	}

	chunk := m.lastChunk()
	start := len(chunk.lines)
	if err := m.readOrCountChunk(chunk, reader, start); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		reader = m.afterEOF(reader)
	}
	return reader, nil
}

// followRead reads lines added to the file while in follow-mode.
func (m *Document) followRead(reader *bufio.Reader) (*bufio.Reader, error) {
	if m.checkClose() {
		return reader, nil
	}
	if !m.FollowMode && !m.FollowAll {
		return reader, nil
	}
	if m.seekable {
		if _, err := m.file.Seek(m.offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek:%w", err)
		}
		reader = bufio.NewReader(m.file)
	}
	if err := m.readAll(reader); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		reader = m.afterEOF(reader)
	}
	return reader, nil
}

// readChunk loads the read contents into chunks.
func (m *Document) readChunk(reader *bufio.Reader, chunkNum int) (*bufio.Reader, error) {
	// non-seekable files are all in memory, so loadControl should not be called.
	if !m.seekable {
		return nil, fmt.Errorf("cannot be loaded")
	}

	chunk := m.chunks[chunkNum]
	if _, err := m.file.Seek(chunk.start, io.SeekStart); err != nil {
		return nil, err
	}
	reader.Reset(m.file)
	start := 0
	if err := m.packChunk(chunk, reader, start, false); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		reader = m.afterEOF(reader)
	}
	return reader, nil
}

func (m *Document) searchChunk(reader *bufio.Reader, chunkNum int) error {
	chunk := m.chunks[chunkNum]
	start := 0
	if m.seekable {
		if _, err := m.file.Seek(chunk.start, io.SeekStart); err != nil {
			return err
		}
		reader.Reset(m.file)
	}
	if err := m.packChunk(chunk, reader, start, false); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
		reader = m.afterEOF(reader)
	}
	return nil
}

// reloadRead performs reload processing
func (m *Document) reloadRead(reader *bufio.Reader) (*bufio.Reader, error) {
	if !m.WatchMode {
		m.reset()
	} else {
		chunk := m.lastChunk()
		m.appendFormFeed(chunk)
	}
	var err error
	reader, err = m.reloadFile(reader)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func (m *Document) readOrCountChunk(chunk *chunk, reader *bufio.Reader, start int) error {
	if !m.seekable {
		return m.packChunk(chunk, reader, start, true)
	}
	return m.countChunk(chunk, reader, start)
}

func (m *Document) reloadFile(reader *bufio.Reader) (*bufio.Reader, error) {
	if !m.seekable {
		m.ClearCache()
		return reader, nil
	}
	atomic.StoreInt32(&m.closed, 1)
	if err := m.file.Close(); err != nil {
		log.Printf("reload: %s", err)
	}
	m.ClearCache()

	atomic.StoreInt32(&m.closed, 0)
	atomic.StoreInt32(&m.eof, 0)
	log.Println("reload", m.FileName)
	r, err := m.openFileReader(m.FileName)
	if err != nil {
		str := fmt.Sprintf("Access is no longer possible: %s", err)
		reader = bufio.NewReader(strings.NewReader(str))
		return reader, nil
	}
	reader = bufio.NewReader(r)
	return reader, nil
}

func (m *Document) afterEOF(reader *bufio.Reader) *bufio.Reader {
	m.offset = m.size
	atomic.StoreInt32(&m.eof, 1)
	if !m.seekable { // for NamedPipe.
		return bufio.NewReader(m.file)
	}
	return reader
}

func (m *Document) fileReader(f *os.File) (io.Reader, error) {
	m.mu.Lock()
	m.file = f

	cFormat, r := uncompressedReader(m.file, m.seekable)
	if cFormat == UNCOMPRESSED {
		if m.seekable {
			f.Seek(0, io.SeekStart)
			r = f
		}
	} else {
		m.seekable = false
	}
	m.CFormat = cFormat
	if STDOUTPIPE != nil {
		r = io.TeeReader(r, STDOUTPIPE)
	}
	m.mu.Unlock()

	return r, nil
}

func (m *Document) openFileReader(fileName string) (io.Reader, error) {
	f, err := open(fileName)
	if err != nil {
		return nil, err
	}
	r, err := m.fileReader(f)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func open(fileName string) (*os.File, error) {
	if fileName == "" {
		if term.IsTerminal(0) {
			return nil, ErrMissingFile
		}
		return os.Stdin, nil
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// readAll reads to the end.
// The read lines are stored in the lines of the Document.
func (m *Document) readAll(reader *bufio.Reader) error {
	chunk := m.lastChunk()
	start := len(chunk.lines)
	for {
		if err := m.packChunk(chunk, reader, start, true); err != nil {
			return err
		}
		chunk = NewChunk(m.size)
		m.mu.Lock()
		m.chunks = append(m.chunks, chunk)
		m.mu.Unlock()
		start = 0
	}
}

// packChunk append lines read from reader into chunks.
func (m *Document) packChunk(chunk *chunk, reader *bufio.Reader, start int, isCount bool) error {
	var line strings.Builder
	var isPrefix bool
	i := start
	for i < ChunkSize {
		if atomic.LoadInt32(&m.readCancel) == 1 {
			break
		}
		buf, err := reader.ReadSlice('\n')
		if err == bufio.ErrBufferFull {
			isPrefix = true
			err = nil
		}
		line.Write(buf)
		if isPrefix {
			isPrefix = false
			continue
		}

		i++
		atomic.StoreInt32(&m.changed, 1)
		if err != nil {
			if line.Len() != 0 {
				if isCount {
					m.append(chunk, line.String())
					m.offset = m.size
				} else {
					m.appendOnly(chunk, line.String())
				}
			}
			return err
		}
		if isCount {
			m.append(chunk, line.String())
			m.offset = m.size
		} else {
			m.appendOnly(chunk, line.String())
		}
		line.Reset()
	}
	return nil
}

func (m *Document) countChunk(chunk *chunk, reader *bufio.Reader, start int) error {
	var isPrefix bool
	i := start
	for i < ChunkSize {
		buf, err := reader.ReadSlice('\n')
		if err == bufio.ErrBufferFull {
			isPrefix = true
			err = nil
		}
		m.size += int64(len(buf))
		if isPrefix {
			isPrefix = false
			continue
		}

		i++
		if len(buf) != 0 {
			m.endNum++
		}
		m.offset = m.size
		atomic.StoreInt32(&m.changed, 1)
		if err != nil {
			return err
		}
	}
	return nil
}

// appendOnly appends to the line of the chunk.
// appendOnly does not updates the number of lines and size.
func (m *Document) appendOnly(chunk *chunk, line string) {
	m.mu.Lock()
	chunk.lines = append(chunk.lines, line)
	m.mu.Unlock()
}

// append appends to the line of the chunk.
// append updates the number of lines and size.
func (m *Document) append(chunk *chunk, line string) {
	m.mu.Lock()
	size := len(line)
	m.size += int64(size)
	m.endNum++
	chunk.lines = append(chunk.lines, line)
	m.mu.Unlock()
}

func (m *Document) appendFormFeed(chunk *chunk) {
	line := ""
	m.mu.Lock()
	if len(chunk.lines) > 0 {
		line = chunk.lines[len(chunk.lines)-1]
	}
	m.mu.Unlock()
	// Do not add if the previous is FormFeed.
	if line != FormFeed {
		m.append(chunk, FormFeed)
	}
}

func (m *Document) lastChunk() *chunk {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.endNum < len(m.chunks)*ChunkSize {
		return m.chunks[len(m.chunks)-1]
	}
	chunk := NewChunk(m.size)
	m.chunks = append(m.chunks, chunk)
	return chunk
}

// reload will read again.
// Regular files are reopened and reread increase.
// The pipe will reset what it has read.
func (m *Document) reload() error {
	atomic.StoreInt32(&m.readCancel, 1)
	sc := controlSpecifier{
		control: reloadControl,
		done:    make(chan struct{}),
	}
	m.ctlCh <- sc
	<-sc.done
	atomic.StoreInt32(&m.readCancel, 0)
	if !m.WatchMode {
		m.topLN = 0
	}

	return nil
}

// reset clears all lines.
func (m *Document) reset() {
	m.mu.Lock()
	m.size = 0
	m.endNum = 0
	m.chunks = []*chunk{
		NewChunk(0),
	}
	m.mu.Unlock()
	atomic.StoreInt32(&m.changed, 1)
	m.ClearCache()
}

// checkClose returns if the file is closed.
func (m *Document) checkClose() bool {
	return atomic.LoadInt32(&m.closed) == 1
}

// Close closes the File.
// Record the last read position.
func (m *Document) close() error {
	if m.checkClose() {
		return nil
	}
	log.Println("close")
	if err := m.file.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	atomic.StoreInt32(&m.eof, 1)
	atomic.StoreInt32(&m.closed, 1)
	atomic.StoreInt32(&m.changed, 1)
	return nil
}

// ReadFile reads file.
// If the file name is empty, read from standard input.
//
// Deprecated:
func (m *Document) ReadFile(fileName string) error {
	r, err := m.openFileReader(fileName)
	if err != nil {
		return err
	}
	return m.ReadAll(r)
}

// ContinueReadAll continues to read even if it reaches EOF.
//
// Deprecated:
func (m *Document) ContinueReadAll(ctx context.Context, r io.Reader) error {
	return m.ReadAll(r)
}

// ReadReader reads reader.
// A wrapper for ReadAll.
//
// Deprecated:
func (m *Document) ReadReader(r io.Reader) error {
	return m.ReadAll(r)
}

// ReadAll reads all from the reader.
// And store it in the lines of the Document.
// ReadAll needs to be notified on eofCh.
//
// Deprecated:
func (m *Document) ReadAll(r io.Reader) error {
	reader := bufio.NewReader(r)
	go func() {
		if m.checkClose() {
			return
		}

		if err := m.readAll(reader); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed) {
				m.offset = m.size
				atomic.StoreInt32(&m.eof, 1)
				return
			}
			log.Printf("error: %v\n", err)
			return
		}
	}()
	return nil
}
