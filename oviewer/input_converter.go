package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// setConvertType sets the convert type input mode.
func (root *Root) setConvertType(context.Context) {
	input := root.input
	input.reset()
	input.Event = newConvertTypeEvent(input.Candidate[ConvertType])
}

// converterCandidate returns the candidate to set to default.
func converterCandidate() *candidate {
	return &candidate{
		list: []string{
			convEscaped,
			convRaw,
			convAlign,
		},
	}
}

// eventViewMode represents the view mode input mode.
type eventConverter struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newConvertEvent returns ConvertTypeEvent.
func newConvertTypeEvent(clist *candidate) *eventConverter {
	return &eventConverter{clist: clist}
}

// Mode returns InputMode.
func (*eventConverter) Mode() InputMode {
	return ConvertType
}

// Prompt returns the prompt string in the input field.
func (*eventConverter) Prompt() string {
	return "Convert:"
}

// Confirm returns the event when the input is confirmed.
func (e *eventConverter) Confirm(str string) tcell.Event {
	e.value = str
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventConverter) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventConverter) Down(_ string) string {
	return e.clist.down()
}
