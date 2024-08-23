package oviewer

import (
	"github.com/mattn/go-runewidth"
)

type align struct {
	es        *escapeSequence
	maxWidths []int // column max width
	orgWidths []int
	WidthF    bool
	delimiter []rune
	delmCount int
	count     int
	columnNum int
}

func newAlignConverter(widthF bool) *align {
	return &align{
		es:        newESConverter(),
		count:     0,
		columnNum: 0,
		WidthF:    widthF,
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

	if a.columnNum >= len(a.maxWidths) {
		return false
	}
	a.count += 1
	if runewidth.RuneWidth(st.mainc) > 1 {
		a.count += 1
	}

	if a.WidthF {
		return a.convertWidth(st)
	}
	return a.convertDelm(st)
}

func (a *align) reset() {
	a.count = 0
	a.columnNum = 0
	a.delmCount = 0
}

// convertWidth accumulates one line and then adds spaces to align the column widths.
func (a *align) convertDelm(st *parseState) bool {
	if a.delmCount < len(a.delimiter) && st.mainc == a.delimiter[a.delmCount] {
		a.delmCount += 1
	} else {
		a.delmCount = 0
		return false
	}
	if a.delmCount > len(a.delimiter) {
		return false
	}
	// Add space to align columns.
	for ; a.count < a.maxWidths[a.columnNum]; a.count++ {
		st.lc = append(st.lc, SpaceContent)
	}
	a.columnNum++
	a.count = 0
	a.delmCount = 0

	return false
}

// convertDelm aligns the column widths by adding spaces when it reaches a delimiter.
func (a *align) convertWidth(st *parseState) bool {
	if st.mainc != '\n' {
		return false
	}
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
