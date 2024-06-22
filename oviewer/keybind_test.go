package oviewer

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
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
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			kb := defaultKeyBinds()
			kbStr := kb.String()
			count := 0
			s := bufio.NewScanner(bytes.NewBufferString(kbStr))
			for s.Scan() {
				if strings.HasPrefix(s.Text(), " ") {
					count++
				}
			}
			if count != len(kb)+1 {
				t.Errorf("number of KeyBind String = %v, want %v", count, len(kb)+1)
			}
		})
	}
}
