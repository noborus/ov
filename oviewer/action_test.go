package oviewer

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

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
	root.toggleColumnWidth(ctx)
	if v != root.Doc.ColumnWidth {
		t.Errorf("toggleColumnWidth() = %v, want %v", root.Doc.ColumnWidth, v)
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
	t.Parallel()
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
		str    string
		height int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test1",
			args: args{
				str:    "1",
				height: 30,
			},
			want: 1,
		},
		{
			name: "test.5",
			args: args{
				str:    ".5",
				height: 30,
			},
			want: 15,
		},
		{
			name: "test20%",
			args: args{
				str:    "20%",
				height: 30,
			},
			want: 6,
		},
		{
			name: "test.3",
			args: args{
				str:    "30%",
				height: 45,
			},
			want: 13.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := calculatePosition(tt.args.str, tt.args.height); got != tt.want {
				t.Errorf("position() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_setJumpTarget(t *testing.T) {
	root := rootHelper(t)
	type fields struct{}
	type args struct {
		input string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantBool bool
		want     int
	}{
		{
			name: "testJumpTargetBlank",
			args: args{
				input: "",
			},
			wantBool: false,
			want:     0,
		},
		{
			name: "testJumpTargetBlank2",
			args: args{
				input: "",
			},
			wantBool: false,
			want:     0,
		},
		{
			name: "testJumpTargetOutOfRange",
			args: args{
				input: "200",
			},
			wantBool: false,
			want:     0,
		},
		{
			name: "testJumpTarget1",
			args: args{
				input: "1",
			},
			wantBool: false,
			want:     1,
		},
		{
			name: "testJumpTarget0",
			args: args{
				input: "0",
			},
			wantBool: false,
			want:     0,
		},
		{
			name: "testJumpTarget section",
			args: args{
				input: "section",
			},
			wantBool: true,
			want:     0,
		},
		{
			name: "testJumpTargetMinus",
			args: args{
				input: "-1",
			},
			wantBool: false,
			want:     23,
		},
	}
	for _, tt := range tests {
		root.prepareScreen()
		t.Run(tt.name, func(t *testing.T) {
			root.setJumpTarget(tt.args.input)
			if root.Doc.jumpTargetSection != tt.wantBool {
				t.Errorf("setJumpTarget() = %v, want %v", root.Doc.JumpTarget, tt.wantBool)
			}
			if root.Doc.jumpTargetHeight != tt.want {
				t.Errorf("setJumpTarget() height = %v, want %v", root.Doc.jumpTargetHeight, tt.want)
			}
		})
	}
}

func Test_jumpPosition(t *testing.T) {
	t.Parallel()
	type args struct {
		str    string
		height int
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
				str:    "1",
				height: 30,
			},
			want:  1,
			want1: false,
		},
		{
			name: "test.3",
			args: args{
				str:    ".3",
				height: 10,
			},
			want:  3,
			want1: false,
		},
		{
			name: "testMinus",
			args: args{
				str:    "-10",
				height: 30,
			},
			want:  19,
			want1: false,
		},
		{
			name: "testInvalid",
			args: args{
				str:    "invalid",
				height: 30,
			},
			want:  0,
			want1: false,
		},
		{
			name: "testInvalid2",
			args: args{
				str:    ".i",
				height: 30,
			},
			want:  0,
			want1: false,
		},
		{
			name: "testInvalid3",
			args: args{
				str:    "p%",
				height: 30,
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, got1 := jumpPosition(tt.args.str, tt.args.height)
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
			root := rootFileReadHelper(t, tt.fields.fileName)
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
			root := rootFileReadHelper(t, tt.fields.fileName)
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
			name: "testSetSkipErr",
			args: args{
				input: "err",
			},
			want: 0,
		},
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
			name: "testSetSectionNumNil",
			args: args{
				input: "",
			},
			want: 0,
		},
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

func TestRoot_searchGo(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName          string
		searchWord        string
		jumpTargetSection bool
		sectionDelimiter  string
	}
	type args struct {
		lN int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "testSearchGo1",
			fields: fields{
				fileName:          filepath.Join(testdata, "MOCK_DATA.csv"),
				searchWord:        "1",
				sectionDelimiter:  "",
				jumpTargetSection: false,
			},
			args: args{
				lN: 1,
			},
			want: 1,
		},
		{
			name: "test section2-1",
			fields: fields{
				fileName:          filepath.Join(testdata, "section2.txt"),
				searchWord:        "5",
				sectionDelimiter:  "^-",
				jumpTargetSection: false,
			},
			args: args{
				lN: 1,
			},
			want: 1,
		},
		{
			name: "test section2-2",
			fields: fields{
				fileName:          filepath.Join(testdata, "section2.txt"),
				searchWord:        "5",
				sectionDelimiter:  "^-",
				jumpTargetSection: true,
			},
			args: args{
				lN: 5,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileName)
			root.Doc.WrapMode = true
			root.prepareScreen()
			ctx := context.Background()
			root.everyUpdate(ctx)
			root.prepareDraw(ctx)
			root.setSearcher(tt.fields.searchWord, false)
			root.Doc.jumpTargetSection = tt.fields.jumpTargetSection
			root.setSectionDelimiter(tt.fields.sectionDelimiter)
			root.Doc.SectionHeader = true
			root.searchGo(ctx, tt.args.lN, root.searcher)
			if root.Doc.topLN != tt.want {
				t.Errorf("searchGo() = %v, want %v", root.Doc.topLN, tt.want)
			}
		})
	}
}

