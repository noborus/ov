package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// inputSectionDelimiter sets the inputMode to SectionDelimiter.
func (root *Root) inputSectionDelimiter(context.Context) {
	input := root.input
	input.reset()
	input.Candidate[SectionDelimiter].toLast(root.Doc.SectionDelimiter)
	input.Event = newSectionDelimiterEvent(input.Candidate[SectionDelimiter])
}

// sectionDelimiterCandidate returns the candidate to set to default.
func sectionDelimiterCandidate() *candidate {
	return &candidate{
		list: []string{
			"^commit",
			"^diff",
			"^#",
			"^$",
			"^\\f",
		},
	}
}

// delimiterInput represents the delimiter input mode.
type eventSectionDelimiter struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newSectionDelimiterEvent returns sectionDelimiterInput.
func newSectionDelimiterEvent(clist *candidate) *eventSectionDelimiter {
	return &eventSectionDelimiter{clist: clist}
}

// Mode returns InputMode.
func (*eventSectionDelimiter) Mode() InputMode {
	return SectionDelimiter
}

// Prompt returns the prompt string in the input field.
func (*eventSectionDelimiter) Prompt() string {
	return "Section delimiter:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventSectionDelimiter) Confirm(str string) tcell.Event {
	e.value = str
	if str != "" {
		e.clist.toLast(str)
	}
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventSectionDelimiter) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventSectionDelimiter) Down(_ string) string {
	return e.clist.down()
}
