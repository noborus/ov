package oviewer

import "github.com/gdamore/tcell/v2"

// normalInput represents the normal input mode.
// This is a dummy as it normally does not accept input.
type normalInput struct {
	tcell.EventTime
}

// newNormalInput returns normalInput.
func newNormalInput() *normalInput {
	return &normalInput{}
}

// Prompt returns the prompt string in the input field.
func (n *normalInput) Prompt() string {
	return ""
}

// Confirm returns the event when the input is confirmed.
func (n *normalInput) Confirm(str string) tcell.Event {
	return nil
}

// Up returns strings when the up key is pressed during input.
func (n *normalInput) Up(str string) string {
	return ""
}

// Down returns strings when the down key is pressed during input.
func (n *normalInput) Down(str string) string {
	return ""
}
