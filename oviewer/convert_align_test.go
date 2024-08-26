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
