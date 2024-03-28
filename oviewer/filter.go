package oviewer

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/gdamore/tcell/v2"
)

// eventInputFilter represents the filter input mode.
type eventInputFilter struct {
	tcell.EventTime
	clist *candidate
	value string
}

// setBackSearchMode sets the inputMode to Backsearch.
func (root *Root) setSearchFilterMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0

	if root.searcher != nil {
		input.SearchCandidate.toLast(root.searcher.String())
	}

	input.Event = newSearchFilterEvent(input.SearchCandidate)
}

// newSearchFilterEvent returns FilterInput.
func newSearchFilterEvent(clist *candidate) *eventInputFilter {
	return &eventInputFilter{
		value:     "",
		clist:     clist,
		EventTime: tcell.EventTime{},
	}
}

// Mode returns InputMode.
func (e *eventInputFilter) Mode() InputMode {
	return Filter
}

// Prompt returns the prompt string in the input field.
func (e *eventInputFilter) Prompt() string {
	return "&"
}

// Confirm returns the event when the input is confirmed.
func (e *eventInputFilter) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventInputFilter) Up(str string) string {
	e.clist.toAddLast(str)
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventInputFilter) Down(str string) string {
	e.clist.toAddTop(str)
	return e.clist.down()
}

// Filter fires the filter event.
func (root *Root) Filter(str string) {
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
	root.setMessagef("filter:%v", word)

	m := root.Doc
	r, w := io.Pipe()
	renderDoc, err := renderDoc(m, r)
	if err != nil {
		log.Println(err)
		return
	}
	renderDoc.FileName = fmt.Sprintf("filter:%s:%v", m.FileName, word)
	renderDoc.Caption = fmt.Sprintf("%s:%v", m.FileName, word)
	root.addDocument(renderDoc.Document)

	renderDoc.writer = w
	renderDoc.Header = m.Header
	renderDoc.SkipLines = m.SkipLines

	// Copy the header
	if renderDoc.Header > 0 {
		for ln := renderDoc.SkipLines; ln < renderDoc.Header; ln++ {
			line, err := m.Line(ln)
			if err != nil {
				break
			}
			renderDoc.lineNumMap.Store(ln, ln)
			renderDoc.writeLine(line)
		}
	}
	go m.searchWriter(ctx, searcher, renderDoc, m.firstLine())
	root.setMessagef("filter:%v", word)
}

// searchWriter searches the document and writes the result to w.
func (m *Document) searchWriter(ctx context.Context, searcher Searcher, renderDoc *renderDocument, ln int) {
	defer renderDoc.writer.Close()
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
		renderDoc.lineNumMap.Store(renderLN, num)
		renderDoc.writeLine(line)
		renderLN++
		originLN = lineNum + 1
	}
}
