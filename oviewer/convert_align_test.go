package oviewer

import (
	"regexp"
	"testing"
)

func Test_align_convert(t *testing.T) {
	type fields struct {
		es           *escapeSequence
		maxWidths    []int
		orgWidths    []int
		WidthF       bool
		delimiter    string
		delimiterReg *regexp.Regexp
		count        int
	}
	type args struct {
		st *parseState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantStr string
	}{
		{
			name: "convertAlignDelm",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{1, 2},
				WidthF:    false,
				delimiter: ",",
				count:     0,
			},
			args: args{
				st: &parseState{
					lc:    StrToContents("a,b,c\n", 8),
					mainc: '\n',
				},
			},
			want:    false,
			wantStr: "a ,b ,c\n",
		},
		{
			name: "convertAlignDelmTab",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{1, 2},
				WidthF:    false,
				delimiter: "\t",
				count:     0,
			},
			args: args{
				st: &parseState{
					lc:    StrToContents("", 8),
					mainc: '\n',
				},
			},
			want:    false,
			wantStr: "",
		},
		{
			name: "convertAlignDelmES",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{1, 2},
				WidthF:    false,
				delimiter: "\t",
				count:     0,
			},
			args: args{
				st: &parseState{
					lc:    StrToContents("", 8),
					mainc: '\x1b',
				},
			},
			want:    true,
			wantStr: "",
		},
		{
			name: "convertAlignNoDelm",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{},
				WidthF:    false,
				delimiter: "\t",
				count:     0,
			},
			args: args{
				st: &parseState{
					lc:    StrToContents("", 8),
					mainc: 'a',
				},
			},
			want:    false,
			wantStr: "",
		},
		{
			name: "convertAlignDelm2",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{1, 2},
				WidthF:    false,
				delimiter: ",",
				count:     0,
			},
			args: args{
				st: &parseState{
					lc:    StrToContents("a,b,", 8),
					mainc: 'あ',
				},
			},
			want:    false,
			wantStr: "a,b,",
		},
		{
			name: "convertAlignWidth",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{3, 3},
				orgWidths: []int{2, 5},
				WidthF:    true,
				count:     0,
			},
			args: args{
				st: &parseState{
					lc:    StrToContents("a  b  c\n", 8),
					mainc: '\n',
				},
			},
			want:    false,
			wantStr: "a   b   c\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &align{
				es:           tt.fields.es,
				maxWidths:    tt.fields.maxWidths,
				orgWidths:    tt.fields.orgWidths,
				WidthF:       tt.fields.WidthF,
				delimiter:    tt.fields.delimiter,
				delimiterReg: tt.fields.delimiterReg,
				count:        tt.fields.count,
			}
			if got := a.convert(tt.args.st); got != tt.want {
				t.Errorf("align.convert() = %v, want %v", got, tt.want)
			}
			goStr, _ := ContentsToStr(tt.args.st.lc)
			if goStr != tt.wantStr {
				t.Errorf("align.convert() = %v, want %v", goStr, tt.wantStr)
			}
		})
	}
}

func Test_align_convertDelm(t *testing.T) {
	type fields struct {
		es           *escapeSequence
		maxWidths    []int
		orgWidths    []int
		shrink       []bool
		rightAlign   []bool
		WidthF       bool
		delimiter    string
		delimiterReg *regexp.Regexp
		count        int
	}
	type args struct {
		src contents
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "convertAlignDelm1",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{1, 2},
				WidthF:    false,
				delimiter: ",",
				count:     0,
			},
			args: args{
				src: StrToContents("a,b,c,d,e,f", 8),
			},
			want: "a ,b ,c,d,e,f",
		},
		{
			name: "convertAlignDelm2",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{1, 2, 2, 2, 2, 2},
				WidthF:    false,
				delimiter: ",",
				count:     0,
			},
			args: args{
				src: StrToContents("a,b,c,d,e,f", 8),
			},
			want: "a ,b ,c ,d ,e ,f",
		},
		{
			name: "convertAlignDelmShrink1",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{1, 2, 2, 2, 2, 2},
				shrink:    []bool{false, true, false, false, false},
				WidthF:    false,
				delimiter: ",",
				count:     0,
			},
			args: args{
				src: StrToContents("a,b,c,d,e,f", 8),
			},
			want: "a ,…,c ,d ,e ,f",
		},
		{
			name: "convertAlignDelmShrink2",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{1, 2, 2, 2, 2, 2},
				shrink:    []bool{true, false, false, false, false},
				WidthF:    false,
				delimiter: ",",
				count:     0,
			},
			args: args{
				src: StrToContents("a,b,c,d,e,f", 8),
			},
			want: "…,b ,c ,d ,e ,f",
		},
		{
			name: "convertAlignDelmShrink3",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{1, 2, 2, 2, 2, 2},
				shrink:    []bool{false, false, false, false, false, true},
				WidthF:    false,
				delimiter: ",",
				count:     0,
			},
			args: args{
				src: StrToContents("a,b,c,d,e,f", 8),
			},
			want: "a ,b ,c ,d ,e ,…",
		},
		{
			name: "convertAlignDelmRight",
			fields: fields{
				es:         newESConverter(),
				maxWidths:  []int{1, 2, 2, 2, 2, 2},
				shrink:     []bool{false, false, false, false, false, true},
				rightAlign: []bool{true, false, false, false, false, false},
				WidthF:     false,
				delimiter:  ",",
			},
			args: args{
				src: StrToContents("a,b,c,d,e,f", 8),
			},
			want: " a,b ,c ,d ,e ,…",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &align{
				es:           tt.fields.es,
				maxWidths:    tt.fields.maxWidths,
				orgWidths:    tt.fields.orgWidths,
				shrink:       tt.fields.shrink,
				rightAlign:   tt.fields.rightAlign,
				WidthF:       tt.fields.WidthF,
				delimiter:    tt.fields.delimiter,
				delimiterReg: tt.fields.delimiterReg,
				count:        tt.fields.count,
			}
			got := a.convertDelm(tt.args.src)
			gotStr, _ := ContentsToStr(got)
			if gotStr != tt.want {
				t.Errorf("align.convertDelm() = %v, want %v", gotStr, tt.want)
			}
		})
	}
}

