package oviewer

import (
	"log"
	"strconv"
	"strings"
)

// toggleStyle toggles the style based on the provided value.
func (root *Root) toggleStyle(input string) {
	stylesLen := root.Doc.styles.Len()
	if stylesLen == 0 {
		return
	}
	input = strings.TrimSpace(input)

	styleIndex, ok := calcStyleIndex(input, stylesLen)
	if !ok {
		return
	}
	k, v, ok := root.Doc.styles.Index(styleIndex)
	if !ok {
		return
	}
	root.Doc.styles.Set(k, !v)
	log.Println("Toggling style index:", styleIndex, "Key:", k, "Value:", v, "OK:", ok)
}

func calcStyleIndex(input string, stylesLen int) (int, bool) {
	if len(input) == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(input)
	if err != nil {
		return 0, false
	}
	if n < 0 || n >= stylesLen {
		return 0, false
	}
	return n, true
}
