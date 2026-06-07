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
	root.Doc.backupStyleFlags()
	input.Event = newStyleToggleEvent(input.Candidate[StyleToggle])
}

// backupStyleFlags backs up the current style flags.
func (m *Document) backupStyleFlags() {
	if m.styles.Len() == 0 {
		m.styleFlags = nil
		return
	}
	m.styleFlags = make([]bool, m.styles.Len())
	for i := 0; i < m.styles.Len(); i++ {
		_, b, ok := m.styles.Index(i)
		if !ok {
			continue
		}
		m.styleFlags[i] = b
	}
}

// restoreStyleFlags restores the style flags from the backup.
func (m *Document) restoreStyleFlags() {
	if m.styleFlags == nil {
		return
	}
	m.styles.SetValues(m.styleFlags)
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