func Test_align_convertWidth(t *testing.T) {
	type fields struct {
		es           *escapeSequence
		maxWidths    []int
		orgWidths    []int
		shrink       []bool
		rightAlign   []bool
		WidthF       bool
		delimiter    string
		delimiterReg *regexp.Regexp
		count        int
	}
	type args struct {
		src contents
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "convertAlignWidth1",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{3, 3},
				orgWidths: []int{2, 5, 8, 11, 14},
				WidthF:    true,
				count:     0,
			},
			args: args{
				src: StrToContents("a  b  c  d  e  f", 8),
			},
			want: "a   b   c d e f",
		},
		{
			name: "convertAlignWidth2",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{3, 3, 3, 3, 3},
				orgWidths: []int{2, 5, 8, 11, 14},
				WidthF:    true,
				count:     0,
			},
			args: args{
				src: StrToContents("a  b  c  d  e  f", 8),
			},
			want: "a   b   c   d   e   f",
		},
		{
			name: "convertAlignWidthShrink1",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{3, 3, 3, 3, 3},
				orgWidths: []int{2, 5, 8, 11, 14},
				shrink:    []bool{false, true, false, false, false},
				WidthF:    true,
				count:     0,
			},
			args: args{
				src: StrToContents("a  b  c  d  e  f", 8),
			},
			want: "a   … c   d   e   f",
		},
		{
			name: "convertAlignWidthShrink2",
			fields: fields{
				es:        newESConverter(),
				maxWidths: []int{3, 3, 3, 3, 3},
				orgWidths: []int{2, 5, 8, 11, 14},
				shrink:    []bool{false, false, false, false, false, true},
				WidthF:    true,
				count:     0,
			},
			args: args{
				src: StrToContents("a  b  c  d  e  f", 8),
			},
			want: "a   b   c   d   e   …",
		},
		{
			name: "convertAlignWidthRight",
			fields: fields{
				es:         newESConverter(),
				maxWidths:  []int{3, 3, 3, 3, 3, 3},
				orgWidths:  []int{2, 5, 8, 11, 14, 17},
				shrink:     []bool{false, false, false, false, false, true},
				rightAlign: []bool{true, false, false, false, false, false},
				WidthF:     true,
				count:      0,
			},
			args: args{
				src: StrToContents("a  b  c  d  e  f", 8),
			},
			want: "  a b   c   d   e   … ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &align{
				es:           tt.fields.es,
				maxWidths:    tt.fields.maxWidths,
				orgWidths:    tt.fields.orgWidths,
				shrink:       tt.fields.shrink,
				rightAlign:   tt.fields.rightAlign,
				WidthF:       tt.fields.WidthF,
				delimiter:    tt.fields.delimiter,
				delimiterReg: tt.fields.delimiterReg,
				count:        tt.fields.count,
			}
			got := a.convertWidth(tt.args.src)
			gotStr, _ := ContentsToStr(got)
			if gotStr != tt.want {
				t.Errorf("align.convertWidth() = %v, want %v", gotStr, tt.want)
			}
		})
	}
}
