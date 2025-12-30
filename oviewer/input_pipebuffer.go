package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v3"
)

// inputPipeBuffer is a wrapper to move to setPipeBufferMode.
func (root *Root) inputPipeBuffer(_ context.Context) {
	input := root.input
	input.reset()
	input.Event = newPipeBufferEvent(input.Candidate[PipeBuffer])
}

// eventPipeBuffer represents the input event for pipe buffer mode.
type eventPipeBuffer struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newPipeBufferEvent returns an eventPipeBuffer for pipe buffer mode.
func newPipeBufferEvent(clist *candidate) *eventPipeBuffer {
	return &eventPipeBuffer{clist: clist}
}

// Mode returns InputMode.
func (*eventPipeBuffer) Mode() InputMode {
	return PipeBuffer
}

// Prompt returns the prompt string in the input field.
func (*eventPipeBuffer) Prompt() string {
	return "(Pipe)command:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventPipeBuffer) Confirm(str string) tcell.Event {
	e.value = str
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventPipeBuffer) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventPipeBuffer) Down(_ string) string {
	return e.clist.down()
}
