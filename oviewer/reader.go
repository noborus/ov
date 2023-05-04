package oviewer

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/term"
)

// FormFeed is the delimiter that separates the sections.
// The default delimiter that separates single output from watch.
const FormFeed = "\f"

// firstRead first reads the file.
// Fill the contents of the read file into the first chunk.
func (m *Document) firstRead(reader *bufio.Reader) (*bufio.Reader, error) {
	chunk := m.chunks[0]
	if err := m.fillChunk(chunk, reader, 0, true); err != nil {
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
	chunk := m.chunkForAdd()
	start := len(chunk.lines)
	if err := m.addChunk(chunk, reader, start); err != nil {
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

	chunk := m.chunkForAdd()
	start := len(chunk.lines)
	if err := m.fillChunk(chunk, reader, start, true); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		reader = m.afterEOF(reader)
	}
	return reader, nil
}

// readChunk loads the read contents into chunks.
func (m *Document) readChunk(reader *bufio.Reader, chunkNum int) (*bufio.Reader, error) {
	chunk := m.chunks[chunkNum]
	if _, err := m.file.Seek(chunk.start, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek:%w", err)
	}
	reader.Reset(m.file)
	start := 0
	if err := m.fillChunk(chunk, reader, start, false); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		reader = m.afterEOF(reader)
	}
	return reader, nil
}

