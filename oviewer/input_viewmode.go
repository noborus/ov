package oviewer

import "github.com/gdamore/tcell/v2"

// setViewModeMode sets the inputMode to ViewMode.
func (root *Root) setViewInputMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.Event = newViewModeEvent(input.ModeCandidate)
}

// viewModeCandidate returns the candidate to set to default.
func viewModeCandidate() *candidate {
	return &candidate{
		list: []string{
			"general",
		},
	}
}

// eventViewMode represents the mode input mode.
type eventViewMode struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newViewModeEvent returns viewModeEvent.
func newViewModeEvent(clist *candidate) *eventViewMode {
	return &eventViewMode{clist: clist}
}

// Mode returns InputMode.
func (e *eventViewMode) Mode() InputMode {
	return ViewMode
}

// Prompt returns the prompt string in the input field.
func (e *eventViewMode) Prompt() string {
	return "Mode:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventViewMode) Confirm(str string) tcell.Event {
	e.value = str
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventViewMode) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventViewMode) Down(str string) string {
	return e.clist.down()
}
