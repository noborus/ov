package oviewer

import (
	"testing"
)

func normalDocument() *Document {
	m, _ := OpenDocument("../testdata/normal.txt")
	for !m.BufEOF() {
	}
	return m
}
func TestSCR_lineRangeToString(t *testing.T) {
	type fields struct {
		numbers []LineNumber
		vWidth  int
		vHeight int
		startX  int
	}
	type args struct {
		m  *Document
		x1 int
		y1 int
		x2 int
		y2 int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				numbers: []LineNumber{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {4, 1}, {5, 0}, {6, 0}, {7, 0}, {8, 0}},
				vWidth:  80,
				vHeight: 24,
				startX:  0,
			},
			args:    args{m: normalDocument(), x1: 0, y1: 0, x2: 10, y2: 0},
			want:    "khaki	med",
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				numbers: []LineNumber{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {4, 1}, {5, 0}, {6, 0}, {7, 0}, {8, 0}},
				vWidth:  80,
				vHeight: 24,
				startX:  0,
			},
			args:    args{m: normalDocument(), x1: 0, y1: 3, x2: 10, y2: 4},
			want:    "darkolivegreen	darkkhaki	orchid	olive	darkslateblue	mediumseagreen\nchocolate	",
			wantErr: false,
		},
		{
			name: "test3",
			fields: fields{
				numbers: []LineNumber{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {4, 1}, {5, 0}, {6, 0}, {7, 0}, {8, 0}},
				vWidth:  80,
				vHeight: 24,
				startX:  0,
			},
			args:    args{m: normalDocument(), x1: 4, y1: 1, x2: 10, y2: 4},
			want:    "brown	khaki	darkkhaki	mediumturquoise	mediumseagreen	darkgoldenrod\nplum	darkslateblue	teal	darkkhaki	turquoise	mediumorchid\ndarkolivegreen	darkkhaki	orchid	olive	darkslateblue	mediumseagreen\nchocolate	",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scr := SCR{
				numbers: tt.fields.numbers,
				vWidth:  tt.fields.vWidth,
				vHeight: tt.fields.vHeight,
				startX:  tt.fields.startX,
			}
			got, err := scr.lineRangeToString(tt.args.m, tt.args.x1, tt.args.y1, tt.args.x2, tt.args.y2)
			if (err != nil) != tt.wantErr {
				t.Errorf("SCR.lineRangeToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SCR.lineRangeToString() = [%v], want [%v]", got, tt.want)
			}
		})
	}
}

func TestSCR_rectangleToString(t *testing.T) {
	type fields struct {
		numbers []LineNumber
		vWidth  int
		vHeight int
		startX  int
	}
	type args struct {
		m  *Document
		x1 int
		y1 int
		x2 int
		y2 int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				numbers: []LineNumber{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {4, 1}, {5, 0}, {6, 0}, {7, 0}, {8, 0}},
				vWidth:  80,
				vHeight: 24,
				startX:  0,
			},
			args:    args{m: normalDocument(), x1: 0, y1: 0, x2: 10, y2: 0},
			want:    "khaki	med\n",
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				numbers: []LineNumber{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {4, 1}, {5, 0}, {6, 0}, {7, 0}, {8, 0}},
				vWidth:  80,
				vHeight: 24,
				startX:  0,
			},
			args:    args{m: normalDocument(), x1: 0, y1: 3, x2: 10, y2: 4},
			want:    "darkolivegr\nchocolate	\n",
			wantErr: false,
		},
		{
			name: "test3",
			fields: fields{
				numbers: []LineNumber{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {4, 1}, {5, 0}, {6, 0}, {7, 0}, {8, 0}},
				vWidth:  80,
				vHeight: 24,
				startX:  0,
			},
			args:    args{m: normalDocument(), x1: 4, y1: 1, x2: 10, y2: 4},
			want:    "brown	\n	dar\nolivegr\nolate	\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scr := SCR{
				numbers: tt.fields.numbers,
				vWidth:  tt.fields.vWidth,
				vHeight: tt.fields.vHeight,
				startX:  tt.fields.startX,
			}
			got, err := scr.rectangleToString(tt.args.m, tt.args.x1, tt.args.y1, tt.args.x2, tt.args.y2)
			if (err != nil) != tt.wantErr {
				t.Errorf("SCR.rectangleToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SCR.rectangleToString() = [%v], want [%v]", got, tt.want)
			}
		})
	}
}

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
		{
			// 2  4  6  8  10
			// あ い う え お
			name:   "testあいうえお",
			fields: fields{startX: 0},
			args: args{
				line: NewLineC("あいうえお", 8),
				x1:   0,
				x2:   4,
			},
			want: "あい",
		},
		{
			name:   "test2あいうえお",
			fields: fields{startX: 0},
			args: args{
				line: NewLineC("あいうえお", 8),
				x1:   0,
				x2:   9,
			},
			want: "あいうえお",
		},
		{
			name:   "test3あいうえお",
			fields: fields{startX: 0},
			args: args{
				line: NewLineC("あいうえお", 8),
				x1:   1,
				x2:   3,
			},
			want: "い",
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
