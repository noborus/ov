package oviewer

import "github.com/gdamore/tcell/v2"

// setTabWidthMode sets the inputMode to TabWidth.
func (root *Root) setTabWidthMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = TabWidth
	input.EventInput = newTabWidthInput(input.TabWidthCandidate)
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

// tabWidthInput represents the TABWidth input mode.
type tabWidthInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newTabWidthInput returns TABWidthInput.
func newTabWidthInput(clist *candidate) *tabWidthInput {
	return &tabWidthInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (t *tabWidthInput) Prompt() string {
	return "TAB width:"
}

// Confirm returns the event when the input is confirmed.
func (t *tabWidthInput) Confirm(str string) tcell.Event {
	t.value = str
	t.clist.list = toLast(t.clist.list, str)
	t.clist.p = 0
	t.SetEventNow()
	return t
}

// Up returns strings when the up key is pressed during input.
func (t *tabWidthInput) Up(str string) string {
	return t.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (t *tabWidthInput) Down(str string) string {
	return t.clist.down()
}
