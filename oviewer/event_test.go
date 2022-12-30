package oviewer

import (
	"bytes"
	"log"
	"os"
	"strings"
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
			root, err := NewRoot(bytes.NewBufferString("test"))
			if err != nil {
				t.Fatal(err)
			}
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
			root, err := NewRoot(bytes.NewBufferString("test"))
			if err != nil {
				t.Fatal(err)
			}
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

func TestRoot_Reload(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()

	tests := []struct {
		name string
		want string
	}{
		{
			name: "test1",
			want: "reload test1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := NewRoot(bytes.NewBufferString("test"))
			if err != nil {
				t.Fatal(err)
			}
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defaultFlags := log.Flags()
			log.SetFlags(0)
			defer func() {
				log.SetOutput(os.Stderr)
				log.SetFlags(defaultFlags)
			}()
			root.Doc.FileName = tt.name
			root.Reload()
			got := strings.TrimRight(buf.String(), "\n")
			if got != tt.want {
				t.Errorf("Root.SetDocument() = [%v], want [%v]", got, tt.want)
			}
		})
	}
}
