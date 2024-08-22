package oviewer

import (
	"log"

	"github.com/mattn/go-runewidth"
)

type align struct {
	es         *escapeSequence
	Columns    []int // column width
	orgColumns []int
	WidthF     bool
	overF      bool
	delimiter  []rune
	delmCount  int
	count      int
	fullCount  int
	columnNum  int
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

	if a.columnNum >= len(a.Columns) {
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
	a.fullCount = 0
	a.count = 0
	a.columnNum = 0
	a.delmCount = 0
	a.overF = false
}

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
	for ; a.count < a.Columns[a.columnNum]; a.count++ {
		st.lc = append(st.lc, DefaultContent)
	}
	a.columnNum++
	a.count = 0
	a.delmCount = 0

	return false
}

func (a *align) convertWidth(st *parseState) bool {
	if st.mainc != '\n' {
		return false
	}
	s := 0
	lc := make(contents, 0, len(st.lc))
	for i := 0; i < len(a.orgColumns); i++ {
		e := findColumnEnd(st.lc, a.orgColumns, i) + 1
		e = min(e, len(st.lc))
		width := e - s
		if s >= e {
			break
		}
		lc = append(lc, st.lc[s:e]...)
		for ; width <= a.Columns[i]; width++ {
			c := DefaultContent
			c.mainc = '_'
			lc = append(lc, c)
		}
		s = e
	}
	lc = append(lc, st.lc[s:]...)
	log.Println("len", len(st.lc), len(lc))
	st.lc = lc
	return false
}
