package oviewer

import "github.com/gdamore/tcell/v2"

// setJumpTargetMode sets the inputMode to JumpTarget.
func (root *Root) setJumpTargetMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.Event = newJumpTargetEvent(input.JumpTargetCandidate)
}

// jumpTargetCandidate returns the candidate to set to default.
func jumpTargetCandidate() *candidate {
	return &candidate{
		list: []string{},
	}
}

// jumpTargetEvent represents the jump target input mode.
type jumpTargetEvent struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newJumpTargetEvent returns jumpTargetEvent.
func newJumpTargetEvent(clist *candidate) *jumpTargetEvent {
	return &jumpTargetEvent{clist: clist}
}

// Mode returns InputMode.
func (e *jumpTargetEvent) Mode() InputMode {
	return JumpTarget
}

// Prompt returns the prompt string in the input field.
func (e *jumpTargetEvent) Prompt() string {
	return "Jump Target line:"
}

// Confirm returns the event when the input is confirmed.
func (e *jumpTargetEvent) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *jumpTargetEvent) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *jumpTargetEvent) Down(str string) string {
	return e.clist.down()
}
