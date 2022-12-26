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
	input.mode = SkipLines
	input.EventInput = newSkipLinesInput()
}

// skipLinesInput represents the goto input mode.
type skipLinesInput struct {
	value string
	tcell.EventTime
}

// Prompt returns the prompt string in the input field.
func (h *skipLinesInput) Prompt() string {
	return "Skip lines:"
}

// newSkipLinesInput returns SkipLinesInput.
func newSkipLinesInput() *skipLinesInput {
	return &skipLinesInput{}
}

// Confirm returns the event when the input is confirmed.
func (h *skipLinesInput) Confirm(str string) tcell.Event {
	h.value = str
	h.SetEventNow()
	return h
}

// Up returns strings when the up key is pressed during input.
func (h *skipLinesInput) Up(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n + 1)
}

// Down returns strings when the down key is pressed during input.
func (h *skipLinesInput) Down(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil || n <= 0 {
		return "0"
	}
	return strconv.Itoa(n - 1)
}
