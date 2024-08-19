package oviewer

import "github.com/mattn/go-runewidth"

type align struct {
	Columns   []int // column width
	delimiter []rune
	delmCount int
	count     int
	columnNum int
}

func newAlignConverter() *align {
	return &align{
		count:     0,
		columnNum: 0,
	}
}

// convert converts only escape sequence codes to display characters and returns them as is.
// Returns true if it is an escape sequence and a non-printing character.
func (a *align) convert(st *parseState) bool {
	if len(st.lc) == 0 {
		a.count = 0
		a.columnNum = 0
		a.delmCount = 0
	}

	a.count += 1
	if runewidth.RuneWidth(st.mainc) > 1 {
		a.count += 1
	}

	if len(a.Columns) <= a.columnNum {
		return false
	}
	if len(a.delimiter) > a.delmCount && st.mainc == a.delimiter[a.delmCount] {
		a.delmCount += 1
	} else {
		a.delmCount = 0
		return false
	}
	if len(a.delimiter) < a.delmCount {
		return false
	}
	// Add space to align columns.
	for ; a.count < a.Columns[a.columnNum]; a.count++ {
		c := DefaultContent
		st.lc = append(st.lc, c)
	}
	a.count = 0
	a.columnNum++
	a.delmCount = 0

	return false
}