// reloadRead performs reload processing.
func (m *Document) reloadRead(reader *bufio.Reader) (*bufio.Reader, error) {
	if !m.WatchMode {
		m.reset()
	} else {
		m.seekable = false
		chunk := m.chunkForAdd()
		m.appendFormFeed(chunk)
	}
	var err error
	reader, err = m.reloadFile(reader)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

// addChunk fills or reserves a chunk.
func (m *Document) addChunk(chunk *chunk, reader *bufio.Reader, start int) error {
	if m.seekable {
		return m.reserveChunk(reader, start)
	}
	return m.fillChunk(chunk, reader, start, true)
}

// reloadFile reloads a file.
func (m *Document) reloadFile(reader *bufio.Reader) (*bufio.Reader, error) {
	if !m.reopenable {
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
	r, err := m.openFileReader(m.FileName)
	if err != nil {
		str := fmt.Sprintf("Access is no longer possible: %s", err)
		reader = bufio.NewReader(strings.NewReader(str))
		return reader, nil
	}
	reader = bufio.NewReader(r)
	return reader, nil
}

// afterEOF does processing after reaching EOF.
func (m *Document) afterEOF(reader *bufio.Reader) *bufio.Reader {
	m.offset = m.size
	atomic.StoreInt32(&m.eof, 1)
	if !m.seekable { // for NamedPipe.
		return bufio.NewReader(m.file)
	}
	return reader
}

// openFileReader opens a file,
// selects a reader according to compression type and returns io.Reader.
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

// fileReader returns a io.Reader.
func (m *Document) fileReader(f *os.File) (io.Reader, error) {
	m.mu.Lock()
	m.file = f

	cFormat, r := uncompressedReader(m.file, m.seekable)
	if cFormat == UNCOMPRESSED {
		if m.seekable {
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				return nil, fmt.Errorf("seek:%w", err)
			}
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

// open opens a file.
func open(fileName string) (*os.File, error) {
	if fileName == "" {
		if term.IsTerminal(0) {
			return nil, ErrMissingFile
		}
		return os.Stdin, nil
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fileName, err)
	}
	return f, nil
}

// readAll reads to the end.
// The read lines are stored in the lines of the Document.
func (m *Document) readAll(reader *bufio.Reader) error {
	chunk := m.chunkForAdd()
	start := len(chunk.lines)
	for {
		if err := m.fillChunk(chunk, reader, start, true); err != nil {
			return err
		}
		chunk = NewChunk(m.size)
		m.mu.Lock()
		m.chunks = append(m.chunks, chunk)
		m.mu.Unlock()
		start = 0
	}
}

// fillChunk append lines read from reader into chunks.
func (m *Document) fillChunk(chunk *chunk, reader *bufio.Reader, start int, isCount bool) error {
	var line bytes.Buffer
	var isPrefix bool
	num := start
	for num < ChunkSize {
		if atomic.LoadInt32(&m.readCancel) == 1 {
			break
		}
		buf, err := reader.ReadSlice('\n')
		if errors.Is(err, bufio.ErrBufferFull) {
			isPrefix = true
			err = nil
		}
		line.Write(buf)
		if isPrefix {
			isPrefix = false
			continue
		}

		num++
		atomic.StoreInt32(&m.changed, 1)
		if err != nil {
			if line.Len() != 0 {
				if isCount {
					m.append(chunk, line.Bytes())
					m.offset = m.size
				} else {
					m.appendOnly(chunk, line.Bytes())
				}
			}
			return err
		}
		if isCount {
			m.append(chunk, line.Bytes())
			m.offset = m.size
		} else {
			m.appendOnly(chunk, line.Bytes())
		}
		line.Reset()
	}
	return nil
}

const bufSize = 4096

// reserveChunk reserves ChunkSize lines.
// read and update size only.
func (m *Document) reserveChunk(reader *bufio.Reader, start int) error {
	num := start
	for {
		buf := make([]byte, bufSize)
		size, err := reader.Read(buf)
		if err != nil {
			return err
		}
		if size == 0 {
			return io.EOF
		}

		count := bytes.Count(buf, []byte("\n"))
		// If it exceeds ChunkSize, Re-aggregate size and count.
		if num+count > ChunkSize {
			start := 0
			left := ChunkSize - num
			for i := 0; i < left; i++ {
				p := bytes.IndexByte(buf[start:], '\n')
				start += p + 1
			}
			size = start
			count = left
		}

		num += count
		m.mu.Lock()
		m.endNum += count
		m.size += int64(size)
		m.offset = m.size
		m.mu.Unlock()
		if num >= ChunkSize {
			break
		}
	}
	atomic.StoreInt32(&m.changed, 1)
	return nil
}

// appendOnly appends to the line of the chunk.
// appendOnly does not updates the number of lines and size.
func (m *Document) appendOnly(chunk *chunk, line []byte) {
	m.mu.Lock()
	size := len(line)
	dst := make([]byte, size)
	copy(dst, line)
	chunk.lines = append(chunk.lines, dst)
	m.mu.Unlock()
}

// append appends to the line of the chunk.
// append updates the number of lines and size.
func (m *Document) append(chunk *chunk, line []byte) {
	m.mu.Lock()
	size := len(line)
	m.size += int64(size)
	m.endNum++
	dst := make([]byte, size)
	copy(dst, line)
	chunk.lines = append(chunk.lines, dst)
	m.mu.Unlock()
}

func (m *Document) appendFormFeed(chunk *chunk) {
	line := ""
	m.mu.Lock()
	if len(chunk.lines) > 0 {
		line = string(chunk.lines[len(chunk.lines)-1])
	}
	m.mu.Unlock()
	// Do not add if the previous is FormFeed(always add for formfeedTime).
	if line != FormFeed {
		feed := FormFeed
		if m.formfeedTime {
			feed = fmt.Sprintf("%sTime: %s", FormFeed, time.Now().Format(time.RFC3339))
		}
		m.append(chunk, []byte(feed))
	}
}

// chunkForAdd is a helper function to get the chunk to add.
func (m *Document) chunkForAdd() *chunk {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.endNum < len(m.chunks)*ChunkSize {
		return m.chunks[len(m.chunks)-1]
	}
	chunk := NewChunk(m.size)
	m.chunks = append(m.chunks, chunk)
	return chunk
}

// lastChunkNum returns the last chunk number.
func (m *Document) lastChunkNum() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.chunks) - 1
}

// reload will read again.
// Regular files are reopened and reread increase.
// The pipe will reset what it has read.
func (m *Document) reload() error {
	if m.preventReload {
		return ErrPreventReload
	}
	// Prevent reload if stdin reaches EOF.
	// Because no more content will be added.
	if m.FileName == "" && m.BufEOF() {
		return ErrEOFreached
	}

	atomic.StoreInt32(&m.readCancel, 1)
	m.requestReload()
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
// Deprecated: Use ControlFile instead.
func (m *Document) ReadFile(fileName string) error {
	r, err := m.openFileReader(fileName)
	if err != nil {
		return err
	}
	return m.ReadAll(r)
}

// ContinueReadAll continues to read even if it reaches EOF.
//
// Deprecated: Use ControlFile instead.
func (m *Document) ContinueReadAll(ctx context.Context, r io.Reader) error {
	return m.ReadAll(r)
}

// ReadReader reads reader.
// A wrapper for ReadAll.
//
// Deprecated: Use ControlReader instead.
func (m *Document) ReadReader(r io.Reader) error {
	return m.ReadAll(r)
}

// ReadAll reads all from the reader.
// And store it in the lines of the Document.
// ReadAll needs to be notified on eofCh.
//
// Deprecated: Use ControlReader instead.
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
