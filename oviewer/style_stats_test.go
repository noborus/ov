package oviewer

import (
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

func TestToggleStyleCommaSeparated(t *testing.T) {
	root := newStyleTestRoot([]bool{false, false, false})

	root.Doc.applyStyleSelection("0, 2")

	_, got0, ok := root.Doc.styles.Index(0)
	if !ok || !got0 {
		t.Fatalf("style 0 = %v, ok=%v; want true", got0, ok)
	}
	_, got1, ok := root.Doc.styles.Index(1)
	if !ok || got1 {
		t.Fatalf("style 1 = %v, ok=%v; want false", got1, ok)
	}
	_, got2, ok := root.Doc.styles.Index(2)
	if !ok || !got2 {
		t.Fatalf("style 2 = %v, ok=%v; want true", got2, ok)
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
		tt := tt
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
