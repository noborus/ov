package oviewer

import (
	"github.com/gdamore/tcell/v2"
)

// setSearchMode sets the inputMode to Search.
func (root *Root) setSearchMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0

	if root.searcher != nil {
		input.SearchCandidate.toLast(root.searcher.String())
	}

	input.Event = newSearchEvent(input.SearchCandidate)
	root.OriginPos = root.Doc.topLN
}

// searchCandidate returns the candidate to set to default.
func searchCandidate() *candidate {
	return &candidate{
		list: []string{},
	}
}

// eventInputSearch represents the search input mode.
type eventInputSearch struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newSearchEvent returns SearchInput.
func newSearchEvent(clist *candidate) *eventInputSearch {
	return &eventInputSearch{
		value:     "",
		clist:     clist,
		EventTime: tcell.EventTime{},
	}
}

// Mode returns InputMode.
func (e *eventInputSearch) Mode() InputMode {
	return Search
}

// Prompt returns the prompt string in the input field.
func (e *eventInputSearch) Prompt() string {
	return "/"
}

// Confirm returns the event when the input is confirmed.
func (e *eventInputSearch) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventInputSearch) Up(str string) string {
	e.clist.toAddLast(str)
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventInputSearch) Down(str string) string {
	e.clist.toAddTop(str)
	return e.clist.down()
}

// setBackSearchMode sets the inputMode to Backsearch.
func (root *Root) setBackSearchMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0

	if root.searcher != nil {
		input.SearchCandidate.toLast(root.searcher.String())
	}

	input.Event = newBackSearchEvent(input.SearchCandidate)
	root.OriginPos = root.Doc.topLN
}

// eventInputBackSearch represents the back search input mode.
type eventInputBackSearch struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newBackSearchEvent returns backSearchEvent.
func newBackSearchEvent(clist *candidate) *eventInputBackSearch {
	return &eventInputBackSearch{clist: clist}
}

// Mode returns InputMode.
func (e *eventInputBackSearch) Mode() InputMode {
	return Backsearch
}

// Prompt returns the prompt string in the input field.
func (e *eventInputBackSearch) Prompt() string {
	return "?"
}

// Confirm returns the event when the input is confirmed.
func (e *eventInputBackSearch) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventInputBackSearch) Up(str string) string {
	e.clist.toAddLast(str)
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventInputBackSearch) Down(str string) string {
	e.clist.toAddTop(str)
	return e.clist.down()
}

// searchCandidates returns the list of search candidates.
func (input *Input) searchCandidates(n int) []string {
	input.SearchCandidate.mux.Lock()
	defer input.SearchCandidate.mux.Unlock()

	listLen := len(input.SearchCandidate.list)
	start := max(0, listLen-n)
	return input.SearchCandidate.list[start:listLen]
}
