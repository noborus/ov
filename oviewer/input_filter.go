package oviewer

import (
	"github.com/gdamore/tcell/v2"
)

// eventInputFilter represents the filter input mode.
type eventInputFilter struct {
	tcell.EventTime
	clist *candidate
	value string
}

// setSearchFilterMode sets the inputMode to Filter.
func (root *Root) setSearchFilterMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0

	root.Doc.nonMatch = false
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
