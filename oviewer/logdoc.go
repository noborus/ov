package oviewer

import (
	"log"
)

// NewLogDoc generates a document for log.
// NewLogDoc makes LogDoc the output destination of go's standard logger.
func NewLogDoc() (*Document, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}
	m.FollowMode = true
	m.FileName = "Log"
	m.seekable = false
	log.SetOutput(m)
	return m, nil
}

// Write matches the interface of io.Writer(so package log is possible).
// Therefore, the log.Print output is displayed by logDoc.
func (m *Document) Write(p []byte) (int, error) {
	chunk := m.lastChunk()
	m.append(chunk, string(p))
	if len(chunk.lines) >= ChunkSize {
		chunk = NewChunk(m.size)
		m.mu.Lock()
		m.chunks = append(m.chunks, chunk)
		m.mu.Unlock()
	}
	return len(p), nil
}
