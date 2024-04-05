package oviewer

import (
	"github.com/gdamore/tcell/v2"
)

// setDelimiterMode sets the inputMode to Delimiter.
func (root *Root) setDelimiterMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0

	input.DelimiterCandidate.toLast(root.Doc.ColumnDelimiter)

	input.Event = newDelimiterEvent(input.DelimiterCandidate)
}

// delimiterCandidate returns the candidate to set to default.
func delimiterCandidate() *candidate {
	return &candidate{
		list: []string{
			"│",
			"\t",
			"|",
			",",
			`/\s+/`,
		},
	}
}

// eventDelimiter represents the delimiter input mode.
type eventDelimiter struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newDelimiterEvent returns delimiterEvent.
func newDelimiterEvent(clist *candidate) *eventDelimiter {
	return &eventDelimiter{clist: clist}
}

// Mode returns InputMode.
func (*eventDelimiter) Mode() InputMode {
	return Delimiter
}

// Prompt returns the prompt string in the input field.
func (*eventDelimiter) Prompt() string {
	return "Delimiter:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventDelimiter) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventDelimiter) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventDelimiter) Down(_ string) string {
	return e.clist.down()
}
