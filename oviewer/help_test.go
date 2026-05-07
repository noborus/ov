package oviewer

import (
	"testing"
)

func TestDuplicateKeyBind(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		k    KeyBind
		want string
	}{
		{
			name: "no duplicate key bindings",
			k:    KeyBind{ /* initialize with no duplicates */ },
			want: "",
		},
		{
			name: "invalid key bindings",
			k: KeyBind{
				actionCancel: {"Escape", "nokey+c"},
			},
			want: "\n\tInvalid key bindings:\nError: invalid key format: key nokey+c for action cancel\n",
		},
		{
			name: "duplicate key bindings",
			k: KeyBind{
				actionExit:   {"Escape", "q"},
				actionCancel: {"q"},
			},
			want: "\n\tDuplicate key bindings:\nDuplicate key [q] for exit, cancel\n",
		},
		{
			name: "DefaultKeyBind",
			k:    DefaultKeyBinds(),
			want: "",
		},
		{
			name: "LessKeyBind",
			k:    LessKeyBinds(),
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DuplicateKeyBind(tt.k)
			if got != tt.want {
				t.Errorf("DuplicateKeyBind() = %v, want %v", got, tt.want)
			}
		})
	}
}
