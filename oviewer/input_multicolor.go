package oviewer

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// searchCandidateListLen is the number of search candidates for multicolor.
const searchCandidateListLen = 10

// setMultiColorMode sets the inputMode to MultiColor.
func (root *Root) setMultiColorMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0

	list := root.searchCandidates(searchCandidateListLen)
	str := strings.Join(list, " ")
	input.MultiColorCandidate.toLast(str)

	input.Event = newMultiColorEvent(input.MultiColorCandidate)
}

// multiColorCandidate returns the candidate to set to default.
func multiColorCandidate() *candidate {
	return &candidate{
		list: []string{
			"error info warn debug",
			"ERROR WARNING NOTICE INFO PANIC FATAL LOG",
		},
	}
}

// eventMultiColor represents the multi color input mode.
type eventMultiColor struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newMultiColorEvent returns multiColorEvent.
func newMultiColorEvent(clist *candidate) *eventMultiColor {
	return &eventMultiColor{clist: clist}
}

// Mode returns InputMode.
func (e *eventMultiColor) Mode() InputMode {
	return MultiColor
}

// Prompt returns the prompt string in the input field.
func (e *eventMultiColor) Prompt() string {
	return "multicolor:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventMultiColor) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventMultiColor) Up(str string) string {
	e.clist.toAddLast(str)
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventMultiColor) Down(str string) string {
	e.clist.toAddTop(str)
	return e.clist.down()
}
