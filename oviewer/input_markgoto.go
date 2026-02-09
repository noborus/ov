package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v3"
)

// inputMarkNumber sets the inputMode to Mark number.
func (root *Root) inputMarkNumber(ctx context.Context) {
	root.previousSidebarMode = root.sidebarMode
	root.openSidebar(ctx, SidebarModeMarks)
	input := root.input
	input.reset()
	input.Event = newrMarkGotoEvent(input.Candidate[MarkNum])
}

// eventMarkGoto represents the rMarkGoto input mode.
type eventMarkGoto struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newrMarkGotoEvent returns rMarkGotoEvent.
func newrMarkGotoEvent(clist *candidate) *eventMarkGoto {
	return &eventMarkGoto{clist: clist}
}

// Mode returns InputMode.
func (*eventMarkGoto) Mode() InputMode {
	return MarkNum
}

// Prompt returns the prompt string in the input field.
func (*eventMarkGoto) Prompt() string {
	return "Mark number:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventMarkGoto) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventMarkGoto) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventMarkGoto) Down(_ string) string {
	return e.clist.down()
}
