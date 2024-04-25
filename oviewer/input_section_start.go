package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// setSectionStartMode sets the inputMode to SectionStart.
func (root *Root) setSectionStartMode(context.Context) {
	input := root.input
	input.reset()
	input.Event = newSectionStartEvent(input.SectionStartCandidate)
}

// sectionStartCandidate returns the candidate to set to default.
func sectionStartCandidate() *candidate {
	return &candidate{
		list: []string{
			"2",
			"1",
			"0",
		},
	}
}

// eventSectionStart represents the section start position input mode.
type eventSectionStart struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newSectionStartEvent returns sectionDelimiterEvent.
func newSectionStartEvent(clist *candidate) *eventSectionStart {
	return &eventSectionStart{clist: clist}
}

// Mode returns InputMode.
func (*eventSectionStart) Mode() InputMode {
	return SectionStart
}

// Prompt returns the prompt string in the input field.
func (*eventSectionStart) Prompt() string {
	return "Section start:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventSectionStart) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventSectionStart) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventSectionStart) Down(_ string) string {
	return e.clist.down()
}
