package oviewer

import (
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
			st.style = csToStyle(st.style, es.parameter.String())
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
func csToStyle(style tcell.Style, params string) tcell.Style {
	if params == "0" || params == "" || params == ";" {
		return tcell.StyleDefault.Normal()
	}
	if s, ok := csiCache.Load(params); ok {
		style = applyStyle(style, s.(OVStyle))
		return style
	}

	s := parseCSI(params)
	csiCache.Store(params, s)
	style = applyStyle(style, s)
	return style
}

// parseCSI actually parses the style and returns ovStyle.
func parseCSI(params string) OVStyle {
	s := OVStyle{}
	fields := strings.Split(params, ";")
	fl := len(fields)
	for index := 0; index < fl; index++ {
		field := fields[index]
		switch field {
		case "1", "01":
			s.Bold = true
		case "2", "02":
			s.Dim = true
		case "3", "03":
			s.Italic = true
		case "4", "04":
			s.Underline = true
		case "5", "05":
			s.Blink = true
		case "6", "06":
			s.Blink = true
		case "7", "07":
			s.Reverse = true
		case "8", "08":
			// Invisible On (not implemented)
		case "9", "09":
			s.StrikeThrough = true
		case "22":
			s.UnBold = true
		case "23":
			s.UnItalic = true
		case "24":
			s.UnUnderline = true
		case "25":
			s.UnBlink = true
		case "27":
			s.UnReverse = true
		case "28":
			// Invisible Off (not implemented)
		case "29":
			s.UnStrikeThrough = true
		case "30", "31", "32", "33", "34", "35", "36", "37":
			colorNumber, _ := strconv.Atoi(field)
			s.Foreground = colorName(int(tcell.Color(colorNumber - 30)))
		case "39":
			s.Foreground = "default"
		case "40", "41", "42", "43", "44", "45", "46", "47":
			colorNumber, _ := strconv.Atoi(field)
			s.Background = colorName(int(tcell.Color(colorNumber - 40)))
		case "49":
			s.Background = "default"
		case "53":
			s.OverLine = true
		case "55":
			s.UnOverLine = true
		case "90", "91", "92", "93", "94", "95", "96", "97":
			colorNumber, _ := strconv.Atoi(field)
			s.Foreground = colorName(int(tcell.Color(colorNumber - 82)))
		case "100", "101", "102", "103", "104", "105", "106", "107":
			colorNumber, _ := strconv.Atoi(field)
			s.Background = colorName(int(tcell.Color(colorNumber - 92)))
		case "38", "48":
			var i int
			i, s = csColor(s, fields[index:])
			index += i
		}
	}
	return s
}

// csColor parses 8-bit color and 24-bit color.
func csColor(s OVStyle, fields []string) (int, OVStyle) {
	if len(fields) < 2 {
		return 1, s
	}

	var color string
	index := 0
	fg := fields[index]
	index++
	if fields[index] == "5" && len(fields) > index+1 { // 8-bit colors.
		index++
		c, _ := strconv.Atoi(fields[index])
		color = colorName(c)
	} else if fields[index] == "2" && len(fields) > index+3 { // 24-bit colors.
		red, _ := strconv.Atoi(fields[index+1])
		green, _ := strconv.Atoi(fields[index+2])
		blue, _ := strconv.Atoi(fields[index+3])
		index += 3
		color = fmt.Sprintf("#%02x%02x%02x", red, green, blue)
	}
	if len(color) > 0 {
		if fg == "38" {
			s.Foreground = color
		} else {
			s.Background = color
		}
	}
	return index, s
}

// colorName returns a string that can be used to specify the color of tcell.
func colorName(colorNumber int) string {
	if colorNumber <= 7 {
		return lookupColor(colorNumber)
	} else if colorNumber <= 15 {
		return lookupColor(colorNumber)
	} else if colorNumber <= 231 {
		red := (colorNumber - 16) / 36
		green := ((colorNumber - 16) / 6) % 6
		blue := (colorNumber - 16) % 6
		return fmt.Sprintf("#%02x%02x%02x", 255*red/5, 255*green/5, 255*blue/5)
	} else if colorNumber <= 255 {
		grey := 255 * (colorNumber - 232) / 23
		return fmt.Sprintf("#%02x%02x%02x", grey, grey, grey)
	}
	return ""
}

// lookupColor returns the color name from the color number.
func lookupColor(colorNumber int) string {
	if colorNumber < 0 || colorNumber > 15 {
		return "black"
	}
	return [...]string{
		"black",
		"maroon",
		"green",
		"olive",
		"navy",
		"purple",
		"teal",
		"silver",
		"gray",
		"red",
		"lime",
		"yellow",
		"blue",
		"fuchsia",
		"aqua",
		"white",
	}[colorNumber]
}
