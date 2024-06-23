package oviewer

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func getContents(t *testing.T, root *Root, Y int, width int) string {
	t.Helper()
	var buf strings.Builder
	for x := 0; x < width; x++ {
		r, _, _, _ := root.Screen.GetContent(x, Y)
		buf.WriteRune(r)
	}
	return buf.String()
}

func TestRoot_draw(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames   []string
		header      int
		wrapMode    bool
		lineNumMode bool
	}
	type args struct {
		ctx context.Context
	}
	type want struct {
		bottomLN int
		bottomLX int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "testDraw1",
			fields: fields{
				fileNames:   []string{filepath.Join(testdata, "test.txt")},
				header:      0,
				wrapMode:    false,
				lineNumMode: true,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN: 24,
				bottomLX: 0,
			},
		},
		{
			name: "testDraw2",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
				header:    0,
				wrapMode:  true,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN: 24,
				bottomLX: 0,
			},
		},
		{
			name: "testDraw-header",
			fields: fields{
				fileNames:   []string{filepath.Join(testdata, "test.txt")},
				header:      1,
				wrapMode:    false,
				lineNumMode: true,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN: 24,
				bottomLX: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileNames...)
			root.Doc.Header = tt.fields.header
			root.Doc.WrapMode = tt.fields.wrapMode
			root.Doc.LineNumMode = tt.fields.lineNumMode
			root.ViewSync(tt.args.ctx)
			root.prepareScreen()
			root.draw(tt.args.ctx)
			if root.Doc.bottomLN != tt.want.bottomLN {
				t.Errorf("Root.draw() bottomLN = %v, want %v", root.Doc.bottomLN, tt.want.bottomLN)
			}
			if root.Doc.bottomLX != tt.want.bottomLX {
				t.Errorf("Root.draw() bottomLX = %v, want %v", root.Doc.bottomLX, tt.want.bottomLX)
			}
		})
	}
}

func TestRoot_draw2(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames        []string
		header           int
		sectionDelimiter string
		hideOtherSection bool
		wrapMode         bool
		lineNumMode      bool
		topLN            int
	}
	type args struct {
		ctx context.Context
	}
	type want struct {
		bottomLN            int
		bottomLX            int
		sectionHeaderHeight int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "testDraw-section-header",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section-header.txt")},
				header:           3,
				sectionDelimiter: "^#",
				hideOtherSection: false,
				wrapMode:         false,
				lineNumMode:      true,
				topLN:            0,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN:            24,
				bottomLX:            0,
				sectionHeaderHeight: 1,
			},
		},
		{
			name: "testDraw-section-header2",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section-header.txt")},
				header:           3,
				sectionDelimiter: "^#",
				hideOtherSection: false,
				wrapMode:         true,
				lineNumMode:      true,
				topLN:            0,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN:            22,
				bottomLX:            0,
				sectionHeaderHeight: 1,
			},
		},
		{
			name: "testDraw-section-header2",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section-header.txt")},
				header:           3,
				sectionDelimiter: "^#",
				hideOtherSection: false,
				wrapMode:         true,
				lineNumMode:      true,
				topLN:            10,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN:            34,
				bottomLX:            0,
				sectionHeaderHeight: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileNames...)
			root.Doc.Header = tt.fields.header
			if tt.fields.sectionDelimiter != "" {
				root.Doc.setSectionDelimiter(tt.fields.sectionDelimiter)
				root.Doc.SectionHeaderNum = 1
				root.Doc.SectionHeader = true
				root.Doc.HideOtherSection = tt.fields.hideOtherSection
			}
			root.Doc.topLN = tt.fields.topLN
			root.Doc.WrapMode = tt.fields.wrapMode
			root.Doc.LineNumMode = tt.fields.lineNumMode
			root.ViewSync(tt.args.ctx)
			root.prepareScreen()
			root.draw(tt.args.ctx)
			if root.Doc.bottomLN != tt.want.bottomLN {
				t.Errorf("Root.draw() bottomLN = %v, want %v", root.Doc.bottomLN, tt.want.bottomLN)
			}
			if root.Doc.bottomLX != tt.want.bottomLX {
				t.Errorf("Root.draw() bottomLX = %v, want %v", root.Doc.bottomLX, tt.want.bottomLX)
			}
			if got := root.Doc.sectionHeaderHeight; got != tt.want.sectionHeaderHeight {
				t.Errorf("Root.draw() sectionHeaderHeight = %v, want %v", got, tt.want.sectionHeaderHeight)
			}
		})
	}
}

