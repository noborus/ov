package zpager

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell"
)

func Test_strToContent(t *testing.T) {
	type args struct {
		line     string
		subStr   string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want []content
	}{
		{
			name: "test1",
			args: args{line: "1", subStr: "", tabWidth: 4},
			want: []content{
				{width: 1, style: tcell.StyleDefault, mainc: rune('1'), combc: []rune{}},
			},
		},
		{
			name: "testASCII",
			args: args{line: "abc", subStr: "", tabWidth: 4},
			want: []content{
				{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: []rune{}},
				{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: []rune{}},
				{width: 1, style: tcell.StyleDefault, mainc: rune('c'), combc: []rune{}},
			},
		},
		{
			name: "testHiragana",
			args: args{line: "あ", subStr: "", tabWidth: 4},
			want: []content{
				{width: 2, style: tcell.StyleDefault, mainc: rune('あ'), combc: []rune{}},
				{width: 0, style: tcell.StyleDefault, mainc: 0, combc: []rune{}},
			},
		},
		{
			name: "testKANJI",
			args: args{line: "漢", subStr: "", tabWidth: 4},
			want: []content{
				{width: 2, style: tcell.StyleDefault, mainc: rune('漢'), combc: []rune{}},
				{width: 0, style: tcell.StyleDefault, mainc: 0, combc: []rune{}},
			},
		},
		{
			name: "testMIX",
			args: args{line: "abc漢", subStr: "", tabWidth: 4},
			want: []content{
				{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: []rune{}},
				{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: []rune{}},
				{width: 1, style: tcell.StyleDefault, mainc: rune('c'), combc: []rune{}},
				{width: 2, style: tcell.StyleDefault, mainc: rune('漢'), combc: []rune{}},
				{width: 0, style: tcell.StyleDefault, mainc: 0, combc: []rune{}},
			},
		},
		{
			name: "testTab",
			args: args{line: "a\tb", subStr: "", tabWidth: 4},
			want: []content{
				{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: []rune{}},
				{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: []rune{}},
				{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: []rune{}},
				{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: []rune{}},
				{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: []rune{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strToContent(tt.args.line, tt.args.subStr, tt.args.tabWidth); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("strToContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
