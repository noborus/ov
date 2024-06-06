package oviewer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

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
			},
			want: want{
				headerHeight:        5,
				sectionHeaderHeight: 0,
				topLN:               4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(filepath.Join(testdata, "section-header.txt"))
			if err != nil {
				t.Fatal(err)
			}
			m := root.Doc
			m.width = 80
			root.scr.vHeight = 24
			m.topLX = 0
			m.SkipLines = tt.fields.skipLines
			m.Header = tt.fields.header
			m.WrapMode = tt.fields.wrapMode
			m.SectionHeader = tt.fields.sectionHeader
			m.setSectionDelimiter(tt.fields.sectionDelimiter)
			m.SectionHeaderNum = tt.fields.sectionHeaderNum
			m.showGotoF = tt.fields.showGotoF
			m.topLN = tt.fields.topLN
			ctx := context.Background()
			for !m.BufEOF() {
			}
			root.scr.contents = make(map[int]LineC)
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
