package oviewer

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func sectionHeader1Helper(t *testing.T) *Root {
	t.Helper()
	return sectionHeaderTextHelper(t, "section-header.txt")
}

func sectionHeader2Helper(t *testing.T) *Root {
	t.Helper()
	return sectionHeaderTextHelper(t, "section2.txt")
}

func sectionHeaderTextHelper(t *testing.T, fileName string) *Root {
	t.Helper()
	root := rootFileReadHelper(t, filepath.Join(testdata, fileName))
	m := root.Doc
	m.width = 80
	root.scr.vHeight = 24
	m.topLX = 0
	root.scr.lines = make(map[int]LineC)
	return root
}

func sectionStrHelper(t *testing.T, root *Root) string {
	t.Helper()
	lines := root.scr.lines
	lNs := lineNumbers(lines)
	var buf bytes.Buffer
	buf.WriteString("|")
	for _, lN := range lNs {
		line := lines[lN]
		buf.WriteString(fmt.Sprintf("(%d)%d-%02d|", lN, line.section, line.sectionNm))
	}
	return buf.String()
}

func TestRoot_prepareDraw_sectionHeader(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		wrapMode         bool
		skipLines        int
		header           int
		sectionHeader    bool
		sectionDelimiter string
		sectionHeaderNum int
		showGotoF        bool
		topLN            int
		jumpTargetHeight int
	}
	type want struct {
		headerHeight        int
		sectionHeaderHeight int
		topLN               int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "Test section-header",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           3,
				sectionHeader:    true,
				sectionDelimiter: "^#",
				sectionHeaderNum: 3,
				showGotoF:        false,
				topLN:            10,
				jumpTargetHeight: 0,
			},
			want: want{
				headerHeight:        5,
				sectionHeaderHeight: 5,
				topLN:               10,
			},
		},
		{
			name: "Test section-Error",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           3,
				sectionHeader:    true,
				sectionDelimiter: "errordelimiter",
				sectionHeaderNum: 3,
				showGotoF:        false,
				topLN:            10,
				jumpTargetHeight: 0,
			},
			want: want{
				headerHeight:        5,
				sectionHeaderHeight: 0,
				topLN:               10,
			},
		},
		{
			name: "Test section-noWrap",
			fields: fields{
				wrapMode:         false,
				skipLines:        0,
				header:           3,
				sectionHeader:    true,
				sectionDelimiter: "^#",
				sectionHeaderNum: 3,
				showGotoF:        false,
				topLN:            10,
				jumpTargetHeight: 0,
			},
			want: want{
				headerHeight:        3,
				sectionHeaderHeight: 3,
				topLN:               10,
			},
		},
		{
			name: "Test section-noWrap2",
			fields: fields{
				wrapMode:         false,
				skipLines:        0,
				header:           3,
				sectionHeader:    true,
				sectionDelimiter: "^#",
				sectionHeaderNum: 3,
				showGotoF:        true,
				topLN:            3,
				jumpTargetHeight: 0,
			},
			want: want{
				headerHeight:        3,
				sectionHeaderHeight: 3,
				topLN:               0,
			},
		},
		{
			name: "Test section-ShowGoto1",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           3,
				sectionHeader:    true,
				sectionDelimiter: "^#",
				sectionHeaderNum: 3,
				showGotoF:        true,
				topLN:            10,
				jumpTargetHeight: 0,
			},
			want: want{
				headerHeight:        5,
				sectionHeaderHeight: 5,
				topLN:               5,
			},
		},
		{
			name: "Test section-ShowGoto2",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           3,
				sectionHeader:    true,
				sectionDelimiter: "^#",
				sectionHeaderNum: 3,
				showGotoF:        true,
				topLN:            4,
				jumpTargetHeight: 0,
			},
			want: want{
				headerHeight:        5,
				sectionHeaderHeight: 5,
				topLN:               1,
			},
		},
		{
			name: "Test section-ShowGoto3",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           3,
				sectionHeader:    true,
				sectionDelimiter: "^#",
				sectionHeaderNum: 3,
				showGotoF:        true,
				topLN:            2,
				jumpTargetHeight: 0,
			},
			want: want{
				headerHeight:        5,
				sectionHeaderHeight: 5,
				topLN:               0,
			},
		},
		{
			name: "Test no-section",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           3,
				sectionHeader:    false,
				sectionDelimiter: "^#",
				sectionHeaderNum: 3,
				showGotoF:        true,
				topLN:            4,
				jumpTargetHeight: 0,
			},
			want: want{
				headerHeight:        5,
				sectionHeaderHeight: 0,
				topLN:               4,
			},
		},
		{
			name: "Test jumpTargetHeight",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           3,
				sectionHeader:    true,
				sectionDelimiter: "^#",
				sectionHeaderNum: 3,
				showGotoF:        true,
				topLN:            3,
				jumpTargetHeight: 5,
			},
			want: want{
				headerHeight:        5,
				sectionHeaderHeight: 5,
				topLN:               3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := sectionHeader1Helper(t)
			m := root.Doc
			m.SkipLines = tt.fields.skipLines
			m.Header = tt.fields.header
			m.WrapMode = tt.fields.wrapMode
			m.SectionHeader = tt.fields.sectionHeader
			m.setSectionDelimiter(tt.fields.sectionDelimiter)
			m.SectionHeaderNum = tt.fields.sectionHeaderNum
			m.showGotoF = tt.fields.showGotoF
			m.topLN = tt.fields.topLN
			m.jumpTargetHeight = tt.fields.jumpTargetHeight

			ctx := context.Background()
			root.prepareDraw(ctx)
			if root.Doc.headerHeight != tt.want.headerHeight {
				t.Errorf("header height got: %d, want: %d", root.Doc.headerHeight, tt.want.headerHeight)
			}
			if root.Doc.sectionHeaderHeight != tt.want.sectionHeaderHeight {
				t.Errorf("section header height got: %d, want: %d", root.Doc.sectionHeaderHeight, tt.want.sectionHeaderHeight)
			}
			if root.Doc.topLN != tt.want.topLN {
				t.Errorf("topLN got: %d, want: %d", root.Doc.topLN, tt.want.topLN)
			}
		})
	}
}

