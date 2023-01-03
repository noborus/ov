package oviewer

import "github.com/gdamore/tcell/v2"

// setSectionStartMode sets the inputMode to SectionStart.
func (root *Root) setSectionStartMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.Event = newSectionStartEvent(input.SectionStartCandidate)
}

// sectionStartCandidate returns the candidate to set to default.
func sectionStartCandidate() *candidate {
	return &candidate{
		list: []string{
			"2",
			"1",
			"0",
		},
	}
}

// eventSectionStart represents the section start position input mode.
type eventSectionStart struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newSectionStartEvent returns sectionDelimiterEvent.
func newSectionStartEvent(clist *candidate) *eventSectionStart {
	return &eventSectionStart{clist: clist}
}

// Mode returns InputMode.
func (e *eventSectionStart) Mode() InputMode {
	return SectionStart
}

// Prompt returns the prompt string in the input field.
func (e *eventSectionStart) Prompt() string {
	return "Section start:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventSectionStart) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventSectionStart) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventSectionStart) Down(str string) string {
	return e.clist.down()
}
