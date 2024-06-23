package oviewer

import (
	"bytes"
	"context"
	"log"
	"os"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestRoot_MoveLine(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()

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
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type args struct {
		docNum int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "set document",
			args: args{
				docNum: 1,
			},
			want: "",
		},
		{
			name: "set document",
			args: args{
				docNum: -1,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defaultFlags := log.Flags()
			log.SetFlags(0)
			defer func() {
				log.SetOutput(os.Stderr)
				log.SetFlags(defaultFlags)
			}()
			root.SetDocument(tt.args.docNum)
			got := buf.String()
			if got != tt.want {
				t.Errorf("Root.SetDocument() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_event(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
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
				ev: tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
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
