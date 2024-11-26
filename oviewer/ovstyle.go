package oviewer

import "github.com/gdamore/tcell/v2"

// OVStyle represents a style in addition to the original style.
type OVStyle struct {
	// Background is a color name string.
	Background string
	// Foreground is a color name string.
	Foreground string
	// UnderlineColor is a underline color name string.
	UnderlineColor string
	// UnderlineType is a underline type.
	UnderlineType int
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
	style = style.StrikeThrough(s.StrikeThrough)
	return style
}

// applyStyle applies the OVStyle to the tcell style.
func applyStyle(style tcell.Style, s OVStyle) tcell.Style {
	if s.Foreground != "" {
		style = style.Foreground(tcell.GetColor(s.Foreground))
	}
	if s.Background != "" {
		style = style.Background(tcell.GetColor(s.Background))
	}
	// tcell does not support underline color.
	// if s.UnderlineColor != "" {
	//	style = style.UnderlineColor(tcell.GetColor(s.UnderlineColor))
	// }
	// tcell does not support underline type.
	// if s.UnderlineType != 0 {
	//	Double,Curly,Dotted,Dashed
	// }
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
