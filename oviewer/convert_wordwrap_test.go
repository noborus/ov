package oviewer

import (
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
