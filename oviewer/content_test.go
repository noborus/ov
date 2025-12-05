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
				{width: 1, str: "t"},
				{width: 1, str: "e"},
				{width: 1, str: "s"},
				{width: 1, str: "t"},
			},
		},
		{
			name: "testASCII",
			args: args{line: "abc", tabWidth: 4},
			want: contents{
				{width: 1, str: "a"},
				{width: 1, str: "b"},
				{width: 1, str: "c"},
			},
		},
		{
			name: "testHiragana",
			args: args{line: "あ", tabWidth: 4},
			want: contents{
				{width: 2, str: "あ"},
				{width: 0, str: ""},
			},
		},
		{
			name: "testKANJI",
			args: args{line: "漢", tabWidth: 4},
			want: contents{
				{width: 2, str: "漢"},
				{width: 0, str: ""},
			},
		},
		{
			name: "testMIX",
			args: args{line: "abc漢", tabWidth: 4},
			want: contents{
				{width: 1, str: "a"},
				{width: 1, str: "b"},
				{width: 1, str: "c"},
				{width: 2, str: "漢"},
				{width: 0, str: ""},
			},
		},
		{
			name: "testTab",
			args: args{line: "a\tb", tabWidth: 4},
			want: contents{
				{width: 1, style: tcell.StyleDefault, str: "a"},
				{width: 1, style: tcell.StyleDefault, str: "\t"},
				{width: 1, style: tcell.StyleDefault, str: ""},
				{width: 1, style: tcell.StyleDefault, str: ""},
				{width: 1, style: tcell.StyleDefault, str: "b"},
			},
		},
		{
			name: "testFormFeed",
			args: args{line: "a\fa", tabWidth: 4},
			want: contents{
				{width: 1, style: tcell.StyleDefault, str: "a"},
				{width: 0, style: tcell.StyleDefault, str: "\f"},
				{width: 1, style: tcell.StyleDefault, str: "a"},
			},
		},
		{
			name: "testDEL",
			args: args{line: "a\u007fa", tabWidth: 4},
			want: contents{
				{width: 1, style: tcell.StyleDefault, str: "a"},
				{width: 0, style: tcell.StyleDefault, str: "\u007f"},
				{width: 1, style: tcell.StyleDefault, str: "a"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StrToContents(tt.args.line, tt.args.tabWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = \n%#v, want \n%#v", got, tt.want)
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
				{width: 1, style: tcell.StyleDefault.Bold(true), str: "a"},
			},
		},
		{
			name: "testOverstrike2",
			args: args{line: "\ba", tabWidth: 8},
			want: contents{
				{width: 1, style: tcell.StyleDefault, str: "a"},
			},
		},
		{
			name: "testOverstrike3",
			args: args{line: "あ\bあ", tabWidth: 8},
			want: contents{
				{width: 2, style: tcell.StyleDefault.Bold(true), str: "あ"},
				{width: 0, style: tcell.StyleDefault, str: ""},
			},
		},
		{
			name: "testOverstrike4",
			args: args{line: "\a", tabWidth: 8},
			want: contents{
				{width: 0, style: tcell.StyleDefault, str: "\a"},
			},
		},
		{
			name: "testOverstrike5",
			args: args{line: "あ\bああ\bあ", tabWidth: 8},
			want: contents{
				{width: 2, style: tcell.StyleDefault.Bold(true), str: "あ"},
				{width: 0, style: tcell.StyleDefault, str: ""},
				{width: 2, style: tcell.StyleDefault.Bold(true), str: "あ"},
				{width: 0, style: tcell.StyleDefault, str: ""},
			},
		},
		{
			name: "testOverstrikeUnderLine",
			args: args{line: "_\ba", tabWidth: 8},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Underline(true), str: "a"},
			},
		},
		{
			name: "testOverstrikeUnderLine2",
			args: args{line: "_\bあ", tabWidth: 8},
			want: contents{
				{width: 2, style: tcell.StyleDefault.Underline(true), str: "あ"},
				{width: 0, style: tcell.StyleDefault, str: ""},
			},
		},
	}
	for _, tt := range tests {
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
				{width: 1, style: tcell.StyleDefault, str: "a"},
				{width: 1, style: tcell.StyleDefault.Reverse(true), str: "\\"},
				{width: 1, style: tcell.StyleDefault.Reverse(true), str: "t"},
				{width: 1, style: tcell.StyleDefault, str: "b"},
			},
		},
		{
			name: "red",
			args: args{
				line: "\x1B[31mred\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorMaroon), str: "r"},
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorMaroon), str: "e"},
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorMaroon), str: "d"},
			},
		},
		{
			name: "bright color",
			args: args{
				line: "\x1B[90mc\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorGray), str: "c"},
			},
		},
		{
			name: "bright color back",
			args: args{
				line: "\x1B[100mc\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Background(tcell.ColorGray), str: "c"},
			},
		},
		{
			name: "256",
			args: args{
				line: "\x1b[38;5;1mc\x1b[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorValid + 1), str: "c"},
			},
		},
		{
			name: "256 both",
			args: args{
				line: "\x1b[38;5;1;48;5;2mc\x1b[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorValid + 1).Background(tcell.ColorValid + 2), str: "c"},
			},
		},
		{
			name: "24bitcolor",
			args: args{
				line: "\x1b[38;2;250;123;250mc\x1b[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.NewRGBColor(250, 123, 250)), str: "c"},
			},
		},
		{
			name: "24bitcolor both",
			args: args{
				line: "\x1b[38;2;255;0;0;48;2;0;0;255mc\x1b[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.NewRGBColor(255, 0, 0)).Background(tcell.NewRGBColor(0, 0, 255)), str: "c"},
			},
		},
		{
			name: "default color",
			args: args{
				line: "\x1B[39md\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault, str: "d"},
			},
		},
		{
			name: "bold",
			args: args{
				line: "\x1B[1mbold\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Bold(true), str: "b"},
				{width: 1, style: tcell.StyleDefault.Bold(true), str: "o"},
				{width: 1, style: tcell.StyleDefault.Bold(true), str: "l"},
				{width: 1, style: tcell.StyleDefault.Bold(true), str: "d"},
			},
		},
		{
			name: "reset",
			args: args{
				line: "\x1B[31mr\x1B[me", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorMaroon), str: "r"},
				{width: 1, style: tcell.StyleDefault, str: "e"},
			},
		},
		{
			name: "reset2",
			args: args{
				line: "\x1B[31mr\x1Bce", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Foreground(tcell.ColorMaroon), str: "r"},
				{width: 1, style: tcell.StyleDefault, str: "e"},
			},
		},
		{
			name: "substring",
			args: args{
				line: "\x1B]sub\x1Bmt", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault, str: "t"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StrToContents(tt.args.line, tt.args.tabWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = \n%#v, want \n%#v", got, tt.want)
			}
		})
	}
}

