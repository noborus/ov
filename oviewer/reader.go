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

// bufSize is the size of the buffer used when reading the file.
// This bufSize is used when only counting.
const bufSize = 4096

// FormFeed is the delimiter that separates the sections.
// The default delimiter that separates single output from watch.
const FormFeed = "\f"

// firstRead first reads the file.
// Fill the contents of the read file into the first chunk.
func (m *Document) firstRead(reader *bufio.Reader) (*bufio.Reader, error) {
	atomic.StoreInt32(&m.noNewlineEOF, 0)
	chunk := m.store.chunks[0]
	if err := m.readLines(chunk, reader, 0, ChunkSize, true); err != nil {
		if !errors.Is(err, io.EOF) {
			atomic.StoreInt32(&m.eof, 1)
			return nil, err
		}
		reader = m.afterEOF(reader)
	}

	if !m.BufEOF() {
		m.requestContinue()
	}
	return reader, nil
}

// continueRead is executed after the second
// and only reads the file or counts the lines of the file.
func (m *Document) continueRead(reader *bufio.Reader) (*bufio.Reader, error) {
	if m.seekable {
		if _, err := m.file.Seek(m.offset, io.SeekStart); err != nil {
			atomic.StoreInt32(&m.eof, 1)
			log.Printf("seek: %s", err)
			m.seekable = false
		} else {
			reader.Reset(m.file)
		}
	}
	chunk := m.chunkForAdd()
	start := len(chunk.lines)
	if err := m.addOrReserveChunk(chunk, reader, start); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("addChunk: %w", err)
		}
		reader = m.afterEOF(reader)
	}

	if !m.BufEOF() {
		m.requestContinue()
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

	reader, err := m.loadRead(reader, m.lastChunkNum())
	if err != nil {
		return reader, err
	}

	atomic.StoreInt32(&m.eof, 0)
	chunk := m.store.chunks[m.lastChunkNum()]
	start := len(chunk.lines) - 1
	if atomic.LoadInt32(&m.noNewlineEOF) == 0 {
		chunk = m.chunkForAdd()
		start = len(chunk.lines)
	}
	if m.seekable {
		if _, err := m.file.Seek(m.offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek: %w", err)
		}
		reader = bufio.NewReader(m.file)
	}

	if err := m.readLines(chunk, reader, start, ChunkSize, true); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		reader = m.afterEOF(reader)
	}

	if !m.BufEOF() {
		m.requestContinue()
	}
	return reader, nil
}

// addOrReserveChunk reads a file to add or reserve a Chunk.
// If it's a seekable file, it just reserves it for later reading.
func (m *Document) addOrReserveChunk(chunk *chunk, reader *bufio.Reader, start int) error {
	if m.seekable {
		return m.reserveChunk(reader, start)
	}
	return m.readLines(chunk, reader, start, ChunkSize, true)
}

// reserveChunk reserves ChunkSize lines.
// read and update size only.
func (m *Document) reserveChunk(reader *bufio.Reader, start int) error {
	count, size, err := m.countLines(reader, start)
	m.mu.Lock()
	m.endNum += count
	m.size += int64(size)
	m.offset = m.size
	m.mu.Unlock()
	atomic.StoreInt32(&m.changed, 1)
	return err
}

// loadChunk actually loads the reserved Chunk.
func (m *Document) loadChunk(reader *bufio.Reader, chunkNum int) (*bufio.Reader, error) {
	chunk, err := m.seekChunk(reader, chunkNum)
	if err != nil {
		return nil, err
	}

	start := 0
	end := ChunkSize
	lastChunk, endNum := chunkLine(m.endNum)
	if chunkNum == lastChunk {
		end = endNum
	}
	// log.Println("fillChunk", chunkNum, start, end)

	if err := m.readLines(chunk, reader, start, end, false); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		reader = m.afterEOF(reader)
	}
	return reader, nil
}

// seekChunk seeks to the start of the chunk.
func (m *Document) seekChunk(reader *bufio.Reader, chunkNum int) (*chunk, error) {
	chunk := m.store.chunks[chunkNum]
	if _, err := m.file.Seek(chunk.start, io.SeekStart); err != nil {
		return chunk, fmt.Errorf("seek: %w", err)
	}
	reader.Reset(m.file)
	return chunk, nil
}

// searchRead searches chunks and loads chunks if found.
func (m *Document) searchRead(reader *bufio.Reader, chunkNum int, searcher Searcher) (*bufio.Reader, error) {
	if _, err := m.searchChunk(chunkNum, searcher); err != nil {
		if errors.Is(err, ErrNotFound) {
			return reader, nil
		}
		return reader, err
	}
	if err := m.evictChunksFile(chunkNum); err != nil {
		log.Println(err)
		return reader, nil
	}
	return m.loadChunk(reader, chunkNum)
}

// loadRead loads the read contents into chunks.
func (m *Document) loadRead(reader *bufio.Reader, chunkNum int) (*bufio.Reader, error) {
	if m.seekable {
		m.currentChunk = chunkNum
		return m.loadReadFile(reader, chunkNum)
	}
	return m.loadReadMem(reader, chunkNum)
}

