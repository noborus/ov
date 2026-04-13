package oviewer

import (
	"bufio"
	"bytes"
	"fmt"
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
			kb := defaultKeyBinds()
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
			k:    defaultKeyBinds(),
		},
		{
			name: "normalize modifier keys for display",
			k: func() KeyBind {
				kb := defaultKeyBinds()
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
