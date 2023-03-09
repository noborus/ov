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
	chunkNum int
	done     chan struct{}
	control  control
}

type control string

const (
	firstControl  = "first"
	readControl   = "read"
	followControl = "follow"
	closeControl  = "close"
	reloadControl = "reload"
	loadControl   = "load"
	searchControl = "search"
)

const FormFeed = "\f"

// ControlFile controls file read and loads in chunks.
func (m *Document) ControlFile(fileName string) error {
	go func() error {
		r, err := m.openFile(fileName)
		if err != nil {
			log.Println(err)
			return err
		}
		atomic.StoreInt32(&m.eof, 0)
		reader := bufio.NewReader(r)
		for sc := range m.ctlCh {
			switch sc.control {
			case firstControl:
				reader, err = m.firstControl(reader)
			case readControl:
				reader, err = m.readControl(reader)
			case followControl:
				reader, err = m.followControl(reader)
			case reloadControl:
				reader, err = m.reloadControl(reader)
			case loadControl:
				reader, err = m.loadControl(reader, sc.chunkNum)
			case searchControl:
				err = m.searchControl(reader, sc.chunkNum)
			default:
				panic(fmt.Sprintf("unexpected %s", sc.control))
			}

			if sc.done != nil {
				close(sc.done)
			}
			if err != nil {
				log.Println(err)
				close(m.ctlCh)
			}
		}
		log.Println("close m.ctlCh")
		return nil
	}()

	m.ctlCh <- controlSpecifier{
		control:  firstControl,
		chunkNum: 0,
		done:     nil,
	}
	return nil
}

// ControlNonFile only supports reload.
func (m *Document) ControlNonFile() error {
	go func() error {
		var err error
		for sc := range m.ctlCh {
			switch sc.control {
			case reloadControl:
				m.reset()
			case loadControl:
				chunk := m.chunks[sc.chunkNum]
				chunk.lines = make([]string, ChunkSize)
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
	return nil
}

// firstControl is executed first.
// Open the file and read the first chunk.
func (m *Document) firstControl(reader *bufio.Reader) (*bufio.Reader, error) {
	if err := m.firstRead(reader); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		reader = m.afterEOF(reader)
	}
	if m.BufEOF() {
		return reader, nil
	}
	m.ctlCh <- controlSpecifier{
		control: readControl,
	}

	return reader, nil
}

// readControl is executed after the second
// and only reads the file or counts the lines of the file.
func (m *Document) readControl(reader *bufio.Reader) (*bufio.Reader, error) {
	if m.seekable {
		if _, err := m.file.Seek(m.offset, io.SeekStart); err != nil {
			return nil, err
		}
		reader.Reset(m.file)
	}

	chunk := m.lastChunk()
	start := 0
	if err := m.readOrCountChunk(chunk, reader, start); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		reader = m.afterEOF(reader)
	}

	if !m.BufEOF() {
		chunk := NewChunk(m.size)
		m.mu.Lock()
		m.chunks = append(m.chunks, chunk)
		m.mu.Unlock()
		m.ctlCh <- controlSpecifier{
			control: readControl,
		}
	}
	return reader, nil
}

// followControl reads lines added to the file while in follow-mode.
func (m *Document) followControl(reader *bufio.Reader) (*bufio.Reader, error) {
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

// loadControl loads the read contents into chunks.
func (m *Document) loadControl(reader *bufio.Reader, chunkNum int) (*bufio.Reader, error) {
	chunk := m.chunks[chunkNum]
	start := 0
	if m.seekable {
		if _, err := m.file.Seek(chunk.start, io.SeekStart); err != nil {
			return nil, err
		}
		reader.Reset(m.file)
	}
	if err := m.packChunk(chunk, reader, start, false); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		reader = m.afterEOF(reader)
	}
	return reader, nil
}

func (m *Document) searchControl(reader *bufio.Reader, chunkNum int) error {
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

// reloadControl performs reload processing
func (m *Document) reloadControl(reader *bufio.Reader) (*bufio.Reader, error) {
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
	err = m.firstRead(reader)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		reader = m.afterEOF(reader)
	}

	return reader, nil
}

func (m *Document) readOrCountChunk(chunk *chunk, reader *bufio.Reader, start int) error {
	if !m.seekable {
		log.Println("pack", len(m.chunks))
		return m.packChunk(chunk, reader, start, true)
	}
	log.Println("count", len(m.chunks))
	return m.countChunk(chunk, reader, start)
}

func (m *Document) reloadFile(reader *bufio.Reader) (*bufio.Reader, error) {
	if !m.seekable {
		m.ClearCache()
		return reader, nil
	}
	if err := m.file.Close(); err != nil {
		log.Printf("read: %s", err)
	}
	m.ClearCache()
	r, err := m.openFile(m.FileName)
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

func (m *Document) openFile(fileName string) (io.Reader, error) {
	f, err := open(fileName)
	if err != nil {
		return nil, err
	}
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

func (m *Document) firstRead(reader *bufio.Reader) error {
	if m.seekable {
		return m.readFirstOnly(reader)
	}
	return m.readFirstOnly(reader)
}

// readAll actually reads everything.
// The read lines are stored in the lines of the Document.
func (m *Document) readAll(reader *bufio.Reader) error {
	chunk := m.chunks[len(m.chunks)-1]
	start := len(chunk.lines)
	for {
		log.Println("loop?")
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

// readAll actually reads everything.
// The read lines are stored in the lines of the Document.
func (m *Document) readFirstOnly(reader *bufio.Reader) error {
	chunk := m.chunks[0]
	start := 0
	if err := m.packChunk(chunk, reader, start, true); err != nil {
		return err
	}
	return nil
}

// packChunk append lines read from reader into chunks.
func (m *Document) packChunk(chunk *chunk, reader *bufio.Reader, start int, isCount bool) error {
	var line strings.Builder
	var isPrefix bool
	i := start
	for i < ChunkSize {
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
	sc := controlSpecifier{
		control: reloadControl,
		done:    make(chan struct{}),
	}
	log.Println("reload send")
	m.ctlCh <- sc
	<-sc.done
	log.Println("receive done")
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
	atomic.StoreInt32(&m.closed, 1)
	atomic.StoreInt32(&m.changed, 1)
	return nil
}

// ReadFile reads file.
// If the file name is empty, read from standard input.
//
// Deprecated:
func (m *Document) ReadFile(fileName string) error {
	r, err := m.openFile(fileName)
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
