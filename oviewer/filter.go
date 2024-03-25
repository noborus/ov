package oviewer

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
)

func (root *Root) filter(ctx context.Context) {
	searcher := root.setSearcher(root.input.value, root.Config.CaseSensitive)
	if searcher == nil {
		if root.Doc.jumpTargetSection {
			root.Doc.jumpTargetNum = 0
		}
		return
	}
	word := root.searcher.String()
	root.setMessagef("filter:%v (%v)Cancel", word, strings.Join(root.cancelKeys, ","))

	m := root.Doc
	r, w := io.Pipe()
	filterDoc, err := renderDoc(m, r)
	if err != nil {
		log.Println(err)
		return
	}
	filterDoc.FileName = fmt.Sprintf("filter:%s:%v", m.FileName, word)
	filterDoc.Caption = fmt.Sprintf("%s:%v", m.FileName, word)
	root.addDocument(filterDoc.Document)

	if m.Header > 0 {
		for ln := m.SkipLines; ln < m.Header; ln++ {
			line, err := m.Line(ln)
			if err != nil {
				break
			}
			filterDoc.lineNumMap.Store(ln, ln)
			w.Write(line)
			w.Write([]byte("\n"))
		}
	}

	filterDoc.writer = w
	filterDoc.Header = m.Header
	filterDoc.SkipLines = m.SkipLines

	go m.searchWriter(ctx, searcher, filterDoc, m.firstLine())
	root.setMessagef("filter:%v", word)
}

// searchWriter searches the document and writes the result to w.
func (m *Document) searchWriter(ctx context.Context, searcher Searcher, renderDoc *renderDocument, ln int) {
	defer renderDoc.writer.Close()
	nextLN := ln
	for {
		lineNum, err := m.searchLine(ctx, searcher, true, nextLN)
		if err != nil {
			break
		}
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
		renderDoc.lineNumMap.Store(ln, num)
		renderDoc.writer.Write(line)
		renderDoc.writer.Write([]byte("\n"))
		nextLN = lineNum + 1
		ln++
	}
}
