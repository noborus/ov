package oviewer

import (
	"log"
)

// NewLogDoc generates a document for log.
func NewLogDoc() (*Document, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}
	m.FollowMode = true
	m.FileName = "Log"
	log.SetOutput(m)
	return m, nil
}

// Write matches the interface of io.Writer.
// Therefore, the log.Print output is displayed by logDoc.
func (m *Document) Write(p []byte) (int, error) {
	str := string(p)
	m.append(str)
	return len(str), nil
}
