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

// bufSize is the size of the buffer used when reading the file.
// This bufSize is used when only counting.
const bufSize = 4096

// tailSize represents the position to start reading backwards from the end position.
const tailSize = 10000

// FormFeed is the delimiter that separates the sections.
// The default delimiter that separates single output from watch.
const FormFeed = "\f"

// firstRead first reads the file.
// Fill the contents of the read file into the first chunk.
func (m *Document) firstRead(reader *bufio.Reader) (*bufio.Reader, error) {
	atomic.StoreInt32(&m.store.noNewlineEOF, 0)
	chunk := m.store.chunks[0]
	if err := m.store.readLines(chunk, reader, 0, ChunkSize, true); err != nil {
		if errors.Is(err, io.EOF) {
			return m.afterEOF(reader), nil
		}
		atomic.StoreInt32(&m.store.eof, 1)
		return nil, err
	}

	m.requestContinue()
	return reader, nil
}

// tmpRead read tail to temporary store.
// It is executed only once if EOF has not been reached after follow-mode is set.
func (m *Document) tmpRead(reader *bufio.Reader) (*bufio.Reader, error) {
	m.followStore = NewStore()
	atomic.StoreInt32(&m.tmpFollow, 1)

	if _, err := m.file.Seek(tailSize*-1, io.SeekEnd); err != nil {
		return reader, fmt.Errorf("tmpFollowRead seek: %w", err)
	}
	reader.Reset(m.file)
	chunk := m.followStore.chunks[0]
	if err := m.followStore.readLines(chunk, reader, 0, ChunkSize, true); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	atomic.StoreInt32(&m.followStore.eof, 1)
	m.requestContinue()
	return reader, nil
}

// continueRead is executed after the second
// and only reads the file or counts the lines of the file.
func (m *Document) continueRead(reader *bufio.Reader) (*bufio.Reader, error) {
	if m.seekable {
		if err := m.seekChunk(reader, m.store.offset); err != nil {
			atomic.StoreInt32(&m.store.eof, 1)
			log.Printf("continueRead: %s", err)
			m.seekable = false
		} else {
			reader.Reset(m.file)
		}
	}
	chunk := m.store.chunkForAdd(m.seekable, m.store.size)
	start := len(chunk.lines)
	if err := m.addOrReserveChunk(chunk, reader, start, ChunkSize); err != nil {
		if errors.Is(err, io.EOF) {
			return m.afterEOF(reader), nil
		}
		return nil, fmt.Errorf("addChunk: %w", err)
	}

	m.requestContinue()
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

	reader, err := m.loadRead(reader, m.store.lastChunkNum())
	if err != nil {
		return reader, err
	}

	atomic.StoreInt32(&m.store.eof, 0)
	chunk := m.store.chunks[m.store.lastChunkNum()]
	start := len(chunk.lines) - 1
	if atomic.LoadInt32(&m.store.noNewlineEOF) == 0 {
		chunk = m.store.chunkForAdd(m.seekable, m.store.size)
		start = len(chunk.lines)
	}
	if m.seekable {
		if err := m.seekChunk(reader, m.store.offset); err != nil {
			return nil, fmt.Errorf("followRead: %w", err)
		}
		reader = bufio.NewReader(m.file)
	}

	if err := m.store.readLines(chunk, reader, start, ChunkSize, true); err != nil {
		if errors.Is(err, io.EOF) {
			return m.afterEOF(reader), nil
		}
		return nil, err
	}

	m.requestContinue()
	return reader, nil
}

// addOrReserveChunk reads a file to add or reserve a Chunk.
// If it's a seekable file, it just reserves it for later reading.
func (m *Document) addOrReserveChunk(chunk *chunk, reader *bufio.Reader, start int, end int) error {
	if m.seekable {
		return m.reserveChunk(reader, start, ChunkSize)
	}
	return m.store.readLines(chunk, reader, start, ChunkSize, true)
}

// reserveChunk reserves ChunkSize lines.
// read and update size only.
func (m *Document) reserveChunk(reader *bufio.Reader, start int, end int) error {
	count, size, err := m.store.countLines(reader, start, end)
	m.store.mu.Lock()
	m.store.size += int64(size)
	m.store.offset = m.store.size
	m.store.mu.Unlock()
	atomic.AddInt32(&m.store.endNum, int32(count))
	atomic.StoreInt32(&m.store.changed, 1)
	return err
}

