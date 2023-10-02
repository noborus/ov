package oviewer

import "github.com/gdamore/tcell/v2"

// setWriteBAMode sets the inputMode to WriteBA.
func (root *Root) setWriteBAMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.Event = newWriteBAEvent(input.WriteBACandidate)
}

// baCandidate returns the candidate to set to default.
func baCandidate() *candidate {
	return &candidate{
		list: []string{},
	}
}

// eventWriteBA represents the writeBA input mode.
type eventWriteBA struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newWriteBAEvent returns writeBAEvent.
func newWriteBAEvent(clist *candidate) *eventWriteBA {
	return &eventWriteBA{clist: clist}
}

// Mode returns InputMode.
func (e *eventWriteBA) Mode() InputMode {
	return WriteBA
}

// Prompt returns the prompt string in the input field.
func (e *eventWriteBA) Prompt() string {
	return "WriteAndQuit Before:After:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventWriteBA) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventWriteBA) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventWriteBA) Down(str string) string {
	return e.clist.down()
}
