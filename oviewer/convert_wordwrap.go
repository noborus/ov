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

// convertWordWrap converts the contents to fit the screen width.
func (c *wordwrapConverter) convertWordWrap(src contents) contents {
	if c.screenWidth <= 0 {
		return src
	}

	dst := make(contents, 0, len(src))
	c.row = 1
	str, pos := ContentsToStr(src)
	start := pos.x(0)
	end := start
	var current string
	var word strings.Builder
	state := -1
	var boundaries int
	srcp := 0
	for len(str) > 0 {
		current, str, boundaries, state = uniseg.StepString(str, state)
		word.WriteString(current)
		if boundaries&uniseg.MaskWord == 0 {
			continue
		}
		wordLen := word.Len()

		srcp += wordLen
		end = pos.x(srcp)
		srcWord := src[start:end]
		word.Reset()

		if len(dst)+len(srcWord) <= c.screenWidth*c.row {
			dst = append(dst, srcWord...)
			start = end
			continue
		}
		if len(srcWord) > c.screenWidth {
			// If the word is longer than the screen width, break it into multiple lines.
			dst = append(dst, srcWord...)
			c.row++
			continue
		}
		addSpaces := c.screenWidth*c.row - len(dst)
		if addSpaces > 0 {
			dst = append(dst, StrToContents(strings.Repeat(" ", addSpaces), addSpaces)...)
		}

		// next line
		c.row++
		if skipSpace(srcWord) {
			start = end
			continue
		}
		dst = append(dst, srcWord...)
		start = end
	}
	return dst
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
