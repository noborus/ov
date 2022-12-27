package oviewer

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
)

// setHeaderMode sets the inputMode to Header.
func (root *Root) setHeaderMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = Header
	input.EventInput = newHeaderInput()
}

// headerInput represents the goto input mode.
type headerInput struct {
	value string
	tcell.EventTime
}

// newHeaderInput returns HeaderInput.
func newHeaderInput() *headerInput {
	return &headerInput{}
}

// Prompt returns the prompt string in the input field.
func (h *headerInput) Prompt() string {
	return "Header length:"
}

// Confirm returns the event when the input is confirmed.
func (h *headerInput) Confirm(str string) tcell.Event {
	h.value = str
	h.SetEventNow()
	return h
}

// Up returns strings when the up key is pressed during input.
func (h *headerInput) Up(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n + 1)
}

// Down returns strings when the down key is pressed during input.
func (h *headerInput) Down(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil || n <= 0 {
		return "0"
	}
	return strconv.Itoa(n - 1)
}
