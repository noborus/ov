package oviewer

import "github.com/gdamore/tcell/v2"

// setSearchMode sets the inputMode to Search.
func (root *Root) setSearchMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.Event = newSearchEvent(input.SearchCandidate)
	root.OriginPos = root.Doc.topLN
}

// searchCandidate returns the candidate to set to default.
func searchCandidate() *candidate {
	return &candidate{
		list: []string{},
	}
}

// inputSearch represents the search input mode.
type inputSearch struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newSearchEvent returns SearchInput.
func newSearchEvent(clist *candidate) *inputSearch {
	return &inputSearch{
		value:     "",
		clist:     clist,
		EventTime: tcell.EventTime{},
	}
}

// Mode returns InputMode.
func (e *inputSearch) Mode() InputMode {
	return Search
}

// Prompt returns the prompt string in the input field.
func (e *inputSearch) Prompt() string {
	return "/"
}

// Confirm returns the event when the input is confirmed.
func (e *inputSearch) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *inputSearch) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *inputSearch) Down(str string) string {
	return e.clist.down()
}

// setBackSearchMode sets the inputMode to Backsearch.
func (root *Root) setBackSearchMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.Event = newBackSearchEvent(input.SearchCandidate)
	root.OriginPos = root.Doc.topLN
}

// inputBackSearch represents the back search input mode.
type inputBackSearch struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newBackSearchEvent returns backSearchEvent.
func newBackSearchEvent(clist *candidate) *inputBackSearch {
	return &inputBackSearch{clist: clist}
}

func (e *inputBackSearch) Mode() InputMode {
	return Backsearch
}

// Prompt returns the prompt string in the input field.
func (e *inputBackSearch) Prompt() string {
	return "?"
}

// Confirm returns the event when the input is confirmed.
func (e *inputBackSearch) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *inputBackSearch) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *inputBackSearch) Down(str string) string {
	return e.clist.down()
}
