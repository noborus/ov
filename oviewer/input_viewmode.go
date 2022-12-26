package oviewer

import "github.com/gdamore/tcell/v2"

// setViewModeMode sets the inputMode to ViewMode.
func (root *Root) setViewInputMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = ViewMode
	input.EventInput = newViewModeInput(input.ModeCandidate)
}

// viewModeInput represents the mode input mode.
type viewModeInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// viewModeCandidate returns the candidate to set to default.
func viewModeCandidate() *candidate {
	return &candidate{
		list: []string{
			"general",
		},
	}
}

// newViewModeInput returns viewModeInput.
func newViewModeInput(clist *candidate) *viewModeInput {
	return &viewModeInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (d *viewModeInput) Prompt() string {
	return "Mode:"
}

// Confirm returns the event when the input is confirmed.
func (d *viewModeInput) Confirm(str string) tcell.Event {
	d.value = str
	d.SetEventNow()
	return d
}

// Up returns strings when the up key is pressed during input.
func (d *viewModeInput) Up(str string) string {
	return d.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (d *viewModeInput) Down(str string) string {
	return d.clist.down()
}
