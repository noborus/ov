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

// writeBAEvent represents the WatchInteval input mode.
type writeBAEvent struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newWriteBAEvent returns writeBAEvent.
func newWriteBAEvent(clist *candidate) *writeBAEvent {
	return &writeBAEvent{clist: clist}
}

// Mode returns InputMode.
func (e *writeBAEvent) Mode() InputMode {
	return WriteBA
}

// Prompt returns the prompt string in the input field.
func (e *writeBAEvent) Prompt() string {
	return "WriteAndQuit Before:After:"
}

// Confirm returns the event when the input is confirmed.
func (e *writeBAEvent) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *writeBAEvent) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *writeBAEvent) Down(str string) string {
	return e.clist.down()
}
