package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// inputGoLine sets the inputMode to Goline.
func (root *Root) inputGoLine(context.Context) {
	input := root.input
	input.reset()
	input.Event = newGotoEvent(input.Candidate[Goline])
}

// eventGoto represents the goto input mode.
type eventGoto struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newGotoEvent returns gotoEvent.
func newGotoEvent(clist *candidate) *eventGoto {
	return &eventGoto{clist: clist}
}

// Mode returns InputMode.
func (*eventGoto) Mode() InputMode {
	return Goline
}

// Prompt returns the prompt string in the input field.
func (*eventGoto) Prompt() string {
	return "Goto line:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventGoto) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventGoto) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventGoto) Down(_ string) string {
	return e.clist.down()
}
