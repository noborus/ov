package oviewer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

// ChunkSize is the unit of number of lines to split the file.
var ChunkSize = 10000

// NewStore returns store.
func NewStore() *store {
	return &store{
		chunks: []*chunk{
			NewChunk(0),
		},
	}
}

// NewChunk returns chunk.
func NewChunk(start int64) *chunk {
	return &chunk{
		lines: make([][]byte, 0, ChunkSize),
		start: start,
	}
}

// setNewLoadChunks creates a new LRU cache.
// Manage chunks loaded in LRU cache.
func (s *store) setNewLoadChunks(capacity int) {
	if capacity <= 0 {
		capacity = 1
	}
	loaded, err := lru.New[int, struct{}](capacity)
	if err != nil {
		log.Panicf("lru new %s", err)
	}
	s.loadedChunks = loaded
}

// loadChunksCapacity creates a new LRU cache.
func loadChunksCapacity(isFile bool) int {
	mlMem := MemoryLimit
	mlFile := MemoryLimitFile
	if mlMem >= 0 {
		if mlMem < 2 {
			mlMem = 2
		}
		mlFile = mlMem
	}

	if mlFile < 2 {
		mlFile = 2
	}

	capacity := mlFile + 1
	if !isFile {
		if mlMem > 0 {
			capacity = mlMem + 1
		}
	}
	return capacity
}

// swapLoadedFile swaps loaded chunks.
// Unload unnecessary Chunks and load new Chunks.
// Already loaded Chunk tells new.
func (s *store) swapLoadedFile(chunkNum int) {
	if chunkNum == 0 {
		return
	}
	if s.loadedChunks.Len() >= MemoryLimitFile {
		k, _, _ := s.loadedChunks.GetOldest()
		if chunkNum != k {
			s.unloadChunk(k)
		}
	}
	if s.loadedChunks.Add(chunkNum, struct{}{}) {
		log.Println("loadChunksFile evicted!")
	}
}

// loadChunksMem adds non-regular file chunks to memory.
func (s *store) loadChunksMem(chunkNum int) {
	if MemoryLimit < 0 {
		return
	}
	if chunkNum == 0 {
		return
	}
	if _, _, evicted := s.loadedChunks.PeekOrAdd(chunkNum, struct{}{}); evicted {
		log.Println("loadChunksMem evicted!")
	}
}

// evictChunksMem evicts non-regular file chunks from memory.
// Change the start position after unloading.
func (s *store) evictChunksMem(chunkNum int) {
	if MemoryLimit < 0 {
		return
	}
	if chunkNum == 0 {
		return
	}
	if s.loadedChunks.Len() < MemoryLimit {
		return
	}
	k, _, _ := s.loadedChunks.GetOldest()
	s.unloadChunk(k)
	atomic.StoreInt32(&s.startNum, int32((k+1)*ChunkSize))
}

// unloadChunk unloads the chunk from memory.
func (s *store) unloadChunk(chunkNum int) {
	s.loadedChunks.Remove(chunkNum)
	s.chunks[chunkNum].lines = nil
}

// lastChunkNum returns the last chunk number.
func (s *store) lastChunkNum() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.chunks) - 1
}

// chunkForAdd returns a Chunk for appending lines.
// Returns a new Chunk if the Chunk is full.
func (s *store) chunkForAdd(isFile bool, start int64) *chunk {
	s.mu.Lock()
	defer s.mu.Unlock()
	endNum := int(atomic.LoadInt32(&s.endNum))
	if endNum < len(s.chunks)*ChunkSize {
		return s.chunks[len(s.chunks)-1]
	}

	chunk := NewChunk(start)
	s.chunks = append(s.chunks, chunk)
	if !isFile {
		s.loadChunksMem(len(s.chunks) - 1)
	}
	return chunk
}

// isLoadedChunk returns true if the chunk is loaded in memory.
func (s *store) isLoadedChunk(chunkNum int, isFile bool) bool {
	if chunkNum == 0 {
		return true
	}
	if !isFile {
		return true
	}
	return s.loadedChunks.Contains(chunkNum)
}

// isContinueRead returns whether to continue reading.
func (s *store) isContinueRead(limit int) bool {
	if limit < 0 {
		return true
	}
	if s.loadedChunks.Len() < limit {
		return true
	}
	return false
}

// chunkRange returns the start and end line numbers of the chunk.
func (s *store) chunkRange(chunkNum int) (int, int) {
	start := 0
	end := ChunkSize
	endNum := int(atomic.LoadInt32(&s.endNum))
	lastChunk, chunkEndNum := chunkLineNum(endNum)
	if chunkNum == lastChunk {
		end = chunkEndNum
	}
	return start, end
}

