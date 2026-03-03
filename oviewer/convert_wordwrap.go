package oviewer

import (
	"unicode"
)

type wordwrapConverter struct {
	es          *escapeSequence
	screenWidth int
	row         int
	leadingFlag bool
}

func newWordwrapConverter(width int) *wordwrapConverter {
	return &wordwrapConverter{
		es:          newESConverter(),
		screenWidth: width,
		row:         1,
		leadingFlag: false,
	}
}

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
	pos := 0
	row := 1
	for pos < len(src) {
		word := contents{src[pos]}
		w := 1
		for pos+w < len(src) {
			t := src[pos+w]
			if isWrapTarget(t.str) {
				break
			}
			w++
			if t.width > 1 {
				w += (t.width - 1)
			}
		}
		word = src[pos : pos+w]
		if len(word) < c.screenWidth && len(dst)+len(word) > c.screenWidth*row {
			left := c.screenWidth*row - len(dst)
			for range left {
				dst = append(dst, SpaceContent)
			}
			if len(dst) <= c.screenWidth*row {
				// word skip space, tab, empty cell
				wp := skipSpace(word, 0)
				word = word[wp:]
				pos += wp
				// src skip space, tab, empty cell
				pos = skipSpace(src, pos)
			}
			row++
		}
		dst = append(dst, word...)
		pos += len(word)
	}
	return dst
}

// isWrapTarget returns true if the string is a wrap target.
func isWrapTarget(str string) bool {
	if len(str) == 0 {
		return true
	}
	r := rune(str[0])
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	switch str {
	case ".", ",", ";", ":", "!", "?":
		return false
	}
	return true
}

func skipSpace(src contents, pos int) int {
	for pos < len(src) {
		if src[pos].width != 0 && src[pos].str != " " && src[pos].str != "\t" && src[pos].str != "" {
			break
		}
		pos++
	}
	return pos
}
