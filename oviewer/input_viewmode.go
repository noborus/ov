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

// viewModeEvent represents the mode input mode.
type viewModeEvent struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newViewModeEvent returns viewModeEvent.
func newViewModeEvent(clist *candidate) *viewModeEvent {
	return &viewModeEvent{clist: clist}
}

// Mode returns InputMode.
func (e *viewModeEvent) Mode() InputMode {
	return ViewMode
}

// Prompt returns the prompt string in the input field.
func (e *viewModeEvent) Prompt() string {
	return "Mode:"
}

// Confirm returns the event when the input is confirmed.
func (e *viewModeEvent) Confirm(str string) tcell.Event {
	e.value = str
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *viewModeEvent) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *viewModeEvent) Down(str string) string {
	return e.clist.down()
}