// loadChunk actually loads the reserved Chunk.
func (m *Document) loadChunk(reader *bufio.Reader, chunkNum int) (*bufio.Reader, error) {
	chunk := m.store.chunks[chunkNum]
	if err := m.seekChunk(reader, chunk.start); err != nil {
		return nil, err
	}

	start, end := m.store.chunkRange(chunkNum)
	if err := m.store.readLines(chunk, reader, start, end, false); err != nil {
		log.Printf("Failed to read the expected number of lines(%d:%d): %s", start, end, err)
		if errors.Is(err, io.EOF) {
			return m.afterEOF(reader), nil
		}
		return nil, err
	}
	return reader, nil
}

// seekChunk seeks to the start of the chunk.
func (m *Document) seekChunk(reader *bufio.Reader, start int64) error {
	if _, err := m.file.Seek(start, io.SeekStart); err != nil {
		return fmt.Errorf("seek: %w", err)
	}
	reader.Reset(m.file)
	return nil
}

// searchRead searches chunks and loads chunks if found.
func (m *Document) searchRead(reader *bufio.Reader, chunkNum int, searcher Searcher) (*bufio.Reader, error) {
	if _, err := m.searchChunk(chunkNum, searcher); err != nil {
		return reader, err
	}
	return m.loadReadFile(reader, chunkNum)
}

// loadRead loads the read contents into chunks.
func (m *Document) loadRead(reader *bufio.Reader, chunkNum int) (*bufio.Reader, error) {
	if m.seekable {
		return m.loadReadFile(reader, chunkNum)
	}
	return m.loadReadMem(reader, chunkNum)
}

// loadReadFile loads the read contents into chunks.
// loadReadFile frees old chunks and loads new chunks.
func (m *Document) loadReadFile(reader *bufio.Reader, chunkNum int) (*bufio.Reader, error) {
	m.store.swapLoadedFile(chunkNum)
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

	if chunkNum < m.store.lastChunkNum() {
		// already loaded.
		// return reader, fmt.Errorf("%w %d", ErrAlreadyLoaded, chunkNum)
		return reader, nil
	}
	m.store.evictChunksMem(chunkNum)
	m.requestContinue()
	return reader, nil
}

// reloadRead performs reload processing.
func (m *Document) reloadRead(reader *bufio.Reader) (*bufio.Reader, error) {
	if !m.WatchMode {
		m.reset()
	} else {
		m.seekable = false
		chunk := m.store.chunkForAdd(m.seekable, m.store.size)
		m.store.appendFormFeed(chunk)
	}

	return m.reloadFile(reader)
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
	atomic.StoreInt32(&m.store.eof, 0)
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
	m.store.offset = m.store.size
	atomic.StoreInt32(&m.store.eof, 1)
	if atomic.SwapInt32(&m.tmpFollow, 0) == 1 {
		m.cache.Purge()
	}
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
	return m.fileReader(f)
}

// fileReader returns a io.Reader.
func (m *Document) fileReader(f *os.File) (io.Reader, error) {
	m.store.mu.Lock()
	defer m.store.mu.Unlock()

	atomic.StoreInt32(&m.closed, 0)
	m.file = f
	cFormat, r := uncompressedReader(m.file, m.seekable)
	if cFormat == UNCOMPRESSED {
		if m.seekable {
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				atomic.StoreInt32(&m.closed, 1)
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

	atomic.StoreInt32(&m.store.readCancel, 1)
	m.requestReload()
	atomic.StoreInt32(&m.store.readCancel, 0)
	if !m.WatchMode {
		m.topLN = 0
	}

	return nil
}

// reset clears all lines.
func (m *Document) reset() {
	m.store = NewStore()
	m.store.setNewLoadChunks(m.seekable)
	atomic.StoreInt32(&m.store.changed, 1)
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

	if err := m.file.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	atomic.StoreInt32(&m.store.eof, 1)
	atomic.StoreInt32(&m.closed, 1)
	atomic.StoreInt32(&m.store.changed, 1)
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
				m.store.offset = m.store.size
				atomic.StoreInt32(&m.store.eof, 1)
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
	chunk := m.store.chunkForAdd(m.seekable, m.store.size)
	start := len(chunk.lines)
	for {
		if err := m.store.readLines(chunk, reader, start, ChunkSize, true); err != nil {
			return err
		}
		chunk = NewChunk(m.store.size)
		m.store.mu.Lock()
		m.store.chunks = append(m.store.chunks, chunk)
		m.store.mu.Unlock()
		start = 0
	}
}
