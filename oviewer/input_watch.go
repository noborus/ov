package oviewer

import "github.com/gdamore/tcell/v2"

// setWatchIntervalMode sets the inputMode to Watch.
func (root *Root) setWatchIntervalMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = Watch
	input.EventInput = newWatchIntevalInput(input.WatchCandidate)
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

// watchIntervalInput represents the WatchInteval input mode.
type watchIntervalInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newWatchIntevalInputt returns WatchIntevalInput.
func newWatchIntevalInput(clist *candidate) *watchIntervalInput {
	return &watchIntervalInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (t *watchIntervalInput) Prompt() string {
	return "Watch interval:"
}

// Confirm returns the event when the input is confirmed.
func (t *watchIntervalInput) Confirm(str string) tcell.Event {
	t.value = str
	t.clist.list = toLast(t.clist.list, str)
	t.clist.p = 0
	t.SetEventNow()
	return t
}

// Up returns strings when the up key is pressed during input.
func (t *watchIntervalInput) Up(str string) string {
	return t.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (t *watchIntervalInput) Down(str string) string {
	return t.clist.down()
}
