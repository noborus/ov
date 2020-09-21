package oviewer

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/gdamore/tcell"
)

func Test_parseString(t *testing.T) {
	type args struct {
		line     string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want lineContents
	}{
		{
			name: "test1",
			args: args{
				line: "test", tabWidth: 8,
			},
			want: lineContents{
				contents: []content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('t'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('e'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('s'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('t'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 2, 3: 3, 4: 4},
				bcw:     []int{0, 1, 2, 3, 4},
			},
		},
		{
			name: "testASCII",
			args: args{line: "abc", tabWidth: 4},
			want: lineContents{
				contents: []content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('c'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 2, 3: 3},
				bcw:     []int{0, 1, 2, 3},
			},
		},
		{
			name: "testHiragana",
			args: args{line: "あ", tabWidth: 4},
			want: lineContents{
				contents: []content{
					{width: 2, style: tcell.StyleDefault, mainc: rune('あ'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				},
				byteMap: map[int]int{0: 0, 3: 2},
				bcw:     []int{0, 2, 2, 2},
			},
		},
		{
			name: "testKANJI",
			args: args{line: "漢", tabWidth: 4},
			want: lineContents{
				contents: []content{
					{width: 2, style: tcell.StyleDefault, mainc: rune('漢'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				},
				byteMap: map[int]int{0: 0, 3: 2},
				bcw:     []int{0, 2, 2, 2},
			},
		},
		{
			name: "testMIX",
			args: args{line: "abc漢", tabWidth: 4},
			want: lineContents{
				contents: []content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('c'), combc: nil},
					{width: 2, style: tcell.StyleDefault, mainc: rune('漢'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 2, 3: 3, 6: 5},
				bcw:     []int{0, 1, 2, 3, 5, 5, 5},
			},
		},
		{
			name: "testTab",
			args: args{line: "a\tb", tabWidth: 4},
			want: lineContents{
				contents: []content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 4, 3: 5},
				bcw:     []int{0, 1, 4, 5},
			},
		},
		{
			name: "testTabMinus",
			args: args{line: "a\tb", tabWidth: -1},
			want: lineContents{
				contents: []content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Reverse(true), mainc: rune('\\'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Reverse(true), mainc: rune('t'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 1, 3: 3, 4: 4},
				bcw:     []int{0, 1, 3, 4},
			},
		},
		{
			name: "red",
			args: args{
				line: "\x1B[31mred\x1B[m", tabWidth: 8,
			},
			want: lineContents{
				contents: []content{
					{width: 1, style: tcell.StyleDefault.Foreground(tcell.Color(1)), mainc: rune('r'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Foreground(tcell.Color(1)), mainc: rune('e'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Foreground(tcell.Color(1)), mainc: rune('d'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 2, 3: 3},
				bcw:     []int{0, 0, 0, 0, 0, 0, 1, 2, 3, 3, 3, 3},
			},
		},
		{
			name: "bold",
			args: args{
				line: "\x1B[1mbold\x1B[m", tabWidth: 8,
			},
			want: lineContents{
				contents: []content{
					{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('b'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('o'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('l'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('d'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 2, 3: 3, 4: 4},
				bcw:     []int{0, 0, 0, 0, 0, 1, 2, 3, 4, 4, 4, 4},
			},
		},
		{
			name: "testOverstrike",
			args: args{line: "a\ba", tabWidth: 8},
			want: lineContents{
				contents: []content{
					{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('a'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1},
				bcw:     []int{0, 1, 0, 1},
			},
		},
		{
			name: "testOverstrike2",
			args: args{line: "\ba", tabWidth: 8},
			want: lineContents{
				contents: []content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 0, 2: 1},
				bcw:     []int{0, 0, 1},
			},
		},
		{
			name: "testOverstrike3",
			args: args{line: "あ\bあ", tabWidth: 8},
			want: lineContents{
				contents: []content{
					{width: 2, style: tcell.StyleDefault.Bold(true), mainc: rune('あ'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				},
				byteMap: map[int]int{0: 0, 3: 2},
				bcw:     []int{0, 2, 2, 2, 0, 2, 2, 2},
			},
		},
		{
			name: "testOverstrike4",
			args: args{line: "\a", tabWidth: 8},
			want: lineContents{
				contents: nil,
				byteMap:  map[int]int{0: 0, 1: 0},
				bcw:      []int{0, 0},
			},
		},
		{
			name: "testOverstrike5",
			args: args{line: "あ\bああ\bあ", tabWidth: 8},
			want: lineContents{
				contents: []content{
					{width: 2, style: tcell.StyleDefault.Bold(true), mainc: rune('あ'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
					{width: 2, style: tcell.StyleDefault.Bold(true), mainc: rune('あ'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				},
				byteMap: map[int]int{0: 0, 3: 2, 6: 4},
				bcw:     []int{0, 2, 2, 2, 0, 2, 2, 2, 4, 4, 4, 2, 4, 4, 4},
			},
		},
		{
			name: "testOverstrikeUnderLine",
			args: args{line: "_\ba", tabWidth: 8},
			want: lineContents{
				contents: []content{
					{width: 1, style: tcell.StyleDefault.Underline(true), mainc: rune('a'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1},
				bcw:     []int{0, 1, 0, 1},
			},
		},
		{
			name: "testOverstrikeUnderLine2",
			args: args{line: "_\bあ", tabWidth: 8},
			want: lineContents{
				contents: []content{
					{width: 2, style: tcell.StyleDefault.Underline(true), mainc: rune('あ'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 3: 2},
				bcw:     []int{0, 1, 0, 2, 2, 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseString(tt.args.line, tt.args.tabWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lastContent(t *testing.T) {
	type args struct {
		contents []content
	}
	tests := []struct {
		name string
		args args
		want content
	}{
		{
			name: "tsetNil",
			args: args{
				contents: nil,
			},
			want: content{},
		},
		{
			name: "tset1",
			args: args{
				contents: []content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('t'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('e'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('s'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('t'), combc: nil},
				},
			},
			want: content{width: 1, style: tcell.StyleDefault, mainc: rune('t'), combc: nil},
		},
		{
			name: "tsetWide",
			args: args{
				contents: []content{
					{width: 2, style: tcell.StyleDefault, mainc: rune('あ'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
					{width: 2, style: tcell.StyleDefault, mainc: rune('い'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
					{width: 2, style: tcell.StyleDefault, mainc: rune('う'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
				},
			},
			want: content{width: 2, style: tcell.StyleDefault, mainc: rune('う'), combc: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lastContent(tt.args.contents); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lastContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_csToStyle(t *testing.T) {
	type args struct {
		style        tcell.Style
		csiParameter *bytes.Buffer
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
				csiParameter: bytes.NewBufferString("38;5;1"),
			},
			want: tcell.StyleDefault.Foreground(tcell.ColorMaroon),
		},
		{
			name: "color8bit2",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: bytes.NewBufferString("38;5;21"),
			},
			want: tcell.StyleDefault.Foreground(tcell.GetColor("#0000ff")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := csToStyle(tt.args.style, tt.args.csiParameter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("csToStyle() = %v, want %v", got, tt.want)
				gfg, gbg, gattr := got.Decompose()
				wfg, wbg, wattr := tt.want.Decompose()
				t.Errorf("csToStyle() = %x,%x,%v, want %x,%x,%v", gfg.Hex(), gbg.Hex(), gattr, wfg.Hex(), wbg.Hex(), wattr)
			}
		})
	}
}

func Test_strToContents(t *testing.T) {
	type args struct {
		line     string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want []content
	}{
		{
			name: "test1",
			args: args{line: "1", tabWidth: 4},
			want: []content{
				{width: 1, style: tcell.StyleDefault, mainc: rune('1'), combc: nil},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strToContents(tt.args.line, tt.args.tabWidth); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("strToContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_contentsByteNum(t *testing.T) {
	type args struct {
		str   string
		width int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test1",
			args: args{
				str:   "a",
				width: 0,
			},
			want: 0,
		},
		{
			name: "test2",
			args: args{
				str:   "a",
				width: 1,
			},
			want: 1,
		},
		{
			name: "test3",
			args: args{
				str:   "abcdefg",
				width: 4,
			},
			want: 4,
		},
		{
			name: "testHiragana",
			args: args{
				str:   "あいう",
				width: 4,
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := parseString(tt.args.str, 8)
			if got := lc.contentsByteNum(tt.args.width); got != tt.want {
				t.Errorf("widthToByte() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_contentsWidth(t *testing.T) {
	type args struct {
		str   string
		nbyte int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test1",
			args: args{
				str:   "a",
				nbyte: 0,
			},
			want: 0,
		},
		{
			name: "test2",
			args: args{
				str:   "a",
				nbyte: 1,
			},
			want: 1,
		},
		{
			name: "test3",
			args: args{
				str:   "abcdefg",
				nbyte: 4,
			},
			want: 4,
		},
		{
			name: "testHiragana",
			args: args{
				str:   "あいう",
				nbyte: 7,
			},
			want: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := parseString(tt.args.str, 8)
			if got := lc.contentsWidth(tt.args.nbyte); got != tt.want {
				t.Errorf("contentsWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}