func TestRoot_prepareDraw_sectionHeader2(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		wrapMode         bool
		skipLines        int
		header           int
		sectionHeader    bool
		sectionDelimiter string
		sectionHeaderNum int
		sectionStartPos  int
		showGotoF        bool
		topLN            int
		jumpTargetHeight int
	}
	type want struct {
		sectionStr string
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "Test section2",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           0,
				sectionHeader:    true,
				sectionDelimiter: "^-",
				sectionHeaderNum: 3,
				sectionStartPos:  0,
				showGotoF:        false,
				topLN:            0,
				jumpTargetHeight: 0,
			},
			want: want{
				sectionStr: "|(0)0-01|(1)1-01|(2)1-02|(3)1-03|(4)1-04|(5)1-05|(6)1-06|(7)1-07|(8)2-01|(9)2-02|(10)2-03|(11)2-04|(12)2-05|(13)2-06|(14)2-07|(15)3-01|(16)3-02|(17)3-03|(18)3-04|(19)3-05|(20)3-06|(21)3-07|(22)4-01|(23)4-02|",
			},
		},
		{
			name: "Test section2+1",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           0,
				sectionHeader:    true,
				sectionDelimiter: "^-",
				sectionHeaderNum: 3,
				sectionStartPos:  1,
				showGotoF:        false,
				topLN:            2,
				jumpTargetHeight: 0,
			},
			want: want{
				sectionStr: "|(2)1-01|(3)1-02|(4)1-03|(5)1-04|(6)1-05|(7)1-06|(8)1-07|(9)2-01|(10)2-02|(11)2-03|(12)2-04|(13)2-05|(14)2-06|(15)2-07|(16)3-01|(17)3-02|(18)3-03|(19)3-04|(20)3-05|(21)3-06|(22)3-07|(23)4-01|(24)4-02|(25)4-03|",
			},
		},
		{
			name: "Test section2-1",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           3,
				sectionHeader:    true,
				sectionDelimiter: "^-",
				sectionHeaderNum: 3,
				sectionStartPos:  -1,
				showGotoF:        false,
				topLN:            2,
				jumpTargetHeight: 0,
			},
			want: want{
				sectionStr: "|(0)1-01|(1)1-02|(2)1-03|(5)1-04|(6)1-05|(7)2-01|(8)2-02|(9)2-03|(10)2-04|(11)2-05|(12)2-06|(13)2-07|(14)3-01|(15)3-02|(16)3-03|(17)3-04|(18)3-05|(19)3-06|(20)3-07|(21)4-01|(22)4-02|(23)4-03|(24)4-04|(25)4-05|(26)4-06|(27)4-07|(28)4-08|",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := sectionHeader2Helper(t)
			m := root.Doc
			m.SkipLines = tt.fields.skipLines
			m.Header = tt.fields.header
			m.WrapMode = tt.fields.wrapMode
			m.SectionHeader = tt.fields.sectionHeader
			m.setSectionDelimiter(tt.fields.sectionDelimiter)
			m.SectionHeaderNum = tt.fields.sectionHeaderNum
			m.SectionStartPosition = tt.fields.sectionStartPos
			m.showGotoF = tt.fields.showGotoF
			m.topLN = tt.fields.topLN
			m.jumpTargetHeight = tt.fields.jumpTargetHeight

			ctx := context.Background()
			root.prepareDraw(ctx)
			sectionStr := sectionStrHelper(t, root)
			if sectionStr != tt.want.sectionStr {
				t.Errorf("sectionStr got: \n%s, want: \n%s", sectionStr, tt.want.sectionStr)
			}
		})
	}
}

