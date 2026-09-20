package oviewer

import (
	"strings"

	"github.com/rivo/uniseg"
)

// convert_wordwrap performs word wrapping while avoiding breaking words in the middle and ensuring that the
// content fits within the specified screen width. It also handles escape sequences and newlines appropriately.

// wordwrapConverter is a converter that converts the contents to fit the screen width.
type wordwrapConverter struct {
	es          *escapeSequence
	screenWidth int
	indentWidth int
}

// newWordwrapConverter creates a new wordwrapConverter.
func newWordwrapConverter(width int) *wordwrapConverter {
	return newWordwrapConverterIndent(width, 0)
}

// newWordwrapConverterIndent creates a wordwrapConverter that indents continuation rows by indent cells.
func newWordwrapConverterIndent(width int, indent int) *wordwrapConverter {
	if indent < 0 {
		indent = 0
	}
	// An indent that leaves no room for the text would wrap forever.
	if indent > width-2 {
		indent = 0
	}
	return &wordwrapConverter{
		es:          newESConverter(),
		screenWidth: width,
		indentWidth: indent,
	}
}

// convert converts the contents to fit the screen width.
func (c *wordwrapConverter) convert(st *parseState) bool {
	if c.es.convert(st) {
		return true
	}
	if st.str != "\n" {
		return false
	}
	st.lc = c.convertWordWrap(st.lc)
	return false
}

// wordWrapProcessor holds the state for word wrapping processing.
type wordWrapProcessor struct {
	dst         contents
	src         contents
	pos         widthPos
	word        string
	start       int
	end         int
	screenWidth int
	indentWidth int
	row         int
}

// convertWordWrap converts the contents to fit the screen width by wrapping words appropriately.
func (c *wordwrapConverter) convertWordWrap(src contents) contents {
	if c.screenWidth <= 0 {
		return src
	}

	str, pos := ContentsToStr(src)

	proc := &wordWrapProcessor{
		dst:         make(contents, 0, len(src)),
		src:         src,
		pos:         pos,
		screenWidth: c.screenWidth,
		indentWidth: c.indentWidth,
		row:         1,
		start:       pos.x(0),
	}
	proc.writeIndent()

	state := -1
	charPos := 0 // accumulated character position in source string.
	for len(str) > 0 {
		proc.word, str, state = uniseg.FirstWordInString(str, state)
		charPos += len(proc.word)

		proc.end = proc.pos.x(charPos)
		srcWord := proc.src[proc.start:proc.end]
		proc.processWord(srcWord)
		proc.start = proc.end
	}
	// The last row is padded to the row end so that the following rows line up.
	// It is trimmed again when the text did not reach the row end by itself.
	lastRow := proc.rowOf(len(proc.dst))
	if lastRow == 1 || proc.indent(proc.row) > proc.indent(lastRow) {
		if pad := proc.padding(); len(pad) < len(proc.dst) {
			proc.dst = proc.dst[:len(proc.dst)-len(pad)]
		}
	}
	return proc.dst
}

// processWord handles the placement of a word in the output.
func (proc *wordWrapProcessor) processWord(srcWord contents) {
	// Word is longer than the room left on the row, so it is split over the
	// following rows. The row is recalculated from the output length instead of
	// being incremented by one.
	if len(srcWord) > proc.screenWidth-proc.indent(proc.row) {
		proc.appendLongWord(srcWord)
		return
	}

	// Word fits in current line.
	if len(proc.dst)+len(srcWord) <= proc.lineEnd(proc.row) {
		proc.dst = append(proc.dst, srcWord...)
		return
	}

	// Finish current line with padding.
	proc.finishLine()
	proc.row++

	// isOnlyWhitespace check is needed to avoid adding unnecessary spaces when the word is only whitespace.
	if isOnlyWhitespace(srcWord) {
		return
	}

	proc.writeIndent()
	proc.dst = append(proc.dst, srcWord...)
}

// finishLine pads the current line with spaces.
func (proc *wordWrapProcessor) finishLine() {
	proc.dst = append(proc.dst, proc.padding()...)
}

// appendLongWord adds a word that is wider than a whole row, breaking it over
// the rows it spans and indenting the rows it continues onto.
func (proc *wordWrapProcessor) appendLongWord(srcWord contents) {
	for len(srcWord) > 0 {
		room := proc.lineEnd(proc.row) - len(proc.dst)
		if room <= 0 {
			proc.row++
			proc.writeIndent()
			continue
		}
		n := min(room, len(srcWord))
		proc.dst = append(proc.dst, srcWord[:n]...)
		srcWord = srcWord[n:]
		// Pad the row before starting the next one, otherwise the indent of the
		// continuation row lands in the middle of this row.
		if len(srcWord) > 0 {
			proc.finishLine()
			proc.row++
			proc.writeIndent()
		}
	}
}

// padding returns the spaces needed to fill the rest of the current row.
func (proc *wordWrapProcessor) padding() contents {
	// The row can already be full when the last word ended on the row boundary.
	if addSpaces := proc.lineEnd(proc.row) - len(proc.dst); addSpaces > 0 {
		return StrToContents(strings.Repeat(" ", addSpaces), addSpaces)
	}
	return nil
}

// indent returns the indent applied to the given row. The first row keeps the
// original line start, only rows the text was wrapped onto are indented.
func (proc *wordWrapProcessor) indent(row int) int {
	if row <= 1 {
		return 0
	}
	return proc.indentWidth
}

// lineEnd returns the destination length at the end of the given row.
func (proc *wordWrapProcessor) lineEnd(row int) int {
	return proc.screenWidth*row + proc.indentWidth*(row-1)
}

// rowOf returns the row a destination length ends on.
func (proc *wordWrapProcessor) rowOf(length int) int {
	if proc.indentWidth == 0 {
		return (length / proc.screenWidth) + 1
	}
	// Needed cells to reach row n are screenWidth*n + indentWidth*(n-1) for n padding
	// and screenWidth*n + indentWidth*(n-2) for n-1 padding.
	row := 1
	for length > proc.lineEnd(row) {
		row++
	}
	return row
}

// writeIndent appends the indent of the current row.
func (proc *wordWrapProcessor) writeIndent() {
	if indent := proc.indent(proc.row); indent > 0 {
		proc.dst = append(proc.dst, StrToContents(strings.Repeat(" ", indent), indent)...)
	}
}

// isOnlyWhitespace returns true if all cells are spaces, tabs, or empty; false otherwise.
func isOnlyWhitespace(src contents) bool {
	for pos := range src {
		if src[pos].width != 0 && src[pos].str != " " && src[pos].str != "\t" && src[pos].str != "" {
			return false
		}
	}
	return true
}
