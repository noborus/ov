package oviewer

import (
	"context"
	"fmt"
	"io"
	"log"
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
	ev := &eventInputFilter{
		value: str,
	}
	root.postEvent(ev)
}

// filter filters the document by the input value.
func (root *Root) filter(ctx context.Context) {
	searcher := root.setSearcher(root.input.value, root.Config.CaseSensitive)
	if searcher == nil {
		if root.Doc.jumpTargetSection {
			root.Doc.jumpTargetNum = 0
		}
		return
	}
	word := root.searcher.String()
	if root.Doc.nonMatch {
		word = fmt.Sprintf("!%s", word)
	}
	root.setMessagef("filter:%v", word)

	m := root.Doc
	r, w := io.Pipe()
	render, err := renderDoc(m, r)
	if err != nil {
		log.Println(err)
		return
	}
	render.documentType = DocFilter
	render.FileName = fmt.Sprintf("filter:%s:%v", m.FileName, word)
	render.Caption = fmt.Sprintf("%s:%v", m.FileName, word)
	root.addDocument(render)
	render.general = mergeGeneral(m.general, render.general)

	render.nonMatch = m.nonMatch
	render.Header = m.Header
	render.SkipLines = m.SkipLines

	filterDoc := &filterDocument{
		Document: render,
		w:        w,
	}

	// Copy the header
	if render.Header > 0 {
		for ln := render.SkipLines; ln < render.Header; ln++ {
			line, err := m.Line(ln)
			if err != nil {
				break
			}
			render.lineNumMap.Store(ln, ln)
			writeLine(w, line)
		}
	}
	go m.filterWriter(ctx, searcher, m.firstLine(), filterDoc)
	root.setMessagef("filter:%v", word)
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
			break
		}
		filterDoc.lineNumMap.Store(renderLN, lineNum)
		writeLine(filterDoc.w, line)
		renderLN++

		originLN = lineNum + 1
	}
}

// closeAllFilter closes all filter documents.
func (root *Root) closeAllFilter() {
	root.closeAllDocument(DocFilter)
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