func TestRoot_setMultiColor(t *testing.T) {
	root := rootHelper(t)
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "testSetMultiColor",
			args: args{
				input: "red,blue",
			},
			want: []string{"red", "blue"},
		},
		{
			name: "testSetMultiColor2",
			args: args{
				input: "red,\"blue\"",
			},
			want: []string{"red", "blue"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.setMultiColor(tt.args.input)
			if reflect.DeepEqual(root.Doc.MultiColorWords, tt.want) {
				t.Errorf("setMultiColor() = %v, want %v", root.Doc.MultiColorWords, tt.want)
			}
		})
	}
}

func TestRoot_setTabWidth(t *testing.T) {
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
			name: "testSetTabWidth4",
			args: args{
				input: "4",
			},
			want: 4,
		},
		{
			name: "testSetTabWidth8",
			args: args{
				input: "8",
			},
			want: 8,
		},
		{
			name: "testSetTabWidth0",
			args: args{
				input: "0",
			},
			want: 0,
		},
		{
			name: "testSetTabWidthInvalid",
			args: args{
				input: "invalid",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.Doc.TabWidth = 0
			root.setTabWidth(tt.args.input)
			if root.Doc.TabWidth != tt.want {
				t.Errorf("got %v, want %v", root.Doc.TabWidth, tt.want)
			}
		})
	}
}

func TestRoot_ColumnDelimiterWrapMode(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		wrapMode     bool
		x            int
		columnCursor int
	}
	type want struct {
		wrapMode     bool
		x            int
		columnCursor int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "testWrapMode0",
			fields: fields{
				x:            0,
				columnCursor: 0,
				wrapMode:     false,
			},
			want: want{
				x:            0,
				columnCursor: 0,
				wrapMode:     true,
			},
		},
		{
			name: "testWrapMode1",
			fields: fields{
				x:            10,
				columnCursor: 1,
				wrapMode:     false,
			},
			want: want{
				x:            10,
				columnCursor: 1,
				wrapMode:     true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, filepath.Join(testdata, "MOCK_DATA.csv"))
			root.prepareScreen()
			ctx := context.Background()
			root.prepareDraw(ctx)
			root.setDelimiter(",")
			root.Doc.ColumnMode = true
			root.Doc.WrapMode = tt.fields.wrapMode
			root.Doc.x = tt.fields.x
			root.Doc.columnCursor = tt.fields.columnCursor
			root.toggleWrapMode(ctx)
			if root.Doc.WrapMode != tt.want.wrapMode {
				t.Errorf("ColumnDelimiter WrapMode() mode = %v, want %v", root.Doc.WrapMode, tt.want.wrapMode)
			}
			if root.Doc.x != tt.want.x {
				t.Errorf("ColumnDelimiter WrapMode() x = %v, want %v", root.Doc.x, tt.want.x)
			}
			if root.Doc.columnCursor != tt.want.columnCursor {
				t.Errorf("ColumnDelimiter WrapMode() columnCursor = %v, want %v", root.Doc.columnCursor, tt.want.columnCursor)
			}
		})
	}
}

