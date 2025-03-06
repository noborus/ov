package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// setViewModeMode sets the inputMode to ViewMode.
func (root *Root) inputViewMode(context.Context) {
	input := root.input
	input.reset()
	input.Event = newViewModeEvent(input.Candidate[ViewMode])
}

// viewModeCandidate returns the candidate to set to default.
func viewModeCandidate() *candidate {
	return &candidate{
		list: []string{
			nameGeneral,
		},
	}
}

// eventViewMode represents the view mode input mode.
type eventViewMode struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newViewModeEvent returns viewModeEvent.
func newViewModeEvent(clist *candidate) *eventViewMode {
	return &eventViewMode{clist: clist}
}

// Mode returns InputMode.
func (*eventViewMode) Mode() InputMode {
	return ViewMode
}

// Prompt returns the prompt string in the input field.
func (*eventViewMode) Prompt() string {
	return "Mode:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventViewMode) Confirm(str string) tcell.Event {
	e.value = str
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventViewMode) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventViewMode) Down(_ string) string {
	return e.clist.down()
}
