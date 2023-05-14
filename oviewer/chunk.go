package oviewer

import (
	"fmt"
	"log"
	"sync/atomic"

	lru "github.com/hashicorp/golang-lru/v2"
)

// setNewLoadChunks creates a new LRU cache.
// Manage chunks loaded in LRU cache.
func (m *Document) setNewLoadChunks() {
	mlFile := MemoryLimitFile
	if MemoryLimit >= 0 {
		if MemoryLimit < 2 {
			MemoryLimit = 2
		}
		mlFile = MemoryLimit
		MemoryLimitFile = MemoryLimit
	}

	if MemoryLimitFile > mlFile {
		MemoryLimitFile = mlFile
	}
	if MemoryLimitFile < 2 {
		MemoryLimitFile = 2
	}

	capacity := MemoryLimitFile + 1
	if !m.seekable {
		if MemoryLimit > 0 {
			capacity = MemoryLimit + 1
		}
	}

	chunks, err := lru.New[int, struct{}](capacity)
	if err != nil {
		log.Panicf("lru new %s", err)
	}
	m.loadedChunks = chunks
}

// addChunksFile adds chunks of a regular file to memory.
func (m *Document) addChunksFile(chunkNum int) {
	if chunkNum == 0 {
		return
	}
	if m.loadedChunks.Add(chunkNum, struct{}{}) {
		log.Println("AddChunksFile evicted!")
	}
}

// addChunksMem adds non-regular file chunks to memory.
func (m *Document) addChunksMem(chunkNum int) {
	if chunkNum == 0 {
		return
	}
	m.loadedChunks.PeekOrAdd(chunkNum, struct{}{})
}

// evictChunksFile evicts chunks of a regular file from memory.
func (m *Document) evictChunksFile(chunkNum int) error {
	if chunkNum == 0 {
		return nil
	}
	if m.loadedChunks.Len() >= MemoryLimitFile {
		k, _, _ := m.loadedChunks.GetOldest()
		if chunkNum != k {
			m.unloadChunk(k)
		}
	}

	chunk := m.chunks[chunkNum]
	if len(chunk.lines) != 0 || atomic.LoadInt32(&m.closed) != 0 {
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
	if (MemoryLimit < 0) || (m.loadedChunks.Len() < MemoryLimit) {
		return
	}
	k, _, _ := m.loadedChunks.GetOldest()
	m.unloadChunk(k)
	m.mu.Lock()
	m.startNum = (k + 1) * ChunkSize
	m.mu.Unlock()
}

// unloadChunk unloads the chunk from memory.
func (m *Document) unloadChunk(chunkNum int) {
	m.loadedChunks.Remove(chunkNum)
	m.chunks[chunkNum].lines = nil
}

// lastChunkNum returns the last chunk number.
func (m *Document) lastChunkNum() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.chunks) - 1
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
	if !m.seekable {
		m.addChunksMem(len(m.chunks) - 1)
	}
	return chunk
}

// isLoadedChunk returns true if the chunk is loaded in memory.
func (m *Document) isLoadedChunk(chunkNum int) bool {
	if chunkNum == 0 {
		return true
	}
	if !m.seekable {
		return true
	}
	return m.loadedChunks.Contains(chunkNum)
}
