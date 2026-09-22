package oviewer

import (
	"strconv"
	"strings"

	"github.com/rivo/uniseg"
)

// convert_wordwrap performs word wrapping while avoiding breaking words in the middle and ensuring that the
// content fits within the specified screen width. It also handles escape sequences and newlines appropriately.

// wordwrapConverter is a converter that converts the contents to fit the screen width.
type wordwrapConverter struct {
	es             *escapeSequence
	screenWidth    int
	indentWidth    int
	indentOffset   int
	relativeIndent bool
}

// newWordwrapConverter creates a new wordwrapConverter.
func newWordwrapConverter(width int, indent string) *wordwrapConverter {
	indentWidth, indentOffset, relativeIndent := parseWrapIndent(indent, width)
	return &wordwrapConverter{
		es:             newESConverter(),
		screenWidth:    width,
		indentWidth:    indentWidth,
		indentOffset:   indentOffset,
		relativeIndent: relativeIndent,
	}
}

func parseWrapIndent(indent string, width int) (int, int, bool) {
	if indent == "L" {
		return 0, 0, true
	}
	if strings.HasPrefix(indent, "L+") || strings.HasPrefix(indent, "L-") {
		offset, err := strconv.Atoi(indent[1:])
		if err != nil {
			return 0, 0, false
		}
		return 0, offset, true
	}
	if strings.HasPrefix(indent, "+") || strings.HasPrefix(indent, "-") {
		offset, err := strconv.Atoi(indent)
		if err != nil {
			return 0, 0, false
		}
		return 0, offset, true
	}

	absolute, err := strconv.Atoi(indent)
	if err != nil || absolute <= 0 || absolute >= width {
		return 0, 0, false
	}
	return absolute, 0, false
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
	indentWidth := c.indentWidth
	if c.relativeIndent {
		indentWidth = leadingIndentWidth(src) + c.indentOffset
		if indentWidth <= 0 || indentWidth >= c.screenWidth {
			indentWidth = 0
		}
	}

	proc := &wordWrapProcessor{
		dst:         make(contents, 0, len(src)),
		src:         src,
		pos:         pos,
		screenWidth: c.screenWidth,
		indentWidth: indentWidth,
		row:         1,
		start:       pos.x(0),
	}

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
	return proc.dst
}

func leadingIndentWidth(src contents) int {
	width := 0
	for _, cell := range src {
		if cell.str != " " && cell.str != "\t" && cell.str != "" {
			break
		}
		width += cell.width
	}
	return width
}

// processWord handles the placement of a word in the output.
func (proc *wordWrapProcessor) processWord(srcWord contents) {
	// Word is longer than screen width, add as-is and move to the row it ends on.
	// The word can span several rows, so the row is recalculated from the output length
	// instead of being incremented by one.
	if len(srcWord) > proc.screenWidth {
		proc.dst = append(proc.dst, srcWord...)
		proc.row = (len(proc.dst) / proc.screenWidth) + 1
		return
	}

	// Word fits in current line.
	if len(proc.dst)+len(srcWord) <= proc.screenWidth*proc.row {
		proc.dst = append(proc.dst, srcWord...)
		return
	}

	// Finish current line with padding.
	proc.finishLine()
	isFit := len(proc.dst) == proc.screenWidth*proc.row

	// wrap to the next line.
	proc.row++

	if isFit {
		// Apply indentation for the new line.
		if proc.indentWidth > 0 {
			proc.dst = append(proc.dst, spaceContents(proc.indentWidth)...)
		}
		// isOnlyWhitespace check is needed to avoid adding unnecessary spaces when the word is only whitespace.
		if isOnlyWhitespace(srcWord) {
			return
		}
	}

	proc.dst = append(proc.dst, srcWord...)
}

// finishLine pads the current line with spaces.
func (proc *wordWrapProcessor) finishLine() {
	addSpaces := proc.screenWidth*proc.row - len(proc.dst)
	if addSpaces > 0 {
		proc.dst = append(proc.dst, spaceContents(addSpaces)...)
	}
}

// spaceContents returns width spaces as contents.
func spaceContents(width int) contents {
	return StrToContents(strings.Repeat(" ", width), width)
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
