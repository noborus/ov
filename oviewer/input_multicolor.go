package oviewer

import "github.com/gdamore/tcell/v2"

// setMultiColorMode sets the inputMode to MultiColor.
func (root *Root) setMultiColorMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = MultiColor
	input.EventInput = newMultiColorInput(input.MultiColorCandidate)
}

// multiColorCandidate returns the candidate to set to default.
func multiColorCandidate() *candidate {
	return &candidate{
		list: []string{
			"error info warn debug",
			"ERROR WARNING NOTICE INFO PANIC FATAL LOG",
		},
	}
}

// multiColorInput represents the multi color input mode.
type multiColorInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

func newMultiColorInput(clist *candidate) *multiColorInput {
	return &multiColorInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (d *multiColorInput) Prompt() string {
	return "multicolor:"
}

// Confirm returns the event when the input is confirmed.
func (d *multiColorInput) Confirm(str string) tcell.Event {
	d.value = str
	d.clist.list = toLast(d.clist.list, str)
	d.clist.p = 0
	d.SetEventNow()
	return d
}

// Up returns strings when the up key is pressed during input.
func (d *multiColorInput) Up(str string) string {
	return d.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (d *multiColorInput) Down(str string) string {
	return d.clist.down()
}
