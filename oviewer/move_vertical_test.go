package oviewer

import (
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
		// TODO: Add test cases.
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
