package oviewer

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/ristretto"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

// The Model structure contains the values
// for the logical screen.
type Model struct {
	// updated by reader goroutine
	buffer []string
	endNum int
	eof    bool

	x          int
	lineNum    int
	yy         int
	header     []string
	beforeSize int
	vWidth     int
	vHight     int
	cache      *ristretto.Cache

	mu sync.Mutex
}

//NewModel returns Model.
func NewModel() *Model {
	return &Model{
		buffer:     make([]string, 0),
		header:     make([]string, 0),
		beforeSize: 1000,
	}
}

// content represents one character on the terminal.
type content struct {
	width int
	style tcell.Style
	mainc rune
	combc []rune
}

type lineContents struct {
	contents []content
	cMap     map[int]int
}

// GetLine returns one line from buffer.
func (m *Model) GetLine(lineNum int) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if lineNum < 0 || lineNum >= len(m.buffer) {
		return ""
	}
	return m.buffer[lineNum]
}

// BufLen return buffer length.
func (m *Model) BufLen() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.buffer)
}

// BuffEndNum return last line number.
func (m *Model) BufEndNum() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.endNum
}

// BufEOF return reaching EOF.
func (m *Model) BufEOF() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.eof
}

// GetContents returns one line of contents from buffer.
func (m *Model) GetContents(lineNum int, tabWidth int) []content {
	lc, err := m.lineToContents(lineNum, tabWidth)
	if err != nil {
		return nil
	}
	return lc.contents
}

// NewCache creates a new cache.
func (m *Model) NewCache() error {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return err
	}
	m.cache = cache
	return nil
}

// ClearCache clears the cache.
func (m *Model) ClearCache() {
	m.cache.Clear()
}

func (m *Model) lineToContents(lineNum int, tabWidth int) (lineContents, error) {
	var lc lineContents
	if lineNum < 0 || lineNum >= m.BufLen() {
		return lc, fmt.Errorf("out of range")
	}
	value, found := m.cache.Get(lineNum)
	if found {
		var ok bool
		if lc, ok = value.(lineContents); !ok {
			return lc, fmt.Errorf("fatal error not lineContents")
		}
	} else {
		lc.contents, lc.cMap = parseString(m.GetLine(lineNum), tabWidth)
		m.cache.Set(lineNum, lc, 1)
	}
	return lc, nil
}

// The states of the ANSI escape code parser.
const (
	ansiText = iota
	ansiEscape
	ansiSubstring
	ansiControlSequence
)

