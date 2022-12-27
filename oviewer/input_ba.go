package oviewer

import "github.com/gdamore/tcell/v2"

// setJumpWriteBAMode sets the inputMode to WriteBA.
func (root *Root) setWriteBAMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = WriteBA
	input.EventInput = newWriteBAInput(input.WriteBACandidate)
}

// baCandidate returns the candidate to set to default.
func baCandidate() *candidate {
	return &candidate{
		list: []string{},
	}
}

// writeBAInput represents the WatchInteval input mode.
type writeBAInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newWatchIntevalInputt returns WatchIntevalInput.
func newWriteBAInput(clist *candidate) *writeBAInput {
	return &writeBAInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (t *writeBAInput) Prompt() string {
	return "WriteAndQuit Before:After:"
}

// Confirm returns the event when the input is confirmed.
func (t *writeBAInput) Confirm(str string) tcell.Event {
	t.value = str
	t.clist.list = toLast(t.clist.list, str)
	t.clist.p = 0
	t.SetEventNow()
	return t
}

// Up returns strings when the up key is pressed during input.
func (t *writeBAInput) Up(str string) string {
	return t.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (t *writeBAInput) Down(str string) string {
	return t.clist.down()
}
