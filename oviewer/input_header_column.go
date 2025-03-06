package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// setHeaderColumnMode sets the inputMode to Vertical Header Column.
func (root *Root) setHeaderColumnMode(context.Context) {
	input := root.input
	input.reset()
	input.Event = newHeaderColumnEvent()
}

// eventHeaderColumn represents the vertical header column input mode.
type eventHeaderColumn struct {
	tcell.EventTime
	value string
}

// newHeaderColumnEvent returns a new vertical header column event.
func newHeaderColumnEvent() *eventHeaderColumn {
	return &eventHeaderColumn{}
}

// Mode returns InputMode.
func (*eventHeaderColumn) Mode() InputMode {
	return HeaderColumn
}

// Prompt returns the prompt string in the input field.
func (*eventHeaderColumn) Prompt() string {
	return "Header column:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventHeaderColumn) Confirm(str string) tcell.Event {
	e.value = str
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (*eventHeaderColumn) Up(str string) string {
	return upNum(str)
}

// Down returns strings when the down key is pressed during input.
func (*eventHeaderColumn) Down(str string) string {
	return downNum(str)
}