func TestRoot_ColumnWidthWrapMode(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := rootFileReadHelper(t, filepath.Join(testdata, "ps.txt"))
	root.prepareScreen()
	type fields struct {
		wrapMode     bool
		x            int
		columnCursor int
	}
	type want struct {
		wrapMode     bool
		x            int
		columnCursor int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "testWrapMode0",
			fields: fields{
				x:            0,
				columnCursor: 0,
				wrapMode:     false,
			},
			want: want{
				x:            0,
				columnCursor: 0,
				wrapMode:     true,
			},
		},
		{
			name: "testWrapMode1",
			fields: fields{
				x:            10,
				columnCursor: 1,
				wrapMode:     false,
			},
			want: want{
				x:            10,
				columnCursor: 1,
				wrapMode:     true,
			},
		},
		{
			name: "testWrapMode2",
			fields: fields{
				x:            10,
				columnCursor: 15,
				wrapMode:     false,
			},
			want: want{
				x:            10,
				columnCursor: 15,
				wrapMode:     true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			root.Doc.ColumnWidth = true
			root.Doc.WrapMode = tt.fields.wrapMode
			root.Doc.x = tt.fields.x
			root.Doc.columnCursor = tt.fields.columnCursor
			root.Doc.setColumnWidths()
			root.prepareDraw(ctx)
			root.toggleWrapMode(ctx)
			if root.Doc.WrapMode != tt.want.wrapMode {
				t.Errorf("ColumnWidth WrapMode() mode = %v, want %v", root.Doc.WrapMode, tt.want.wrapMode)
			}
			if root.Doc.x != tt.want.x {
				t.Errorf("ColumnWidth WrapMode() x = %v, want %v", root.Doc.x, tt.want.x)
			}
			if root.Doc.columnCursor != tt.want.columnCursor {
				t.Errorf("ColumnWidth WrapMode() columnCursor = %v, want %v", root.Doc.columnCursor, tt.want.columnCursor)
			}
		})
	}
}

func TestRoot_Mark(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := rootFileReadHelper(t, filepath.Join(testdata, "test3.txt"))
	t.Run("TestMark", func(t *testing.T) {
		root.prepareScreen()
		ctx := context.Background()
		root.Doc.topLN = 1
		root.draw(ctx)
		root.addMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, []int{1}) {
			t.Errorf("addMark() = %#v, want %#v", root.Doc.marked, []int{1})
		}
		root.Doc.topLN = 10
		root.draw(ctx)
		root.addMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, []int{1, 10}) {
			t.Errorf("addMark() = %#v, want %#v", root.Doc.marked, []int{1, 10})
		}
		root.Doc.topLN = 1
		root.draw(ctx)
		root.removeMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, []int{10}) {
			t.Errorf("removeAllMark() = %#v, want %#v", root.Doc.marked, []int{10})
		}
		root.removeMark(ctx)
		root.Doc.topLN = 2
		root.draw(ctx)
		root.addMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, []int{10, 2}) {
			t.Errorf("addMark() = %#v, want %#v", root.Doc.marked, []int{10, 2})
		}
		root.removeAllMark(ctx)
		if !reflect.DeepEqual(root.Doc.marked, []int(nil)) {
			t.Errorf("removeAllMark() = %#v, want %#v", root.Doc.marked, []int(nil))
		}
	})
}

func TestRoot_nextMark(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := rootFileReadHelper(t, filepath.Join(testdata, "test3.txt"))
	root.prepareScreen()
	tests := []struct {
		name        string
		markedPoint int
		wantLine    int
	}{
		{
			name:        "testMarkNext1",
			markedPoint: 0,
			wantLine:    3,
		},
		{
			name:        "testMarkNext2",
			markedPoint: 1,
			wantLine:    5,
		},
		{
			name:        "testMarkNext3",
			markedPoint: 2,
			wantLine:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.nextMark(context.Background()) // no marked
			root.Doc.marked = []int{1, 3, 5}
			root.Doc.markedPoint = tt.markedPoint
			root.nextMark(context.Background())
			if root.Doc.topLN != tt.wantLine {
				t.Errorf("got line %d, want line %d", root.Doc.topLN, tt.wantLine)
			}
		})
	}
}

func TestRoot_prevMark(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := rootFileReadHelper(t, filepath.Join(testdata, "test3.txt"))
	root.prepareScreen()
	tests := []struct {
		name        string
		markedPoint int
		wantLine    int
	}{
		{
			name:        "testMarkPrev1",
			markedPoint: 2,
			wantLine:    3,
		},
		{
			name:        "testMarkPrev2",
			markedPoint: 1,
			wantLine:    1,
		},
		{
			name:        "testMarkPrev3",
			markedPoint: 0,
			wantLine:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root.prevMark(context.Background()) // no marked
			root.Doc.marked = []int{1, 3, 5}
			root.Doc.markedPoint = tt.markedPoint
			root.prevMark(context.Background())
			if root.Doc.topLN != tt.wantLine {
				t.Errorf("got line %d, want line %d", root.Doc.topLN, tt.wantLine)
			}
		})
	}
}

