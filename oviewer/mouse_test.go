package oviewer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// setupSCRHelper is a helper function to setup SCR for testing.
func setupSCRHelper(t *testing.T, m *Document) SCR {
	t.Helper()
	scr := SCR{
		numbers: []LineNumber{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {4, 1}, {5, 0}, {6, 0}, {7, 0}, {8, 0}},
		vWidth:  80,
		vHeight: 24,
		lines:   make(map[int]LineC, 0),
	}
	for i := range 10 {
		line := m.getLineC(i)
		scr.lines[i] = line
	}
	return scr
}

func NewLineC(t *testing.T, str string, tabWidth int) LineC {
	t.Helper()
	lineC := LineC{}
	lc := StrToContents(str, tabWidth)
	lineC.lc = lc
	lineC.str, lineC.pos = ContentsToStr(lc)
	return lineC
}

func TestRoot_mouseEvent(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		ev *tcell.EventMouse
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "testWheelUp",
			fields: fields{
				ev: tcell.NewEventMouse(0, 0, tcell.WheelUp, tcell.ModNone),
			},
			want: false,
		},
		{
			name: "testWheelDown",
			fields: fields{
				ev: tcell.NewEventMouse(0, 0, tcell.WheelDown, tcell.ModNone),
			},
			want: false,
		},
		{
			name: "testWheelLeft",
			fields: fields{
				ev: tcell.NewEventMouse(0, 0, tcell.WheelLeft, tcell.ModNone),
			},
			want: false,
		},
		{
			name: "testWheelRight",
			fields: fields{
				ev: tcell.NewEventMouse(0, 0, tcell.WheelRight, tcell.ModNone),
			},
			want: false,
		},
		{
			name: "testShiftWheelUp",
			fields: fields{
				ev: tcell.NewEventMouse(0, 0, tcell.WheelUp, tcell.ModShift),
			},
			want: false,
		},
		{
			name: "testShiftWheelDown",
			fields: fields{
				ev: tcell.NewEventMouse(0, 0, tcell.WheelDown, tcell.ModShift),
			},
			want: false,
		},
		{
			name: "testButtonNone",
			fields: fields{
				ev: tcell.NewEventMouse(0, 0, tcell.ButtonNone, tcell.ModNone),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, filepath.Join(testdata, "normal.txt"))
			root.prepareScreen()
			ctx := context.Background()
			root.prepareDraw(ctx)
			root.draw(ctx)
			root.mouseEvent(context.Background(), tt.fields.ev)
			if got := root.skipDraw; got != tt.want {
				t.Errorf("Root.mouseEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_rangeToString(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		rectangle bool
	}
	type args struct {
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
			name:    "test1",
			fields:  fields{rectangle: false},
			args:    args{x1: 0, y1: 0, x2: 10, y2: 0},
			want:    "khaki	med",
			wantErr: false,
		},
		{
			name:    "test1-1",
			fields:  fields{rectangle: false},
			args:    args{x1: 10, y1: 0, x2: 0, y2: 0},
			want:    "khaki	med",
			wantErr: false,
		},
		{
			name:    "test2-1",
			fields:  fields{rectangle: false},
			args:    args{x1: 0, y1: 0, x2: 10, y2: 1},
			want:    "khaki	mediumseagreen	steelblue	forestgreen	royalblue	mediumseagreen\nrosybrown	",
			wantErr: false,
		},
		{
			name:    "test2-2",
			fields:  fields{rectangle: false},
			args:    args{x1: 10, y1: 1, x2: 0, y2: 0},
			want:    "khaki	mediumseagreen	steelblue	forestgreen	royalblue	mediumseagreen\nrosybrown	",
			wantErr: false,
		},
		{
			name:    "test-rectangle1",
			fields:  fields{rectangle: true},
			args:    args{x1: 0, y1: 0, x2: 10, y2: 0},
			want:    "khaki	med\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, filepath.Join(testdata, "normal.txt"))
			root.prepareScreen()
			ctx := context.Background()
			root.prepareDraw(ctx)
			root.scr.mouseRectangle = tt.fields.rectangle
			root.draw(ctx)
			got, err := root.rangeToString(tt.args.x1, tt.args.y1, tt.args.x2, tt.args.y2)
			if (err != nil) != tt.wantErr {
				t.Errorf("Root.rangeToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Root.rangeToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSCR_lineRangeToString(t *testing.T) {
	t.Parallel()
	type args struct {
		x1 int
		y1 int
		x2 int
		y2 int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "test1",
			args:    args{x1: 0, y1: 0, x2: 10, y2: 0},
			want:    "khaki	med",
			wantErr: false,
		},
		{
			name:    "test2",
			args:    args{x1: 0, y1: 3, x2: 10, y2: 4},
			want:    "darkolivegreen	darkkhaki	orchid	olive	darkslateblue	mediumseagreen\nchocolate	",
			wantErr: false,
		},
		{
			name:    "test3",
			args:    args{x1: 4, y1: 1, x2: 10, y2: 4},
			want:    "brown	khaki	darkkhaki	mediumturquoise	mediumseagreen	darkgoldenrod\nplum	darkslateblue	teal	darkkhaki	turquoise	mediumorchid\ndarkolivegreen	darkkhaki	orchid	olive	darkslateblue	mediumseagreen\nchocolate	",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := docFileReadHelper(t, filepath.Join(testdata, "normal.txt"))
			scr := setupSCRHelper(t, m)
			got, err := scr.lineRangeToString(m, tt.args.x1, tt.args.y1, tt.args.x2, tt.args.y2)
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
	t.Parallel()
	type args struct {
		x1 int
		y1 int
		x2 int
		y2 int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "test1",
			args:    args{x1: 0, y1: 0, x2: 10, y2: 0},
			want:    "khaki	med\n",
			wantErr: false,
		},
		{
			name:    "test2",
			args:    args{x1: 0, y1: 3, x2: 10, y2: 4},
			want:    "darkolivegr\nchocolate	\n",
			wantErr: false,
		},
		{
			name:    "test3",
			args:    args{x1: 4, y1: 1, x2: 10, y2: 4},
			want:    "brown	\n	dar\nolivegr\nolate	\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := docFileReadHelper(t, filepath.Join(testdata, "normal.txt"))
			scr := setupSCRHelper(t, m)
			got, err := scr.rectangleToString(m, tt.args.x1, tt.args.y1, tt.args.x2, tt.args.y2)
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

func TestSCR_selectLine(t *testing.T) {
	t.Parallel()
	type args struct {
		line LineC
		x1   int
		x2   int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				line: NewLineC(t, "test", 8),
				x1:   0,
				x2:   -1,
			},
			want: "test",
		},
		{
			name: "testParts",
			args: args{
				line: NewLineC(t, "test", 8),
				x1:   1,
				x2:   2,
			},
			want: "e",
		},
		{
			// 2  4  6  8  10
			// あ い う え お
			name: "testあいうえお",
			args: args{
				line: NewLineC(t, "あいうえお", 8),
				x1:   0,
				x2:   4,
			},
			want: "あい",
		},
		{
			name: "test2あいうえお",
			args: args{
				line: NewLineC(t, "あいうえお", 8),
				x1:   0,
				x2:   9,
			},
			want: "あいうえお",
		},
		{
			name: "test3あいうえお",
			args: args{
				line: NewLineC(t, "あいうえお", 8),
				x1:   1,
				x2:   3,
			},
			want: "い",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := selectLine(tt.args.line, 0, tt.args.x1, tt.args.x2); got != tt.want {
				t.Errorf("SCR.selectLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCharTypeAt(t *testing.T) {
	type args struct {
		line     LineC
		contentX int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Whitespace-space",
			args: args{
				line: func() LineC {
					lc := StrToContents(" ", 8)
					lineC := LineC{lc: lc}
					return lineC
				}(),
				contentX: 0,
			},
			want: charTypeWhitespace,
		},
		{
			name: "Whitespace-tab",
			args: args{
				line: func() LineC {
					lc := StrToContents("\t", 8)
					lineC := LineC{lc: lc}
					return lineC
				}(),
				contentX: 0,
			},
			want: charTypeWhitespace,
		},
		{
			name: "Alphanumeric-letter",
			args: args{
				line: func() LineC {
					lc := StrToContents("a", 8)
					lineC := LineC{lc: lc}
					return lineC
				}(),
				contentX: 0,
			},
			want: charTypeAlphanumeric,
		},
		{
			name: "Alphanumeric-digit",
			args: args{
				line: func() LineC {
					lc := StrToContents("5", 8)
					lineC := LineC{lc: lc}
					return lineC
				}(),
				contentX: 0,
			},
			want: charTypeAlphanumeric,
		},
		{
			name: "Alphanumeric-underscore",
			args: args{
				line: func() LineC {
					lc := StrToContents("_", 8)
					lineC := LineC{lc: lc}
					return lineC
				}(),
				contentX: 0,
			},
			want: charTypeAlphanumeric,
		},
		{
			name: "Other-symbol",
			args: args{
				line: func() LineC {
					lc := StrToContents("!", 8)
					lineC := LineC{lc: lc}
					return lineC
				}(),
				contentX: 0,
			},
			want: charTypeOther,
		},
		{
			name: "Other-non-ascii",
			args: args{
				line: func() LineC {
					lc := StrToContents("あ", 8)
					lineC := LineC{lc: lc}
					return lineC
				}(),
				contentX: 0,
			},
			want: charTypeAlphanumeric,
		},
		{
			name: "Other-non-ascii2",
			args: args{
				line: func() LineC {
					lc := StrToContents("あ", 8)
					lineC := LineC{lc: lc}
					return lineC
				}(),
				contentX: 1,
			},
			want: charTypeAlphanumeric,
		},
		{
			name: "OutOfBounds-negative",
			args: args{
				line: func() LineC {
					lc := StrToContents("abc", 8)
					lineC := LineC{lc: lc}
					return lineC
				}(),
				contentX: -1,
			},
			want: charTypeWhitespace,
		},
		{
			name: "OutOfBounds-tooLarge",
			args: args{
				line: func() LineC {
					lc := StrToContents("abc", 8)
					lineC := LineC{lc: lc}
					return lineC
				}(),
				contentX: 3,
			},
			want: charTypeWhitespace,
		},
		{
			name: "EmptyLine",
			args: args{
				line:     LineC{lc: []content{}},
				contentX: 0,
			},
			want: charTypeWhitespace,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCharTypeAt(tt.args.line, tt.args.contentX); got != tt.want {
				t.Errorf("getCharTypeAt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_branchWidth(t *testing.T) {
	type args struct {
		lc     contents
		branch int
		width  int
		offset int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "No wrap, single branch",
			args: args{
				lc: contents{
					{str: "a", width: 1},
					{str: "b", width: 1},
					{str: "c", width: 1},
				},
				branch: 0,
				width:  10,
				offset: 0,
			},
			want: 0,
		},
		{
			name: "Wrap after width, branch 1",
			args: args{
				lc: contents{
					{str: "a", width: 1},
					{str: "b", width: 1},
					{str: "c", width: 1},
				},
				branch: 1,
				width:  10,
				offset: 0,
			},
			want: 3,
		},
		{
			name: "Offset applied, branch 1",
			args: args{
				lc: contents{
					{str: "a", width: 1},
					{str: "b", width: 1},
					{str: "c", width: 1},
				},
				branch: 1,
				width:  10,
				offset: 1,
			},
			want: 3,
		},
		{
			name: "Wide characters, wrap at width",
			args: args{
				lc: contents{
					{str: "あ", width: 2},
					{str: "い", width: 2},
					{str: "う", width: 2},
					{str: "え", width: 2},
				},
				branch: 1,
				width:  10,
				offset: 0,
			},
			want: 8,
		},
		{
			name: "Empty contents",
			args: args{
				lc:     contents{},
				branch: 0,
				width:  10,
				offset: 0,
			},
			want: 0,
		},
		{
			name: "Branch greater than possible wraps",
			args: args{
				lc: contents{
					{str: "a", width: 2},
					{str: "b", width: 2},
				},
				branch: 5,
				width:  10,
				offset: 0,
			},
			want: 4,
		},
		{
			name: "Offset nonzero, no wrap",
			args: args{
				lc: contents{
					{str: "a", width: 1},
					{str: "b", width: 1},
				},
				branch: 0,
				width:  10,
				offset: 2,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := branchWidth(tt.args.lc, tt.args.branch, tt.args.width, tt.args.offset); got != tt.want {
				t.Errorf("branchWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}
