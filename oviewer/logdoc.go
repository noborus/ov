package oviewer

import (
	"log"
	"sync/atomic"
)

// NewLogDoc generates a document for log.
// NewLogDoc makes LogDoc the output destination of go's standard logger.
func NewLogDoc() (*Document, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}
	m.FollowMode = true
	m.Caption = "Log"
	m.seekable = false
	atomic.StoreInt32(&m.closed, 1)
	if err := m.ControlLog(); err != nil {
		return nil, err
	}
	log.SetOutput(m)
	return m, nil
}

// Write matches the interface of io.Writer(so package log is possible).
// Therefore, the log.Print output is displayed by logDoc.
func (m *Document) Write(p []byte) (int, error) {
	s := m.store
	chunk := s.chunkForAdd(false, s.size)
	s.append(chunk, true, p)
	if len(chunk.lines) < ChunkSize {
		return len(p), nil
	}
	chunk = NewChunk(s.size)
	s.mu.Lock()
	if len(s.chunks) > 2 {
		s.chunks[len(s.chunks)-2].lines = nil
		atomic.StoreInt32(&s.startNum, int32(ChunkSize*(len(s.chunks)-1)))
	}
	s.chunks = append(s.chunks, chunk)
	s.mu.Unlock()
	return len(p), nil
}
