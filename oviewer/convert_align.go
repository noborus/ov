package oviewer

import (
	"regexp"

	"github.com/mattn/go-runewidth"
)

// align is a converter that aligns columns.
// It is used to align columns when the delimiter is reached or to align columns by adding spaces to the end of the line.
type align struct {
	es           *escapeSequence
	maxWidths    []int // Maximum width of each column.
	orgWidths    []int
	shrink       []bool // Shrink column.
	rightAlign   []bool // Right align column.
	WidthF       bool
	delimiter    string
	delimiterReg *regexp.Regexp
	count        int
}

func newAlignConverter(widthF bool) *align {
	return &align{
		es:         newESConverter(),
		maxWidths:  []int{},
		shrink:     []bool{},
		rightAlign: []bool{},
		count:      0,
		WidthF:     widthF,
	}
}

// convert converts only escape sequence codes to display characters and returns them as is.
// Returns true if it is an escape sequence and a non-printing character.
func (a *align) convert(st *parseState) bool {
	if len(st.lc) == 0 {
		if !a.WidthF && a.delimiter == "\t" {
			st.tabWidth = 1
		}
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
		st.lc = a.convertWidth(st.lc)
	} else {
		st.lc = a.convertDelm(st.lc)
	}
	return false
}

func (a *align) reset() {
	a.count = 0
}

// convertDelm aligns the column widths by adding spaces when it reaches a delimiter.
// convertDelm works line by line.
func (a *align) convertDelm(src contents) contents {
	str, pos := ContentsToStr(src)
	indexes := allIndex(str, a.delimiter, a.delimiterReg)
	if len(indexes) == 0 {
		return src
	}
	delmWidth := runewidth.StringWidth(a.delimiter)
	dst := make(contents, 0, len(src))

	s := 0
	for c := 0; c < len(indexes); c++ {
		e := pos.x(indexes[c][0])
		width := e - s
		if width <= 0 {
			break
		}

		// shrink column.
		if a.isShrink(c) {
			if s == 0 {
				dst = appendShrink(dst, nil)
			} else {
				dst = appendShrink(dst, src[s:s+delmWidth])
			}
			s = e
			continue
		}

		addSpace := 0
		if c < len(a.maxWidths) {
			addSpace = ((a.maxWidths[c] + 1) - width)
		}

		if a.isRightAlign(c) {
			dst = appendSpaces(dst, addSpace)
			dst = append(dst, src[s:e]...)
		} else {
			dst = append(dst, src[s:e]...)
			dst = appendSpaces(dst, addSpace)
		}
		s = e
	}
	if !a.isShrink(len(indexes)) {
		dst = append(dst, src[s:]...)
		return dst
	}
	dst = appendShrink(dst, src[s:s+delmWidth])
	return dst
}

// convertWidth accumulates one line and then adds spaces to align the column widths.
// convertWidth works line by line.
func (a *align) convertWidth(src contents) contents {
	dst := make(contents, 0, len(src))

	s := 0
	for c := 0; c < len(a.orgWidths); c++ {
		e := findColumnEnd(src, a.orgWidths, c) + 1
		e = min(e, len(src))

		if a.isShrink(c) {
			dst = appendShrink(dst, nil)
			dst = append(dst, SpaceContent)
			a.maxWidths[c] = runewidth.RuneWidth(Shrink)
			s = e
			continue
		}

		ss := countLeftSpaces(src, s)
		ee := countRightSpaces(src, e)
		addSpace := 0
		if c < len(a.maxWidths) {
			addSpace = (a.maxWidths[c] - (ee - ss))
		}
		s = e

		if a.isRightAlign(c) {
			// Add left space to align columns.
			dst = appendSpaces(dst, addSpace)
			// Add content.
			dst = append(dst, src[ss:ee]...)
		} else {
			// Add content.
			dst = append(dst, src[ss:ee]...)
			// Add right space to align columns.
			dst = appendSpaces(dst, addSpace)
		}
		dst = append(dst, SpaceContent)

	}
	if !a.isShrink(len(a.orgWidths)) {
		dst = append(dst, src[s:]...)
		return dst
	}
	dst = appendShrink(dst, nil)
	return dst
}

func appendShrink(lc contents, delm contents) contents {
	lc = append(lc, delm...)
	lc = append(lc, ShrinkContent)
	if ShrinkContent.width > 1 {
		lc = append(lc, SpaceContent)
	}
	return lc
}

func appendSpaces(lc contents, num int) contents {
	for i := 0; i < num; i++ {
		lc = append(lc, SpaceContent)
	}
	return lc
}

func (a *align) isShrink(col int) bool {
	if len(a.shrink) <= col {
		return false
	}
	return a.shrink[col]
}

func (a *align) isRightAlign(col int) bool {
	if len(a.rightAlign) <= col {
		return false
	}
	return a.rightAlign[col]
}

func countLeftSpaces(lc contents, s int) int {
	for ; s < len(lc) && lc[s].mainc == ' '; s++ {
	}
	return s
}

func countRightSpaces(lc contents, e int) int {
	for ; e > 0 && lc[e-1].mainc == ' '; e-- {
	}
	return e
}
