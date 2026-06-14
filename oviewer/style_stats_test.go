package oviewer

import (
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/noborus/ov/indexmap"
)

func newStyleTestRoot(states []bool) *Root {
	root := &Root{
		Doc: &Document{
			styles: indexmap.NewIndexMap[tcell.Style, bool](),
		},
	}

	styleKeys := []tcell.Style{
		tcell.StyleDefault,
		tcell.StyleDefault.Bold(true),
		tcell.StyleDefault.Italic(true),
	}

	for i, key := range styleKeys {
		if i >= len(states) {
			break
		}
		root.Doc.styles.Set(key, states[i])
	}

	return root
}

func styleStates(m *Document) []bool {
	states := make([]bool, m.styles.Len())
	for i := 0; i < m.styles.Len(); i++ {
		_, v, ok := m.styles.Index(i)
		if !ok {
			continue
		}
		states[i] = v
	}
	return states
}

func TestDocument_applyStyleSelection(t *testing.T) {
	tests := []struct {
		name  string
		input string
		args  []bool
		want  []bool
	}{
		{
			name:  "no styles does nothing",
			input: "0",
			args:  []bool{},
			want:  []bool{},
		},
		{
			name:  "invalid input does nothing",
			input: ",,,",
			args:  []bool{false, false, false},
			want:  []bool{false, false, false},
		},
		{
			name:  "empty input does nothing",
			input: "",
			args:  []bool{false, false, false},
			want:  []bool{false, false, false},
		},
		{
			name:  "invalid token does nothing",
			input: "x",
			args:  []bool{false, false, false},
			want:  []bool{false, false, false},
		},
		{
			name:  "single index toggles style",
			input: "1",
			args:  []bool{false, false, false},
			want:  []bool{false, true, false},
		},
		{
			name:  "comma-separated indices toggle styles",
			input: "0,2",
			args:  []bool{false, false, false},
			want:  []bool{true, false, true},
		},
		{
			name:  "range toggles styles",
			input: "0-1",
			args:  []bool{false, false, false},
			want:  []bool{true, true, false},
		},
		{
			name:  "invalid range does nothing",
			input: "1-",
			args:  []bool{false, false, false},
			want:  []bool{false, false, false},
		},
		{
			name:  "invalid range format",
			input: "1-2-3",
			args:  []bool{false, false, false},
			want:  []bool{false, false, false},
		},
	}
	for _, tt := range tests {
		root := newStyleTestRoot(tt.args)
		t.Run(tt.name, func(t *testing.T) {
			m := root.Doc
			m.applyStyleSelection(tt.input)
			got := styleStates(root.Doc)
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("input=%q index=%d got=%v want=%v", tt.input, i, got[i], tt.want[i])
				}
			}
			if len(tt.want) == 0 {
				for i := 0; i < m.styles.Len(); i++ {
					_, v, ok := m.styles.Index(i)
					if !ok || v {
						t.Fatalf("input=%q index=%d got=%v, ok=%v; want false", tt.input, i, v, ok)
					}
				}
			}
		})
	}
}

func TestProcessingToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		initial []bool
		token   string
		want    []bool
	}{
		{
			name:    "enable all with a",
			initial: []bool{false, false, false},
			token:   "a",
			want:    []bool{true, true, true},
		},
		{
			name:    "invert all with i",
			initial: []bool{false, true, false},
			token:   "i",
			want:    []bool{true, false, true},
		},
		{
			name:    "toggle single index",
			initial: []bool{false, false, false},
			token:   "1",
			want:    []bool{false, true, false},
		},
		{
			name:    "toggle range",
			initial: []bool{false, false, false},
			token:   "0-2",
			want:    []bool{true, true, true},
		},
		{
			name:    "range end is clamped",
			initial: []bool{false, false, false},
			token:   "1-10",
			want:    []bool{false, true, true},
		},
		{
			name:    "o prefix disables all then toggles",
			initial: []bool{true, true, true},
			token:   "o1",
			want:    []bool{false, true, false},
		},
		{
			name:    "invalid token does nothing",
			initial: []bool{true, false, true},
			token:   "x",
			want:    []bool{true, false, true},
		},
		{
			name:    "out of range index does nothing",
			initial: []bool{true, false, true},
			token:   "3",
			want:    []bool{true, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := newStyleTestRoot(tt.initial)
			root.Doc.applyStyleToken(tt.token, root.Doc.styles.Len())

			got := styleStates(root.Doc)
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("token=%q index=%d got=%v want=%v", tt.token, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestStyleStats_FromAnsiSample40Patterns(t *testing.T) {
	root := rootFileReadHelper(t, filepath.Join(testdata, "ansiescape_style_stats_screen.txt"))
	m := root.Doc

	for i := 0; i < m.BufEndNum(); i++ {
		_ = m.getLineC(i)
	}

	const wantPatterns = 40
	if got := m.styles.Len(); got != wantPatterns {
		t.Fatalf("styles.Len()=%d want=%d", got, wantPatterns)
	}

	seen := make(map[string]struct{}, wantPatterns)
	for i := 0; i < m.styles.Len(); i++ {
		style, _, ok := m.styles.Index(i)
		if !ok {
			t.Fatalf("styles.Index(%d) not found", i)
		}
		s := styleString(style)
		if s == "" {
			t.Fatalf("styleString() is empty at index=%d", i)
		}
		if _, dup := seen[s]; dup {
			t.Fatalf("duplicate styleString=%q", s)
		}
		seen[s] = struct{}{}
	}
}
