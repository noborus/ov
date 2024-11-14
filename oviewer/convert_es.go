package oviewer

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
)

// The states of the ANSI escape code parser.
const (
	ansiText = iota
	ansiEscape
	ansiSubstring
	ansiControlSequence
	otherSequence
	systemSequence
	oscHyperLink
	oscParameter
	oscURL
)

// escape sequence states.
type escapeSequence struct {
	parameter strings.Builder
	url       strings.Builder
	state     int
}

// newESConverter returns a new escape sequence converter.
func newESConverter() *escapeSequence {
	return &escapeSequence{
		parameter: strings.Builder{},
		url:       strings.Builder{},
		state:     ansiText,
	}
}

// csiCache caches escape sequences.
var csiCache sync.Map

// convert parses an escape sequence and changes state.
// Returns true if it is an escape sequence and a non-printing character.
func (es *escapeSequence) convert(st *parseState) bool {
	mainc := st.mainc
	switch es.state {
	case ansiEscape:
		switch mainc {
		case '[': // Control Sequence Introducer.
			es.parameter.Reset()
			es.state = ansiControlSequence
			return true
		case 'c': // Reset.
			st.style = tcell.StyleDefault
			es.state = ansiText
			return true
		case ']': // Operating System Command Sequence.
			es.state = systemSequence
			return true
		case 'P', 'X', '^', '_': // Substrings and commands.
			es.state = ansiSubstring
			return true
		case '(':
			es.state = otherSequence
			return true
		default: // Ignore.
			es.state = ansiText
		}
	case ansiSubstring:
		if mainc == 0x1b {
			es.state = ansiControlSequence
		}
		return true
	case ansiControlSequence:
		switch {
		case mainc == 'm':
			style, err := csToStyle(st.style, es.parameter.String())
			if err != nil {
				es.state = ansiText
				return false
			}
			st.style = style
		case mainc == 'K':
			// CSI 0 K or CSI K maintains the style after the newline
			// (can change the background color of the line).
			params := es.parameter.String()
			if params == "" || params == "0" {
				st.eolStyle = st.style
			}
		case mainc >= 'A' && mainc <= 'T':
			// Ignore.
		case mainc >= '0' && mainc <= 'f':
			es.parameter.WriteRune(mainc)
			return true
		case mainc < 0x40:
			es.parameter.WriteRune(mainc)
			return true
		}
		es.state = ansiText
		return true
	case otherSequence:
		es.state = ansiEscape
		return true
	case systemSequence:
		switch mainc {
		case '8':
			es.state = oscHyperLink
			return true
		case '\\':
			es.state = ansiText
			return true
		case 0x1b:
			// unknown but for compatibility.
			es.state = ansiControlSequence
			return true
		case 0x07:
			es.state = ansiText
			return true
		}
		log.Printf("invalid char %c", mainc)
		return true
	case oscHyperLink:
		switch mainc {
		case ';':
			es.state = oscParameter
			return true
		}
		es.state = ansiText
		return false
	case oscParameter:
		if mainc != ';' {
			es.parameter.WriteRune(mainc)
			return true
		}
		urlID := es.parameter.String()
		if urlID != "" {
			st.style = st.style.UrlId(urlID)
		}
		es.parameter.Reset()
		es.state = oscURL
		return true
	case oscURL:
		switch mainc {
		case 0x1b:
			st.style = st.style.Url(es.url.String())
			es.url.Reset()
			es.state = systemSequence
			return true
		case 0x07:
			st.style = st.style.Url(es.url.String())
			es.url.Reset()
			es.state = ansiText
			return true
		}
		es.url.WriteRune(mainc)
		return true
	}
	switch mainc {
	case 0x1b:
		es.state = ansiEscape
		return true
	case '\n':
		return false
	}
	return false
}

// csToStyle returns tcell.Style from the control sequence.
func csToStyle(style tcell.Style, params string) (tcell.Style, error) {
	if params == "0" || params == "" || params == ";" {
		return tcell.StyleDefault.Normal(), nil
	}

	if s, ok := csiCache.Load(params); ok {
		style = applyStyle(style, s.(OVStyle))
		return style, nil
	}

	s, err := parseCSI(params)
	if err != nil {
		return style, err
	}
	csiCache.Store(params, s)
	return applyStyle(style, s), nil
}