func TestRoot_setWriteBA(t *testing.T) {
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
		want   bool
	}{
		{
			name: "testWriteBA",
			fields: fields{
				fileName: filepath.Join(testdata, "test3.txt"),
			},
			args: args{
				input: "1:2",
			},
			want: true,
		},
		{
			name: "testWriteBA",
			fields: fields{
				fileName: filepath.Join(testdata, "test3.txt"),
			},
			args: args{
				input: "err",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileName)
			ctx := context.Background()
			root.setWriteBA(ctx, tt.args.input)
			if root.IsWriteOriginal != tt.want {
				t.Errorf("setWriteBA() = %v, want %v", root.IsWriteOriginal, tt.want)
			}
		})
	}
}

func TestRoot_modeConfig(t *testing.T) {
	type args struct {
		modeName string
	}
	tests := []struct {
		name    string
		args    args
		want    general
		wantErr bool
	}{
		{
			name: "testModeConfig error",
			args: args{
				modeName: "test",
			},
			want:    general{},
			wantErr: true,
		},
		{
			name: "testModeConfig general",
			args: args{
				modeName: nameGeneral,
			},
			want:    general{},
			wantErr: false,
		},
		{
			name: "testModeConfig a",
			args: args{
				modeName: "a",
			},
			want:    general{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			root.Config = Config{
				General: general{},
				Mode: map[string]general{
					"a": {},
					"b": {},
				},
			}
			got, err := root.modeConfig(tt.args.modeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Root.modeConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Root.modeConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_follow(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		wrapMode bool
	}
	type args struct {
		fileNames []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "testFollow1",
			fields: fields{
				wrapMode: false,
			},
			args: args{
				fileNames: []string{filepath.Join(testdata, "test3.txt")},
			},
			want: 12322,
		},
		{
			name: "testFollow2",
			fields: fields{
				wrapMode: true,
			},
			args: args{
				fileNames: []string{filepath.Join(testdata, "test3.txt")},
			},
			want: 12322,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.args.fileNames...)
			root.Doc.WrapMode = tt.fields.wrapMode
			ctx := context.Background()
			root.prepareScreen()
			root.everyUpdate(ctx)
			root.follow(ctx)
			if root.Doc.topLN != tt.want {
				t.Errorf("follow() topLN = %v, want %v", root.Doc.topLN, tt.want)
			}
			root.Cancel(ctx)
		})
	}
}

func TestRoot_followAll(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		wrapMode bool
	}
	type args struct {
		fileNames []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "testFollowAll1",
			fields: fields{
				wrapMode: false,
			},
			args: args{
				fileNames: []string{
					filepath.Join(testdata, "test3.txt"),
					filepath.Join(testdata, "test.txt"),
				},
			},
			want: 12322,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.args.fileNames...)
			root.General.FollowAll = true
			root.Doc.WrapMode = tt.fields.wrapMode
			ctx := context.Background()
			root.prepareScreen()
			root.everyUpdate(ctx)
			root.followAll(ctx)
			if root.Doc.topLN != tt.want {
				t.Errorf("follow() topLN = %v, want %v", root.Doc.topLN, tt.want)
			}
		})
	}
}

