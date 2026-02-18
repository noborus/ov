package oviewer

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
)

// OVStyle represents a style in addition to the original style.
type OVStyle struct {
	// Background is a color name string.
	Background string
	// Foreground is a color name string.
	Foreground string
	// UnderlineColor is a underline color name string.
	UnderlineColor string
	// UnderlineStyle is a underline style.
	UnderlineStyle string
	// VerticalAlignType is a vertical align type.
	VerticalAlignType int
	// If true, add blink.
	Blink bool
	// If true, add bold.
	Bold bool
	// If true, add dim.
	Dim bool
	// If true, add italic.
	Italic bool
	// If true, add reverse.
	Reverse bool
	// If true, add underline.
	Underline bool
	// If true, add strike through.
	StrikeThrough bool
	// If true, add overline (not yet supported).
	OverLine bool
	// If true, sub blink.
	UnBlink bool
	// If true, sub bold.
	UnBold bool
	// If true, sub dim.
	UnDim bool
	// If true, sub italic.
	UnItalic bool
	// If true, sub reverse.
	UnReverse bool
	// If true, sub underline.
	UnUnderline bool
	// If true, sub strike through.
	UnStrikeThrough bool
	// if true, sub overline (not yet supported).
	UnOverLine bool
	// If true, reset all styles.
	Reset bool
}

// ToTcellStyle convert from OVStyle to tcell style.
func ToTcellStyle(s OVStyle) tcell.Style {
	style := tcell.StyleDefault
	style = style.Foreground(tcell.GetColor(s.Foreground))
	style = style.Background(tcell.GetColor(s.Background))
	style = style.Blink(s.Blink)
	style = style.Bold(s.Bold)
	style = style.Dim(s.Dim)
	style = style.Italic(s.Italic)
	style = style.Reverse(s.Reverse)
	style = style.Underline(s.Underline)
	if s.UnderlineStyle != "" {
		style = style.Underline(underLineStyle(s.UnderlineStyle))
	}
	if s.UnderlineColor != "" {
		style = style.Underline(tcell.GetColor(s.UnderlineColor))
	}
	style = style.StrikeThrough(s.StrikeThrough)
	return style
}

// applyStyle applies the OVStyle to the tcell style.
func applyStyle(style tcell.Style, s OVStyle) tcell.Style {
	if s.Reset {
		style = tcell.StyleDefault
	}
	if s.Foreground != "" {
		style = style.Foreground(tcell.GetColor(s.Foreground))
	}
	if s.Background != "" {
		style = style.Background(tcell.GetColor(s.Background))
	}
	// tcell does not support vertical align type.
	// if s.VerticalAlignType != 0 {
	//	Top,Middle,Bottom
	// }

	if s.Blink {
		style = style.Blink(true)
	}
	if s.Bold {
		style = style.Bold(true)
	}
	if s.Dim {
		style = style.Dim(true)
	}
	if s.Italic {
		style = style.Italic(true)
	}
	if s.Reverse {
		style = style.Reverse(true)
	}
	if s.Underline {
		style = style.Underline(true)
	}
	if s.UnderlineStyle != "" {
		style = style.Underline(underLineStyle(s.UnderlineStyle))
	}
	if s.UnderlineColor != "" {
		style = style.Underline(tcell.GetColor(s.UnderlineColor))
	}
	if s.StrikeThrough {
		style = style.StrikeThrough(true)
	}
	// tcell does not support overline.
	// if s.OverLine {
	//	style = style.Overline(true)
	// }

	if s.UnBlink {
		style = style.Blink(false)
	}
	if s.UnBold {
		style = style.Bold(false)
	}
	if s.UnDim {
		style = style.Dim(false)
	}
	if s.UnItalic {
		style = style.Italic(false)
	}
	if s.UnReverse {
		style = style.Reverse(false)
	}
	if s.UnUnderline {
		style = style.Underline(false)
	}
	if s.UnStrikeThrough {
		style = style.StrikeThrough(false)
	}
	// tcell does not support overline.
	// if s.UnOverLine {
	//	style = style.Overline(false)
	// }

	return style
}

// underLineStyle sets the tcell.UnderlineStyle from the string.
// only support 0-5.
// 0: None, 1: Single, 2: Double, 3: Curly, 4: Dotted, 5: Dashed
func underLineStyle(ustyle string) tcell.UnderlineStyle {
	val, err := strconv.ParseUint(ustyle, 10, 8)
	if err != nil {
		return tcell.UnderlineStyleNone
	}
	us := tcell.UnderlineStyle(val)
	if us < tcell.UnderlineStyleNone || us > tcell.UnderlineStyleDashed {
		// Note: It is not appropriate to turn off the underline style for out-of-range values.
		// It is out of specification.
		return tcell.UnderlineStyleNone
	}
	return us
}
