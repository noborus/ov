package oviewer

import (
	"io"

	"github.com/noborus/ov/biomap"
)

// renderDoc returns a new document with the reader.
// The reader is rendered and returned as a new document.
func renderDoc(parent *Document, reader io.Reader) (*Document, error) {
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
	return doc, nil
}
