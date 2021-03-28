package oviewer

import (
	"log"
)

type LogDocument = Document

// NewLogDoc generates a document for log.
func NewLogDoc() (*LogDocument, error) {
	logDoc, err := NewDocument()
	if err != nil {
		return nil, err
	}
	logDoc.FollowMode = true
	logDoc.FileName = "Log"
	log.SetOutput(logDoc)
	return logDoc, nil
}

// Write matches the interface of io.Writer.
// Therefore, the log.Print output is displayed by logDoc.
func (logDoc *LogDocument) Write(p []byte) (int, error) {
	str := string(p)
	logDoc.append(str)
	return len(str), nil
}
