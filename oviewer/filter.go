package oviewer

import (
	"context"
	"io"
	"log"
	"strings"
)

// filterDocument is a document for filtering.
type filterDocument struct {
	*Document
	w io.WriteCloser
}

// Filter fires the filter event.
func (root *Root) Filter(str string, nonMatch bool) {
	root.Doc.nonMatch = nonMatch
	root.input.value = str
	ev := &eventInputSearch{
		searchType: filter,
		value:      str,
	}
	root.postEvent(ev)
}

// filter filters the document by the input value.
func (root *Root) filter(ctx context.Context, str string) {
	searcher := root.setSearcher(str, root.Config.CaseSensitive)
	if searcher == nil {
		return
	}
	root.filterDocument(ctx, searcher)
}

// filterDocument filters the document by the searcher.
// It creates a new document and writes the filtered lines to it.
func (root *Root) filterDocument(ctx context.Context, searcher Searcher) {
	m := root.Doc
	r, w := io.Pipe()
	render, err := renderDoc(m, r)
	if err != nil {
		log.Println(err)
		return
	}
	render.documentType = DocFilter
	match := searcher.String()
	if m.nonMatch {
		match = "!" + match
	}
	render.Caption = "filter:" + match
	msg := "search:" + match
	root.insertDocument(ctx, root.CurrentDoc, render)
	render.RunTimeSettings = updateRunTimeSettings(render.RunTimeSettings, root.Config.General)
	render.regexpCompile()

	render.nonMatch = m.nonMatch
	render.Header = m.Header
	render.SkipLines = m.SkipLines
	filterDoc := &filterDocument{
		Document: render,
		w:        w,
	}

	// Copy the header
	for ln := 0; ln < render.firstLine(); ln++ {
		line, err := m.Line(ln)
		if err != nil {
			log.Println(err)
			break
		}
		render.lineNumMap.Store(ln, ln)
		writeLine(w, line)
	}
	go m.filterWriter(ctx, searcher, m.firstLine(), filterDoc)
	root.setMessage(msg)
}

// filterWriter searches and writes to filterDoc.
func (m *Document) filterWriter(ctx context.Context, searcher Searcher, startLN int, filterDoc *filterDocument) {
	defer filterDoc.w.Close()
	for originLN, renderLN := startLN, startLN; ; {
		select {
		case <-ctx.Done():
			return
		default:
		}
		lineNum, err := m.searchLine(ctx, searcher, true, originLN)
		if err != nil {
			// Not found
			break
		}
		// Found
		line, err := m.Line(lineNum)
		if err != nil {
			// deleted?
			log.Println(err)
			break
		}
		filterDoc.lineNumMap.Store(renderLN, lineNum)
		writeLine(filterDoc.w, line)
		renderLN++

		originLN = lineNum + 1
	}
}

// closeAllFilter closes all filter documents.
func (root *Root) closeAllFilter(ctx context.Context) {
	if root.DocumentLen() == 1 {
		root.setMessage("only this document")
		return
	}

	docNum, closed := root.closeAllDocumentsOfType(DocFilter)
	root.setDocumentNum(ctx, docNum)

	if len(closed) == 0 {
		root.setMessage("no document to close")
		return
	}
	root.setMessageLogf("close %s", strings.Join(closed, ", "))
}

// writeLine writes a line to w.
// It adds a newline to the end of the line.
func writeLine(w io.Writer, line []byte) {
	if _, err := w.Write(line); err != nil {
		panic(err)
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		panic(err)
	}
}
