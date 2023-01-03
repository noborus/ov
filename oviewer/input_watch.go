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

// eventWatchInterval represents the WatchInteval input mode.
type eventWatchInterval struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newWatchIntevalInputt returns watchIntervalEvent.
func newWatchIntevalEvent(clist *candidate) *eventWatchInterval {
	return &eventWatchInterval{clist: clist}
}

// Mode returns InputMode.
func (e *eventWatchInterval) Mode() InputMode {
	return Watch
}

// Prompt returns the prompt string in the input field.
func (e *eventWatchInterval) Prompt() string {
	return "Watch interval:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventWatchInterval) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.list = toLast(e.clist.list, str)
	e.clist.p = 0
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventWatchInterval) Up(str string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventWatchInterval) Down(str string) string {
	return e.clist.down()
}