func TestRoot_prepareDraw_sectionStart(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		wrapMode         bool
		skipLines        int
		header           int
		sectionHeader    bool
		sectionDelimiter string
		sectionHeaderNum int
		sectionStart     int
		showGotoF        bool
		topLN            int
		jumpTargetHeight int
	}
	type want struct {
		headerHeight    int
		sectionHeaderLN int
		topLN           int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "Test section-start1",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           0,
				sectionHeader:    true,
				sectionDelimiter: "^#",
				sectionHeaderNum: 3,
				sectionStart:     1,
				showGotoF:        false,
				topLN:            10,
				jumpTargetHeight: 0,
			},
			want: want{
				headerHeight:    0,
				sectionHeaderLN: 4,
				topLN:           10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := sectionHeader1Helper(t)
			m := root.Doc
			m.SkipLines = tt.fields.skipLines
			m.Header = tt.fields.header
			m.WrapMode = tt.fields.wrapMode
			m.SectionHeader = tt.fields.sectionHeader
			m.setSectionDelimiter(tt.fields.sectionDelimiter)
			m.SectionHeaderNum = tt.fields.sectionHeaderNum
			m.SectionStartPosition = tt.fields.sectionStart
			m.showGotoF = tt.fields.showGotoF
			m.topLN = tt.fields.topLN
			m.jumpTargetHeight = tt.fields.jumpTargetHeight

			ctx := context.Background()
			root.prepareDraw(ctx)
			if root.Doc.headerHeight != tt.want.headerHeight {
				t.Errorf("header height got: %d, want: %d", root.Doc.headerHeight, tt.want.headerHeight)
			}
			if root.scr.sectionHeaderLN != tt.want.sectionHeaderLN {
				t.Errorf("section header LineNumber got: %d, want: %d", root.scr.sectionHeaderLN, tt.want.sectionHeaderLN)
			}
			if root.Doc.topLN != tt.want.topLN {
				t.Errorf("topLN got: %d, want: %d", root.Doc.topLN, tt.want.topLN)
			}
		})
	}
}

func TestRoot_prepareLines(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		wrapMode         bool
		skipLines        int
		header           int
		sectionHeader    bool
		sectionDelimiter string
		sectionHeaderNum int
	}
	type want struct {
		num int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "Test prepareLines",
			fields: fields{
				wrapMode:         true,
				skipLines:        0,
				header:           3,
				sectionHeader:    true,
				sectionDelimiter: "^#",
				sectionHeaderNum: 3,
			},
			want: want{
				num: 24,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := sectionHeader1Helper(t)
			m := root.Doc
			m.SkipLines = tt.fields.skipLines
			m.Header = tt.fields.header
			m.WrapMode = tt.fields.wrapMode
			m.SectionHeader = tt.fields.sectionHeader
			m.setSectionDelimiter(tt.fields.sectionDelimiter)
			m.SectionHeaderNum = tt.fields.sectionHeaderNum
			root.scr.lines = make(map[int]LineC)
			root.prepareLines(root.scr.lines)
			if len(root.scr.lines) != tt.want.num {
				t.Errorf("screen lines len got: %d, want: %d", len(root.scr.lines), tt.want.num)
			}
			root.prepareLines(root.scr.lines)
			if len(root.scr.lines) != tt.want.num {
				t.Errorf("screen lines len got: %d, want: %d", len(root.scr.lines), tt.want.num)
			}
		})
	}
}

