package oviewer

import (
	"fmt"
	"log"

	lru "github.com/hashicorp/golang-lru/v2"
)

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

// addChunksMem adds non-regular file chunks to memory.
func (s *store) addChunksMem(chunkNum int) {
	if chunkNum == 0 {
		return
	}
	s.loadedChunks.PeekOrAdd(chunkNum, struct{}{})
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

// evictChunksMem evicts non-regular file chunks from memory.
// Change the start position after unloading.
func (m *Document) evictChunksMem(chunkNum int) {
	if chunkNum == 0 {
		return
	}
	if (MemoryLimit < 0) || (m.store.loadedChunks.Len() < MemoryLimit) {
		return
	}
	k, _, _ := m.store.loadedChunks.GetOldest()
	m.store.unloadChunk(k)
	m.store.mu.Lock()
	m.store.startNum = (k + 1) * ChunkSize
	m.store.mu.Unlock()
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
func (m *Document) chunkForAdd(isFile bool) *chunk {
	m.store.mu.Lock()
	defer m.store.mu.Unlock()

	if m.store.endNum < len(m.store.chunks)*ChunkSize {
		return m.store.chunks[len(m.store.chunks)-1]
	}
	chunk := NewChunk(m.size)
	m.store.chunks = append(m.store.chunks, chunk)
	if !isFile {
		m.store.addChunksMem(len(m.store.chunks) - 1)
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