// parseCSI actually parses the style and returns ovStyle.
func parseCSI(params string) (OVStyle, error) {
	s := OVStyle{}
	fields := strings.Split(params, ";")
	for index := 0; index < len(fields); index++ {
		field := fields[index]
		num, err := toESCode(field)
		if err != nil {
			if errors.Is(err, ErrNotSuuport) {
				return s, nil
			}
			return s, err
		}
		switch num {
		case 0:
			s = OVStyle{}
		case 1:
			s.Bold = true
		case 2:
			s.Dim = true
		case 3:
			s.Italic = true
		case 4:
			s.Underline = true
		case 5:
			s.Blink = true
		case 6:
			s.Blink = true
		case 7:
			s.Reverse = true
		case 8:
			// Invisible On (not implemented)
		case 9:
			s.StrikeThrough = true
		case 22:
			s.UnBold = true
		case 23:
			s.UnItalic = true
		case 24:
			s.UnUnderline = true
		case 25:
			s.UnBlink = true
		case 27:
			s.UnReverse = true
		case 28:
			// Invisible Off (not implemented)
		case 29:
			s.UnStrikeThrough = true
		case 30, 31, 32, 33, 34, 35, 36, 37:
			s.Foreground = colorName(num - 30)
		case 38:
			i, color := csColor(fields[index:])
			if i == 0 {
				return s, nil
			}
			index += i
			s.Foreground = color
		case 39:
			s.Foreground = "default"
		case 40, 41, 42, 43, 44, 45, 46, 47:
			s.Background = colorName(num - 40)
		case 48:
			i, color := csColor(fields[index:])
			if i == 0 {
				return s, nil
			}
			index += i
			s.Background = color
		case 49:
			s.Background = "default"
		case 53:
			s.OverLine = true
		case 55:
			s.UnOverLine = true
		case 90, 91, 92, 93, 94, 95, 96, 97:
			s.Foreground = colorName(num - 82)
		case 100, 101, 102, 103, 104, 105, 106, 107:
			s.Background = colorName(num - 92)
		}
	}
	return s, nil
}

// toESCode converts a string to an integer.
// If the code is smaller than 0x40 for compatibility,
// return -1 instead of an error.
func toESCode(str string) (int, error) {
	num, err := strconv.Atoi(str)
	if err != nil {
		for _, char := range str {
			if char >= 0x40 {
				return 0, ErrInvalidCSI
			}
		}
		return 0, ErrNotSuuport
	}
	return num, nil
}

// csColor parses 8-bit color and 24-bit color.
func csColor(fields []string) (int, string) {
	if len(fields) < 2 {
		return 0, ""
	}
	if fields[1] == "" {
		return 1, ""
	}
	ex, err := strconv.Atoi(fields[1])
	if err != nil {
		return 1, ""
	}
	switch ex {
	case 5: // 8-bit colors.
		if len(fields) < 3 {
			return 1, ""
		}
		if fields[2] == "" {
			return len(fields), ""
		}

		color, err := parse8BitColor(fields[2])
		if err != nil {
			return 0, color
		}
		return 2, color
	case 2: // 24-bit colors.
		if len(fields) < 5 {
			return len(fields), ""
		}
		for i := 2; i < 5; i++ {
			if fields[i] == "" {
				return i, ""
			}
		}

		color, err := parseRGBColor(fields[2:5])
		if err != nil {
			return 0, color
		}
		return 4, color
	}
	return 1, ""
}

func parse8BitColor(field string) (string, error) {
	c, err := strconv.Atoi(field)
	if err != nil {
		return "", fmt.Errorf("invalid 8-bit color value: %v", field)
	}
	color := colorName(c)
	return color, nil
}

func parseRGBColor(fields []string) (string, error) {
	red, err1 := strconv.Atoi(fields[0])
	green, err2 := strconv.Atoi(fields[1])
	blue, err3 := strconv.Atoi(fields[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return "", fmt.Errorf("invalid RGB color values: %v, %v, %v", fields[0], fields[1], fields[2])
	}
	color := fmt.Sprintf("#%02x%02x%02x", red, green, blue)
	return color, nil
}

// colorName returns a string that can be used to specify the color of tcell.
func colorName(colorNumber int) string {
	if colorNumber < 0 || colorNumber > 255 {
		return ""
	}
	return tcell.PaletteColor(colorNumber).String()
}
