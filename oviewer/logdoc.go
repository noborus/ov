package oviewer

import (
	"log"
	"sync/atomic"
)

type LogDocument struct {
	*Document
	channel chan []byte
}

// NewLogDoc generates a document for log.
// NewLogDoc makes LogDoc the output destination of go's standard logger.
func NewLogDoc() (*LogDocument, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}
	m.documentType = DocLog
	m.FollowMode = true
	m.Caption = "Log"
	m.seekable = false
	atomic.StoreInt32(&m.closed, 1)
	if err := m.ControlLog(); err != nil {
		return nil, err
	}
	logDoc := &LogDocument{
		Document: m,
		channel:  make(chan []byte, 1),
	}
	go logDoc.chanWrite()
	log.SetOutput(logDoc)
	return logDoc, nil
}

// Write matches the interface of io.Writer(so package log is possible).
// Therefore, the log.Print output is displayed by logDoc.
func (logDoc *LogDocument) Write(p []byte) (int, error) {
	line := make([]byte, len(p))
	copy(line, p)
	logDoc.channel <- line
	return len(line), nil
}

// chanWrite writes the log output to the logDoc.
func (logDoc *LogDocument) chanWrite() {
	for p := range logDoc.channel {
		s := logDoc.store
		chunk := s.chunkForAdd(false, s.size)
		s.append(chunk, true, p)
		if len(chunk.lines) < ChunkSize {
			continue
		}
		chunk = NewChunk(s.size)

		s.mu.Lock()
		if len(s.chunks) > 2 {
			s.chunks[len(s.chunks)-2].lines = nil
			atomic.StoreInt32(&s.startNum, int32(ChunkSize*(len(s.chunks)-1)))
		}
		s.chunks = append(s.chunks, chunk)
		s.mu.Unlock()
	}
}
