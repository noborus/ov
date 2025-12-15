package oviewer

import "github.com/gdamore/tcell/v3"

// eventNormal represents the normal input mode.
// This is a dummy as it normally does not accept input.
type eventNormal struct {
	tcell.EventTime
}

// normal returns normalEvent.
func normal() *eventNormal {
	return &eventNormal{}
}

// Mode returns InputMode.
func (*eventNormal) Mode() InputMode {
	return Normal
}

// Prompt returns the prompt string in the input field.
func (*eventNormal) Prompt() string {
	return ""
}

// Confirm returns the event when the input is confirmed.
func (*eventNormal) Confirm(_ string) tcell.Event {
	return nil
}

// Up returns strings when the up key is pressed during input.
func (*eventNormal) Up(_ string) string {
	return ""
}

// Down returns strings when the down key is pressed during input.
func (*eventNormal) Down(_ string) string {
	return ""
}
