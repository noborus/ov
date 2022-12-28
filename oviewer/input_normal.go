package oviewer

import "github.com/gdamore/tcell/v2"

// normalEvent represents the normal input mode.
// This is a dummy as it normally does not accept input.
type normalEvent struct {
	tcell.EventTime
}

// normal returns normalEvent.
func normal() *normalEvent {
	return &normalEvent{}
}

// Mode returns InputMode.
func (e *normalEvent) Mode() InputMode {
	return Normal
}

// Prompt returns the prompt string in the input field.
func (e *normalEvent) Prompt() string {
	return ""
}

// Confirm returns the event when the input is confirmed.
func (e *normalEvent) Confirm(str string) tcell.Event {
	return nil
}

// Up returns strings when the up key is pressed during input.
func (e *normalEvent) Up(str string) string {
	return ""
}

// Down returns strings when the down key is pressed during input.
func (e *normalEvent) Down(str string) string {
	return ""
}
