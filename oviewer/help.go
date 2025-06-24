package oviewer

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/jwalton/gchalk"
)

// NewHelp generates a document for help.
func NewHelp(k KeyBind) (*Document, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}
	m.documentType = DocHelp
	helpStr := bytes.NewBufferString("\t\t\t" + gchalk.WithUnderline().Bold("ov help"))
	helpStr.WriteString("\n")
	helpStr.WriteString(KeyBindString(k))
	// Duplicate key bindings.
	helpStr.WriteString(DuplicateKeyBind(k))

	m.Caption = "Help"
	m.store.eof = 1
	m.preventReload = true
	m.seekable = false
	m.Header = 2
	m.SectionHeader = true
	m.setSectionDelimiter("\t")
	m.SectionHeaderNum = 1
	m.Style = NewHelpStyle()
	atomic.StoreInt32(&m.closed, 1)
	if err := m.ControlReader(helpStr, nil); err != nil {
		return nil, err
	}
	return m, err
}

// KeyBindString returns keybind as a string for help.
func KeyBindString(k KeyBind) string {
	s := bufio.NewScanner(bytes.NewBufferString(k.String()))
	var buf bytes.Buffer
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "\t") {
			line = gchalk.Bold(line)
		}
		fmt.Fprintln(&buf, line)
	}
	return buf.String()
}

// DuplicateKeyBind returns a string representing duplicate key bindings.
func DuplicateKeyBind(k KeyBind) string {
	w := &strings.Builder{}
	dupkey := findDuplicateKeyBind(k)
	if len(dupkey) == 0 {
		return ""
	}
	fmt.Fprintf(w, "\n\tDuplicate key bindings:\n")
	for _, v := range dupkey {
		v.key = strings.TrimPrefix(v.key, "input_")
		fmt.Fprintf(w, "%s [%s] for %s\n", gchalk.Red("Duplicate key"), v.key, strings.Join(v.action, ", "))
	}
	return w.String()
}

// NewHelpStyle returns a Style for the help document.
func NewHelpStyle() Style {
	return Style{
		Header: OVStyle{
			Foreground: "gold",
			Background: "darkblue",
			Bold:       true,
		},
		SectionLine: OVStyle{
			Background: "slateblue",
			Bold:       true,
		},
		Alternate: OVStyle{
			Background: "gray",
		},
		LineNumber: OVStyle{
			Bold: true,
		},
		SearchHighlight: OVStyle{
			Reverse: true,
		},
		MarkLine: OVStyle{
			Background: "darkgoldenrod",
		},
		MultiColorHighlight: []OVStyle{
			{Foreground: "red"},
			{Foreground: "aqua"},
			{Foreground: "yellow"},
			{Foreground: "fuchsia"},
			{Foreground: "lime"},
			{Foreground: "blue"},
			{Foreground: "grey"},
		},
		JumpTargetLine: OVStyle{
			Underline: true,
		},
	}
}
