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

	m.Caption = "Help"
	m.store.eof = 1
	m.preventReload = true
	m.seekable = false
	m.SectionHeader = true
	m.setSectionDelimiter("\t")
	m.SectionHeaderNum = 1
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
