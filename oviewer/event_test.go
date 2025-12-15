package oviewer

import (
	"context"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestRoot_MoveLine(t *testing.T) {
	type args struct {
		num int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "move line",
			args: args{
				num: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			root.MoveLine(tt.args.num)
		})
	}
}

func TestRoot_SetDocument(t *testing.T) {
	type args struct {
		docNum int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid document number",
			args: args{
				docNum: 0,
			},
			wantErr: false,
		},
		{
			name: "invalid document number - negative",
			args: args{
				docNum: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid document number - out of range",
			args: args{
				docNum: 100,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			if err := root.SetDocument(tt.args.docNum); (err != nil) != tt.wantErr {
				t.Errorf("Root.SetDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoot_event(t *testing.T) {
	type args struct {
		ev tcell.Event
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "event app quit",
			args: args{
				ev: &eventAppQuit{},
			},
			want: true,
		},
		{
			name: "event update end num",
			args: args{
				ev: &eventUpdateEndNum{},
			},
			want: false,
		},
		{
			name: "event document",
			args: args{
				ev: &eventDocument{},
			},
			want: false,
		},
		{
			name: "event resize",
			args: args{
				ev: tcell.NewEventResize(20, 20),
			},
			want: false,
		},
		{
			name: "event mouse",
			args: args{
				ev: tcell.NewEventMouse(0, 0, tcell.Button1, tcell.ModNone),
			},
			want: false,
		},
		{
			name: "event key",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyRune, "a", tcell.ModNone),
			},
			want: false,
		},
		{
			name: "event nil",
			args: args{
				ev: nil,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			ctx := context.Background()
			if got := root.event(ctx, tt.args.ev); got != tt.want {
				t.Errorf("Root.event() = %v, want %v", got, tt.want)
			}
		})
	}
}
