package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// setVerticalHeaderMode sets the inputMode to Vertical Header.
func (root *Root) setVerticalHeaderMode(context.Context) {
	input := root.input
	input.reset()
	input.Event = newVerticalHeaderEvent()
}

// eventVerticalHeader represents the vertical header input mode.
type eventVerticalHeader struct {
	tcell.EventTime
	value string
}

// newVerticalHeaderEvent returns a new vertical header event.
func newVerticalHeaderEvent() *eventVerticalHeader {
	return &eventVerticalHeader{}
}

// Mode returns InputMode.
func (*eventVerticalHeader) Mode() InputMode {
	return VerticalHeader
}

// Prompt returns the prompt string in the input field.
func (*eventVerticalHeader) Prompt() string {
	return "Vertical header length:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventVerticalHeader) Confirm(str string) tcell.Event {
	e.value = str
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (*eventVerticalHeader) Up(str string) string {
	return upNum(str)
}

// Down returns strings when the down key is pressed during input.
func (*eventVerticalHeader) Down(str string) string {
	return downNum(str)
}
