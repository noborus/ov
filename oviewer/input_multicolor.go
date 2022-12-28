package oviewer

import "github.com/gdamore/tcell/v2"

// setMultiColorMode sets the inputMode to MultiColor.
func (root *Root) setMultiColorMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.Event = newMultiColorEvent(input.MultiColorCandidate)
}

// multiColorCandidate returns the candidate to set to default.
func multiColorCandidate() *candidate {
	return &candidate{
		list: []string{
			"error info warn debug",
			"ERROR WARNING NOTICE INFO PANIC FATAL LOG",
		},
	}
}

// multiColorEvent represents the multi color input mode.
type multiColorEvent struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newMultiColorEvent returns multiColorEvent.
func newMultiColorEvent(clist *candidate) *multiColorEvent {
	return &multiColorEvent{clist: clist}
}

// Mode returns InputMode.
func (e *multiColorEvent) Mode() InputMode {
	return MultiColor
}

// Prompt returns the prompt string in the input field.
func (e *multiColorEvent) Prompt() string {
	return "multicolor:"
}

// Confirm returns the event when the input is confirmed.
func (e *multiColorEvent) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *multiColorEvent) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *multiColorEvent) Down(str string) string {
	return e.clist.down()
}
