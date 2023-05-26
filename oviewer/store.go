package oviewer

import (
	"log"
	"sync/atomic"

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
func (s *store) setNewLoadChunks(isFile bool) {
	capacity := newLoadChunks(isFile)
	loaded, err := lru.New[int, struct{}](capacity)
	if err != nil {
		log.Panicf("lru new %s", err)
	}
	s.loadedChunks = loaded
}

// newLoadChunks creates a new LRU cache.
func newLoadChunks(isFile bool) int {
	mlFile := MemoryLimitFile
	if MemoryLimit >= 0 {
		if MemoryLimit < 2 {
			MemoryLimit = 2
		}
		MemoryLimitFile = MemoryLimit
	}

	if MemoryLimitFile > mlFile {
		MemoryLimitFile = mlFile
	}
	if MemoryLimitFile < 2 {
		MemoryLimitFile = 2
	}

	capacity := MemoryLimitFile + 1
	if !isFile {
		if MemoryLimit > 0 {
			capacity = MemoryLimit + 1
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
	s.mu.Lock()
	s.startNum = (k + 1) * ChunkSize
	s.mu.Unlock()
}

// unloadChunk unloads the chunk from memory.
func (s *store) unloadChunk(chunkNum int) {
	s.loadedChunks.Remove(chunkNum)
	s.chunks[chunkNum].lines = nil
}

// lastChunkNum returns the last chunk number.
func (s *store) lastChunkNum() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.chunks) - 1
}

// chunkForAdd returns a Chunk for appending lines.
// Returns a new Chunk if the Chunk is full.
func (s *store) chunkForAdd(isFile bool, start int64) *chunk {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.endNum < len(s.chunks)*ChunkSize {
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
func (s *store) isContinueRead(isFile bool) bool {
	if isFile {
		return true
	}
	if MemoryLimit < 0 {
		return true
	}
	if s.loadedChunks.Len() < MemoryLimit {
		return true
	}
	return false
}

// chunkRange returns the start and end line numbers of the chunk.
func (s *store) chunkRange(chunkNum int) (int, int) {
	start := 0
	end := ChunkSize
	lastChunk, endNum := chunkLineNum(s.endNum)
	if chunkNum == lastChunk {
		end = endNum
	}
	return start, end
}

// chunkLineNum returns chunkNum and chunk line number from line number.
func chunkLineNum(n int) (int, int) {
	chunkNum := n / ChunkSize
	cn := n % ChunkSize
	return chunkNum, cn
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
	s.endNum++
	dst := make([]byte, size)
	copy(dst, line)
	chunk.lines = append(chunk.lines, dst)
}

// joinLast joins the new content to the last line.
// This is used when the last line is added without a newline and EOF.
func (s *store) joinLast(chunk *chunk, line []byte) bool {
	size := len(line)
	if size == 0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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
