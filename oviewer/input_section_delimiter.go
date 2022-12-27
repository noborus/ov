package oviewer

import "github.com/gdamore/tcell/v2"

// setSectionDelimiterMode sets the inputMode to SectionDelimiter.
func (root *Root) setSectionDelimiterMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = SectionDelimiter
	input.EventInput = newSectionDelimiterInput(input.SectionDelmCandidate)
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
type sectionDelimiterInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newSectionDelimiterInput returns sectionDelimiterInput.
func newSectionDelimiterInput(clist *candidate) *sectionDelimiterInput {
	return &sectionDelimiterInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (d *sectionDelimiterInput) Prompt() string {
	return "Section delimiter:"
}

// Confirm returns the event when the input is confirmed.
func (d *sectionDelimiterInput) Confirm(str string) tcell.Event {
	d.value = str
	d.clist.list = toLast(d.clist.list, str)
	d.clist.p = 0
	d.SetEventNow()
	return d
}

// Up returns strings when the up key is pressed during input.
func (d *sectionDelimiterInput) Up(str string) string {
	return d.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (d *sectionDelimiterInput) Down(str string) string {
	return d.clist.down()
}
