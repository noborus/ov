package oviewer

import (
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

	if input == "a" {
		// Toggle all styles
		for i := range stylesLen {
			root.toggleStyleIdx(i)
		}
		return
	}

	styleIndex, ok := calcStyleIndex(input)
	if !ok {
		return
	}
	if styleIndex < 0 || styleIndex >= stylesLen {
		return
	}

	root.toggleStyleIdx(styleIndex)
}

func (root *Root) toggleStyleIdx(idx int) {
	k, v, ok := root.Doc.styles.Index(idx)
	if !ok {
		return
	}
	root.Doc.styles.Set(k, !v)
}

func calcStyleIndex(input string) (int, bool) {
	if len(input) == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(input)
	if err != nil {
		return 0, false
	}
	return n, true
}
