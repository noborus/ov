package oviewer

import (
	"context"
	"strconv"

	"github.com/gdamore/tcell/v3"
)

// inputTabWidth sets the inputMode to TabWidth.
func (root *Root) inputTabWidth(context.Context) {
	input := root.input
	input.reset()
	input.Candidate[TabWidth].toLast(strconv.Itoa(root.Doc.TabWidth))
	input.Event = newTabWidthEvent(input.Candidate[TabWidth])
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

// eventTabWidth represents the TABWidth input mode.
type eventTabWidth struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newTabWidthEvent returns tabWidthEvent.
func newTabWidthEvent(clist *candidate) *eventTabWidth {
	return &eventTabWidth{clist: clist}
}

// Mode returns InputMode.
func (*eventTabWidth) Mode() InputMode {
	return TabWidth
}

// Prompt returns the prompt string in the input field.
func (*eventTabWidth) Prompt() string {
	return "TAB width:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventTabWidth) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventTabWidth) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventTabWidth) Down(_ string) string {
	return e.clist.down()
}
