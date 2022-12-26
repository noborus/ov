package oviewer

import "github.com/gdamore/tcell/v2"

// setDelimiterMode sets the inputMode to Delimiter.
func (root *Root) setDelimiterMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = Delimiter
	input.EventInput = newDelimiterInput(input.DelimiterCandidate)
}

// delimiterCandidate returns the candidate to set to default.
func delimiterCandidate() *candidate {
	return &candidate{
		list: []string{
			"│",
			"\t",
			"|",
			",",
		},
	}
}

// delimiterInput represents the delimiter input mode.
type delimiterInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newDelimiterInput returns DelimiterInput.
func newDelimiterInput(clist *candidate) *delimiterInput {
	return &delimiterInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (d *delimiterInput) Prompt() string {
	return "Delimiter:"
}

// Confirm returns the event when the input is confirmed.
func (d *delimiterInput) Confirm(str string) tcell.Event {
	d.value = str
	d.clist.list = toLast(d.clist.list, str)
	d.clist.p = 0
	d.SetEventNow()
	return d
}

// Up returns strings when the up key is pressed during input.
func (d *delimiterInput) Up(str string) string {
	return d.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (d *delimiterInput) Down(str string) string {
	return d.clist.down()
}