func TestRoot_styleContent(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		wrapMode        bool
		PlainMode       bool
		ColumnMode      bool
		ColumnWidth     bool
		ColumnDelimiter string
		multiColorWords []string
	}
	type args struct {
		lineNum int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Test styleContent1",
			fields: fields{
				wrapMode:        true,
				PlainMode:       true,
				ColumnMode:      false,
				ColumnWidth:     false,
				multiColorWords: []string{"1", "2", "3"},
			},
		},
		{
			name: "Test styleContent2",
			fields: fields{
				wrapMode:        true,
				PlainMode:       false,
				ColumnMode:      true,
				ColumnWidth:     false,
				multiColorWords: nil,
			},
		},
		{
			name: "Test styleContent3",
			fields: fields{
				wrapMode:        true,
				PlainMode:       false,
				ColumnMode:      true,
				ColumnWidth:     false,
				ColumnDelimiter: " ",
				multiColorWords: nil,
			},
		},
		{
			name: "Test styleContent4",
			fields: fields{
				wrapMode:        true,
				PlainMode:       true,
				ColumnMode:      true,
				ColumnWidth:     true,
				multiColorWords: []string{"1", "2", "3"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := sectionHeader1Helper(t)
			m := root.Doc
			m.PlainMode = tt.fields.PlainMode
			m.ColumnMode = tt.fields.ColumnMode
			m.ColumnWidth = tt.fields.ColumnWidth
			m.setMultiColorWords(tt.fields.multiColorWords)
			m.setDelimiter(tt.fields.ColumnDelimiter)
			m.setColumnWidths()
			root.scr.lines = make(map[int]LineC)
			root.prepareDraw(context.Background())
			line := m.getLineC(tt.args.lineNum)
			if line.lc == nil {
				t.Fatal("line is nil")
			}
		})
	}
}

func TestRoot_searchHighlight(t *testing.T) {
	tcellNewScreen = fakeScreen
	searchHighlight := tcell.StyleDefault.Reverse(true)
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		searcher Searcher
	}
	type args struct {
		lineNum int
	}
	type want struct {
		str   string
		start int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Test searchHighlight",
			fields: fields{
				searcher: NewSearcher("dy", regexpCompile("dy", false), false, false),
			},
			args: args{
				lineNum: 6,
			},
			want: want{
				str:   "body 1",
				start: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := sectionHeader1Helper(t)

			line := root.Doc.getLineC(tt.args.lineNum)
			root.searcher = tt.fields.searcher
			root.StyleSearchHighlight = OVStyle{Reverse: true}
			root.searchHighlight(line)
			if line.str != tt.want.str {
				t.Errorf("\nline: %v\nwant: %v\n", line.str, tt.want.str)
			}
			if line.lc[tt.want.start].style != searchHighlight {
				t.Errorf("style got: %v want: %v", line.lc[tt.want.start].style, searchHighlight)
			}
		})
	}
}

func TestRoot_columnDelimiterHighlight(t *testing.T) {
	tcellNewScreen = fakeScreen
	columnHighlight := tcell.StyleDefault.Bold(true)
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		columnDelimiter string
		columnCursor    int
	}
	type args struct {
		lineNum int
	}
	type want struct {
		str   string
		start int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Test columnDelimiterHighlight1",
			fields: fields{
				columnDelimiter: "|",
				columnCursor:    0,
			},
			args: args{
				lineNum: 2,
			},
			want: want{
				str:   "| 4     | 5     | 6     |",
				start: 1,
			},
		},
		{
			name: "Test columnDelimiterHighlight2",
			fields: fields{
				columnDelimiter: "|",
				columnCursor:    1,
			},
			args: args{
				lineNum: 2,
			},
			want: want{
				str:   "| 4     | 5     | 6     |",
				start: 11,
			},
		},
		{
			name: "Test columnDelimiterHighlight3",
			fields: fields{
				columnDelimiter: "|",
				columnCursor:    2,
			},
			args: args{
				lineNum: 2,
			},
			want: want{
				str:   "| 4     | 5     | 6     |",
				start: 19,
			},
		},
		{
			name: "Test columnDelimiterHighlight4",
			fields: fields{
				columnDelimiter: "|",
				columnCursor:    3,
			},
			args: args{
				lineNum: 0,
			},
			want: want{
				str:   "| test1 | test2 | test3 |a",
				start: 25,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, filepath.Join(testdata, "column.txt"))
			m := root.Doc
			m.ColumnDelimiter = tt.fields.columnDelimiter
			m.ColumnDelimiterReg = condRegexpCompile(m.ColumnDelimiter)
			m.columnCursor = tt.fields.columnCursor
			root.StyleColumnHighlight = OVStyle{Bold: true}
			line := root.Doc.getLineC(tt.args.lineNum)
			root.columnDelimiterHighlight(line)
			if line.str != tt.want.str {
				t.Errorf("\nline: %v\nwant: %v\n", line.str, tt.want.str)
			}
			if line.lc[tt.want.start].style != columnHighlight {
				t.Errorf("style got: %v want: %v", line.lc[tt.want.start].style, columnHighlight)
			}
		})
	}
}

