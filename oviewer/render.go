package oviewer

import (
	"io"

	"github.com/noborus/ov/biomap"
)

type renderDocument struct {
	*Document
	writer io.WriteCloser
}

// renderDoc returns a new document with the reader.
// The reader is rendered and returned as a new document.
func renderDoc(parent *Document, reader io.Reader) (*renderDocument, error) {
	doc, err := NewDocument()
	if err != nil {
		return nil, err
	}
	doc.parent = parent
	doc.lineNumMap = biomap.NewMap[int, int]()
	doc.preventReload = true
	doc.seekable = false
	if err := doc.ControlReader(reader, nil); err != nil {
		return nil, err
	}
	return &renderDocument{Document: doc}, nil
}

func (render *renderDocument) writeLine(line []byte) {
	if _, err := render.writer.Write(line); err != nil {
		panic(err)
	}
	if _, err := render.writer.Write([]byte("\n")); err != nil {
		panic(err)
	}
}
