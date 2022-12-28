package oviewer

import "github.com/gdamore/tcell/v2"

// setSectionDelimiterMode sets the inputMode to SectionDelimiter.
func (root *Root) setSectionDelimiterMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.Event = newSectionDelimiterEvent(input.SectionDelmCandidate)
}

// sectionDelimiterCandidate returns the candidate to set to default.
func sectionDelimiterCandidate() *candidate {
	return &candidate{
		list: []string{
			"^commit",
			"^diff",
			"^#",
			"^$",
			"^\\f",
		},
	}
}

// delimiterInput represents the delimiter input mode.
type sectionDelimiterEvent struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newSectionDelimiterEvent returns sectionDelimiterInput.
func newSectionDelimiterEvent(clist *candidate) *sectionDelimiterEvent {
	return &sectionDelimiterEvent{clist: clist}
}

// Mode returns InputMode.
func (e *sectionDelimiterEvent) Mode() InputMode {
	return SectionDelimiter
}

// Prompt returns the prompt string in the input field.
func (e *sectionDelimiterEvent) Prompt() string {
	return "Section delimiter:"
}

// Confirm returns the event when the input is confirmed.
func (e *sectionDelimiterEvent) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *sectionDelimiterEvent) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *sectionDelimiterEvent) Down(str string) string {
	return e.clist.down()
}
