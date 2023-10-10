package oviewer

import "github.com/gdamore/tcell/v2"

// setSectionDelimiterMode sets the inputMode to SectionDelimiter.
func (root *Root) setSectionDelimiterMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0

	input.SectionDelmCandidate.toLast(root.Doc.SectionDelimiter)

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
type eventSectionDelimiter struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newSectionDelimiterEvent returns sectionDelimiterInput.
func newSectionDelimiterEvent(clist *candidate) *eventSectionDelimiter {
	return &eventSectionDelimiter{clist: clist}
}

// Mode returns InputMode.
func (e *eventSectionDelimiter) Mode() InputMode {
	return SectionDelimiter
}

// Prompt returns the prompt string in the input field.
func (e *eventSectionDelimiter) Prompt() string {
	return "Section delimiter:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventSectionDelimiter) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventSectionDelimiter) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventSectionDelimiter) Down(str string) string {
	return e.clist.down()
}
