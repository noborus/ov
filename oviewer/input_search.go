package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// searchType is the type of search.
type searchType int

// Responsible for forward search / backward search / filter search input.
const (
	forward searchType = iota
	backward
	filter
)

// setSearchMode sets the inputMode to Search.
func (root *Root) setCommonSearchMode() {
	input := root.input
	input.reset()

	if root.searcher != nil {
		input.SearchCandidate.toLast(root.searcher.String())
	}

	root.Doc.nonMatch = false
	root.OriginPos = root.Doc.topLN
}

// setSearchMode sets the inputMode to Search.
func (root *Root) setSearchMode(context.Context) {
	root.setCommonSearchMode()
	root.input.Event = newSearchEvent(root.input.SearchCandidate, forward)
}

// setBackSearchMode sets the inputMode to Backsearch.
func (root *Root) setBackSearchMode(context.Context) {
	root.setCommonSearchMode()
	root.input.Event = newSearchEvent(root.input.SearchCandidate, backward)
}

// setSearchFilterMode sets the inputMode to Filter.
func (root *Root) setSearchFilterMode(context.Context) {
	root.setCommonSearchMode()
	root.input.Event = newSearchEvent(root.input.SearchCandidate, filter)
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
	clist      *candidate
	value      string
	searchType searchType
}

// newSearchEvent returns SearchInput.
func newSearchEvent(clist *candidate, searchType searchType) *eventInputSearch {
	return &eventInputSearch{
		value:      "",
		searchType: searchType,
		clist:      clist,
		EventTime:  tcell.EventTime{},
	}
}

// Mode returns InputMode.
func (e *eventInputSearch) Mode() InputMode {
	switch e.searchType {
	case forward:
		return Search
	case backward:
		return Backsearch
	case filter:
		return Filter
	}
	panic("invalid searchType")
}

// Prompt returns the prompt string in the input field.
func (e *eventInputSearch) Prompt() string {
	switch e.searchType {
	case forward:
		return "/"
	case backward:
		return "?"
	case filter:
		return "&"
	}
	panic("invalid searchType")
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

// searchCandidates returns the list of search candidates.
func (input *Input) searchCandidates(n int) []string {
	input.SearchCandidate.mux.Lock()
	defer input.SearchCandidate.mux.Unlock()

	listLen := len(input.SearchCandidate.list)
	start := max(0, listLen-n)
	return input.SearchCandidate.list[start:listLen]
}
