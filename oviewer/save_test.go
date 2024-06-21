package oviewer

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func Test_saveConfirmKey(t *testing.T) {
	type args struct {
		ev *tcell.EventKey
	}
	tests := []struct {
		name string
		args args
		want saveSelection
	}{
		{
			name: "saveOverWrite",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyRune, 'O', tcell.ModNone),
			},
			want: saveOverWrite,
		},
		{
			name: "saveAppend",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone),
			},
			want: saveAppend,
		},
		{
			name: "saveCancel",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyRune, 'N', tcell.ModNone),
			},
			want: saveCancel,
		},
		{
			name: "saveCancel2",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone),
			},
			want: saveCancel,
		},
		{
			name: "saveIgnore",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyRune, 'I', tcell.ModNone),
			},
			want: saveIgnore,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := saveConfirmKey(tt.args.ev); got != tt.want {
				t.Errorf("saveConfirmKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
