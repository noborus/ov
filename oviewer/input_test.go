package oviewer

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestInput_keyEvent(t *testing.T) {
	type fields struct {
		value   string
		cursorX int
	}
	type args struct {
		evKey *tcell.EventKey
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      bool
		wantValue string
	}{
		{
			name: "key enter",
			fields: fields{
				value:   "test",
				cursorX: 4,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyEnter, 0, 0),
			},
			want:      true,
			wantValue: "test",
		},
		{
			name: "key cancel",
			fields: fields{
				value:   "test",
				cursorX: 4,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyEscape, 0, 0),
			},
			want:      false,
			wantValue: "",
		},
		{
			name: "key rune",
			fields: fields{
				value:   "test",
				cursorX: 4,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyRune, 'a', 0),
			},
			want:      false,
			wantValue: "testa",
		},
		{
			name: "key tab",
			fields: fields{
				value:   "test",
				cursorX: 4,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyTAB, 0, 0),
			},
			want:      false,
			wantValue: "test\t",
		},
		{
			name: "key backspace",
			fields: fields{
				value:   "test",
				cursorX: 4,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyBackspace, 0, 0),
			},
			want:      false,
			wantValue: "tes",
		},
		{
			name: "key delete",
			fields: fields{
				value:   "test",
				cursorX: 0,
			},
			args: args{
				evKey: tcell.NewEventKey(tcell.KeyDelete, 0, 0),
			},
			want:      false,
			wantValue: "est",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &Input{
				value:   tt.fields.value,
				cursorX: tt.fields.cursorX,
			}
			got := input.keyEvent(tt.args.evKey)
			if got != tt.want {
				t.Errorf("Input.keyEvent() = %v, want %v", got, tt.want)
			}
			if tt.wantValue != input.value {
				t.Errorf("Input.value = %v, want %v", input.value, tt.wantValue)
			}
		})
	}
}
