package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// inputWatchInterval sets the inputMode to Watch.
func (root *Root) inputWatchInterval(context.Context) {
	input := root.input
	input.reset()
	input.Event = newWatchIntervalEvent(input.Candidate[Watch])
}

// watchCandidate returns the candidate to set to default.
func watchCandidate() *candidate {
	return &candidate{
		list: []string{
			"3",
			"2",
			"1",
		},
	}
}

// eventWatchInterval represents the WatchInterval input mode.
type eventWatchInterval struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newWatchIntervalInput returns watchIntervalEvent.
func newWatchIntervalEvent(clist *candidate) *eventWatchInterval {
	return &eventWatchInterval{clist: clist}
}

// Mode returns InputMode.
func (*eventWatchInterval) Mode() InputMode {
	return Watch
}

// Prompt returns the prompt string in the input field.
func (*eventWatchInterval) Prompt() string {
	return "Watch interval:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventWatchInterval) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventWatchInterval) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventWatchInterval) Down(_ string) string {
	return e.clist.down()
}