func Test_StrToContentUnStyle(t *testing.T) {
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
				{width: 1, style: tcell.StyleDefault.Bold(true), str: "b"},
				{width: 1, style: tcell.StyleDefault.Bold(true), str: "o"},
				{width: 1, style: tcell.StyleDefault.Bold(false), str: "l"},
				{width: 1, style: tcell.StyleDefault.Bold(false), str: "d"},
			},
		},
		{
			name: "unItalic",
			args: args{
				line: "\x1B[3mab\x1B[23mcd\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Italic(true), str: "a"},
				{width: 1, style: tcell.StyleDefault.Italic(true), str: "b"},
				{width: 1, style: tcell.StyleDefault.Italic(false), str: "c"},
				{width: 1, style: tcell.StyleDefault.Italic(false), str: "d"},
			},
		},
		{
			name: "unUnderline",
			args: args{
				line: "\x1B[4mab\x1B[24mcd\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Underline(true), str: "a"},
				{width: 1, style: tcell.StyleDefault.Underline(true), str: "b"},
				{width: 1, style: tcell.StyleDefault.Underline(false), str: "c"},
				{width: 1, style: tcell.StyleDefault.Underline(false), str: "d"},
			},
		},
		{
			name: "unBlink",
			args: args{
				line: "\x1B[5mab\x1B[25mcd\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Blink(true), str: "a"},
				{width: 1, style: tcell.StyleDefault.Blink(true), str: "b"},
				{width: 1, style: tcell.StyleDefault.Blink(false), str: "c"},
				{width: 1, style: tcell.StyleDefault.Blink(false), str: "d"},
			},
		},
		{
			name: "unReverse",
			args: args{
				line: "\x1B[7mab\x1B[27mcd\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Reverse(true), str: "a"},
				{width: 1, style: tcell.StyleDefault.Reverse(true), str: "b"},
				{width: 1, style: tcell.StyleDefault.Reverse(false), str: "c"},
				{width: 1, style: tcell.StyleDefault.Reverse(false), str: "d"},
			},
		},
		{
			name: "unStrikeThrough",
			args: args{
				line: "\x1B[9mab\x1B[29mcd\x1B[m", tabWidth: 8,
			},
			want: contents{
				{width: 1, style: tcell.StyleDefault.StrikeThrough(true), str: "a"},
				{width: 1, style: tcell.StyleDefault.StrikeThrough(true), str: "b"},
				{width: 1, style: tcell.StyleDefault.StrikeThrough(false), str: "c"},
				{width: 1, style: tcell.StyleDefault.StrikeThrough(false), str: "d"},
			},
		},
	}
	for _, tt := range tests {
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
				{width: 2, style: tcell.StyleDefault, str: "\u908a\U000e0104"},
				{width: 0, style: tcell.StyleDefault, str: ""},
			},
		},
		{
			name: "testEmojiCombiningSequence",
			args: args{line: string([]rune{'1', '\ufe0f', '\u20e3'}), tabWidth: 4},
			want: contents{
				{width: 1, style: tcell.StyleDefault, str: "1\ufe0f\u20e3"},
			},
		},
		{
			name: "testEmojiModifierSequence",
			args: args{line: string([]rune{'\u261d', '\U0001f3fb'}), tabWidth: 4},
			want: contents{
				{width: 1, style: tcell.StyleDefault, str: "\u261d\U0001f3fb"},
			},
		},
		{
			name: "testEmojiFlagSequence",
			args: args{line: string([]rune{'\U0001f1ef', '\U0001f1f5'}), tabWidth: 4},
			want: contents{
				{width: 2, style: tcell.StyleDefault, str: "\U0001f1ef\U0001f1f5"},
				{width: 0, style: tcell.StyleDefault, str: ""},
			},
		},
		{
			name: "testEmojiZWJSequence",
			args: args{line: string([]rune{'\U0001f468', '\u200d', '\U0001f468', '\u200d', '\U0001f466'}), tabWidth: 4},
			want: contents{
				{width: 2, style: tcell.StyleDefault, str: "\U0001f468\u200d\U0001f468\u200d\U0001f466"},
				{width: 0, style: tcell.StyleDefault, str: ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StrToContents(tt.args.line, tt.args.tabWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = \n%#v, want \n%#v", got, tt.want)
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
				{width: 1, style: tcell.StyleDefault.Url("http://example.com"), str: "l"},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com"), str: "i"},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com"), str: "n"},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com"), str: "k"},
			},
		},
		{
			name: "testHyperLinkID",
			args: args{line: "\x1b]8;1;http://example.com\x1b\\link\x1b]8;;\x1b\\", tabWidth: 8},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Url("http://example.com").UrlId("1"), str: "l"},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com").UrlId("1"), str: "i"},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com").UrlId("1"), str: "n"},
				{width: 1, style: tcell.StyleDefault.Url("http://example.com").UrlId("1"), str: "k"},
			},
		},
		{
			name: "testHyperLinkfile",
			args: args{line: "\x1b]8;;file:///file\afile\x1b]8;;\x1b\\", tabWidth: 8},
			want: contents{
				{width: 1, style: tcell.StyleDefault.Url("file:///file"), str: "f"},
				{width: 1, style: tcell.StyleDefault.Url("file:///file"), str: "i"},
				{width: 1, style: tcell.StyleDefault.Url("file:///file"), str: "l"},
				{width: 1, style: tcell.StyleDefault.Url("file:///file"), str: "e"},
			},
		},
	}
	for _, tt := range tests {
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
					{width: 1, style: tcell.StyleDefault, str: "t"},
					{width: 1, style: tcell.StyleDefault, str: "e"},
					{width: 1, style: tcell.StyleDefault, str: "s"},
					{width: 1, style: tcell.StyleDefault, str: "t"},
				},
			},
			want: content{width: 1, style: tcell.StyleDefault, str: "t"},
		},
		{
			name: "tsetWide",
			args: args{
				lc: contents{
					{width: 2, style: tcell.StyleDefault, str: "あ"},
					{width: 1, style: tcell.StyleDefault, str: " "},
					{width: 2, style: tcell.StyleDefault, str: "い"},
					{width: 1, style: tcell.StyleDefault, str: " "},
					{width: 2, style: tcell.StyleDefault, str: "う"},
					{width: 1, style: tcell.StyleDefault, str: " "},
				},
			},
			want: content{width: 2, style: tcell.StyleDefault, str: "う"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.args.lc.last(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lastContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
