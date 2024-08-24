package oviewer

import (
	"regexp"

	"github.com/mattn/go-runewidth"
)

// align is a converter that aligns columns.
// It is used to align columns when the delimiter is reached or to align columns by adding spaces to the end of the line.
type align struct {
	es           *escapeSequence
	maxWidths    []int // column max width
	orgWidths    []int
	WidthF       bool
	delimiter    string
	delimiterReg *regexp.Regexp
	count        int
}

func newAlignConverter(widthF bool) *align {
	return &align{
		es:     newESConverter(),
		count:  0,
		WidthF: widthF,
	}
}

// convert converts only escape sequence codes to display characters and returns them as is.
// Returns true if it is an escape sequence and a non-printing character.
func (a *align) convert(st *parseState) bool {
	if len(st.lc) == 0 {
		a.reset()
	}
	if a.es.convert(st) {
		return true
	}

	if len(a.maxWidths) == 0 {
		return false
	}
	a.count += 1
	if runewidth.RuneWidth(st.mainc) > 1 {
		a.count += 1
	}

	if st.mainc != '\n' {
		return false
	}

	if a.WidthF {
		return a.convertWidth(st)
	}
	return a.convertDelm(st)
}

func (a *align) reset() {
	a.count = 0
}

// convertDelm aligns the column widths by adding spaces when it reaches a delimiter.
// convertDelm works line by line.
func (a *align) convertDelm(st *parseState) bool {
	str, pos := ContentsToStr(st.lc)
	indexes := allIndex(str, a.delimiter, a.delimiterReg)
	if len(indexes) == 0 {
		return false
	}
	s := 0
	lc := make(contents, 0, len(st.lc))
	for c := 0; c < len(indexes); c++ {
		e := pos.x(indexes[c][0])
		lc = append(lc, st.lc[s:e]...)
		width := e - s
		// Add space to align columns.
		for ; width < a.maxWidths[c]; width++ {
			lc = append(lc, SpaceContent)
		}
		s = e
	}
	lc = append(lc, st.lc[s:]...)
	st.lc = lc
	return false
}

// convertWidth accumulates one line and then adds spaces to align the column widths.
// convertWidth works line by line.
func (a *align) convertWidth(st *parseState) bool {
	s := 0
	lc := make(contents, 0, len(st.lc))
	for i := 0; i < len(a.orgWidths); i++ {
		e := findColumnEnd(st.lc, a.orgWidths, i) + 1
		e = min(e, len(st.lc))
		width := e - s
		if s >= e {
			break
		}
		lc = append(lc, st.lc[s:e]...)
		// Add space to align columns.
		for ; width <= a.maxWidths[i]; width++ {
			lc = append(lc, SpaceContent)
		}
		s = e
	}
	lc = append(lc, st.lc[s:]...)
	st.lc = lc
	return false
}
