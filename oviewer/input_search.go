package oviewer

import (
	"context"
	"strings"

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

// inputForwardSearch sets the inputMode to Forwardsearch.
func (root *Root) inputForwardSearch(context.Context) {
	root.setSearchMode(forward)
}

// inputBackSearch sets the inputMode to Backsearch.
func (root *Root) inputBackSearch(context.Context) {
	root.setSearchMode(backward)
}

// inputSearchFilter sets the inputMode to Filter.
func (root *Root) inputSearchFilter(context.Context) {
	root.setSearchMode(filter)
}

// setSearchMode sets the inputMode to Search.
func (root *Root) setSearchMode(searchType searchType) {
	input := root.input
	input.reset()
	input.Event = newSearchEvent(input.Candidate[Search], searchType)

	if root.searcher != nil {
		input.Candidate[Search].toLast(root.searcher.String())
	}

	root.Doc.nonMatch = false // Reset nonMatch.
	root.OriginPos = root.Doc.topLN
	root.setPromptOpt()
}

// setPromptOpt returns a string describing the input field.
func (root *Root) setPromptOpt() {
	mode := root.input.Event.Mode()

	if mode != Search && mode != Backsearch && mode != Filter {
		root.searchOpt = ""
		return
	}

	var opt strings.Builder
	if mode == Filter && root.Doc.nonMatch {
		opt.WriteString("Non-match")
	}
	if root.Config.RegexpSearch {
		opt.WriteString("(R)")
	}
	if mode != Filter && root.Config.Incsearch {
		opt.WriteString("(I)")
	}
	if root.Config.SmartCaseSensitive {
		opt.WriteString("(S)")
	} else if root.Config.CaseSensitive {
		opt.WriteString("(Aa)")
	}
	root.searchOpt = opt.String()
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
	e.value = stripBackSlash(str)
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
	input.Candidate[Search].mux.Lock()
	defer input.Candidate[Search].mux.Unlock()

	listLen := len(input.Candidate[Search].list)
	start := max(0, listLen-n)
	return input.Candidate[Search].list[start:listLen]
}

// stripBackSlash removes the backslash from the string
// when the backslash is followed by an exclamation mark.
func stripBackSlash(str string) string {
	if len(str) < 2 {
		return str
	}
	idx := strings.Index(str, "\\!")
	if idx < 0 {
		return str
	}
	str = str[:idx] + str[idx+1:]
	return str
}
