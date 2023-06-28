package oviewer

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestDocument_moveLine(t *testing.T) {
	type args struct {
		lN int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "moveNoLine",
			args: args{
				lN: 10,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			if got := m.moveLine(tt.args.lN); got != tt.want {
				t.Errorf("Document.moveLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_moveLineFile(t *testing.T) {
	type fields struct {
		fileName string
	}
	type args struct {
		lN int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "moveLine",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
			},
			args: args{
				lN: 10,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := OpenDocument(tt.fields.fileName)
			if err != nil {
				t.Fatal(err)
			}
			for !m.BufEOF() {
			}
			if got := m.moveLine(tt.args.lN); got != tt.want {
				t.Errorf("Document.moveLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_moveTopFile(t *testing.T) {
	type fields struct {
		fileName string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "moveTop",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := OpenDocument(tt.fields.fileName)
			if err != nil {
				t.Fatal(err)
			}
			for !m.BufEOF() {
			}
			m.moveTop()
			if got := m.topLN; got != tt.want {
				t.Errorf("Document.moveLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_moveLineNth(t *testing.T) {
	type fields struct {
		fileName string
		width    int
		wrap     bool
	}
	type args struct {
		lN  int
		nTh int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		want1  int
	}{
		{
			name: "moveNoLineNth",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				width:    10,
				wrap:     true,
			},
			args: args{
				lN:  1,
				nTh: 6,
			},
			want:  1,
			want1: 6,
		},
		{
			name: "moveNoLineNthNone",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				width:    10,
				wrap:     true,
			},
			args: args{
				lN:  3,
				nTh: 6,
			},
			want:  3,
			want1: 0,
		},
		{
			name: "moveNoLineNthNoWrap",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				width:    10,
				wrap:     false,
			},
			args: args{
				lN:  1,
				nTh: 6,
			},
			want:  1,
			want1: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := OpenDocument(tt.fields.fileName)
			if err != nil {
				t.Fatal(err)
			}
			for !m.BufEOF() {
			}
			m.width = tt.fields.width
			m.WrapMode = tt.fields.wrap
			got, got1 := m.moveLineNth(tt.args.lN, tt.args.nTh)
			if got != tt.want {
				t.Errorf("Document.moveLineNth() LN = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Document.moveLineNth() nTh = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestDocument_moveBottomFile(t *testing.T) {
	type fields struct {
		fileName string
		width    int
		height   int
		wrap     bool
	}
	tests := []struct {
		name   string
		fields fields
		want   int
		want1  int
	}{
		{
			name: "moveBottomNoWrap",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				width:    10,
				height:   10,
				wrap:     false,
			},
			want:  0,
			want1: 33,
		},
		{
			name: "moveBottomWrap",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				width:    10,
				height:   10,
				wrap:     true,
			},
			want:  70,
			want1: 40,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := OpenDocument(tt.fields.fileName)
			if err != nil {
				t.Fatal(err)
			}
			m.width = tt.fields.width
			m.height = tt.fields.height
			m.WrapMode = tt.fields.wrap
			for !m.BufEOF() {
			}
			m.moveBottom()
			if got := m.topLX; got != tt.want {
				t.Errorf("Document.moveLine() LX = %v, want %v", got, tt.want)
			}
			if got := m.topLN; got != tt.want1 {
				t.Errorf("Document.moveLine() LN = %v, want %v", got, tt.want1)
			}
		})
	}
}

func Test_leftMostX(t *testing.T) {
	type args struct {
		width int
		lc    contents
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "oneLine",
			args: args{
				width: 10,
				lc:    parseString("0123456789", 8),
			},
			want: []int{0},
		},
		{
			name: "twoLine",
			args: args{
				width: 10,
				lc:    parseString("000000000111111111", 8),
			},
			want: []int{0, 10},
		},
		{
			name: "fullWidth",
			args: args{
				width: 10,
				// 0あいうえ
				// おかきくけ
				// こ
				lc: parseString("0あいうえおかきくけこ", 8),
			},
			want: []int{0, 9, 19},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := leftMostX(tt.args.width, tt.args.lc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("leftMostX() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_moveYUp(t *testing.T) {
	type fields struct {
		fileName string
		lX       int
		lN       int
		width    int
		height   int
		wrap     bool
	}
	type args struct {
		moveY int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		want1  int
	}{
		{
			name: "moveYUp",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				lX:       0,
				lN:       3,
				width:    10,
				height:   10,
				wrap:     false,
			},
			args: args{
				moveY: 1,
			},
			want:  0,
			want1: 2,
		},
		{
			name: "moveYUpWrap",
			fields: fields{
				fileName: filepath.Join(testdata, "move.txt"),
				lX:       40,
				lN:       1,
				width:    10,
				height:   10,
				wrap:     true,
			},
			args: args{
				moveY: 1,
			},
			want:  30,
			want1: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := OpenDocument(tt.fields.fileName)
			if err != nil {
				t.Fatal(err)
			}
			for !m.BufEOF() {
			}
			m.width = tt.fields.width
			m.height = tt.fields.height
			m.WrapMode = tt.fields.wrap
			m.topLN = tt.fields.lN
			m.topLX = tt.fields.lX
			m.moveYUp(tt.args.moveY)
			if m.topLX != tt.want {
				t.Errorf("Document.moveYUp() LX = %v, want %v", m.topLX, tt.want)
			}
			if m.topLN != tt.want1 {
				t.Errorf("Document.moveYUp() LN = %v, want %v", m.topLN, tt.want1)
			}
		})
	}
}