// loadReadFile loads the read contents into chunks.
// loadReadFile frees old chunks and loads new chunks.
func (m *Document) loadReadFile(reader *bufio.Reader, chunkNum int) (*bufio.Reader, error) {
	_ = m.evictChunksFile(chunkNum)
	m.addChunksFile(chunkNum)
	if len(m.store.chunks[chunkNum].lines) != 0 {
		// already loaded.
		return reader, nil
	}
	return m.loadChunk(reader, chunkNum)
}

// loadReadMem loads the read contents into chunks.
// loadReadMem frees the memory behind and reads forward.
func (m *Document) loadReadMem(reader *bufio.Reader, chunkNum int) (*bufio.Reader, error) {
	if m.BufEOF() {
		return reader, nil
	}

	if chunkNum < m.lastChunkNum() {
		// already loaded.
		// return reader, fmt.Errorf("%w %d", ErrAlreadyLoaded, chunkNum)
		return reader, nil
	}
	m.evictChunksMem(chunkNum)
	m.requestContinue()
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
				return nil, fmt.Errorf("seek: %w", err)
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
		fd := os.Stdin.Fd()
		if term.IsTerminal(int(fd)) {
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

// readLines append lines read from reader into chunks.
// Read and fill the number of lines from start to end in chunk.
// If addLines is true, increment the number of lines read (update endNum).
func (m *Document) readLines(chunk *chunk, reader *bufio.Reader, start int, end int, addLines bool) error {
	var line bytes.Buffer
	var isPrefix bool
	for num := start; num < end; {
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
				m.append(chunk, addLines, line.Bytes())
				m.offset = m.size
				atomic.StoreInt32(&m.noNewlineEOF, 1)
			}
			return err
		}
		m.append(chunk, addLines, line.Bytes())
		m.offset = m.size
		line.Reset()
	}
	return nil
}

// countLines counts the number of lines and the size of the buffer.
func (m *Document) countLines(reader *bufio.Reader, start int) (int, int, error) {
	num := start
	count := 0
	size := 0
	buf := make([]byte, bufSize)
	for {
		bufLen, err := reader.Read(buf)
		if err != nil {
			return count, size, fmt.Errorf("read: %w", err)
		}
		if bufLen == 0 {
			return count, size, io.EOF
		}

		lSize := bufLen
		lCount := bytes.Count(buf[:bufLen], []byte("\n"))
		// If it exceeds ChunkSize, Re-aggregate size and count.
		if num+lCount > ChunkSize {
			lSize = 0
			lCount = ChunkSize - num
			for i := 0; i < lCount; i++ {
				p := bytes.IndexByte(buf[lSize:bufLen], '\n')
				lSize += p + 1
			}
		}

		num += lCount
		count += lCount
		size += lSize
		if num >= ChunkSize {
			// no newline at the end of the file.
			if bufLen < bufSize {
				p := bytes.LastIndex(buf[:bufLen], []byte("\n"))
				size -= bufLen - p - 1
			}
			break
		}
		// no newline at the end of the file.
		if bufLen < bufSize {
			p := bytes.LastIndex(buf[:bufLen], []byte("\n"))
			if p+1 < bufLen {
				count++
				atomic.StoreInt32(&m.noNewlineEOF, 1)
			}
		}
	}
	return count, size, nil
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

// joinLast joins the new content to the last line.
// This is used when the last line is added without a newline and EOF.
func (m *Document) joinLast(chunk *chunk, line []byte) bool {
	if len(chunk.lines) == 0 {
		return false
	}
	size := len(line)

	m.mu.Lock()
	num := len(chunk.lines) - 1
	buf := chunk.lines[num]
	dst := make([]byte, 0, len(buf)+size)
	dst = append(dst, buf...)
	dst = append(dst, line...)
	m.size += int64(size)
	chunk.lines[num] = dst
	m.mu.Unlock()

	if line[len(line)-1] == '\n' {
		atomic.StoreInt32(&m.noNewlineEOF, 0)
	}
	m.cache.Remove(m.endNum - 1)
	return true
}

// append appends a line to the chunk.
func (m *Document) append(chunk *chunk, isCount bool, line []byte) {
	if isCount {
		m.appendLine(chunk, line)
		return
	}

	m.appendOnly(chunk, line)
}

// appendLine appends to the line of the chunk.
// appendLine updates the number of lines and size.
func (m *Document) appendLine(chunk *chunk, line []byte) {
	if atomic.SwapInt32(&m.noNewlineEOF, 0) == 1 {
		m.joinLast(chunk, line)
		return
	}
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
		m.appendLine(chunk, []byte(feed))
	}
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
	m.store.chunks = []*chunk{
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

// readAll reads to the end.
// The read lines are stored in the lines of the Document.
func (m *Document) readAll(reader *bufio.Reader) error {
	chunk := m.chunkForAdd()
	start := len(chunk.lines)
	for {
		if err := m.readLines(chunk, reader, start, ChunkSize, true); err != nil {
			return err
		}
		chunk = NewChunk(m.size)
		m.mu.Lock()
		m.store.chunks = append(m.store.chunks, chunk)
		m.mu.Unlock()
		start = 0
	}
}