// chunkLineNum returns chunkNum and chunk line number from line number.
func chunkLineNum(n int) (int, int) {
	chunkNum := n / ChunkSize
	cn := n % ChunkSize
	return chunkNum, cn
}

// readLines append lines read from reader into chunks.
// Read and fill the number of lines from start to end in chunk.
// If addLines is true, increment the number of lines read (update endNum).
func (s *store) readLines(chunk *chunk, reader *bufio.Reader, start int, end int, updateNum bool) error {
	var line bytes.Buffer
	var isPrefix bool
	for num := start; num < end; {
		if atomic.LoadInt32(&s.readCancel) == 1 {
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
		atomic.StoreInt32(&s.changed, 1)
		if err != nil {
			if line.Len() != 0 {
				s.append(chunk, updateNum, line.Bytes())
				atomic.StoreInt32(&s.noNewlineEOF, 1)
			}
			return err
		}
		s.append(chunk, updateNum, line.Bytes())
		line.Reset()
	}
	return nil
}

// countLines counts the number of lines and the size of the buffer.
func (s *store) countLines(reader *bufio.Reader, start int, end int) (int, int, error) {
	count := 0
	size := 0
	buf := make([]byte, bufSize)
	for num := start; num < end; {
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
			for range lCount {
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
				atomic.StoreInt32(&s.noNewlineEOF, 1)
			}
		}
	}
	return count, size, nil
}

// append appends a line to the chunk.
func (s *store) append(chunk *chunk, updateNum bool, line []byte) {
	if updateNum {
		s.appendLine(chunk, line)
	} else {
		s.appendOnly(chunk, line)
	}
	s.offset = s.size
}

// appendOnly appends to the line of the chunk.
// appendOnly does not updates the number of lines and size.
func (s *store) appendOnly(chunk *chunk, line []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	size := len(line)
	dst := make([]byte, size)
	copy(dst, line)
	chunk.lines = append(chunk.lines, dst)
}

// appendLine appends to the line of the chunk.
// appendLine updates the number of lines and size.
func (s *store) appendLine(chunk *chunk, line []byte) {
	if atomic.SwapInt32(&s.noNewlineEOF, 0) == 1 {
		s.joinLast(chunk, line)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	size := len(line)
	s.size += int64(size)
	atomic.AddInt32(&s.endNum, 1)
	dst := make([]byte, size)
	copy(dst, line)
	chunk.lines = append(chunk.lines, dst)
}

// joinLast joins the new content to the last line.
// This is used when the last line is added without a newline and EOF.
func (s *store) joinLast(chunk *chunk, line []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	size := len(line)
	if size == 0 {
		return false
	}

	num := len(chunk.lines) - 1
	buf := chunk.lines[num]
	dst := make([]byte, 0, len(buf)+size)
	dst = append(dst, buf...)
	dst = append(dst, line...)
	s.size += int64(size)
	chunk.lines[num] = dst

	if line[len(line)-1] == '\n' {
		atomic.StoreInt32(&s.noNewlineEOF, 0)
	}
	return true
}

// appendFormFeed appends a formfeed to the chunk.
func (s *store) appendFormFeed(chunk *chunk) {
	line := ""
	s.mu.Lock()
	if len(chunk.lines) > 0 {
		line = string(chunk.lines[len(chunk.lines)-1])
	}
	s.mu.Unlock()
	// Do not add if the previous is FormFeed(always add for formfeedTime).
	if line != FormFeed {
		feed := FormFeed
		if s.formfeedTime {
			feed = fmt.Sprintf("%sTime: %s", FormFeed, time.Now().Format(time.RFC3339))
		}
		s.appendLine(chunk, []byte(feed))
	}
}

func (s *store) export(w io.Writer, chunk *chunk, start int, end int) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start = max(0, start)
	end = min(len(chunk.lines), end)
	for i := start; i < end; i++ {
		if _, err := w.Write(chunk.lines[i]); err != nil {
			return err
		}
	}
	return nil
}

// exportPlain exports the plain text without ANSI escape sequences.
func (s *store) exportPlain(w io.Writer, chunk *chunk, start int, end int) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start = max(0, start)
	end = min(len(chunk.lines), end)
	for i := start; i < end; i++ {
		plain := stripEscapeSequenceBytes(chunk.lines[i])
		if _, err := w.Write(plain); err != nil {
			return err
		}
	}
	return nil
}
