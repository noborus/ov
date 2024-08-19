package oviewer

import (
	"os"
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func setup() {
	OverStrikeStyle = ToTcellStyle(OVStyle{Bold: true})
	OverLineStyle = ToTcellStyle(OVStyle{Underline: true})
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	os.Exit(ret)
}

func Test_StrToContentsNormal(t *testing.T) {
	t.Parallel()
	type args struct {
		line     string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want contents
	}{
		{
			name: "test1",
			args: args{
				line: "test", tabWidth: 8,
			},
			want: contents{
				{width: 1, mainc: rune('t')},
				{width: 1, mainc: rune('e')},
				{width: 1, mainc: rune('s')},
				{width: 1, mainc: rune('t')},
			},
		},
		{
			name: "testASCII",
			args: args{line: "abc", tabWidth: 4},
			want: contents{
				{width: 1, mainc: rune('a')},
				{width: 1, mainc: rune('b')},
				{width: 1, mainc: rune('c')},
			},
		},
		{
			name: "testHiragana",
			args: args{line: "あ", tabWidth: 4},
			want: contents{
				{width: 2, mainc: rune('あ')},
				{width: 0, mainc: 0},
			},
		},
		{
			name: "testKANJI",
			args: args{line: "漢", tabWidth: 4},
			want: contents{
				{width: 2, mainc: rune('漢')},
				{width: 0, mainc: 0},
			},
		},
		{
			name: "testMIX",
			args: args{line: "abc漢", tabWidth: 4},
			want: contents{
				{width: 1, mainc: rune('a')},
				{width: 1, mainc: rune('b')},
				{width: 1, mainc: rune('c')},
				{width: 2, mainc: rune('漢')},
				{width: 0, mainc: 0},
			},
		},
		{
			name: "testTab",
			args: args{line: "a\tb", tabWidth: 4},
			want: contents{
				{width: 1, style: tcell.StyleDefault, mainc: rune('a')},
				{width: 1, style: tcell.StyleDefault, mainc: rune('\t')},
				{width: 1, style: tcell.StyleDefault, mainc: rune(0)},
				{width: 1, style: tcell.StyleDefault, mainc: rune(0)},
				{width: 1, style: tcell.StyleDefault, mainc: rune('b')},
			},
		},
		{
			name: "testFormFeed",
			args: args{line: "a\fa", tabWidth: 4},
			want: contents{
				{width: 1, style: tcell.StyleDefault, mainc: rune('a')},
				{width: 0, style: tcell.StyleDefault, mainc: rune('\f')},
				{width: 1, style: tcell.StyleDefault, mainc: rune('a')},
			},
		},
		{
			name: "testDEL",
			args: args{line: "a\u007fa", tabWidth: 4},
			want: contents{
				{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: []rune{0x7f}},
				{width: 1, style: tcell.StyleDefault, mainc: rune('a')},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StrToContents(tt.args.line, tt.args.tabWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_StrToContentsOverlapping(t *testing.T) {
	t.Parallel()
	type args struct {
		line     string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want contents
	}{
		{
			name: "testOverstrike",
			args: args{line: "a\ba", tabWidth: 8},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('a'), combc: nil},
			},
		},
		{
			name: "testOverstrike2",
			args: args{line: "\ba", tabWidth: 8},
			want: contents{
				{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
			},
		},
		{
			name: "testOverstrike3",
			args: args{line: "あ\bあ", tabWidth: 8},
			want: contents{
				{width: 2, style: tcell.StyleDefault.Bold(true), mainc: rune('あ'), combc: nil},
				{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
			},
		},
		{
			name: "testOverstrike4",
			args: args{line: "\a", tabWidth: 8},
			want: contents{
				{width: 0, style: tcell.StyleDefault, mainc: '\a', combc: nil},
			},
		},
		{
			name: "testOverstrike5",
			args: args{line: "あ\bああ\bあ", tabWidth: 8},
			want: contents{
				{width: 2, style: tcell.StyleDefault.Bold(true), mainc: rune('あ'), combc: nil},
				{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				{width: 2, style: tcell.StyleDefault.Bold(true), mainc: rune('あ'), combc: nil},
				{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
			},
		},
		{
			name: "testOverstrikeUnderLine",
			args: args{line: "_\ba", tabWidth: 8},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Underline(true), mainc: rune('a'), combc: nil},
			},
		},
		{
			name: "testOverstrikeUnderLine2",
			args: args{line: "_\bあ", tabWidth: 8},
			want: contents{
				{width: 2, style: tcell.StyleDefault.Underline(true), mainc: rune('あ'), combc: nil},
				{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StrToContents(tt.args.line, tt.args.tabWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_StrToContentsStyle1(t *testing.T) {
	t.Parallel()
	type args struct {
		line     string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want contents
	}{
		{
			name: "testTabMinus",
			args: args{line: "a\tb", tabWidth: -1},
			want: contents{
				{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Reverse(true), mainc: rune('\\'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Reverse(true), mainc: rune('t'), combc: nil},
				{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: nil},
			},
		},
		{
			name: "red",
			args: args{
				line: "\x1B[31mred\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorMaroon), mainc: rune('r'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorMaroon), mainc: rune('e'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorMaroon), mainc: rune('d'), combc: nil},
			},
		},
		{
			name: "bright color",
			args: args{
				line: "\x1B[90mc\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorGray), mainc: rune('c'), combc: nil},
			},
		},
		{
			name: "bright color back",
			args: args{
				line: "\x1B[100mc\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Background(tcell.ColorGray), mainc: rune('c'), combc: nil},
			},
		},
		{
			name: "216 color",
			args: args{
				line: "\x1B[38;5;31mc\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.NewRGBColor(0, 102, 153)), mainc: 'c', combc: nil},
			},
		},
		{
			name: "256",
			args: args{
				line: "\x1b[38;5;1mc\x1b[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorValid + 1), mainc: 'c', combc: nil},
			},
		},
		{
			name: "256 both",
			args: args{
				line: "\x1b[38;5;1;48;5;2mc\x1b[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorValid + 1).Background(tcell.ColorValid + 2), mainc: 'c', combc: nil},
			},
		},
		{
			name: "24bitcolor",
			args: args{
				line: "\x1b[38;2;250;123;250mc\x1b[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.NewRGBColor(250, 123, 250)), mainc: rune('c'), combc: nil},
			},
		},
		{
			name: "24bitcolor both",
			args: args{
				line: "\x1b[38;2;255;0;0;48;2;0;0;255mc\x1b[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.NewRGBColor(255, 0, 0)).Background(tcell.NewRGBColor(0, 0, 255)), mainc: rune('c'), combc: nil},
			},
		},
		{
			name: "default color",
			args: args{
				line: "\x1B[39md\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault, mainc: rune('d'), combc: nil},
			},
		},
		{
			name: "bold",
			args: args{
				line: "\x1B[1mbold\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('b'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('o'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('l'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('d'), combc: nil},
			},
		},
		{
			name: "reset",
			args: args{
				line: "\x1B[31mr\x1B[me", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorMaroon), mainc: rune('r'), combc: nil},
				{width: 1, style: tcell.StyleDefault, mainc: rune('e'), combc: nil},
			},
		},
		{
			name: "reset2",
			args: args{
				line: "\x1B[31mr\x1Bce", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorMaroon), mainc: rune('r'), combc: nil},
				{width: 1, style: tcell.StyleDefault, mainc: rune('e'), combc: nil},
			},
		},
		{
			name: "substring",
			args: args{
				line: "\x1B]sub\x1Bmt", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault, mainc: rune('t'), combc: nil},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StrToContents(tt.args.line, tt.args.tabWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = \n%#v, want \n%#v", got, tt.want)
			}
		})
	}
}

func Test_StrToContentsStyle2(t *testing.T) {
	type args struct {
		str      string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want contents
	}{
		{
			name: "OverLine",
			args: args{
				str: "\x1B[53mol\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Underline(true), mainc: rune('o'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Underline(true), mainc: rune('l'), combc: nil},
			},
		},
		{
			name: "UnOverLine",
			args: args{
				str: "\x1B[53mo\x1B[m\x1B[55mu\x1B[m\x1B[53ml\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Underline(true), mainc: rune('o'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Underline(false), mainc: rune('u'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Underline(true), mainc: rune('l'), combc: nil},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrToContents(tt.args.str, tt.args.tabWidth); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_parseStringUnStyle(t *testing.T) {
	t.Parallel()
	type args struct {
		line     string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want contents
	}{
		{
			name: "unBold",
			args: args{
				line: "\x1B[1mbo\x1B[22mld\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('b'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('o'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Bold(false), mainc: rune('l'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Bold(false), mainc: rune('d'), combc: nil},
			},
		},
		{
			name: "unItalic",
			args: args{
				line: "\x1B[3mab\x1B[23mcd\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Italic(true), mainc: rune('a'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Italic(true), mainc: rune('b'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Italic(false), mainc: rune('c'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Italic(false), mainc: rune('d'), combc: nil},
			},
		},
		{
			name: "unUnderline",
			args: args{
				line: "\x1B[4mab\x1B[24mcd\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Underline(true), mainc: rune('a'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Underline(true), mainc: rune('b'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Underline(false), mainc: rune('c'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Underline(false), mainc: rune('d'), combc: nil},
			},
		},
		{
			name: "unBlink",
			args: args{
				line: "\x1B[5mab\x1B[25mcd\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Blink(true), mainc: rune('a'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Blink(true), mainc: rune('b'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Blink(false), mainc: rune('c'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Blink(false), mainc: rune('d'), combc: nil},
			},
		},
		{
			name: "unReverse",
			args: args{
				line: "\x1B[7mab\x1B[27mcd\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Reverse(true), mainc: rune('a'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Reverse(true), mainc: rune('b'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Reverse(false), mainc: rune('c'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Reverse(false), mainc: rune('d'), combc: nil},
			},
		},
		{
			name: "unStrikeThrough",
			args: args{
				line: "\x1B[9mab\x1B[29mcd\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.StrikeThrough(true), mainc: rune('a'), combc: nil},
				{width: 1, style: tcell.StyleDefault.StrikeThrough(true), mainc: rune('b'), combc: nil},
				{width: 1, style: tcell.StyleDefault.StrikeThrough(false), mainc: rune('c'), combc: nil},
				{width: 1, style: tcell.StyleDefault.StrikeThrough(false), mainc: rune('d'), combc: nil},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StrToContents(tt.args.line, tt.args.tabWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = \n%#v, want \n%#v", got, tt.want)
			}
		})
	}
}

func Test_StrToContentsCombining(t *testing.T) {
	t.Parallel()
	type args struct {
		line     string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want contents
	}{
		{
			name: "testIVS",
			args: args{line: string([]rune{'\u908a', '\U000e0104'}), tabWidth: 4},
			want: contents{
				{width: 2, style: tcell.StyleDefault, mainc: rune('邊'), combc: []rune{0x000e0104}},
				{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
			},
		},
		{
			name: "testEmojiCombiningSequence",
			args: args{line: string([]rune{'1', '\ufe0f', '\u20e3'}), tabWidth: 4},
			want: contents{
				{width: 1, style: tcell.StyleDefault, mainc: rune('1'), combc: []rune{0xfe0f, 0x20e3}},
			},
		},
		{
			name: "testEmojiModifierSequence",
			args: args{line: string([]rune{'\u261d', '\U0001f3fb'}), tabWidth: 4},
			want: contents{
				{width: 1, style: tcell.StyleDefault, mainc: rune('☝'), combc: []rune{0x0001f3fb}},
			},
		},
		{
			name: "testEmojiFlagSequence",
			args: args{line: string([]rune{'\U0001f1ef', '\U0001f1f5'}), tabWidth: 4},
			want: contents{
				{width: 1, style: tcell.StyleDefault, mainc: rune('🇯'), combc: []rune{'🇵'}},
			},
		},
		{
			name: "testEmojiZWJSequence",
			args: args{line: string([]rune{'\U0001f468', '\u200d', '\U0001f468', '\u200d', '\U0001f466'}), tabWidth: 4},
			want: contents{
				{width: 2, style: tcell.StyleDefault, mainc: rune('👨'), combc: []rune{'\u200d', '\U0001f468', '\u200d', '\U0001f466'}},
				{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StrToContents(tt.args.line, tt.args.tabWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_StrToContentsHyperlink(t *testing.T) {
	t.Parallel()
	type args struct {
		line     string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want contents
	}{
		{
			name: "testHyperLink",
			args: args{line: "\x1b]8;;http://example.com\x1b\\link\x1b]8;;\x1b\\", tabWidth: 8},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Url("http://example.com"), mainc: rune('l'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com"), mainc: rune('i'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com"), mainc: rune('n'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com"), mainc: rune('k'), combc: nil},
			},
		},
		{
			name: "testHyperLinkID",
			args: args{line: "\x1b]8;1;http://example.com\x1b\\link\x1b]8;;\x1b\\", tabWidth: 8},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Url("http://example.com").UrlId("1"), mainc: rune('l'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com").UrlId("1"), mainc: rune('i'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com").UrlId("1"), mainc: rune('n'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com").UrlId("1"), mainc: rune('k'), combc: nil},
			},
		},
		{
			name: "testHyperLinkfile",
			args: args{line: "\x1b]8;;file:///file\afile\x1b]8;;\a", tabWidth: 8},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Url("file:///file"), mainc: rune('f'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Url("file:///file"), mainc: rune('i'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Url("file:///file"), mainc: rune('l'), combc: nil},
				{width: 1, style: tcell.StyleDefault.Url("file:///file"), mainc: rune('e'), combc: nil},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StrToContents(tt.args.line, tt.args.tabWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = \n%#v, want \n%#v", got, tt.want)
			}
		})
	}
}

func Test_lastContent(t *testing.T) {
	t.Parallel()
	type args struct {
		lc contents
	}
	tests := []struct {
		name string
		args args
		want content
	}{
		{
			name: "tsetNil",
			args: args{
				lc: nil,
			},
			want: content{},
		},
		{
			name: "tset1",
			args: args{
				lc: contents{
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
				lc: contents{
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.args.lc.last(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lastContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ContentsToStr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		str   string
		want1 string
		want2 widthPos
	}{
		{
			name:  "test1",
			str:   "test",
			want1: "test",
			want2: widthPos{0, 1, 2, 3, 4},
		},
		{
			name:  "testTab",
			str:   "t\test",
			want1: "t\test",
			want2: widthPos{0, 1, 8, 9, 10, 11},
		},
		{
			name:  "testBackSpace",
			str:   "t\btest",
			want1: "test",
			want2: widthPos{0, 1, 2, 3, 4},
		},
		{
			name:  "testEscape",
			str:   "\x1b[40;38;5;82mHello \x1b[30;48;5;82mWorld\x1b[0m",
			want1: "Hello World",
			want2: widthPos{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
		{
			name:  "testMultiByte",
			str:   "あいうえお倍",
			want1: "あいうえお倍",
			want2: widthPos{0, 2, 2, 2, 4, 4, 4, 6, 6, 6, 8, 8, 8, 10, 10, 10, 12, 12, 12},
		},
		{
			name:  "testEmojiZWJSequence",
			str:   string([]rune{'\U0001f468', '\u200d', '\U0001f468', '\u200d', '\U0001f466'}),
			want1: "👨‍👨‍👦",
			want2: widthPos{0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lc := StrToContents(tt.str, 8)
			got1, got2 := ContentsToStr(lc)
			if got1 != tt.want1 {
				t.Errorf("contentsToStr() = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("contentsToStr() = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_widthPos_x(t *testing.T) {
	t.Parallel()
	type args struct {
		x int
	}
	tests := []struct {
		name string
		pos  widthPos
		args args
		want int
	}{
		{
			name: "\ttest",
			pos:  widthPos{0, 1, 8, 9, 10, 11},
			args: args{2},
			want: 8,
		},
		{
			name: "あいうえお",
			pos:  widthPos{0, 2, 2, 2, 4, 4, 4, 6, 6, 6, 8, 8, 8, 10, 10, 10},
			args: args{12},
			want: 8,
		},
		{
			name: "あいうえお2",
			pos:  widthPos{0, 2, 2, 2, 4, 4, 4, 6, 6, 6, 8, 8, 8, 10, 10, 10},
			args: args{20},
			want: 10,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.pos.x(tt.args.x); got != tt.want {
				t.Errorf("widthPos.x() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_widthPos_n(t *testing.T) {
	t.Parallel()
	type args struct {
		w int
	}
	tests := []struct {
		name string
		pos  widthPos
		args args
		want int
	}{
		{
			name: "\ttest",
			pos:  widthPos{0, 1, 8, 9, 10, 11},
			args: args{8},
			want: 2,
		},
		{
			name: "あいうえお",
			pos:  widthPos{0, 2, 2, 2, 4, 4, 4, 6, 6, 6, 8, 8, 8, 10, 10, 10},
			args: args{8},
			want: 12,
		},
		{
			name: "あいうえお2",
			pos:  widthPos{0, 2, 2, 2, 4, 4, 4, 6, 6, 6, 8, 8, 8, 10, 10, 10},
			args: args{9},
			want: 15,
		},
		{
			name: "あいうえお3",
			pos:  widthPos{0, 2, 2, 2, 4, 4, 4, 6, 6, 6, 8, 8, 8, 10, 10, 10},
			args: args{20},
			want: 15,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.pos.n(tt.args.w); got != tt.want {
				t.Errorf("widthPos.n() = %v, want %v", got, tt.want)
			}
		})
	}
}
