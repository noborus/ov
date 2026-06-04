package oviewer

import (
	"strconv"
	"strings"
)

// validateStyle is a placeholder function for confirming the style input.
func (root *Root) validateStyle(input string) {
	// This function is currently a placeholder and does not perform any validation.
}

// applyStyleSelection toggles the style based on the provided value.
func (m *Document) applyStyleSelection(input string) {
	stylesLen := m.styles.Len()
	if stylesLen == 0 {
		return
	}

	tokens, ok := parseInputStyles(input)
	if !ok {
		return
	}
	for _, token := range tokens {
		m.applyStyleToken(token, stylesLen)
	}
}

// parseInputStyles splits the input string by commas and trims whitespace from each token.
func parseInputStyles(input string) ([]string, bool) {
	tokens := strings.Split(input, ",")
	if len(tokens) == 0 {
		return nil, false
	}
	for i := range tokens {
		tokens[i] = strings.TrimSpace(tokens[i])
	}
	return tokens, true
}

// applyStyleToken processes a single token to toggle styles based on the defined rules.
func (m *Document) applyStyleToken(token string, stylesLen int) {
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
		startIdx, err1 := strconv.Atoi(parts[0])
		endIdx, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || startIdx < 0 || endIdx < 0 || startIdx >= stylesLen {
			return
		}
		endIdx = min(endIdx, stylesLen-1)
		for i := startIdx; i <= endIdx; i++ {
			m.toggleStyleIdx(i)
		}
		return
	}

	styleIndex, err := strconv.Atoi(token)
	if err != nil {
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
