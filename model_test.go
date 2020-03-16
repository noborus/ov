package zpager

import (
	"reflect"
	"testing"
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
				{mainc: rune('1'), combc: []rune{}, width: 1},
			},
		},
		{
			name: "testASCII",
			args: args{line: "abc", subStr: "", tabWidth: 4},
			want: []content{
				{mainc: rune('a'), combc: []rune{}, width: 1},
				{mainc: rune('b'), combc: []rune{}, width: 1},
				{mainc: rune('c'), combc: []rune{}, width: 1},
			},
		},
		{
			name: "testHiragana",
			args: args{line: "あ", subStr: "", tabWidth: 4},
			want: []content{
				{mainc: rune('あ'), combc: []rune{}, width: 2},
				{mainc: 0, combc: []rune{}, width: 0},
			},
		},
		{
			name: "testKANJI",
			args: args{line: "漢", subStr: "", tabWidth: 4},
			want: []content{
				{mainc: rune('漢'), combc: []rune{}, width: 2},
				{mainc: 0, combc: []rune{}, width: 0},
			},
		},
		{
			name: "testMIX",
			args: args{line: "abc漢", subStr: "", tabWidth: 4},
			want: []content{
				{mainc: rune('a'), combc: []rune{}, width: 1},
				{mainc: rune('b'), combc: []rune{}, width: 1},
				{mainc: rune('c'), combc: []rune{}, width: 1},
				{mainc: rune('漢'), combc: []rune{}, width: 2},
				{mainc: 0, combc: []rune{}, width: 0},
			},
		},
		{
			name: "testTab",
			args: args{line: "a\tb", subStr: "", tabWidth: 4},
			want: []content{
				{mainc: rune('a'), combc: []rune{}, width: 1},
				{mainc: rune(' '), combc: []rune{}, width: 1},
				{mainc: rune(' '), combc: []rune{}, width: 1},
				{mainc: rune(' '), combc: []rune{}, width: 1},
				{mainc: rune('b'), combc: []rune{}, width: 1},
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
