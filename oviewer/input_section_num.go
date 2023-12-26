package oviewer

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
)

// setSectionNumMode sets the inputMode to SectionNum.
func (root *Root) setSectionNumMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.Event = newSectionNumEvent()
}

// eventSectionNum represents the section num input mode.
type eventSectionNum struct {
	tcell.EventTime
	value string
}

// newSectionNumEvent returns Event.
func newSectionNumEvent() *eventSectionNum {
	return &eventSectionNum{}
}

// Mode returns InputMode.
func (e *eventSectionNum) Mode() InputMode {
	return SectionNum
}

// Prompt returns the prompt string in the input field.
func (e *eventSectionNum) Prompt() string {
	return "Section Num:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventSectionNum) Confirm(str string) tcell.Event {
	e.value = str
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventSectionNum) Up(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n + 1)
}

// Down returns strings when the down key is pressed during input.
func (e *eventSectionNum) Down(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil || n <= 0 {
		return "0"
	}
	return strconv.Itoa(n - 1)
}
