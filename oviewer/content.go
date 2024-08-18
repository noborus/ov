package oviewer

import (
	"log"
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
	otherSequence
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

// csiCache caches escape sequences.
var csiCache sync.Map

// parseState represents the affected state after parsing.
type parseState struct {
	lc        contents
	style     tcell.Style
	bsContent content
	state     int
	tabWidth  int
	tabx      int
	bsFlag    bool // backspace(^H) flag
}

// Converter is an interface for converting escape sequences, etc.
type Converter interface {
	convert(*parseState, rune, []rune) bool
}

type rawConverter struct{}

func newRawConverter() Converter {
	return rawConverter{}
}
func (rawConverter) convert(st *parseState, mainc rune, _ []rune) bool {
	return false
}

// parseString converts a string to lineContents.
// parseString includes escape sequences and tabs.
// If tabwidth is set to -1, \t is displayed instead of functioning as a tab.
func parseString(conv Converter, str string, tabWidth int) contents {
	st := &parseState{
		lc:        make(contents, 0, len(str)),
		state:     ansiText,
		style:     tcell.StyleDefault,
		tabWidth:  tabWidth,
		tabx:      0,
		bsFlag:    false,
		bsContent: DefaultContent,
	}

	gr := uniseg.NewGraphemes(str)
	for gr.Next() {
		r := gr.Runes()
		mainc := r[0]
		combc := r[1:]

		if conv.convert(st, mainc, combc) {
			continue
		}
		st.parseString(mainc, combc)
	}
	return st.lc
}

func (st *parseState) parseString(mainc rune, combc []rune) {
	c := DefaultContent
	switch runewidth.RuneWidth(mainc) {
	case 0:
		switch {
		case mainc == '\t': // TAB
			switch {
			case st.tabWidth > 0:
				tabStop := st.tabWidth - (st.tabx % st.tabWidth)
				c.width = 1
				c.style = st.style
				c.mainc = rune('\t')
				st.lc = append(st.lc, c)
				st.tabx++
				c.mainc = 0
				for i := 0; i < tabStop-1; i++ {
					st.lc = append(st.lc, c)
					st.tabx++
				}
			case st.tabWidth < 0:
				c.width = 1
				c.style = st.style.Reverse(true)
				c.mainc = rune('\\')
				st.lc = append(st.lc, c)
				c.mainc = rune('t')
				st.lc = append(st.lc, c)
				st.tabx += 2
			default:
			}
			return
		case mainc == '\b': // BackSpace
			if len(st.lc) == 0 {
				return
			}
			st.bsFlag = true
			st.bsContent = st.lc.last()
			if st.bsContent.width > 1 {
				st.lc = st.lc[:len(st.lc)-2]
			} else {
				st.lc = st.lc[:len(st.lc)-1]
			}
			return
		case mainc < 0x20: // control character
			if mainc == '\r' { // CR
				return
			}
			c.mainc = mainc
			c.width = 0
			st.lc = append(st.lc, c)
			return
		}
		lastC := st.lc.last()
		lastC.combc = append(lastC.combc, mainc)
		n := len(st.lc) - lastC.width
		if n >= 0 && len(st.lc) > n {
			st.lc[n] = lastC
		}
	case 1:
		c.mainc = mainc
		if len(combc) > 0 {
			c.combc = combc
		}
		c.width = 1
		c.style = st.overstrike(c, st.style)
		st.lc = append(st.lc, c)
		st.tabx++
	case 2:
		c.mainc = mainc
		if len(combc) > 0 {
			c.combc = combc
		}
		c.width = 2
		c.style = st.overstrike(c, st.style)
		st.lc = append(st.lc, c, DefaultContent)
		st.tabx += 2
	}
}

// overstrike set style for overstrike.
func (es *parseState) overstrike(m content, style tcell.Style) tcell.Style {
	if !es.bsFlag {
		return style
	}

	if es.bsContent.mainc == m.mainc {
		style = OverStrikeStyle
	} else if es.bsContent.mainc == '_' {
		style = OverLineStyle
	}
	es.bsFlag = false
	es.bsContent = DefaultContent
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

// StrToContents converts a single-line string into a one line of contents.
// Parse escape sequences, etc.
// 1 Content matches the characters displayed on the screen.
func StrToContents(str string, tabWidth int) contents {
	es := newEscapeSequence()
	return parseString(es, str, tabWidth)
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
		bn += writeRune(&buff, c.mainc)
		for _, r := range c.combc {
			bn += writeRune(&buff, r)
		}
	}

	str := buff.String()
	for ; i <= bn; i++ {
		pos = append(pos, len(lc))
	}
	return str, pos
}

func writeRune(w *strings.Builder, r rune) int {
	n, err := w.WriteRune(r)
	if err != nil {
		log.Fatal(err)
	}
	return n
}

// x returns the x position on the screen.
// [n]byte -> x.
func (pos widthPos) x(n int) int {
	if n < len(pos) {
		return pos[n]
	}
	return pos[len(pos)-1]
}

// n return string position from content.
// x -> [n]byte.
func (pos widthPos) n(w int) int {
	x := w
	for _, c := range pos {
		if c >= w {
			x = c
			break
		}
	}
	// It should return the last byte of a multi-byte character.
	for i := len(pos) - 1; i >= 0; i-- {
		if pos[i] == x {
			return i
		}
	}
	return len(pos) - 1
}
