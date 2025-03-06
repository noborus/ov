package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// setSaveBuffer is a wrapper to move to setSaveBufferMode.
func (root *Root) inputSaveBuffer(ctx context.Context) {
	if root.Doc.seekable {
		root.setMessage("Does not support saving regular files")
		return
	}
	input := root.input
	input.reset()
	input.Event = newSaveBufferEvent(input.Candidate[SaveBuffer])
}

// eventSaveBuffer represents the mode input mode.
type eventSaveBuffer struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newSaveBufferEvent returns SaveBufferModeEvent.
func newSaveBufferEvent(clist *candidate) *eventSaveBuffer {
	return &eventSaveBuffer{clist: clist}
}

// Mode returns InputMode.
func (*eventSaveBuffer) Mode() InputMode {
	return SaveBuffer
}

// Prompt returns the prompt string in the input field.
func (*eventSaveBuffer) Prompt() string {
	return "(Save)file:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventSaveBuffer) Confirm(str string) tcell.Event {
	e.value = str
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventSaveBuffer) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventSaveBuffer) Down(_ string) string {
	return e.clist.down()
}
