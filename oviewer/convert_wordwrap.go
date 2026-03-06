package oviewer

import (
	"strings"

	"github.com/rivo/uniseg"
)

// wordwrapConverter is a converter that converts the contents to fit the screen width.
type wordwrapConverter struct {
	es          *escapeSequence
	screenWidth int
	row         int
	leadingFlag bool
}

// newWordwrapConverter creates a new wordwrapConverter.
func newWordwrapConverter(width int) *wordwrapConverter {
	return &wordwrapConverter{
		es:          newESConverter(),
		screenWidth: width,
		row:         1,
		leadingFlag: false,
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
	word        strings.Builder
	start       int
	end         int
	screenWidth int
	row         int
}

// convertWordWrap converts the contents to fit the screen width.
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
		row:         1,
		start:       pos.x(0),
	}

	state := -1
	charPos := 0 // accumulated character position in source string
	for len(str) > 0 {
		segment, remainder, boundaries, newState := uniseg.StepString(str, state)
		str = remainder
		state = newState

		proc.word.WriteString(segment)
		charPos += len(segment)

		if boundaries&uniseg.MaskWord == 0 {
			continue
		}

		proc.end = proc.pos.x(charPos)
		srcWord := proc.src[proc.start:proc.end]
		proc.word.Reset()
		proc.processWord(srcWord)
		proc.start = proc.end
	}
	return proc.dst
}

// processWord handles the placement of a word in the output.
func (proc *wordWrapProcessor) processWord(srcWord contents) {
	// Word is longer than screen width, add as-is and move to next line
	if len(srcWord) > proc.screenWidth {
		proc.dst = append(proc.dst, srcWord...)
		proc.row++
		return
	}

	// Word fits in current line
	if len(proc.dst)+len(srcWord) <= proc.screenWidth*proc.row {
		proc.dst = append(proc.dst, srcWord...)
		return
	}

	// Finish current line with padding
	proc.finishLine()
	proc.row++

	// Skip space characters
	if skipSpace(srcWord) {
		return
	}

	proc.dst = append(proc.dst, srcWord...)
}

// finishLine pads the current line with spaces.
func (proc *wordWrapProcessor) finishLine() {
	addSpaces := proc.screenWidth*proc.row - len(proc.dst)
	if addSpaces > 0 {
		proc.dst = append(proc.dst, StrToContents(strings.Repeat(" ", addSpaces), addSpaces)...)
	}
}

// skipSpace returns the position of the first non-space, non-tab, non-empty cell in the contents.
func skipSpace(src contents) bool {
	for pos := 0; pos < len(src); pos++ {
		if src[pos].width != 0 && src[pos].str != " " && src[pos].str != "\t" && src[pos].str != "" {
			return false
		}
		pos++
	}
	return true
}