func TestRoot_section2(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames        []string
		header           int
		sectionDelimiter string
		sectionHeaderNum int
		hideOtherSection bool
		wrapMode         bool
		lineNumMode      bool
		topLN            int
	}
	type args struct {
		ctx context.Context
	}
	type want struct {
		bottomLN            int
		bottomLX            int
		sectionHeaderHeight int
		wantY               int
		wantStr             string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "testDraw-section2-1",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section2.txt")},
				header:           0,
				sectionDelimiter: "^-",
				sectionHeaderNum: 3,
				hideOtherSection: false,
				wrapMode:         true,
				lineNumMode:      true,
				topLN:            0,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN:            24,
				bottomLX:            0,
				sectionHeaderHeight: 0,
				wantY:               0,
				wantStr:             " 1 test1  ",
			},
		},
		{
			name: "testDraw-section2-2",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section2.txt")},
				header:           0,
				sectionDelimiter: "^-",
				sectionHeaderNum: 1,
				hideOtherSection: false,
				wrapMode:         true,
				lineNumMode:      true,
				topLN:            4,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN:            28,
				bottomLX:            0,
				sectionHeaderHeight: 1,
				wantY:               8,
				wantStr:             "13        ",
			},
		},
		{
			name: "testDraw-section2-3",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section2.txt")},
				header:           0,
				sectionDelimiter: "^-",
				sectionHeaderNum: 1,
				hideOtherSection: true,
				wrapMode:         true,
				lineNumMode:      true,
				topLN:            3,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN:            27,
				bottomLX:            0,
				sectionHeaderHeight: 1,
				wantY:               10,
				wantStr:             "          ",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileNames...)
			root.Doc.Header = tt.fields.header
			if tt.fields.sectionDelimiter != "" {
				root.Doc.setSectionDelimiter(tt.fields.sectionDelimiter)
				root.Doc.SectionHeaderNum = tt.fields.sectionHeaderNum
				root.Doc.SectionHeader = true
				root.Doc.HideOtherSection = tt.fields.hideOtherSection
			}
			root.Doc.topLN = tt.fields.topLN
			root.Doc.WrapMode = tt.fields.wrapMode
			root.Doc.LineNumMode = tt.fields.lineNumMode
			root.Doc.HideOtherSection = tt.fields.hideOtherSection
			root.ViewSync(tt.args.ctx)
			root.prepareScreen()
			root.draw(tt.args.ctx)
			if root.Doc.bottomLN != tt.want.bottomLN {
				t.Errorf("Root.draw() bottomLN = %v, want %v", root.Doc.bottomLN, tt.want.bottomLN)
			}
			if root.Doc.bottomLX != tt.want.bottomLX {
				t.Errorf("Root.draw() bottomLX = %v, want %v", root.Doc.bottomLX, tt.want.bottomLX)
			}
			if got := root.Doc.sectionHeaderHeight; got != tt.want.sectionHeaderHeight {
				t.Errorf("Root.draw() sectionHeaderHeight = %v, want %v", got, tt.want.sectionHeaderHeight)
			}
			str := getContents(t, root, tt.want.wantY, 10)
			if got := str; got != tt.want.wantStr {
				t.Errorf("Root.draw() = [%v], want [%v]", got, tt.want.wantStr)
			}
		})
	}
}