func lookupColor(colorNumber int, bright bool) string {
	if colorNumber < 0 || colorNumber > 7 {
		return "black"
	}
	if bright {
		colorNumber += 8
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

func csToStyle(csiParameter *bytes.Buffer, runeValue rune) tcell.Style {
	style := tcell.StyleDefault
	fields := strings.Split(csiParameter.String(), ";")
	if len(fields) == 0 || len(fields) == 1 && fields[0] == "0" {
		style = style.Normal()
	}
FieldLoop:
	for index, field := range fields {
		switch field {
		case "1", "01":
			style = style.Bold(true)
		case "2", "02":
			style = style.Dim(true)
		case "4", "04":
			style = style.Underline(true)
		case "5", "05":
			style = style.Blink(true)
		case "7", "07":
			style = style.Reverse(true)
		case "22", "24", "25", "27":
			style = style.Normal()
		case "30", "31", "32", "33", "34", "35", "36", "37":
			colorNumber, _ := strconv.Atoi(field)
			style = style.Foreground(tcell.Color(colorNumber - 30))
		case "39":
			style = style.Foreground(tcell.ColorDefault)
		case "40", "41", "42", "43", "44", "45", "46", "47":
			colorNumber, _ := strconv.Atoi(field)
			style = style.Background(tcell.Color(colorNumber - 40))
		case "49":
			style = style.Background(tcell.ColorDefault)
		case "90", "91", "92", "93", "94", "95", "96", "97":
			colorNumber, _ := strconv.Atoi(field)
			style = style.Foreground(tcell.Color(colorNumber - 82))
		case "100", "101", "102", "103", "104", "105", "106", "107":
			colorNumber, _ := strconv.Atoi(field)
			style = style.Background(tcell.Color(colorNumber - 92))
		case "38", "48":
			var color string
			if len(fields) > index+1 {
				if fields[index+1] == "5" && len(fields) > index+2 { // 8-bit colors.
					colorNumber, _ := strconv.Atoi(fields[index+2])
					if colorNumber <= 7 {
						color = lookupColor(colorNumber, false)
					} else if colorNumber <= 15 {
						color = lookupColor(colorNumber, true)
					} else if colorNumber <= 231 {
						red := (colorNumber - 16) / 36
						green := ((colorNumber - 16) / 6) % 6
						blue := (colorNumber - 16) % 6
						color = fmt.Sprintf("#%02x%02x%02x", 255*red/5, 255*green/5, 255*blue/5)
					} else if colorNumber <= 255 {
						grey := 255 * (colorNumber - 232) / 23
						color = fmt.Sprintf("#%02x%02x%02x", grey, grey, grey)
					}
				} else if fields[index+1] == "2" && len(fields) > index+4 { // 24-bit colors.
					red, _ := strconv.Atoi(fields[index+2])
					green, _ := strconv.Atoi(fields[index+3])
					blue, _ := strconv.Atoi(fields[index+4])
					color = fmt.Sprintf("#%02x%02x%02x", red, green, blue)
				}
			}
			if len(color) > 0 {
				if field == "38" {
					style = style.Foreground(tcell.GetColor(color))
				} else {
					style = style.Background(tcell.GetColor(color))
				}
				break FieldLoop
			}
		}
	}
	return style
}

// strToContents converts a single-line string into a content array.
func strToContents(line string, tabWidth int) []content {
	contents, _ := parseString(line, tabWidth)
	return contents
}

func parseString(line string, tabWidth int) ([]content, map[int]int) {
	contents := []content{}
	defaultStyle := tcell.StyleDefault
	defaultContent := content{
		mainc: 0,
		combc: []rune{},
		width: 0,
		style: defaultStyle,
	}

	state := ansiText
	csiParameter := new(bytes.Buffer)
	style := defaultStyle
	x := 0
	byteMaps := make(map[int]int)
	n := 0
	for _, runeValue := range line {
		c := defaultContent
		byteMaps[n] = len(contents)
		n += len(string(runeValue))
		switch state {
		case ansiEscape:
			switch runeValue {
			case '[': // Control Sequence Introducer.
				csiParameter.Reset()
				state = ansiControlSequence
				continue
			case 'c': // Reset.
				style = defaultStyle
				state = ansiText
				continue
			case 'P', ']', 'X', '^', '_': // Substrings and commands.
				state = ansiSubstring
				continue
			default: // Ignore.
				state = ansiText
			}
		case ansiSubstring:
			if runeValue == 0x1b {
				state = ansiEscape
				continue
			}
		case ansiControlSequence:
			if runeValue == 'm' {
				style = csToStyle(csiParameter, runeValue)
			} else if runeValue >= 'A' && runeValue <= 'T' {
				// Ignore.
			} else {
				if runeValue >= 0x30 && runeValue <= 0x3f {
					csiParameter.WriteRune(runeValue)
					continue
				}
			}
			state = ansiText
			continue
		}

		switch runeValue {
		case 0x1b:
			state = ansiEscape
			continue
		case '\n':
			continue
		case '\t':
			tabStop := tabWidth - (x % tabWidth)
			c.mainc = rune(' ')
			c.width = 1
			c.style = style
			for i := 0; i < tabStop; i++ {
				contents = append(contents, c)
			}
			continue
		}

		switch runewidth.RuneWidth(runeValue) {
		case 0:
			if len(contents) > 0 {
				c2 := contents[len(contents)-1]
				c2.combc = append(c2.combc, runeValue)
				contents[len(contents)-1] = c2
			}
		case 1:
			c.mainc = runeValue
			c.width = 1
			c.style = style
			contents = append(contents, c)
			x++
		case 2:
			c.mainc = runeValue
			c.width = 2
			c.style = style
			contents = append(contents, c)
			contents = append(contents, defaultContent)
			x += 2
		}
	}
	byteMaps[len(line)] = len(contents)
	return contents, byteMaps
}
