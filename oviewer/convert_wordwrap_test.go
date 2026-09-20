package oviewer

import (
	"reflect"
	"strings"
	"testing"
)

func TestConvertWordwrap(t *testing.T) {
	tests := []struct {
		name        string
		screenWidth int
		str         string
		tabWidth    int
		wantStr     string
	}{
		{
			name:        "empty content",
			screenWidth: 10,
			str:         "",
			tabWidth:    4,
			wantStr:     "",
		},
		{
			name:        "zero screen width ",
			screenWidth: 0,
			str:         "abcdefghij",
			tabWidth:    4,
			wantStr:     "abcdefghij",
		},
		{
			name:        "count exceeds screen width",
			screenWidth: 10,
			str:         "abcdefghij",
			tabWidth:    4,
			wantStr:     "abcdefghij",
		},
		{
			name:        "word wrap without wrap target",
			screenWidth: 10,
			str:         "abcdefghijklmno",
			tabWidth:    4,
			wantStr:     "abcdefghijklmno",
		},
		{
			name:        "word wrap with wrap target",
			screenWidth: 10,
			str:         "abcdef ghijkl mnd",
			tabWidth:    4,
			wantStr:     "abcdef    ghijkl mnd",
		},
		{
			name:        "word wrap with hyphen wrap target",
			screenWidth: 5,
			str:         "abcd-efg",
			tabWidth:    4,
			wantStr:     "abcd-efg",
		},
		{
			name:        "no wrap with dot punctuation",
			screenWidth: 5,
			str:         "abcd.efgh",
			tabWidth:    4,
			wantStr:     "abcd.efgh",
		},
		{
			name:        "word wrap with Japanese punctuation wrap target",
			screenWidth: 5,
			str:         "abcd、efg",
			tabWidth:    4,
			wantStr:     "abcd 、efg",
		},
		{
			name:        "mixed Japanese and English with space wrap target",
			screenWidth: 5,
			str:         "abc あいう",
			tabWidth:    4,
			wantStr:     "abc  あい う",
		},
		{
			name:        "mixed Japanese and English with Japanese punctuation wrap target",
			screenWidth: 5,
			str:         "abc、あいう",
			tabWidth:    4,
			wantStr:     "abc、あい う",
		},
		{
			name:        "Japanese first then English word separated by space",
			screenWidth: 5,
			str:         "あいう abcd",
			tabWidth:    4,
			wantStr:     "あい う   abcd",
		},
		{
			name:        "just fit with space wrap target",
			screenWidth: 10,
			str:         "0123456789 0123456789",
			tabWidth:    4,
			wantStr:     "01234567890123456789",
		},
		{
			name:        "long word spanning multiple rows followed by a word",
			screenWidth: 5,
			str:         "abcdefghijklmn op",
			tabWidth:    4,
			wantStr:     "abcdefghijklmn op",
		},
		{
			name:        "long word spanning multiple rows then wrap to a new row",
			screenWidth: 10,
			str:         "aa bbbbbbbbbbbbbbbbbbbbbbbbb cc",
			tabWidth:    4,
			wantStr:     "aa bbbbbbbbbbbbbbbbbbbbbbbbb  cc",
		},
		{
			name:        "escape sequence",
			str:         "abc\x1b[31mdef\x1b[0mghi",
			tabWidth:    4,
			screenWidth: 5,
			wantStr:     "abcdefghi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := newWordwrapConverter(tt.screenWidth)
			result, _ := parseLine(converter, tt.str, tt.tabWidth)

			if result.String() != tt.wantStr {
				t.Errorf("expected string %q, got %q", tt.wantStr, result.String())
			}
		})
	}
}

// rowStrings splits the converted contents into one string per terminal row.
// The first row is as wide as the screen, the rows it wraps onto are narrower
// by the indent they start with.
func rowStrings(lc contents, screenWidth, indent int) []string {
	rows := []string{}
	row := 1
	for len(lc) > 0 {
		width := screenWidth
		if row > 1 {
			width += indent
		}
		if width > len(lc) {
			width = len(lc)
		}
		var sb strings.Builder
		for _, c := range lc[:width] {
			sb.WriteString(c.str)
		}
		rows = append(rows, sb.String())
		lc = lc[width:]
		row++
	}
	return rows
}

func TestConvertWordwrapIndent(t *testing.T) {
	tests := []struct {
		name        string
		screenWidth int
		indent      int
		str         string
		wantRows    []string
	}{
		{
			name:        "wrapped line is indented",
			screenWidth: 10,
			indent:      2,
			str:         "abcdef ghijkl mnd",
			wantRows:    []string{"abcdef    ", "  ghijkl mnd"},
		},
		{
			name:        "word longer than the screen is indented on the rows it spans",
			screenWidth: 10,
			indent:      2,
			str:         "abcdefghijklmno",
			wantRows:    []string{"abcdefghij", "  klmno"},
		},
		{
			name:        "long word followed by a word",
			screenWidth: 10,
			indent:      2,
			str:         "aa bbbbbbbbbbbbbbbbbbbbbbbbb cc",
			wantRows:    []string{"aa bbbbbbb", "  bbbbbbbbbb", "  bbbbbbbb  ", "  cc"},
		},
		{
			name:        "line shorter than the screen keeps no indent",
			screenWidth: 10,
			indent:      2,
			str:         "short",
			wantRows:    []string{"short"},
		},
		{
			name:        "indent larger than the screen falls back to no indent",
			screenWidth: 10,
			indent:      12,
			str:         "abcdef ghijkl mnd",
			wantRows:    []string{"abcdef    ", "ghijkl mnd"},
		},
		{
			name:        "japanese text is indented by the same cell count",
			screenWidth: 5,
			indent:      2,
			str:         "abc あいう",
			wantRows:    []string{"abc  ", "  あい ", "  う"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := newWordwrapConverterIndent(tt.screenWidth, tt.indent)
			result, _ := parseLine(converter, tt.str, 4)

			rows := rowStrings(result, tt.screenWidth, tt.indent)
			if !reflect.DeepEqual(rows, tt.wantRows) {
				t.Errorf("expected rows %q, got %q (full %q)", tt.wantRows, rows, result.String())
			}
		})
	}
}

// The wrapped rows only add leading spaces, so removing the spaces from the
// result must give back the input with the spaces removed.
func TestConvertWordwrapIndentKeepsContent(t *testing.T) {
	tests := []struct {
		screenWidth int
		indent      int
		str         string
	}{
		{10, 2, "abcdef ghijkl mnd"},
		{10, 2, "abcdefghijklmno"},
		{10, 2, "aa bbbbbbbbbbbbbbbbbbbbbbbbb cc"},
		{5, 2, "abc あいう"},
		{5, 1, "abcdefghijklmn op"},
	}

	for _, tt := range tests {
		converter := newWordwrapConverterIndent(tt.screenWidth, tt.indent)
		result, _ := parseLine(converter, tt.str, 4)

		got := strings.ReplaceAll(result.String(), " ", "")
		want := strings.ReplaceAll(tt.str, " ", "")
		if got != want {
			t.Errorf("screenWidth %d indent %d, expected %q, got %q", tt.screenWidth, tt.indent, want, got)
		}
	}
}
