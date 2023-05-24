package oviewer

import (
	"fmt"
	"log"
	"sync/atomic"

	lru "github.com/hashicorp/golang-lru/v2"
)

// ChunkSize is the unit of number of lines to split the file.
var ChunkSize = 10000

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

	chunks, err := lru.New[int, struct{}](capacity)
	if err != nil {
		log.Panicf("lru new %s", err)
	}
	s.loadedChunks = chunks
}

// addChunksFile adds chunks of a regular file to memory.
func (s *store) addChunksFile(chunkNum int) {
	if chunkNum == 0 {
		return
	}
	if s.loadedChunks.Add(chunkNum, struct{}{}) {
		log.Println("AddChunksFile evicted!")
	}
}

// evictChunksFile evicts chunks of a regular file from memory.
func (s *store) evictChunksFile(chunkNum int) error {
	if chunkNum == 0 {
		return nil
	}
	if s.loadedChunks.Len() >= MemoryLimitFile {
		k, _, _ := s.loadedChunks.GetOldest()
		if chunkNum != k {
			s.unloadChunk(k)
		}
	}

	chunk := s.chunks[chunkNum]
	if len(chunk.lines) != 0 {
		return fmt.Errorf("%w %d", ErrAlreadyLoaded, chunkNum)
	}
	return nil
}

// addChunksMem adds non-regular file chunks to memory.
func (s *store) addChunksMem(chunkNum int) {
	if MemoryLimit < 0 {
		return
	}
	if chunkNum == 0 {
		return
	}
	if _, _, evicted := s.loadedChunks.PeekOrAdd(chunkNum, struct{}{}); evicted {
		log.Println("AddChunksMem evicted!")
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

// chunkForAdd is a helper function to get the chunk to add.
func (s *store) chunkForAdd(isFile bool, start int64) *chunk {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.endNum < len(s.chunks)*ChunkSize {
		return s.chunks[len(s.chunks)-1]
	}
	chunk := NewChunk(start)
	s.chunks = append(s.chunks, chunk)
	if !isFile {
		s.addChunksMem(len(s.chunks) - 1)
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

// chunkLine returns chunkNum and chunk line number from line number.
func chunkLine(n int) (int, int) {
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
