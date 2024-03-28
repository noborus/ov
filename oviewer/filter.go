package oviewer

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	// Recheck interval after reaching the last line.
	recheckInterval = 1 * time.Second
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
			w.Write(line)
			w.Write([]byte("\n"))
		}
	}

	go m.searchWriter(ctx, searcher, filterDoc, m.firstLine())

	root.setMessagef("filter:%v", word)
}

// searchWriter searches the document and writes the result to w.
func (m *Document) searchWriter(ctx context.Context, searcher Searcher, renderDoc *renderDocument, ln int) {
	defer renderDoc.writer.Close()
	originLN := ln
	renderLN := ln
	for {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			lineNum, err := m.searchLine(ctx, searcher, true, originLN)
			if err != nil {
				// Not found
				originLN = lineNum
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
			renderDoc.writer.Write(line)
			renderDoc.writer.Write([]byte("\n"))
			renderLN++
			originLN = lineNum + 1
		}
		// Wait for the buffer to be updated.
		for originLN >= m.BufEndNum() {
			time.Sleep(recheckInterval)
		}
	}
}
