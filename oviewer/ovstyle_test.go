package oviewer

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestToTcellStyle(t *testing.T) {
	tests := []struct {
		name string
		s    OVStyle
		want tcell.Style
	}{
		{
			name: "default style",
			s:    OVStyle{},
			want: tcell.StyleDefault,
		},
		{
			name: "foreground and background",
			s: OVStyle{
				Foreground: "red",
				Background: "blue",
			},
			want: tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlue),
		},
		{
			name: "all styles",
			s: OVStyle{
				Foreground:     "red",
				Background:     "blue",
				Blink:          true,
				Bold:           true,
				Dim:            true,
				Italic:         true,
				Reverse:        true,
				Underline:      true,
				UnderlineStyle: "1",
				UnderlineColor: "green",
				StrikeThrough:  true,
			},
			want: tcell.StyleDefault.Foreground(tcell.ColorRed).
				Background(tcell.ColorBlue).
				Blink(true).
				Bold(true).
				Dim(true).
				Italic(true).
				Reverse(true).
				Underline(true).
				Underline(tcell.UnderlineStyleSolid).
				Underline(tcell.ColorGreen).
				StrikeThrough(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToTcellStyle(tt.s); got != tt.want {
				t.Errorf("ToTcellStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyStyle(t *testing.T) {
	tests := []struct {
		name string
		s    OVStyle
		want tcell.Style
	}{
		{
			name: "default style",
			s:    OVStyle{},
			want: tcell.StyleDefault,
		},
		{
			name: "foreground and background",
			s: OVStyle{
				Foreground: "red",
				Background: "blue",
			},
			want: tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlue),
		},
		{
			name: "all styles",
			s: OVStyle{
				Foreground:     "red",
				Background:     "blue",
				Blink:          true,
				Bold:           true,
				Dim:            true,
				Italic:         true,
				Reverse:        true,
				Underline:      true,
				UnderlineStyle: "1",
				UnderlineColor: "green",
				StrikeThrough:  true,
			},
			want: tcell.StyleDefault.Foreground(tcell.ColorRed).
				Background(tcell.ColorBlue).
				Blink(true).
				Bold(true).
				Dim(true).
				Italic(true).
				Reverse(true).
				Underline(true).
				Underline(tcell.UnderlineStyleSolid).
				Underline(tcell.ColorGreen).
				StrikeThrough(true),
		},
		{
			name: "unset styles",
			s: OVStyle{
				UnBlink:         true,
				UnBold:          true,
				UnDim:           true,
				UnItalic:        true,
				UnReverse:       true,
				UnUnderline:     true,
				UnStrikeThrough: true,
			},
			want: tcell.StyleDefault.Blink(false).
				Bold(false).
				Dim(false).
				Italic(false).
				Reverse(false).
				Underline(false).
				StrikeThrough(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := applyStyle(tcell.StyleDefault, tt.s); got != tt.want {
				t.Errorf("applyStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnderLineStyle(t *testing.T) {
	tests := []struct {
		name   string
		ustyle string
		want   tcell.UnderlineStyle
	}{
		{
			name:   "valid underline style 0",
			ustyle: "0",
			want:   tcell.UnderlineStyleNone,
		},
		{
			name:   "valid underline style 1",
			ustyle: "1",
			want:   tcell.UnderlineStyleSolid,
		},
		{
			name:   "valid underline style 2",
			ustyle: "2",
			want:   tcell.UnderlineStyleDouble,
		},
		{
			name:   "valid underline style 3",
			ustyle: "3",
			want:   tcell.UnderlineStyleCurly,
		},
		{
			name:   "valid underline style 4",
			ustyle: "4",
			want:   tcell.UnderlineStyleDotted,
		},
		{
			name:   "valid underline style 5",
			ustyle: "5",
			want:   tcell.UnderlineStyleDashed,
		},
		{
			name:   "invalid underline style -1",
			ustyle: "-1",
			want:   tcell.UnderlineStyleNone,
		},
		{
			name:   "invalid underline style 6",
			ustyle: "6",
			want:   tcell.UnderlineStyleNone,
		},
		{
			name:   "non-numeric underline style",
			ustyle: "abc",
			want:   tcell.UnderlineStyleNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := underLineStyle(tt.ustyle); got != tt.want {
				t.Errorf("underLineStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}
