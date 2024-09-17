package oviewer

import (
	"regexp"

	"github.com/mattn/go-runewidth"
)

// align is a converter that aligns columns.
// It is used to align columns when the delimiter is reached or to align columns by adding spaces to the end of the line.
type align struct {
	es           *escapeSequence
	orgWidths    []int // Original width of each column. This is the width determined by Guesswidth.
	maxWidths    []int // Maximum width of each column.
	columnAttrs  []columnAttribute
	WidthF       bool
	delimiter    string
	delimiterReg *regexp.Regexp
	count        int
}

// columnAttribute is a structure that holds the attributes of a column.
type columnAttribute struct {
	shrink     bool // Shrink column.
	rightAlign bool // Right align column.
}

func newAlignConverter(widthF bool) *align {
	return &align{
		es:          newESConverter(),
		maxWidths:   []int{},
		columnAttrs: []columnAttribute{},
		count:       0,
		WidthF:      widthF,
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

	start := 0
	for c := 0; c < len(a.orgWidths); c++ {
		end := findColumnEnd(src, a.orgWidths, c, start) + 1
		end = min(end, len(src))

		if a.isShrink(c) {
			dst = appendShrink(dst, nil)
			dst = append(dst, SpaceContent)
			a.maxWidths[c] = runewidth.RuneWidth(Shrink)
			start = end
			continue
		}

		tStart := findStartWithTrim(src, start)
		tEnd := findEndWidthTrim(src, end)
		// If the column width is 0, skip.
		if tStart >= tEnd {
			start = end
			continue
		}

		padding := 0
		if c < len(a.maxWidths) {
			padding = (a.maxWidths[c] - (tEnd - tStart))
		}
		start = end

		if a.isRightAlign(c) {
			// Add left space to align columns.
			dst = appendSpaces(dst, padding)
			// Add content.
			dst = append(dst, src[tStart:tEnd]...)
		} else {
			// Add content.
			dst = append(dst, src[tStart:tEnd]...)
			// Add right space to align columns.
			dst = appendSpaces(dst, padding)
		}
		dst = append(dst, SpaceContent)
	}

	// Add the remaining content.
	if !a.isShrink(len(a.orgWidths)) {
		dst = append(dst, src[start:]...)
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
	if col >= 0 && col < len(a.columnAttrs) {
		return a.columnAttrs[col].shrink
	}
	return false
}

func (a *align) isRightAlign(col int) bool {
	if col >= 0 && col < len(a.columnAttrs) {
		return a.columnAttrs[col].rightAlign
	}
	return false
}

func findStartWithTrim(lc contents, s int) int {
	for ; s < len(lc) && lc[s].mainc == ' '; s++ {
	}
	return s
}

func findEndWidthTrim(lc contents, e int) int {
	for ; e > 0 && lc[e-1].mainc == ' '; e-- {
	}
	return e
}
