package oviewer

import (
	"context"
	"strconv"

	"github.com/gdamore/tcell/v3"
)

// inputSignMode starts input for the display marker bitmask.
func (root *Root) inputSignMode(context.Context) {
	input := root.input
	input.reset()
	input.Candidate[SignModeInput].toLast(strconv.Itoa(root.Doc.SignMode))
	input.Event = newSignModeEvent(input.Candidate[SignModeInput])
}

func signModeCandidate() *candidate {
	return &candidate{list: []string{"0", "1", "2", "3", "4", "5", "6", "7"}}
}

// eventSignMode represents the display marker bitmask input mode.
type eventSignMode struct {
	tcell.EventTime
	clist *candidate
	value string
}

func newSignModeEvent(clist *candidate) *eventSignMode {
	return &eventSignMode{clist: clist}
}

func (*eventSignMode) Mode() InputMode { return SignModeInput }

func (*eventSignMode) Prompt() string { return "Sign mode (bitmask):" }

func (e *eventSignMode) Confirm(value string) tcell.Event {
	e.value = value
	e.clist.toLast(value)
	e.SetEventNow()
	return e
}

func (e *eventSignMode) Up(_ string) string { return e.clist.up() }

func (e *eventSignMode) Down(_ string) string { return e.clist.down() }
