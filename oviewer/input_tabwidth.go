package oviewer

import "github.com/gdamore/tcell/v2"

// setTabWidthMode sets the inputMode to TabWidth.
func (root *Root) setTabWidthMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0

	input.Event = newTabWidthEvent(input.TabWidthCandidate)
}

// tabWidthCandidate returns the candidate to set to default.
func tabWidthCandidate() *candidate {
	return &candidate{
		list: []string{
			"3",
			"2",
			"4",
			"8",
		},
	}
}

// tabWidthEvent represents the TABWidth input mode.
type tabWidthEvent struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newTabWidthEvent returns tabWidthEvent.
func newTabWidthEvent(clist *candidate) *tabWidthEvent {
	return &tabWidthEvent{clist: clist}
}

// Mode returns InputMode.
func (e *tabWidthEvent) Mode() InputMode {
	return TabWidth
}

// Prompt returns the prompt string in the input field.
func (e *tabWidthEvent) Prompt() string {
	return "TAB width:"
}

// Confirm returns the event when the input is confirmed.
func (e *tabWidthEvent) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *tabWidthEvent) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *tabWidthEvent) Down(str string) string {
	return e.clist.down()
}
