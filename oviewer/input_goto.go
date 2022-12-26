package oviewer

import "github.com/gdamore/tcell/v2"

// setGoLineMode sets the inputMode to Goline.
func (root *Root) setGoLineMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = Goline
	input.EventInput = newGotoInput(input.GoCandidate)
}

// gotoCandidate returns the candidate to set to default.
func gotoCandidate() *candidate {
	return &candidate{
		list: []string{},
	}
}

// gotoInput represents the goto input mode.
type gotoInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newGotoInput returns GotoInput.
func newGotoInput(clist *candidate) *gotoInput {
	return &gotoInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (g *gotoInput) Prompt() string {
	return "Goto line:"
}

// Confirm returns the event when the input is confirmed.
func (g *gotoInput) Confirm(str string) tcell.Event {
	g.value = str
	g.clist.list = toLast(g.clist.list, str)
	g.clist.p = 0
	g.SetEventNow()
	return g
}

// Up returns strings when the up key is pressed during input.
func (g *gotoInput) Up(str string) string {
	return g.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (g *gotoInput) Down(str string) string {
	return g.clist.down()
}
