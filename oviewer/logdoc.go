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
	chunk := m.chunkForAdd()
	m.appendLine(chunk, p)
	if len(chunk.lines) >= ChunkSize {
		chunk = NewChunk(m.size)
		m.mu.Lock()
		if len(m.chunks) > 2 {
			m.chunks[len(m.chunks)-2].lines = nil
			m.startNum = ChunkSize * (len(m.chunks) - 1)
		}
		m.chunks = append(m.chunks, chunk)
		m.mu.Unlock()
	}
	return len(p), nil
}
