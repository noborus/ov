package oviewer

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func rootHelper(t *testing.T) *Root {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root, err := NewRoot(bytes.NewBufferString("test"))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestRoot_toggle(t *testing.T) {
	root := rootHelper(t)
	ctx := context.Background()
	var v bool
	v = root.Doc.ColumnMode
	root.toggleColumnMode(ctx)
	if v == root.Doc.ColumnMode {
		t.Errorf("toggleColumnMode() = %v, want %v", root.Doc.ColumnMode, !v)
	}
	v = root.Doc.WrapMode
	root.toggleWrapMode(ctx)
	if v == root.Doc.WrapMode {
		t.Errorf("toggleWrapMode() = %v, want %v", root.Doc.WrapMode, !v)
	}
	v = root.Doc.LineNumMode
	root.toggleLineNumMode(ctx)
	if v == root.Doc.LineNumMode {
		t.Errorf("toggleLineNumberMode() = %v, want %v", root.Doc.LineNumMode, !v)
	}
	v = root.Doc.ColumnWidth
	root.toggleColumnWidth(ctx)
	if v == root.Doc.ColumnWidth {
		t.Errorf("toggleColumnWidth() = %v, want %v", root.Doc.ColumnWidth, !v)
	}
	v = root.Doc.AlternateRows
	root.toggleAlternateRows(ctx)
	if v == root.Doc.AlternateRows {
		t.Errorf("toggleAlternateRows() = %v, want %v", root.Doc.AlternateRows, !v)
	}
	v = root.Doc.PlainMode
	root.togglePlain(ctx)
	if v == root.Doc.PlainMode {
		t.Errorf("togglePlainMode() = %v, want %v", root.Doc.PlainMode, !v)
	}
	v = root.Doc.ColumnRainbow
	root.toggleRainbow(ctx)
	if v == root.Doc.ColumnRainbow {
		t.Errorf("toggleRainbow() = %v, want %v", root.Doc.ColumnRainbow, !v)
	}
	v = root.Doc.FollowMode
	root.toggleFollowMode(ctx)
	if v == root.Doc.FollowMode {
		t.Errorf("toggleFollow() = %v, want %v", root.Doc.FollowMode, !v)
	}
	v = root.General.FollowAll
	root.toggleFollowAll(ctx)
	if v == root.General.FollowAll {
		t.Errorf("toggleFollowAll() = %v, want %v", root.General.FollowAll, !v)
	}
	v = root.Doc.FollowSection
	root.toggleFollowSection(ctx)
	if v == root.Doc.FollowSection {
		t.Errorf("toggleFollowSection() = %v, want %v", root.Doc.FollowSection, !v)
	}
	v = root.Doc.HideOtherSection
	root.toggleHideOtherSection(ctx)
	if v == root.Doc.HideOtherSection {
		t.Errorf("toggleHideOtherSection() = %v, want %v", root.Doc.HideOtherSection, !v)
	}
}

func Test_rangeBA(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   int
		wantErr bool
	}{
		{
			name: "testInvalid",
			args: args{
				str: "invalid",
			},
			want:    0,
			want1:   0,
			wantErr: true,
		},
		{
			name: "testInvalid2",
			args: args{
				str: "1:invalid",
			},
			want:    1,
			want1:   0,
			wantErr: true,
		},
		{
			name: "testBefore",
			args: args{
				str: "1",
			},
			want:    1,
			want1:   0,
			wantErr: false,
		},
		{
			name: "testBA",
			args: args{
				str: "1:1",
			},
			want:    1,
			want1:   1,
			wantErr: false,
		},
		{
			name: "testOnlyAfter",
			args: args{
				str: ":1",
			},
			want:    0,
			want1:   1,
			wantErr: false,
		},
		{
			name: "testOnlyBefore",
			args: args{
				str: "1:",
			},
			want:    1,
			want1:   0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := rangeBA(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("rangeBA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("rangeBA() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("rangeBA() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_position(t *testing.T) {
	t.Parallel()
	type args struct {
		height int
		str    string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test1",
			args: args{
				height: 30,
				str:    "1",
			},
			want: 1,
		},
		{
			name: "test.5",
			args: args{
				height: 30,
				str:    ".5",
			},
			want: 15,
		},
		{
			name: "test20%",
			args: args{
				height: 30,
				str:    "20%",
			},
			want: 6,
		},
		{
			name: "test.3",
			args: args{
				height: 45,
				str:    "30%",
			},
			want: 13.5,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := calculatePosition(tt.args.height, tt.args.str); got != tt.want {
				t.Errorf("position() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jumpPosition(t *testing.T) {
	t.Parallel()
	type args struct {
		height int
		str    string
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 bool
	}{
		{
			name: "test1",
			args: args{
				height: 30,
				str:    "1",
			},
			want:  1,
			want1: false,
		},
		{
			name: "test.3",
			args: args{
				height: 10,
				str:    ".3",
			},
			want:  3,
			want1: false,
		},
		{
			name: "testMinus",
			args: args{
				height: 30,
				str:    "-10",
			},
			want:  19,
			want1: false,
		},
		{
			name: "testInvalid",
			args: args{
				height: 30,
				str:    "invalid",
			},
			want:  0,
			want1: false,
		},
		{
			name: "testInvalid2",
			args: args{
				height: 30,
				str:    ".i",
			},
			want:  0,
			want1: false,
		},
		{
			name: "testInvalid3",
			args: args{
				height: 30,
				str:    "p%",
			},
			want:  0,
			want1: false,
		},
		{
			name: "testSection",
			args: args{
				height: 30,
				str:    "s",
			},
			want:  0,
			want1: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, got1 := jumpPosition(tt.args.height, tt.args.str)
			if got != tt.want {
				t.Errorf("jumpPosition() = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("jumpPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_toggleColumnMode(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName    string
		wrapMode    bool
		columnWidth bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test column noWrap",
			fields: fields{
				fileName:    filepath.Join(testdata, "column.txt"),
				wrapMode:    false,
				columnWidth: false,
			},
		},
		{
			name: "test column wrap",
			fields: fields{
				fileName:    filepath.Join(testdata, "column.txt"),
				wrapMode:    true,
				columnWidth: false,
			},
		},
		{
			name: "test column width noWrap",
			fields: fields{
				fileName:    filepath.Join(testdata, "ps.txt"),
				wrapMode:    false,
				columnWidth: true,
			},
		},
		{
			name: "test column width wrap",
			fields: fields{
				fileName:    filepath.Join(testdata, "ps.txt"),
				wrapMode:    true,
				columnWidth: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.fields.fileName)
			if err != nil {
				t.Fatal(err)
			}
			root.Doc.ColumnMode = false
			root.Doc.WrapMode = tt.fields.wrapMode
			root.Doc.ColumnWidth = tt.fields.columnWidth
			root.prepareScreen()
			ctx := context.Background()
			root.everyUpdate(ctx)
			root.toggleColumnMode(ctx)
			if root.Doc.ColumnMode == false {
				t.Errorf("toggleColumnMode() = %v, want %v", root.Doc.ColumnMode, true)
			}
		})
	}
}

func TestRoot_toggleMouse(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		disableMouse bool
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "testDisable",
			fields: fields{disableMouse: true},
			want:   false,
		},
		{
			name:   "testEnable",
			fields: fields{disableMouse: false},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			root.Config.DisableMouse = tt.fields.disableMouse
			root.toggleMouse(context.Background())
			if root.Config.DisableMouse != tt.want {
				t.Errorf("toggleMouse() = %v, want %v", root.Config.DisableMouse, tt.want)
			}
		})
	}
}

func TestRoot_goLine(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName string
	}
	type args struct {
		input string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "testGoLine10",
			fields: fields{
				fileName: filepath.Join(testdata, "MOCK_DATA.csv"),
			},
			args: args{
				input: "10",
			},
			want: 9,
		},
		{
			name: "testGoLine.5",
			fields: fields{
				fileName: filepath.Join(testdata, "MOCK_DATA.csv"),
			},
			args: args{
				input: ".5",
			},
			want: 499,
		},
		{
			name: "testGoLine20%",
			fields: fields{
				fileName: filepath.Join(testdata, "MOCK_DATA.csv"),
			},
			args: args{
				input: "20%",
			},
			want: 199,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.fields.fileName)
			if err != nil {
				t.Fatal(err)
			}
			for !root.Doc.BufEOF() {
			}
			root.goLine(tt.args.input)
			if root.Doc.topLN != tt.want {
				t.Errorf("goLine() = %v, want %v", root.Doc.topLN, tt.want)
			}
		})
	}
}

func TestRoot_setHeader(t *testing.T) {
	root := rootHelper(t)
	root.prepareScreen()
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "testSetHeaderNil",
			args: args{
				input: "",
			},
			want: 0,
		},
		{
			name: "testSetHeaderMinus",
			args: args{
				input: "-1",
			},
			want: 0,
		},
		{
			name: "testSetHeader1",
			args: args{
				input: "1",
			},
			want: 1,
		},
		{
			name: "testSetHeaderNoChange",
			args: args{
				input: "1",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.setHeader(tt.args.input)
			if root.Doc.Header != tt.want {
				t.Errorf("setHeader() = %v, want %v", root.Doc.Header, tt.want)
			}
		})
	}
}
func TestRoot_setSkipLines(t *testing.T) {
	root := rootHelper(t)
	root.prepareScreen()
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "testSetSkipLines1",
			args: args{
				input: "1",
			},
			want: 1,
		},
		{
			name: "testSetSkipLinesNoChange",
			args: args{
				input: "1",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.setSkipLines(tt.args.input)
			if root.Doc.SkipLines != tt.want {
				t.Errorf("setSkipLines() = %v, want %v", root.Doc.SkipLines, tt.want)
			}
		})
	}
}
func TestRoot_setSectioNum(t *testing.T) {
	root := rootHelper(t)
	root.prepareScreen()
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "testSetSectionNum1",
			args: args{
				input: "1",
			},
			want: 1,
		},
		{
			name: "testSetSectionNumNoChange",
			args: args{
				input: "1",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.setSectionNum(tt.args.input)
			if root.Doc.SectionHeaderNum != tt.want {
				t.Errorf("setSectionNum() = %v, want %v", root.Doc.SectionHeaderNum, tt.want)
			}
		})
	}
}
func TestRoot_setSectionStart(t *testing.T) {
	root := rootHelper(t)
	root.prepareScreen()
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "testSetSectionStartNil",
			args: args{
				input: "",
			},
			want: 0,
		},
		{
			name: "testSetSectionStartMinus",
			args: args{
				input: "-1",
			},
			want: -1,
		},
		{
			name: "testSetSectionStart1",
			args: args{
				input: "1",
			},
			want: 1,
		},
		{
			name: "testSetSectionStartOutOfrange",
			args: args{
				input: "100",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.setSectionStart(tt.args.input)
			if root.Doc.SectionStartPosition != tt.want {
				t.Errorf("setSectionStart() = %v, want %v", root.Doc.SectionStartPosition, tt.want)
			}
		})
	}
}
