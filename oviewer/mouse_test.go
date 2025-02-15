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
		startX:  0,
		lines:   make(map[int]LineC, 0),
	}
	for i := 0; i < 10; i++ {
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
				line: NewLineC(t, "test", 8),
				x1:   0,
				x2:   -1,
			},
			want: "test",
		},
		{
			name:   "testParts",
			fields: fields{startX: 0},
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
			name:   "testあいうえお",
			fields: fields{startX: 0},
			args: args{
				line: NewLineC(t, "あいうえお", 8),
				x1:   0,
				x2:   4,
			},
			want: "あい",
		},
		{
			name:   "test2あいうえお",
			fields: fields{startX: 0},
			args: args{
				line: NewLineC(t, "あいうえお", 8),
				x1:   0,
				x2:   9,
			},
			want: "あいうえお",
		},
		{
			name:   "test3あいうえお",
			fields: fields{startX: 0},
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
			scr := SCR{
				startX: tt.fields.startX,
			}
			if got := scr.selectLine(tt.args.line, tt.args.x1, tt.args.x2); got != tt.want {
				t.Errorf("SCR.selectLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
