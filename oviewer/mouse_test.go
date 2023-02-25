package oviewer

import "testing"

func NewLineC(str string, tabWidth int) LineC {
	line := LineC{}
	lc := StrToContents(str, tabWidth)
	line.lc = lc
	line.str, line.pos = ContentsToStr(lc)
	return line
}

func TestSCR_selectLine(t *testing.T) {
	type fields struct {
		startX int
	}
	type args struct {
		line LineC
		x1   int
		x2   int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "test1",
			fields: fields{startX: 0},
			args: args{
				line: NewLineC("test", 8),
				x1:   0,
				x2:   -1,
			},
			want: "test",
		},
		{
			name:   "testParts",
			fields: fields{startX: 0},
			args: args{
				line: NewLineC("test", 8),
				x1:   1,
				x2:   2,
			},
			want: "e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scr := SCR{
				startX: tt.fields.startX,
			}
			if got := scr.selectLine(tt.args.line, tt.args.x1, tt.args.x2); got != tt.want {
				t.Errorf("SCR.selectLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
