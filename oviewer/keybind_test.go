package oviewer

import (
	"bufio"
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func Test_defaultKeyBinds(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "KeyBind",
		},
	}
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := NewRoot(bytes.NewBufferString("test"))
			if err != nil {
				t.Fatal("NewRoot failed: ", err)
			}
			hm := root.handlers()
			kb := DefaultKeyBinds()
			if len(hm) != len(kb) {
				t.Errorf("number of KeyBind = %v, want %v", len(hm), len(kb))
			}
		})
	}
}

func TestKeyBind_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		k    KeyBind
	}{
		{
			name: "KeyBind String",
			k:    DefaultKeyBinds(),
		},
		{
			name: "normalize modifier keys for display",
			k: func() KeyBind {
				kb := DefaultKeyBinds()
				kb[actionCancel] = []string{"CTRL+C", "alt+SHIFT+f8"}
				return kb
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			kbStr := tt.k.String()
			count := 0
			s := bufio.NewScanner(bytes.NewBufferString(kbStr))
			for s.Scan() {
				if strings.HasPrefix(s.Text(), " ") {
					count++
				}
			}
			if count != len(tt.k)+1 {
				t.Errorf("number of KeyBind String = %v, want %v", count, len(tt.k)+1)
			}

			if tt.name != "normalize modifier keys for display" {
				return
			}

			expectedCtrlC, err := normalizeKey("CTRL+C")
			if err != nil {
				t.Fatal(err)
			}
			expectedAltShiftF8, err := normalizeKey("alt+SHIFT+f8")
			if err != nil {
				t.Fatal(err)
			}
			expected := fmt.Sprintf("[%s], [%s]", expectedCtrlC, expectedAltShiftF8)
			if !strings.Contains(kbStr, expected) {
				t.Fatalf("KeyBind String() = %q, want to contain %q", kbStr, expected)
			}
		})
	}
}

func TestLessKeyBinds(t *testing.T) {
	t.Parallel()

	less := LessKeyBinds()

	tests := []struct {
		action string
		want   []string
	}{
		{action: actionEdit, want: []string{"v"}},
		{action: actionGoLine, want: []string{":"}},
		{action: actionMovePgDn, want: []string{"PageDown", "ctrl+v", "alt+space", "f", "z"}},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			t.Parallel()
			if !slices.Equal(less[tt.action], tt.want) {
				t.Fatalf("LessKeyBinds()[%s] = %v, want %v", tt.action, less[tt.action], tt.want)
			}
		})
	}
}

func TestGetKeyBindsDefaultPreset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config Config
		action string
		want   []string
	}{
		{
			name:   "default",
			config: Config{},
			action: actionEdit,
			want:   []string{"alt+v"},
		},
		{
			name:   "less",
			config: Config{DefaultKeyBind: "less"},
			action: actionEdit,
			want:   []string{"v"},
		},
		{
			name:   "disable",
			config: Config{DefaultKeyBind: "disable", Keybind: map[string][]string{actionEdit: {"x"}}},
			action: actionEdit,
			want:   []string{"x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GetKeyBinds(tt.config)
			if !slices.Equal(got[tt.action], tt.want) {
				t.Fatalf("GetKeyBinds(%s)[%s] = %v, want %v", tt.name, tt.action, got[tt.action], tt.want)
			}
		})
	}
}
