package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v3"
)

// inputBreakIndent starts input for the indentation of wrapped lines.
func (root *Root) inputBreakIndent(context.Context) {
	input := root.input
	input.reset()
	input.Candidate[BreakIndentInput].toLast(root.Doc.BreakIndent)
	input.Event = newBreakIndentEvent(input.Candidate[BreakIndentInput])
}

// eventBreakIndent represents the wrapped-line indentation input mode.
type eventBreakIndent struct {
	tcell.EventTime
	clist *candidate
	value string
}

func newBreakIndentEvent(clist *candidate) *eventBreakIndent {
	return &eventBreakIndent{clist: clist}
}

func (*eventBreakIndent) Mode() InputMode { return BreakIndentInput }

func (*eventBreakIndent) Prompt() string { return "Break indent:" }

func (e *eventBreakIndent) Confirm(value string) tcell.Event {
	e.value = value
	e.clist.toLast(value)
	e.SetEventNow()
	return e
}

func (e *eventBreakIndent) Up(_ string) string { return e.clist.up() }

func (e *eventBreakIndent) Down(_ string) string { return e.clist.down() }
