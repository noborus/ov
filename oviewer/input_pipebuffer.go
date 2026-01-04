package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v3"
)

const (
	pipeBufferPrompt    = "(Pipe)command:"
	pipeBufferDocPrompt = "(PipeDoc)command:"
)

// inputPipeBuffer is a wrapper to move to pipe buffer mode (terminal output).
func (root *Root) inputPipeBuffer(_ context.Context) {
	input := root.input
	input.reset()
	input.Event = newPipeBufferEvent(input.Candidate[PipeBuffer], pipeBufferPrompt, false)
}

// inputPipeBufferDoc is a wrapper to move to pipe buffer mode (new document output).
func (root *Root) inputPipeBufferDoc(_ context.Context) {
	input := root.input
	input.reset()
	input.Event = newPipeBufferEvent(input.Candidate[PipeBuffer], pipeBufferDocPrompt, true)
}

// eventPipeBuffer represents the input event for pipe buffer mode.
type eventPipeBuffer struct {
	tcell.EventTime
	clist  *candidate
	value  string
	prompt string
	toDoc  bool
}

// newPipeBufferEvent returns an eventPipeBuffer for pipe buffer mode.
func newPipeBufferEvent(clist *candidate, prompt string, toDoc bool) *eventPipeBuffer {
	return &eventPipeBuffer{
		clist:  clist,
		prompt: prompt,
		toDoc:  toDoc,
	}
}

// Mode returns InputMode.
func (*eventPipeBuffer) Mode() InputMode {
	return PipeBuffer
}

// Prompt returns the prompt string in the input field.
func (e *eventPipeBuffer) Prompt() string {
	return e.prompt
}

// Confirm returns the event when the input is confirmed.
func (e *eventPipeBuffer) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
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
