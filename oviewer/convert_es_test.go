package oviewer

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func Test_escapeSequence_convert(t *testing.T) {
	type fields struct {
		state int
	}
	type args struct {
		st *parseState
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      bool
		wantState int
	}{
		{
			name: "test-escapeSequence",
			fields: fields{
				state: ansiText,
			},
			args: args{
				st: &parseState{
					str: "\x1b",
				},
			},
			want:      true,
			wantState: ansiEscape,
		},
		{
			name: "test-SubString",
			fields: fields{
				state: ansiEscape,
			},
			args: args{
				st: &parseState{
					str: "P",
				},
			},
			want:      true,
			wantState: ansiSubstring,
		},
		{
			name: "test-SubString2",
			fields: fields{
				state: ansiSubstring,
			},
			args: args{
				st: &parseState{
					str: "\x1b",
				},
			},
			want:      true,
			wantState: ansiControlSequence,
		},
		{
			name: "test-OtherSequence",
			fields: fields{
				state: ansiEscape,
			},
			args: args{
				st: &parseState{
					str: "(",
				},
			},
			want:      true,
			wantState: otherSequence,
		},
		{
			name: "test-Other",
			fields: fields{
				state: ansiEscape,
			},
			args: args{
				st: &parseState{
					str: "@",
				},
			},
			want:      true,
			wantState: ansiText,
		},
		{
			name: "test-OtherSequence2",
			fields: fields{
				state: otherSequence,
			},
			args: args{
				st: &parseState{
					str: "a",
				},
			},
			want:      true,
			wantState: otherSequence,
		},
		{
			name: "test-OtherSequence3",
			fields: fields{
				state: otherSequence,
			},
			args: args{
				st: &parseState{
					str: "B",
				},
			},
			want:      true,
			wantState: ansiText,
		},
		{
			name: "test-ControlSequence",
			fields: fields{
				state: ansiControlSequence,
			},
			args: args{
				st: &parseState{
					str: "m",
				},
			},
			want:      true,
			wantState: ansiText,
		},
		{
			name: "test-ControlSequence2",
			fields: fields{
				state: ansiControlSequence,
			},
			args: args{
				st: &parseState{
					str: "A",
				},
			},
			want:      true,
			wantState: ansiText,
		},
		{
			name: "test-ControlSequenceEnd",
			fields: fields{
				state: ansiControlSequence,
			},
			args: args{
				st: &parseState{
					str: "?",
				},
			},
			want:      true,
			wantState: ansiControlSequence,
		},
		{
			name: "test-ControlSequenceOver",
			fields: fields{
				state: ansiControlSequence,
			},
			args: args{
				st: &parseState{
					str: "@",
				},
			},
			want:      true,
			wantState: ansiText,
		},
		{
			name: "test-SysSequence",
			fields: fields{
				state: systemSequence,
			},
			args: args{
				st: &parseState{
					str: string(rune(0x07)),
				},
			},
			want:      true,
			wantState: ansiText,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := newESConverter()
			es.state = tt.fields.state
			if got := es.convert(tt.args.st); got != tt.want {
				t.Errorf("escapeSequence.convert() = %v, want %v", got, tt.want)
			}
			if es.state != tt.wantState {
				t.Errorf("escapeSequence.convert() = %v, want %v", es.state, tt.wantState)
			}
		})
	}
}

func Test_sgrStyle(t *testing.T) {
	t.Parallel()
	type args struct {
		style        tcell.Style
		csiParameter string
	}
	tests := []struct {
		name    string
		args    args
		want    tcell.Style
		wantErr bool
	}{
		{
			name: "color8bit",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "38;5;1",
			},
			want:    tcell.StyleDefault.Foreground(tcell.ColorMaroon),
			wantErr: false,
		},
		{
			name: "color8bit2",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "38;5;21",
			},
			want:    tcell.StyleDefault.Foreground(tcell.GetColor("#0000ff")),
			wantErr: false,
		},
		{
			name: "colorTrueColor",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "38;2;255;0;0",
			},
			want:    tcell.StyleDefault.Foreground(tcell.GetColor("#FF0000")),
			wantErr: false,
		},
		{
			name: "colorTrueColorInvalid",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "38;2;255;",
			},
			want: tcell.StyleDefault,
		},
		{
			name: "colorTrueColorInvalid2",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "38;2;a;b;c",
			},
			want:    tcell.StyleDefault,
			wantErr: false,
		},
		{
			name: "attributes",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "2;3;4;5;6;7;8;9",
			},
			want:    tcell.StyleDefault.Dim(true).Italic(true).Underline(true).Blink(true).Reverse(true).StrikeThrough(true),
			wantErr: false,
		},
		{
			name: "UnderLine double",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "4:2",
			},
			want:    tcell.StyleDefault.Underline(true).Underline(underLineStyle("2")),
			wantErr: false,
		},
		{
			name: "UnderLine double2",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "21",
			},
			want:    tcell.StyleDefault.Underline(true).Underline(underLineStyle("2")),
			wantErr: false,
		},
		{
			name: "UnderLine color",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: "4;58;2;255;0;0",
			},
			want:    tcell.StyleDefault.Underline(true).Underline(tcell.GetColor("#FF0000")),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := sgrStyle(tt.args.style, tt.args.csiParameter)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("csToStyle() = %v, want %v", got, tt.want)
				gfg, gbg, gattr := got.Decompose()
				wfg, wbg, wattr := tt.want.Decompose()
				t.Errorf("csToStyle() = %x,%x,%v, want %x,%x,%v", gfg.Hex(), gbg.Hex(), gattr, wfg.Hex(), wbg.Hex(), wattr)
			}
		})
	}
}

