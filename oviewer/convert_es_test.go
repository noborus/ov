package oviewer

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func Test_csToStyle(t *testing.T) {
	t.Parallel()
	type args struct {
		style        tcell.Style
		csiParameter string
	}
	tests := []struct {
		name string
		args args
		want tcell.Style
	}{
		{
			name: "color8bit",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "38;5;1",
			},
			want: tcell.StyleDefault.Foreground(tcell.ColorMaroon),
		},
		{
			name: "color8bit2",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "38;5;21",
			},
			want: tcell.StyleDefault.Foreground(tcell.GetColor("#0000ff")),
		},
		{
			name: "attributes",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "2;3;4;5;6;7;8;9",
			},
			want: tcell.StyleDefault.Dim(true).Italic(true).Underline(true).Blink(true).Reverse(true).StrikeThrough(true),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := csToStyle(tt.args.style, tt.args.csiParameter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("csToStyle() = %v, want %v", got, tt.want)
				gfg, gbg, gattr := got.Decompose()
				wfg, wbg, wattr := tt.want.Decompose()
				t.Errorf("csToStyle() = %x,%x,%v, want %x,%x,%v", gfg.Hex(), gbg.Hex(), gattr, wfg.Hex(), wbg.Hex(), wattr)
			}
		})
	}
}