func TestRoot_columnWidthHighlight(t *testing.T) {
	tcellNewScreen = fakeScreen
	columnHighlight := tcell.StyleDefault.Bold(true)
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		columnCursor int
	}
	type args struct {
		lineNum int
	}
	type want struct {
		str   string
		start int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Test columnWidthHighlight1",
			fields: fields{
				columnCursor: 0,
			},
			args: args{
				lineNum: 2,
			},
			want: want{
				str:   "root           2  0.0  0.0      0     0 ?        S    Mar11   0:00 [kthreadd]",
				start: 1,
			},
		},
		{
			name: "Test columnWidthHighlight2",
			fields: fields{
				columnCursor: 10,
			},
			args: args{
				lineNum: 2,
			},
			want: want{
				str:   "root           2  0.0  0.0      0     0 ?        S    Mar11   0:00 [kthreadd]",
				start: 67,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, filepath.Join(testdata, "ps.txt"))
			root.StyleColumnHighlight = OVStyle{Bold: true}
			m := root.Doc
			m.ColumnWidth = true
			m.ColumnMode = true
			root.prepareScreen()
			root.prepareDraw(context.Background())
			m.setColumnWidths()
			t.Log(m.columnWidths)
			m.columnCursor = tt.fields.columnCursor
			line := root.Doc.getLineC(tt.args.lineNum)
			root.columnWidthHighlight(line)
			if line.str != tt.want.str {
				t.Errorf("\nline: %v\nwant: %v\n", line.str, tt.want.str)
			}
			if line.lc[tt.want.start].style != columnHighlight {
				v := bytes.Buffer{}
				for i, l := range line.lc {
					v.WriteString(fmt.Sprintf("%d:%v", i, l.style))
				}
				t.Logf("style: %v", v.String())
				t.Errorf("style got: %v want: %v", line.lc[tt.want.start].style, columnHighlight)
			}
		})
	}
}

func TestRoot_sectionNum(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := sectionHeader1Helper(t)
	root.prepareScreen()
	root.prepareDraw(context.Background())
	if got := root.sectionNum(root.scr.lines); !reflect.DeepEqual(got, root.scr.lines) {
		t.Errorf("Root.sectionNum() = %v, want %v", got, root.scr.lines)
	}
	root.Doc.SectionDelimiter = "errordelimiter"
	if got := root.sectionNum(root.scr.lines); !reflect.DeepEqual(got, root.scr.lines) {
		t.Errorf("Root.sectionNum() = %v, want %v", got, root.scr.lines)
	}
}

