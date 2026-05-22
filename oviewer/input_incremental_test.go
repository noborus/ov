package oviewer

import (
	"context"
	"reflect"
	"testing"
)

func TestParseMultiColorWords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "space split",
			input: "error info warn",
			want:  []string{"error", "info", "warn"},
		},
		{
			name:  "quoted word",
			input: "error \"warn notice\"",
			want:  []string{"error", "warn notice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseMultiColorWords(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseMultiColorWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatMultiColorWords(t *testing.T) {
	tests := []struct {
		name  string
		words []string
		want  string
	}{
		{
			name:  "without spaces",
			words: []string{"error", "warn", "info"},
			want:  "error warn info",
		},
		{
			name:  "with spaces",
			words: []string{"error", "warn notice", "panic"},
			want:  "error \"warn notice\" panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatMultiColorWords(tt.words); got != tt.want {
				t.Fatalf("formatMultiColorWords() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMultiColorWordsParseFormatRoundTrip(t *testing.T) {
	words := []string{"error", "warn notice", "debug"}
	formatted := formatMultiColorWords(words)
	if got := parseMultiColorWords(formatted); !reflect.DeepEqual(got, words) {
		t.Fatalf("parseMultiColorWords(formatMultiColorWords()) = %v, want %v", got, words)
	}
}

func TestRoot_incrementalInput_filterSearchHighlight(t *testing.T) {
	root := rootHelper(t)
	root.Config.CaseSensitive = false

	root.input.Event = newSearchEvent(root.input.Candidate[Search], filter)
	root.input.value = "test"

	root.incrementalInput(context.Background())

	// Check that searcher is set (not nil)
	if root.searcher == nil {
		t.Fatalf("incrementalInput Filter: searcher should be set, got nil")
	}

	// Check that searcher string matches input value
	if got, want := root.searcher.String(), "test"; got != want {
		t.Fatalf("incrementalInput Filter: searcher string = %v, want %v", got, want)
	}
}
