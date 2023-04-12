package oviewer

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/jwalton/gchalk"
)

// NewHelp generates a document for help.
func NewHelp(k KeyBind) (*Document, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}

	m.append(m.chunks[len(m.chunks)-1], []byte("\t\t\t"+gchalk.WithUnderline().Bold("ov help")))

	str := strings.Split(KeyBindString(k), "\n")
	for _, s := range str {
		m.append(m.chunks[len(m.chunks)-1], []byte(s))
	}

	m.FileName = "Help"
	m.eof = 1
	m.preventReload = true
	m.seekable = false
	m.setSectionDelimiter("\t")
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
