package oviewer

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func sectionHeaderText(t *testing.T) *Root {
	t.Helper()
	root, err := Open(filepath.Join(testdata, "section-header.txt"))
	if err != nil {
		t.Fatal(err)
	}
	m := root.Doc
	m.width = 80
	root.scr.vHeight = 24
	m.topLX = 0
	root.scr.contents = make(map[int]LineC)
	for !m.BufEOF() {
	}
	return root
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
			root := sectionHeaderText(t)
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
			root := sectionHeaderText(t)
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

func TestRoot_screenContents(t *testing.T) {
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
			name: "Test screenContents",
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
			root := sectionHeaderText(t)
			m := root.Doc
			m.SkipLines = tt.fields.skipLines
			m.Header = tt.fields.header
			m.WrapMode = tt.fields.wrapMode
			m.SectionHeader = tt.fields.sectionHeader
			m.setSectionDelimiter(tt.fields.sectionDelimiter)
			m.SectionHeaderNum = tt.fields.sectionHeaderNum
			root.scr.contents = make(map[int]LineC)
			root.screenContents(root.scr.contents)
			if len(root.scr.contents) != tt.want.num {
				t.Errorf("screen contents len got: %d, want: %d", len(root.scr.contents), tt.want.num)
			}
			root.screenContents(root.scr.contents)
			if len(root.scr.contents) != tt.want.num {
				t.Errorf("screen contents len got: %d, want: %d", len(root.scr.contents), tt.want.num)
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
			root := sectionHeaderText(t)
			m := root.Doc
			m.PlainMode = tt.fields.PlainMode
			m.ColumnMode = tt.fields.ColumnMode
			m.ColumnWidth = tt.fields.ColumnWidth
			m.setMultiColorWords(tt.fields.multiColorWords)
			m.setDelimiter(tt.fields.ColumnDelimiter)
			m.setColumnWidths()
			root.scr.contents = make(map[int]LineC)
			root.prepareDraw(context.Background())
			line := m.getLineC(tt.args.lineNum, m.TabWidth)
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
			root := sectionHeaderText(t)

			line := root.Doc.getLineC(tt.args.lineNum, root.Doc.TabWidth)
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
			root, err := Open(filepath.Join(testdata, "column.txt"))
			if err != nil {
				t.Fatal(err)
			}
			m := root.Doc
			for !m.BufEOF() {
			}
			m.ColumnDelimiter = tt.fields.columnDelimiter
			m.ColumnDelimiterReg = condRegexpCompile(m.ColumnDelimiter)
			m.columnCursor = tt.fields.columnCursor
			root.StyleColumnHighlight = OVStyle{Bold: true}
			line := root.Doc.getLineC(tt.args.lineNum, root.Doc.TabWidth)
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
			root, err := Open(filepath.Join(testdata, "ps.txt"))
			if err != nil {
				t.Fatal(err)
			}
			root.StyleColumnHighlight = OVStyle{Bold: true}
			m := root.Doc
			m.ColumnWidth = true
			m.ColumnMode = true
			for !m.BufEOF() {
			}
			m.setColumnWidths()
			m.columnCursor = tt.fields.columnCursor
			line := root.Doc.getLineC(tt.args.lineNum, root.Doc.TabWidth)
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
