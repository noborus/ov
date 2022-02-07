package oviewer

import (
	"log"
	"strings"
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
	str := strings.TrimSuffix(string(p), "\n")
	m.append(str)
	return len(str), nil
}
