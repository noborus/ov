package oviewer

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// content represents one character on the terminal.
// content is a value that can be set in SetContent() of tcell.
type content struct {
	style tcell.Style
	combc []rune
	width int
	mainc rune
}

// contents represents one line of contents.
type contents []content

// The states of the ANSI escape code parser.
const (
	ansiText = iota
	ansiEscape
	ansiSubstring
	ansiControlSequence
	systemSequence
	oscHyperLink
	oscParameter
	oscURL
)

// DefaultContent is a blank Content.
var DefaultContent = content{
	mainc: 0,
	combc: nil,
	width: 0,
	style: tcell.StyleDefault,
}

// EOFC is the EOF character.
const EOFC rune = '~'

// EOFContent is EOFC only.
var EOFContent = content{
	mainc: EOFC,
	combc: nil,
	width: 1,
	style: tcell.StyleDefault.Foreground(tcell.ColorGray),
}

// csicache caches escape sequences.
var csiCache sync.Map

// parseState represents the affected state after parsing.
type parseState struct {
	state     int
	parameter strings.Builder
	url       strings.Builder
	style     tcell.Style
	tabx      int
	bsFlag    bool // backspace(^H) flag
	bsContent content
}

// parseString converts a string to lineContents.
// parseString includes escape sequences and tabs.
// If tabwidth is set to -1, \t is displayed instead of functioning as a tab.
func parseString(str string, tabWidth int) contents {
	lc := make(contents, 0, len(str))
	state := &parseState{
		state:     ansiText,
		parameter: strings.Builder{},
		url:       strings.Builder{},
		style:     tcell.StyleDefault,
		tabx:      0,
		bsFlag:    false,
		bsContent: DefaultContent,
	}

	gr := uniseg.NewGraphemes(str)
	for gr.Next() {
		r := gr.Runes()
		mainc := r[0]
		combc := r[1:]

		if state.parseEscapeSequence(mainc) {
			continue
		}

		c := DefaultContent
		switch runewidth.RuneWidth(mainc) {
		case 0:
			switch {
			case mainc == '\t': // TAB
				switch {
				case tabWidth > 0:
					tabStop := tabWidth - (state.tabx % tabWidth)
					c.width = 1
					c.style = state.style
					c.mainc = rune('\t')
					lc = append(lc, c)
					state.tabx++
					c.mainc = 0
					for i := 0; i < tabStop-1; i++ {
						lc = append(lc, c)
						state.tabx++
					}
				case tabWidth < 0:
					c.width = 1
					c.style = state.style.Reverse(true)
					c.mainc = rune('\\')
					lc = append(lc, c)
					c.mainc = rune('t')
					lc = append(lc, c)
					state.tabx += 2
				default:
				}
				continue
			case mainc == '\b': // BackSpace
				if len(lc) == 0 {
					continue
				}
				state.bsFlag = true
				state.bsContent = lc.last()
				if state.bsContent.width > 1 {
					lc = lc[:len(lc)-2]
				} else {
					lc = lc[:len(lc)-1]
				}
				continue
			case mainc < 0x20: // control character
				c.mainc = mainc
				c.width = 0
				lc = append(lc, c)
				continue
			}
			lastC := lc.last()
			lastC.combc = append(lastC.combc, mainc)
			n := len(lc) - lastC.width
			if n >= 0 && len(lc) > n {
				lc[n] = lastC
			}
		case 1:
			c.mainc = mainc
			if len(combc) > 0 {
				c.combc = combc
			}
			c.width = 1
			c.style = state.overstrike(c, state.style)
			lc = append(lc, c)
			state.tabx++
		case 2:
			c.mainc = mainc
			if len(combc) > 0 {
				c.combc = combc
			}
			c.width = 2
			c.style = state.overstrike(c, state.style)
			lc = append(lc, c, DefaultContent)
			state.tabx += 2
		}
	}
	return lc
}

// parseEscapeSequence parses an escape sequence and changes state.
// Returns true if it is an escape sequence and a non-printing character.
func (es *parseState) parseEscapeSequence(mainc rune) bool {
	switch es.state {
	case ansiEscape:
		switch mainc {
		case '[': // Control Sequence Introducer.
			es.parameter.Reset()
			es.state = ansiControlSequence
			return true
		case 'c': // Reset.
			es.style = tcell.StyleDefault
			es.state = ansiText
			return true
		case ']': // Operting System Command Sequence.
			es.state = systemSequence
			return true
		case 'P', 'X', '^', '_': // Substrings and commands.
			es.state = ansiSubstring
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
		if mainc == 'm' {
			es.style = csToStyle(es.style, es.parameter.String())
		} else if mainc >= 'A' && mainc <= 'T' {
			// Ignore.
		} else {
			if mainc >= 0x30 && mainc <= 0x3f {
				es.parameter.WriteRune(mainc)
				return true
			}
		}
		es.state = ansiText
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
		urlid := es.parameter.String()
		if urlid != "" {
			es.style = es.style.UrlId(urlid)
		}
		es.parameter.Reset()
		es.state = oscURL
		return true
	case oscURL:
		switch mainc {
		case 0x1b:
			es.style = es.style.Url(es.url.String())
			es.url.Reset()
			es.state = systemSequence
			return true
		case 0x07:
			es.style = es.style.Url(es.url.String())
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
		return true
	}
	return false
}

// overstrike set style for overstrike.
func (state *parseState) overstrike(m content, style tcell.Style) tcell.Style {
	if !state.bsFlag {
		return style
	}

	if state.bsContent.mainc == m.mainc {
		style = OverStrikeStyle
	} else if state.bsContent.mainc == '_' {
		style = OverLineStyle
	}
	state.bsFlag = false
	state.bsContent = DefaultContent
	return style
}

// last returns the last character of Contents.
func (lc contents) last() content {
	n := len(lc)
	if n == 0 {
		return content{}
	}
	if (n > 1) && (lc[n-2].width > 1) {
		return lc[n-2]
	}
	return lc[n-1]
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

// StrToContents converts a single-line string into a one line of contents.
// Parse escape sequences, etc.
// 1 Content matches the characters displayed on the screen.
func StrToContents(str string, tabWidth int) contents {
	return parseString(str, tabWidth)
}

type widthPos []int

// ContentsToStr returns a converted string
// and byte position, as well as the content position conversion table.
func ContentsToStr(lc contents) (string, widthPos) {
	var buff strings.Builder
	pos := make(widthPos, 0, len(lc)*4)

	i, bn := 0, 0
	for n, c := range lc {
		if c.mainc == 0 {
			continue
		}
		for ; i <= bn; i++ {
			pos = append(pos, n)
		}
		mn, err := buff.WriteRune(c.mainc)
		if err != nil {
			log.Println(err)
		}
		bn += mn
		for _, r := range c.combc {
			cn, err := buff.WriteRune(r)
			if err != nil {
				log.Println(err)
			}
			bn += cn
		}
	}
	str := buff.String()
	for ; i <= bn; i++ {
		pos = append(pos, len(lc))
	}
	return str, pos
}

// x returns the x position on the screen.
func (pos widthPos) x(x int) int {
	if x < len(pos) {
		return pos[x]
	}
	return pos[len(pos)-1]
}
