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

	tokens, ok := parseInputTokens(input)
	if !ok {
		return
	}
	for _, token := range tokens {
		root.Doc.processingToken(token, stylesLen)
	}
}

// parseInputTokens splits the input string by commas and trims whitespace from each token.
func parseInputTokens(input string) ([]string, bool) {
	tokens := strings.Split(input, ",")
	if len(tokens) == 0 {
		return nil, false
	}
	for i := range tokens {
		tokens[i] = strings.TrimSpace(tokens[i])
	}
	return tokens, true
}

// processingToken processes a single token to toggle styles based on the defined rules.
func (m *Document) processingToken(token string, stylesLen int) {
	switch token {
	case "a":
		m.enableAllStyles()
		return
	case "n":
		m.disableAllStyles()
		return
	case "i":
		m.toggleAllStyles()
		return
	}

	if strings.HasPrefix(token, "o") {
		token = token[1:]
		m.disableAllStyles()
	}
	if strings.HasPrefix(token, "t") {
		token = token[1:]
		m.enableAllStyles()
	}

	if strings.Contains(token, "-") {
		parts := strings.SplitN(token, "-", 2)
		if len(parts) != 2 {
			return
		}
		startIdx, ok1 := calcStyleIndex(parts[0])
		endIdx, ok2 := calcStyleIndex(parts[1])
		if !ok1 || !ok2 || startIdx < 0 || endIdx < 0 || startIdx >= stylesLen {
			return
		}
		endIdx = min(endIdx, stylesLen-1)
		for i := startIdx; i <= endIdx; i++ {
			m.toggleStyleIdx(i)
		}
		return
	}

	styleIndex, ok := calcStyleIndex(token)
	if !ok {
		return
	}
	if styleIndex < 0 || styleIndex >= stylesLen {
		return
	}

	m.toggleStyleIdx(styleIndex)
}

// enableAllStyles sets all styles to true.
func (m *Document) enableAllStyles() {
	for i := 0; i < m.styles.Len(); i++ {
		k, _, ok := m.styles.Index(i)
		if !ok {
			continue
		}
		m.styles.Set(k, true)
	}
	m.isStylesEnabled = true
}

// disableAllStyles sets all styles to false.
func (m *Document) disableAllStyles() {
	for i := 0; i < m.styles.Len(); i++ {
		k, _, ok := m.styles.Index(i)
		if !ok {
			continue
		}
		m.styles.Set(k, false)
	}
	m.isStylesEnabled = false
}

// toggleAllStyles toggles all styles by inverting their current state.
func (m *Document) toggleAllStyles() {
	for i := 0; i < m.styles.Len(); i++ {
		m.toggleStyleIdx(i)
	}
}

// toggleStyleIdx toggles the style at the specified index by inverting its current state.
func (m *Document) toggleStyleIdx(idx int) {
	k, v, ok := m.styles.Index(idx)
	if !ok {
		return
	}
	m.styles.Set(k, !v)
}

// calcStyleIndex converts a token to an integer index, returning the index and a boolean indicating success.
func calcStyleIndex(token string) (int, bool) {
	if len(token) == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(token)
	if err != nil {
		return 0, false
	}
	return n, true
}
