package oviewer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v3"
)

// validateStyle confirms the style input and applies the selection.
func (root *Root) validateStyle(input string) {
	root.Doc.applyStyleSelection(input)
}

// applyStyleSelection toggles the style based on the provided value.
func (m *Document) applyStyleSelection(input string) {
	m.restoreStyleFlags()
	stylesLen := m.styles.Len()
	if stylesLen == 0 {
		return
	}

	for _, token := range parseInputStyles(input) {
		m.applyStyleToken(token, stylesLen)
	}
}

// parseInputStyles splits the input string by commas and trims whitespace from each token.
func parseInputStyles(input string) []string {
	tokens := strings.Split(input, ",")
	for i := range tokens {
		tokens[i] = strings.TrimSpace(tokens[i])
	}
	return tokens
}

// applyStyleToken processes a single token to toggle styles based on the defined rules.
func (m *Document) applyStyleToken(token string, stylesLen int) {
	if token == "" {
		return
	}
	firstChar := token[:1]
	switch firstChar {
	case "i":
		token = token[1:]
		m.toggleAllStyles()
	case "o":
		token = token[1:]
		m.disableAllStyles()
	case "a":
		token = token[1:]
		m.enableAllStyles()
	}

	if strings.Contains(token, "-") {
		parts := strings.SplitN(token, "-", 2)
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
	for i := range m.styles.Len() {
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
	for i := range m.styles.Len() {
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
	for i := range m.styles.Len() {
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

func styleString(style tcell.Style) string {
	fg := style.GetForeground()
	bg := style.GetBackground()
	attrs := style.GetAttributes()
	uStyle := style.GetUnderlineStyle()
	uColor := style.GetUnderlineColor()
	defaultFG := tcell.StyleDefault.GetForeground()
	defaultBG := tcell.StyleDefault.GetBackground()
	defaultAttrs := tcell.StyleDefault.GetAttributes()

	parts := make([]string, 0, 3)
	if fg != defaultFG {
		parts = append(parts, fmt.Sprintf("FG=%v", fg.String()))
	}
	if bg != defaultBG {
		parts = append(parts, fmt.Sprintf("BG=%v", bg.String()))
	}
	if attrs != defaultAttrs {
		parts = append(parts, fmt.Sprintf("%v", attrString(attrs)))
	}
	if uStyle != tcell.UnderlineStyleNone {
		parts = append(parts, fmt.Sprintf("%v", ustyleString(uStyle)))
	}
	if uColor != defaultFG {
		parts = append(parts, fmt.Sprintf("UnderlineColor=%v", uColor.String()))
	}
	return strings.Join(parts, ", ")
}

func attrString(attrs tcell.AttrMask) string {
	var parts []string
	if attrs&tcell.AttrBold != 0 {
		parts = append(parts, "Bold")
	}
	if attrs&tcell.AttrBlink != 0 {
		parts = append(parts, "Blink")
	}
	if attrs&tcell.AttrReverse != 0 {
		parts = append(parts, "Reverse")
	}
	if attrs&tcell.AttrDim != 0 {
		parts = append(parts, "Dim")
	}
	if attrs&tcell.AttrItalic != 0 {
		parts = append(parts, "Italic")
	}
	if attrs&tcell.AttrStrikeThrough != 0 {
		parts = append(parts, "StrikeThrough")
	}
	return strings.Join(parts, ", ")
}

func ustyleString(uStyle tcell.UnderlineStyle) string {
	switch uStyle {
	case tcell.UnderlineStyleSolid:
		return "Underline"
	case tcell.UnderlineStyleDouble:
		return "Underline=Double"
	case tcell.UnderlineStyleCurly:
		return "Underline=Curly"
	case tcell.UnderlineStyleDotted:
		return "Underline=Dotted"
	case tcell.UnderlineStyleDashed:
		return "Underline=Dashed"
	default:
		return "Underline=None"
	}
}