func Test_findColumnEnd(t *testing.T) {
	type args struct {
		str string
		pos []int
		n   int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Test findColumnEnd1Over",
			args: args{
				str: "012345678901234567890123",
				pos: []int{7, 15},
				n:   0,
			},
			want: 24,
		},
		{
			name: "Test findColumnEnd2",
			args: args{
				str: "header1 header2 header3",
				pos: []int{7, 15},
				n:   0,
			},
			want: 7,
		},
		{
			name: "Test findColumnEnd3",
			args: args{
				str: "1       2       3",
				pos: []int{7, 15},
				n:   0,
			},
			want: 7,
		},
		{
			name: "Test findColumnEnd4",
			args: args{
				str: "     1       2        3",
				pos: []int{7, 15},
				n:   0,
			},
			want: 7,
		},
		{
			name: "Test findColumnEnd6Over1",
			args: args{
				str: "123   456789012 345678901234",
				pos: []int{7, 15},
				n:   0,
			},
			want: 5,
		},
		{
			name: "Test findColumnEnd6Over2",
			args: args{
				str: "123   456789012 345678901234",
				pos: []int{7, 15},
				n:   1,
			},
			want: 15,
		},
		{
			name: "Test findColumnEnd7Over1",
			args: args{
				str: "abedefghi jkujik mnoopqr",
				pos: []int{7, 15},
				n:   0,
			},
			want: 9,
		},
		{
			name: "Test findColumnEnd7Over2",
			args: args{
				str: "abedefghi jkujikl mnoopqr",
				pos: []int{7, 15},
				n:   1,
			},
			want: 17,
		},
		{
			name: "Test findColumnEnd8Over1",
			args: args{
				str: "abedefghi jkujikl mnoopqr",
				pos: []int{7, 15},
				n:   0,
			},
			want: 9,
		},
		{
			name: "Test findColumnEnd8Over2",
			args: args{
				str: "あいうえお かきくけこ さしすせそ",
				pos: []int{7, 15},
				n:   1,
			},
			want: 21,
		},
		{
			name: "Test findColumnEnd9Over",
			args: args{
				str: "abedefg hijkujiklmnoopqrstuvxyz",
				pos: []int{7, 15},
				n:   1,
			},
			want: 31,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := StrToContents(tt.args.str, 8)
			if got := findColumnEnd(lc, tt.args.pos, tt.args.n); got != tt.want {
				t.Errorf("findColumnEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_sectionHeaderTimeOut(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := sectionHeader2Helper(t)
	root.prepareScreen()
	ctx := context.Background()
	root.prepareDraw(ctx)
	root.Doc.setSectionDelimiter("^-")
	root.Doc.SectionHeader = true
	lx, ln := root.sectionHeader(ctx, 0, sectionTimeOut)
	if lx != 0 || ln != 0 {
		t.Errorf("lx: %d, ln: %d", lx, ln)
	}
}

func TestRoot_sectionHeader(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		sectionDelimiter string
	}
	type args struct {
		ctx     context.Context
		lN      int
		timeOut time.Duration
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wants     int
		wante     int
		wantDelim string
	}{
		{
			name: "Test searchSectionHeader1",
			fields: fields{
				sectionDelimiter: "^-",
			},
			args: args{
				ctx:     context.Background(),
				lN:      4,
				timeOut: 1000,
			},
			wants:     1,
			wante:     1,
			wantDelim: "^-",
		},
		{
			name: "Test searchSectionHeaderNoDelimiter",
			fields: fields{
				sectionDelimiter: "",
			},
			args: args{
				ctx:     context.Background(),
				lN:      4,
				timeOut: 1000,
			},
			wants:     0,
			wante:     0,
			wantDelim: "",
		},
		{
			name: "Test searchSectionHeaderTimeOut",
			fields: fields{
				sectionDelimiter: "^-",
			},
			args: args{
				ctx:     context.Background(),
				lN:      4,
				timeOut: 0,
			},
			wants:     0,
			wante:     0,
			wantDelim: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := sectionHeader2Helper(t)
			root.prepareScreen()
			m := root.Doc
			m.SectionHeader = true
			m.setSectionDelimiter(tt.fields.sectionDelimiter)
			got1, got2 := root.sectionHeader(tt.args.ctx, tt.args.lN, tt.args.timeOut)
			if got1 != tt.wants {
				t.Errorf("Root.sectionHeader() = %v, want %v", got1, tt.wants)
			}
			if got2 != tt.wante {
				t.Errorf("Root.sectionHeader() = %v, want %v", got2, tt.wante)
			}
			if m.SectionDelimiter != tt.wantDelim {
				t.Errorf("Root.sectionHeader() = %v, want %v", m.SectionDelimiter, tt.wantDelim)
			}
		})
	}
}

func TestRoot_setAlignConverter(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName        string
		header          int
		columnWidth     bool
		columnDelimiter string
	}
	tests := []struct {
		name   string
		fields fields
		want   []int
	}{
		{
			name: "Test setAlignConverter1",
			fields: fields{
				fileName:    "width-column.txt",
				header:      1,
				columnWidth: true,
			},
			want: []int{3, 5},
		},
		{
			name: "Test setAlignConverterDelimiter",
			fields: fields{
				fileName:        "column.txt",
				header:          1,
				columnWidth:     false,
				columnDelimiter: "|",
			},
			want: []int{-1, 7, 7, 7},
		},
		{
			name: "Test setAlignConverterDelimiter2",
			fields: fields{
				fileName:        "test.csv",
				header:          1,
				columnWidth:     false,
				columnDelimiter: ",",
			},
			want: []int{0, 5, 1},
		},
		{
			name: "Test setAlignNoColumn",
			fields: fields{
				fileName:    "test.txt",
				header:      1,
				columnWidth: true,
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, filepath.Join(testdata, tt.fields.fileName))
			m := root.Doc
			m.width = 80
			root.scr.vHeight = 24
			m.topLX = 0
			m.Header = tt.fields.header
			m.ColumnMode = true
			m.ColumnWidth = tt.fields.columnWidth
			m.ColumnDelimiter = tt.fields.columnDelimiter
			if m.ColumnDelimiter != "" {
				m.ColumnDelimiterReg = condRegexpCompile(m.ColumnDelimiter)
			}
			root.scr.lines = make(map[int]LineC)
			root.prepareScreen()
			ctx := context.Background()
			root.Doc.ColumnRainbow = true
			root.Doc.Converter = convAlign
			root.Doc.alignConv = newAlignConverter(true)
			root.prepareDraw(ctx)
			if got := root.Doc.alignConv.maxWidths; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Root.setAlignConverter() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_maxWidthsDelm(t *testing.T) {
	type args struct {
		maxWidths    []int
		delimiter    string
		delimiterReg *regexp.Regexp
	}
	tests := []struct {
		name string
		str  string
		args args
		want []int
	}{
		{
			name: "Test maxWidthsDelm1",
			str:  "a, b, c, d, e, f, g",
			args: args{
				maxWidths:    []int{0, 0, 0, 0, 0},
				delimiter:    ",",
				delimiterReg: regexpCompile(",", false),
			},
			want: []int{0, 2, 2, 2, 2, 2},
		},
		{
			name: "Test maxWidthsDelm2",
			str:  "no delimiter",
			args: args{
				maxWidths:    []int{2, 2, 2, 2, 2, 2},
				delimiter:    ",",
				delimiterReg: regexpCompile(",", false),
			},
			want: []int{2, 2, 2, 2, 2, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := StrToContents(tt.str, 8)
			if got := maxWidthsDelm(lc, tt.args.maxWidths, tt.args.delimiter, tt.args.delimiterReg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("maxWidthsDelm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_maxWidthsWidth(t *testing.T) {
	type args struct {
		maxWidths []int
		widths    []int
		addRight  []int
	}
	tests := []struct {
		name  string
		str   string
		args  args
		want  []int
		want2 []int
	}{
		{
			name: "Test maxWidthsWidth1",
			str:  " 1   tsst        21",
			args: args{
				maxWidths: []int{0},
				widths:    []int{2, 11},
				addRight:  []int{0, 0},
			},
			want:  []int{1, 4},
			want2: []int{0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := StrToContents(tt.str, 8)
			got, got2 := maxWidthsWidth(lc, tt.args.maxWidths, tt.args.widths, tt.args.addRight)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("maxWidthsWidth() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("maxWidthsWidth() addRight = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_trimWidth(t *testing.T) {
	type args struct {
		lc contents
	}
	tests := []struct {
		name      string
		args      args
		want      int
		wantRight int
	}{
		{
			name: "Test trimLC1",
			args: args{
				lc: StrToContents("test", 8),
			},
			want:      4,
			wantRight: 0,
		},
		{
			name: "Test trimLC2",
			args: args{
				lc: StrToContents("    test   ", 8),
			},
			want:      4,
			wantRight: 1,
		},
		{
			name: "Test trimLC3",
			args: args{
				lc: StrToContents("    123.0 ", 8),
			},
			want:      5,
			wantRight: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotLeft := trimWidth(tt.args.lc)
			if got != tt.want {
				t.Errorf("trimLC() got = %v, want %v", got, tt.want)
			}
			if gotLeft != tt.wantRight {
				t.Errorf("trimLC() gotLeft = %v, want %v", gotLeft, tt.wantRight)
			}
		})
	}
}
