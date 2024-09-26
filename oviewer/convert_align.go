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

// convert aligns the columns.
// convert also shrinks and right-aligns the columns.
// (It only processes escape sequences until the end of the line is reached.)
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
	dst := make(contents, 0, len(src))
	start := pos.x(0)
	for columnNum := 0; columnNum < len(indexes); columnNum++ {
		end := pos.x(indexes[columnNum][0]) // Start of the delimiter.
		delmEnd := pos.x(indexes[columnNum][1])
		if a.isShrink(columnNum) {
			// shrink column.
			dst = appendShrink(dst)
		} else {
			// content column.
			dst = a.appendColumn(dst, columnNum, src[start:end])
		}
		// Add the delimiter.
		dst = append(dst, src[end:delmEnd]...)
		start = delmEnd
	}

	// Add the remaining content.
	if a.isShrink(len(indexes)) {
		dst = appendShrink(dst)
		return dst
	}
	dst = append(dst, src[start:]...)
	return dst
}

// convertWidth accumulates one line and then adds spaces to align the column widths.
// convertWidth works line by line.
func (a *align) convertWidth(src contents) contents {
	dst := make(contents, 0, len(src))

	start := 0
	for columnNum := 0; columnNum < len(a.orgWidths); columnNum++ {
		end := findColumnEnd(src, a.orgWidths, columnNum, start) + 1
		end = min(end, len(src))

		// shrink column.
		if a.isShrink(columnNum) {
			dst = appendShrink(dst)
			dst = append(dst, SpaceContent)
			a.maxWidths[columnNum] = runewidth.RuneWidth(Shrink)
			start = end
			continue
		}

		// content column.
		tStart := findStartWithTrim(src, start)
		tEnd := findEndWidthTrim(src, end)
		// If the column width is 0, skip.
		if tStart >= tEnd {
			start = end
			continue
		}
		dst = a.appendColumn(dst, columnNum, src[tStart:tEnd])
		dst = append(dst, SpaceContent)
		start = end
	}

	// Add the remaining content.
	if a.isShrink(len(a.orgWidths)) {
		dst = appendShrink(dst)
		return dst
	}
	dst = append(dst, src[start:]...)
	return dst
}

// appendColumn adds column content to the lc.
func (a *align) appendColumn(lc contents, columnNum int, column contents) contents {
	padding := 0
	if columnNum < len(a.maxWidths) {
		padding = (a.maxWidths[columnNum] - len(column))
	}

	if a.isRightAlign(columnNum) {
		// Add left space to align columns.
		lc = appendPaddings(lc, padding)
		// Add content.
		lc = append(lc, column...)
		return lc
	}
	// Add content.
	lc = append(lc, column...)
	// Add right space to align columns.
	lc = appendPaddings(lc, padding)
	return lc
}

// appendShrink adds a shrink column to the lc.
func appendShrink(lc contents) contents {
	lc = append(lc, ShrinkContent)
	if ShrinkContent.width > 1 {
		lc = append(lc, DefaultContent)
	}
	return lc
}

// appendPaddings adds spaces to the lc.
func appendPaddings(lc contents, num int) contents {
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
