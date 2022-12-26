package oviewer

import "github.com/gdamore/tcell/v2"

// setSectionStartMode sets the inputMode to SectionStart.
func (root *Root) setSectionStartMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = SectionStart
	input.EventInput = newSectionStartInput(input.SectionStartCandidate)
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

// sectionStartInput represents the section start position input mode.
type sectionStartInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newSectionDelimiterInput returns sectionDelimiterInput.
func newSectionStartInput(clist *candidate) *sectionStartInput {
	return &sectionStartInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (d *sectionStartInput) Prompt() string {
	return "Section start:"
}

// Confirm returns the event when the input is confirmed.
func (d *sectionStartInput) Confirm(str string) tcell.Event {
	d.value = str
	d.clist.list = toLast(d.clist.list, str)
	d.clist.p = 0
	d.SetEventNow()
	return d
}

// Up returns strings when the up key is pressed during input.
func (d *sectionStartInput) Up(str string) string {
	return d.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (d *sectionStartInput) Down(str string) string {
	return d.clist.down()
}
