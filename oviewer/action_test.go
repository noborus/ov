package oviewer

import (
	"bytes"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestRoot_toggleWrapMode(t *testing.T) {
	tests := []struct {
		name       string
		wrapMode   bool
		columnMode bool
		x          int
		wantx      int
	}{
		{
			name:       "test1",
			wrapMode:   true,
			columnMode: false,
			x:          0,
			wantx:      0,
		},
		{
			name:       "test2",
			wrapMode:   false,
			columnMode: false,
			x:          10,
			wantx:      0,
		},
		{
			name:       "test3",
			wrapMode:   false,
			columnMode: true,
			x:          10,
			wantx:      10,
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
				t.Fatal(err)
			}
			root.Doc.WrapMode = tt.wrapMode
			root.Doc.ColumnMode = tt.columnMode
			root.toggleWrapMode()
			if root.Doc.WrapMode == tt.wrapMode {
				t.Errorf("root.toggleWrapMode() = %v, want %v", root.Doc.WrapMode, !tt.wrapMode)
			}
			if root.Doc.x != tt.wantx {
				t.Errorf("root.toggleWrapMode() = %v, want %v", root.Doc.x, tt.wantx)
			}
		})
	}
}

func TestRoot_toggleColumnMode(t *testing.T) {
	tests := []struct {
		name       string
		columnMode bool
	}{
		{
			name:       "test1",
			columnMode: false,
		},
		{
			name:       "test2",
			columnMode: true,
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
				t.Fatal(err)
			}
			root.Doc.ColumnMode = tt.columnMode
			root.toggleColumnMode()
			if root.Doc.ColumnMode == tt.columnMode {
				t.Errorf("root.toggleColumnMode() = %v, want %v", root.Doc.ColumnMode, !tt.columnMode)
			}
		})
	}
}

func Test_position(t *testing.T) {
	type args struct {
		hight int
		str   string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test1",
			args: args{
				hight: 30,
				str:   "1",
			},
			want: 1,
		},
		{
			name: "test.5",
			args: args{
				hight: 30,
				str:   ".5",
			},
			want: 15,
		},
		{
			name: "test20%",
			args: args{
				hight: 30,
				str:   "20%",
			},
			want: 6,
		},
		{
			name: "test.3",
			args: args{
				hight: 45,
				str:   "30%",
			},
			want: 13.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := position(tt.args.hight, tt.args.str); got != tt.want {
				t.Errorf("position() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jumpPosition(t *testing.T) {
	type args struct {
		hight int
		str   string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test1",
			args: args{
				hight: 30,
				str:   "1",
			},
			want: 1,
		},
		{
			name: "test.3",
			args: args{
				hight: 10,
				str:   ".3",
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jumpPosition(tt.args.hight, tt.args.str); got != tt.want {
				t.Errorf("jumpPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}
