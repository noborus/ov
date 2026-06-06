package oviewer

import (
	"context"

	"github.com/gdamore/tcell/v3"
)

// inputStyleToggle sets the inputMode to style highlight suppression toggle.
func (root *Root) inputStyleToggle(ctx context.Context) {
	root.openSidebar(ctx, SidebarModeStyles)
	input := root.input
	input.reset()
	root.backupStyleFlags()
	input.Event = newStyleToggleEvent(input.Candidate[StyleToggle])
}

// backupStyleFlags backs up the current style flags.
func (root *Root) backupStyleFlags() {
	if root.Doc.styles.Len() == 0 {
		return
	}
	root.Doc.backupStyleFlags = make([]bool, root.Doc.styles.Len())
	for i := 0; i < root.Doc.styles.Len(); i++ {
		_, b, ok := root.Doc.styles.Index(i)
		if !ok {
			continue
		}
		root.Doc.backupStyleFlags[i] = b
	}
}

// eventStyleToggle represents the style highlight suppression input mode.
type eventStyleToggle struct {
	tcell.EventTime
	clist *candidate
	value string
}

// newStyleToggleEvent returns a new eventStyleToggle with the given candidate list.
func newStyleToggleEvent(clist *candidate) *eventStyleToggle {
	return &eventStyleToggle{clist: clist}
}

// Mode returns InputMode.
func (*eventStyleToggle) Mode() InputMode {
	return StyleToggle
}

// Prompt returns the prompt string in the input field.
func (*eventStyleToggle) Prompt() string {
	return "Style number(suppress highlight):"
}

// Confirm returns the event when the input is confirmed.
func (e *eventStyleToggle) Confirm(str string) tcell.Event {
	e.value = str
	e.clist.toLast(str)
	e.SetEventNow()
	return e
}

// Up returns strings when the up key is pressed during input.
func (e *eventStyleToggle) Up(_ string) string {
	return e.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (e *eventStyleToggle) Down(_ string) string {
	return e.clist.down()
}
