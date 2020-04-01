package oviewer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

// The Model structure contains the values
// for the logical screen.
type Model struct {
	x          int
	lineNum    int
	yy         int
	endNum     int
	eof        bool
	header     []string
	buffer     []string
	beforeSize int
	vWidth     int
	vHight     int
	cache      *ristretto.Cache
}

//NewModel returns Model.
func NewModel() *Model {
	return &Model{
		buffer:     make([]string, 0),
		header:     make([]string, 0),
		beforeSize: 1000,
	}
}

// ReadAll reads all from the reader to the buffer.
// It returns if beforeSize is accumulated in buffer
// before the end of read.
func (m *Model) ReadAll(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		for scanner.Scan() {
			m.buffer = append(m.buffer, scanner.Text())
			m.endNum++
			if m.endNum == m.beforeSize {
				ch <- struct{}{}
			}
		}
		if err := scanner.Err(); err != nil {
			m.eof = false
			return
		}
		m.eof = true
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(500 * time.Millisecond):
		return nil
	}
}

// content represents one character on the terminal.
type content struct {
	width int
	style tcell.Style
	mainc rune
	combc []rune
}

// getContents returns one row of contents from buffer.
func (m *Model) getContents(lineNum int, tabWidth int) []content {
	var contents []content
	if lineNum < 0 || lineNum >= len(m.buffer) {
		return nil
	}
	value, found := m.cache.Get(lineNum)
	if found {
		var ok bool
		if contents, ok = value.([]content); !ok {
			return nil
		}
	} else {
		contents = strToContents(m.buffer[lineNum], tabWidth)
		m.cache.Set(lineNum, contents, 1)
	}
	return contents
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

func csToStyle(csiParameter *bytes.Buffer, runeValue rune) (*bytes.Buffer, tcell.Style) {
	style := tcell.StyleDefault
	if runeValue == 'm' {
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
				style = style.Foreground(tcell.GetColor(lookupColor(colorNumber-30, false)))
			case "39":
				style = style.Foreground(tcell.ColorDefault)
			case "40", "41", "42", "43", "44", "45", "46", "47":
				colorNumber, _ := strconv.Atoi(field)
				style = style.Background(tcell.GetColor(lookupColor(colorNumber-40, false)))
			case "49":
				style = style.Background(tcell.ColorDefault)
			case "90", "91", "92", "93", "94", "95", "96", "97":
				colorNumber, _ := strconv.Atoi(field)
				style = style.Foreground(tcell.GetColor(lookupColor(colorNumber-90, true)))
			case "100", "101", "102", "103", "104", "105", "106", "107":
				colorNumber, _ := strconv.Atoi(field)
				style = style.Background(tcell.GetColor(lookupColor(colorNumber-100, true)))
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
	}
	return csiParameter, style
}

// strToContents converts a single-line string into a content array.
func strToContents(line string, tabWidth int) []content {
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
	for _, runeValue := range line {
		c := defaultContent

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
			if runeValue >= 0x30 && runeValue <= 0x3f {
				csiParameter.WriteRune(runeValue)
				continue
			}
			csiParameter, style = csToStyle(csiParameter, runeValue)
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
	return contents
}

// contentsToStr converts a content array into a single-line string.
func contentsToStr(contents []content) (string, map[int]int) {
	buf := make([]rune, 0)
	cIndex := make(map[int]int)
	byteLen := 0
	for n := range contents {
		if contents[n].width > 0 {
			cIndex[byteLen] = n
			buf = append(buf, contents[n].mainc)
			b := string(contents[n].mainc)
			byteLen += len(b)
		}
	}
	s := string(buf)
	cIndex[len(s)] = len(contents)
	return s, cIndex
}
