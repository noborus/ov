package oviewer

import "github.com/gdamore/tcell/v2"

// setJumpTargetMode sets the inputMode to JumpTarget.
func (root *Root) setJumpTargetMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = JumpTarget
	input.EventInput = newJumpTargetInput(input.JumpTargetCandidate)
}

// jumpTargetCandidate returns the candidate to set to default.
func jumpTargetCandidate() *candidate {
	return &candidate{
		list: []string{},
	}
}

// jumpTargetInput represents the jump target input mode.
type jumpTargetInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newJumpTargetInput returns JumpTargetInput.
func newJumpTargetInput(clist *candidate) *jumpTargetInput {
	return &jumpTargetInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (j *jumpTargetInput) Prompt() string {
	return "Jump Target line:"
}

// Confirm returns the event when the input is confirmed.
func (j *jumpTargetInput) Confirm(str string) tcell.Event {
	j.value = str
	j.clist.list = toLast(j.clist.list, str)
	j.clist.p = 0
	j.SetEventNow()
	return j
}

// Up returns strings when the up key is pressed during input.
func (g *jumpTargetInput) Up(str string) string {
	return g.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (g *jumpTargetInput) Down(str string) string {
	return g.clist.down()
}
