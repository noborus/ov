package oviewer

import "log"

// NewLogDoc generates a document for log.
func NewLogDoc() (*Document, error) {
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
func (logDoc *Document) Write(p []byte) (int, error) {
	str := string(p)

	logDoc.mu.Lock()
	logDoc.lines = append(logDoc.lines, str)
	logDoc.endNum = len(logDoc.lines)
	logDoc.mu.Unlock()

	return len(str), nil
}
