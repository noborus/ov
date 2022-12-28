package oviewer

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
)

// setSkipLinesMode sets the inputMode to SkipLines.
func (root *Root) setSkipLinesMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.Event = newSkipLinesEvent()
}

// skipLinesEvent represents the goto input mode.
type skipLinesEvent struct {
	value string
	tcell.EventTime
}

// newSkipLinesEvent returns skipLinesEvent.
func newSkipLinesEvent() *skipLinesEvent {
	return &skipLinesEvent{}
}

// Mode returns InputMode.
func (e *skipLinesEvent) Mode() InputMode {
	return SkipLines
}

// Prompt returns the prompt string in the input field.
func (e *skipLinesEvent) Prompt() string {
	return "Skip lines:"
}

// Confirm returns the event when the input is confirmed.
func (e *skipLinesEvent) Confirm(str string) tcell.Event {
	e.value = str
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *skipLinesEvent) Up(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n + 1)
}

// Down returns strings when the down key is pressed during input.
func (e *skipLinesEvent) Down(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil || n <= 0 {
		return "0"
	}
	return strconv.Itoa(n - 1)
}