func Test_parseSGR(t *testing.T) {
	type args struct {
		params string
	}
	tests := []struct {
		name    string
		args    args
		want    OVStyle
		wantErr bool
	}{
		{
			name: "test-attributes",
			args: args{
				params: "2;3;4;5;6;7;8;9",
			},
			want: OVStyle{
				Dim:           true,
				Italic:        true,
				Underline:     true,
				Blink:         true,
				Reverse:       true,
				StrikeThrough: true,
			},
			wantErr: false,
		},
		{
			name: "test-attributesErr",
			args: args{
				params: "38;38;38",
			},
			want: OVStyle{
				Dim:           false,
				Italic:        false,
				Underline:     false,
				Blink:         false,
				Reverse:       false,
				StrikeThrough: false,
			},
			wantErr: false,
		},
		{
			name: "test-attributesNone",
			args: args{
				params: "28",
			},
			want: OVStyle{
				Dim:           false,
				Italic:        false,
				Underline:     false,
				Blink:         false,
				Reverse:       false,
				StrikeThrough: false,
			},
			wantErr: false,
		},
		{
			name: "test-Default",
			args: args{
				params: "49",
			},
			want: OVStyle{
				Background:    "default",
				Dim:           false,
				Italic:        false,
				Underline:     false,
				Blink:         false,
				Reverse:       false,
				StrikeThrough: false,
			},
			wantErr: false,
		},
		{
			name: "test-foreground2",
			args: args{
				params: "038;05;02",
			},
			want: OVStyle{
				Foreground: "green",
			},
		},
		{
			name: "test-foreground216",
			args: args{
				params: "38;5;216",
			},
			want: OVStyle{
				Foreground: "#FFAF87",
			},
			wantErr: false,
		},
		{
			name: "test-foreground216_Underline",
			args: args{
				params: "38;5;216;4",
			},
			want: OVStyle{
				Foreground: "#FFAF87",
				Underline:  true,
			},
		},
		{
			name: "test-reset_Underline",
			args: args{
				params: "38;5;216;0;4",
			},
			want: OVStyle{
				Reset:     true,
				Underline: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSGR(tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSGI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseSGR2(t *testing.T) {
	type args struct {
		params string
	}
	tests := []struct {
		name string
		args args
		want OVStyle
	}{
		{
			name: "test-Colon1",
			args: args{
				params: "38:5:1",
			},
			want: OVStyle{
				Foreground: "maroon",
			},
		},
		{
			name: "test-Colon2",
			args: args{
				params: "48:2:255:0:0",
			},
			want: OVStyle{
				Background: "#ff0000",
			},
		},
		{
			name: "test-Colon3",
			args: args{
				params: "48:2::255:0:0",
			},
			want: OVStyle{
				Background: "#ff0000",
			},
		},
		{
			name: "test-Underline-colon",
			args: args{
				params: "4:0",
			},
			want: OVStyle{
				Underline:      true,
				UnderlineStyle: "0",
			},
		},
		{
			name: "test-invalid1",
			args: args{
				params: "38:5:-",
			},
			want: OVStyle{},
		},
		{
			name: "test-invalid2",
			args: args{
				params: "38:5:999",
			},
			want: OVStyle{},
		},
		{
			name: "test-invalid3",
			args: args{
				params: "38:5",
			},
			want: OVStyle{},
		},
		{
			name: "test-valid",
			args: args{
				params: "38:5:0",
			},
			want: OVStyle{
				Foreground: "black",
			},
		},
		{
			name: "test-rgb-valid",
			args: args{
				params: "4;38:2:255:0:0",
			},
			want: OVStyle{
				Underline:  true,
				Foreground: "#ff0000",
			},
		},
		{
			name: "test-rgb-invalid",
			args: args{
				params: "4;38:2:255:0:-",
			},
			want: OVStyle{},
		},
		{
			name: "test-rgb-invalid2",
			args: args{
				params: "4;48:2:255:0:-",
			},
			want: OVStyle{},
		},
		{
			name: "test-rgb-over",
			args: args{
				params: "4;38:2:255:0:999",
			},
			want: OVStyle{
				Underline: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSGR(tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSGR() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_parseSGR3(t *testing.T) {
	type args struct {
		params string
	}
	tests := []struct {
		name string
		args args
		want OVStyle
	}{
		{
			name: "test-not-supported",
			args: args{
				params: "99;a:4",
			},
			want: OVStyle{},
		},
		{
			name: "test-overline-on",
			args: args{
				params: "53",
			},
			want: OVStyle{
				OverLine: true,
			},
		},
		{
			name: "test-overline-off",
			args: args{
				params: "55",
			},
			want: OVStyle{
				OverLine:   false,
				UnOverLine: true,
			},
		},
		{
			name: "test-underline-color-error",
			args: args{
				params: "58;a",
			},
			want: OVStyle{},
		},
		{
			name: "test-underline-color-default",
			args: args{
				params: "59",
			},
			want: OVStyle{
				UnderlineColor: "default",
			},
		},
		{
			name: "test-vertical-line-on",
			args: args{
				params: "74",
			},
			want: OVStyle{
				VerticalAlignType: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSGR(tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSGR() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_colorName(t *testing.T) {
	type args struct {
		colorNumber int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test-ColorName1",
			args: args{
				colorNumber: 1,
			},
			want: "maroon",
		},
		{
			name: "test-ColorName249",
			args: args{
				colorNumber: 249,
			},
			want: "#B2B2B2",
		},
		{
			name: "test-ColorNameNotImplemented",
			args: args{
				colorNumber: 999,
			},
			want: "",
		},
		{
			name: "test-ColorNameMinus",
			args: args{
				colorNumber: -1,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := colorName(tt.args.colorNumber); got != tt.want {
				t.Errorf("colorName() = %v, want %v", got, tt.want)
			}
		})
	}
}
