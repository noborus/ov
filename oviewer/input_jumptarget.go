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
		list: []string{
			"section",
		},
	}
}

// eventJumpTarget represents the jump target input mode.
type eventJumpTarget struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newJumpTargetEvent returns jumpTargetEvent.
func newJumpTargetEvent(clist *candidate) *eventJumpTarget {
	return &eventJumpTarget{clist: clist}
}

// Mode returns InputMode.
func (e *eventJumpTarget) Mode() InputMode {
	return JumpTarget
}

// Prompt returns the prompt string in the input field.
func (e *eventJumpTarget) Prompt() string {
	return "Jump Target line:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventJumpTarget) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventJumpTarget) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventJumpTarget) Down(str string) string {
	return e.clist.down()
}
