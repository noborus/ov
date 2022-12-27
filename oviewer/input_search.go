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

// searchEvent represents the search input mode.
type searchEvent struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newSearchEvent returns SearchInput.
func newSearchEvent(clist *candidate) *searchEvent {
	return &searchEvent{
		value:     "",
		clist:     clist,
		EventTime: tcell.EventTime{},
	}
}

// Mode returns InputMode.
func (e *searchEvent) Mode() InputMode {
	return Search
}

// Prompt returns the prompt string in the input field.
func (e *searchEvent) Prompt() string {
	return "/"
}

// Confirm returns the event when the input is confirmed.
func (e *searchEvent) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *searchEvent) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *searchEvent) Down(str string) string {
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

// backSearchEvent represents the back search input mode.
type backSearchEvent struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newBackSearchEvent returns backSearchEvent.
func newBackSearchEvent(clist *candidate) *backSearchEvent {
	return &backSearchEvent{clist: clist}
}

func (e *backSearchEvent) Mode() InputMode {
	return Backsearch
}

// Prompt returns the prompt string in the input field.
func (e *backSearchEvent) Prompt() string {
	return "?"
}

// Confirm returns the event when the input is confirmed.
func (e *backSearchEvent) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *backSearchEvent) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *backSearchEvent) Down(str string) string {
	return e.clist.down()
}