func TestRoot_WriteQuit(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames        []string
		sectionDelimiter string
		HideOtherSection bool
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "testWriteQuit1",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section.txt")},
				sectionDelimiter: "",
				HideOtherSection: false,
			},
			want: 0,
		},
		{
			name: "testWriteQuit2",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section.txt")},
				sectionDelimiter: "",
				HideOtherSection: true,
			},
			want: 0,
		},
		{
			name: "testWriteQuit3",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section.txt")},
				sectionDelimiter: "^#",
				HideOtherSection: true,
			},
			want: 2,
		},
		{
			name: "testWriteQuit4",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section.txt")},
				sectionDelimiter: "^notfound",
				HideOtherSection: true,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileNames...)
			root.prepareScreen()
			ctx := context.Background()
			root.setSectionDelimiter(tt.fields.sectionDelimiter)
			root.prepareDraw(context.Background())
			root.Doc.HideOtherSection = tt.fields.HideOtherSection
			root.WriteQuit(ctx)
			if got := root.AfterWriteOriginal; got != tt.want {
				t.Errorf("WriteQuit() AfterWriteOriginal = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_tailSection(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames        []string
		sectionDelimiter string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "testTailSection",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section.txt")},
				sectionDelimiter: "^#",
			},
			want: 8,
		},
		{
			name: "testTailSectionErr",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section.txt")},
				sectionDelimiter: "^testtesttest",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileNames...)
			root.prepareScreen()
			ctx := context.Background()
			root.setSectionDelimiter(tt.fields.sectionDelimiter)
			root.tailSection(ctx)
			if got := root.Doc.lastSectionPosNum; got != tt.want {
				t.Errorf("tailSection() lastSectionPosNum = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_setConverter(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want Converter
	}{
		{
			name: "testSetConverterEscape",
			args: args{
				name: convEscaped,
			},
			want: newESConverter(),
		},
		{
			name: "testSetConverterRaw",
			args: args{
				name: convRaw,
			},
			want: newRawConverter(),
		},
		{
			name: "testSetConverterAlign",
			args: args{
				name: convAlign,
			},
			want: newAlignConverter(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			ctx := context.Background()
			root.setConverter(ctx, tt.args.name)
			got := root.Doc.conv
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setConverter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_ShrinkColumn(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName  string
		converter string
	}
	type args struct {
		cursor int
		shrink bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "testShrinkColumn",
			fields: fields{
				fileName:  filepath.Join(testdata, "MOCK_DATA.csv"),
				converter: convAlign,
			},
			args: args{
				cursor: 0,
				shrink: true,
			},
			wantErr: false,
		},
		{
			name: "testShrinkColumnFalse",
			fields: fields{
				fileName:  filepath.Join(testdata, "MOCK_DATA.csv"),
				converter: convEscaped,
			},
			args: args{
				cursor: 0,
				shrink: true,
			},
			wantErr: true,
		},
		{
			name: "testShrinkColumnOver",
			fields: fields{
				fileName:  filepath.Join(testdata, "MOCK_DATA.csv"),
				converter: convAlign,
			},
			args: args{
				cursor: 20,
				shrink: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileName)
			ctx := context.Background()
			root.prepareScreen()
			root.Doc.Converter = tt.fields.converter
			root.Doc.ColumnDelimiter = ","
			root.prepareDraw(ctx)
			if err := root.Doc.shrinkColumn(tt.args.cursor, tt.args.shrink); (err != nil) != tt.wantErr {
				t.Errorf("Root.ShrinkColumn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestDocument_specifiedAlign(t *testing.T) {
	type fields struct {
		fileName    string
		columnAttrs []columnAttribute
		converter   string
	}
	type args struct {
		cursor int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    specifiedAlign
		wantErr bool
	}{
		{
			name: "testValidColumn",
			fields: fields{
				fileName: filepath.Join(testdata, "MOCK_DATA.csv"),
				columnAttrs: []columnAttribute{
					{shrink: false, specifiedAlign: LeftAlign},
					{shrink: false, specifiedAlign: LeftAlign},
				},
				converter: convAlign,
			},
			args: args{
				cursor: 0,
			},
			want:    RightAlign,
			wantErr: false,
		},
		{
			name: "testInvalidColumn",
			fields: fields{
				fileName: filepath.Join(testdata, "MOCK_DATA.csv"),
				columnAttrs: []columnAttribute{
					{shrink: false, specifiedAlign: LeftAlign},
					{shrink: false, specifiedAlign: LeftAlign},
				},
				converter: convAlign,
			},
			args: args{
				cursor: 11,
			},
			want:    Unspecified,
			wantErr: true,
		},
		{
			name: "testNotAlignMode",
			fields: fields{
				fileName: filepath.Join(testdata, "MOCK_DATA.csv"),
				columnAttrs: []columnAttribute{
					{shrink: false, specifiedAlign: LeftAlign},
					{shrink: false, specifiedAlign: LeftAlign},
				},
				converter: convEscaped,
			},
			args: args{
				cursor: 0,
			},
			want:    Unspecified,
			wantErr: true,
		},
		{
			name: "testCycleAlign",
			fields: fields{
				fileName: filepath.Join(testdata, "MOCK_DATA.csv"),
				columnAttrs: []columnAttribute{
					{shrink: false, specifiedAlign: LeftAlign},
					{shrink: false, specifiedAlign: LeftAlign},
				},
				converter: convAlign,
			},
			args: args{
				cursor: 0,
			},
			want:    RightAlign,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := OpenDocument(tt.fields.fileName)
			if err != nil {
				t.Fatal(err)
			}
			m.Converter = tt.fields.converter
			m.alignConv.columnAttrs = tt.fields.columnAttrs
			got, err := m.specifiedAlign(tt.args.cursor)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.specifiedAlign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.specifiedAlign() = %v, want %v", got, tt.want)
			}
		})
	}
}
