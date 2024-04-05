package oviewer

import (
	"context"
	"fmt"
	"io"
	"log"
)

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
	filterDoc, err := renderDoc(m, r)
	if err != nil {
		log.Println(err)
		return
	}
	filterDoc.documentType = DocFilter
	filterDoc.FileName = fmt.Sprintf("filter:%s:%v", m.FileName, word)
	filterDoc.Caption = fmt.Sprintf("%s:%v", m.FileName, word)
	root.addDocument(filterDoc.Document)
	filterDoc.Document.general = mergeGeneral(m.general, filterDoc.Document.general)

	filterDoc.nonMatch = m.nonMatch
	filterDoc.writer = w
	filterDoc.Header = m.Header
	filterDoc.SkipLines = m.SkipLines

	// Copy the header
	if filterDoc.Header > 0 {
		for ln := filterDoc.SkipLines; ln < filterDoc.Header; ln++ {
			line, err := m.Line(ln)
			if err != nil {
				break
			}
			filterDoc.lineNumMap.Store(ln, ln)
			filterDoc.writeLine(line)
		}
	}
	go m.searchWriter(ctx, searcher, filterDoc, m.firstLine())
	root.setMessagef("filter:%v", word)
}

// searchWriter searches the document and writes the result to w.
func (m *Document) searchWriter(ctx context.Context, searcher Searcher, filterDoc *renderDocument, ln int) {
	defer filterDoc.writer.Close()
	for originLN, renderLN := ln, ln; ; {
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
			break
		}
		num := lineNum
		if m.lineNumMap != nil {
			if n, ok := m.lineNumMap.LoadForward(num); ok {
				num = n
			}
		}
		filterDoc.lineNumMap.Store(renderLN, num)
		filterDoc.writeLine(line)
		renderLN++
		originLN = lineNum + 1
	}
}

func (m *Document) isFilterDocument() bool {
	return m.documentType == DocFilter
}
