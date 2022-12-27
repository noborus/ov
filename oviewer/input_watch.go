package oviewer

import "github.com/gdamore/tcell/v2"

// setWatchIntervalMode sets the inputMode to Watch.
func (root *Root) setWatchIntervalMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.Event = newWatchIntevalEvent(input.WatchCandidate)
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

// watchIntervalEvent represents the WatchInteval input mode.
type watchIntervalEvent struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newWatchIntevalInputt returns watchIntervalEvent.
func newWatchIntevalEvent(clist *candidate) *watchIntervalEvent {
	return &watchIntervalEvent{clist: clist}
}

// Mode returns InputMode.
func (e *watchIntervalEvent) Mode() InputMode {
	return Watch
}

// Prompt returns the prompt string in the input field.
func (e *watchIntervalEvent) Prompt() string {
	return "Watch interval:"
}

// Confirm returns the event when the input is confirmed.
func (e *watchIntervalEvent) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *watchIntervalEvent) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *watchIntervalEvent) Down(str string) string {
	return e.clist.down()
}
