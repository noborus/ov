package oviewer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestRoot_draw(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames        []string
		header           int
		sectionDelimiter string
		wrapMode         bool
		lineNumMode      bool
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
			name: "testDraw1",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "test.txt")},
				header:           0,
				sectionDelimiter: "",
				wrapMode:         false,
				lineNumMode:      true,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN:            24,
				bottomLX:            0,
				sectionHeaderHeight: 0,
			},
		},
		{
			name: "testDraw2",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "test.txt")},
				header:           0,
				sectionDelimiter: "",
				wrapMode:         true,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN:            24,
				bottomLX:            0,
				sectionHeaderHeight: 0,
			},
		},
		{
			name: "testDraw-header",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "test.txt")},
				header:           1,
				sectionDelimiter: "",
				wrapMode:         false,
				lineNumMode:      true,
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				bottomLN:            24,
				bottomLX:            0,
				sectionHeaderHeight: 0,
			},
		},
		{
			name: "testDraw-section-header",
			fields: fields{
				fileNames:        []string{filepath.Join(testdata, "section-header.txt")},
				header:           3,
				sectionDelimiter: "^#",
				wrapMode:         false,
				lineNumMode:      true,
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
				wrapMode:         true,
				lineNumMode:      true,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.fields.fileNames...)
			if err != nil {
				t.Fatal(err)
			}

			root.Doc.Header = tt.fields.header
			if tt.fields.sectionDelimiter != "" {
				root.Doc.setSectionDelimiter(tt.fields.sectionDelimiter)
				root.Doc.SectionHeaderNum = 1
				root.Doc.SectionHeader = true
			}
			// Wait for loading the file.
			for !root.Doc.BufEOF() {
			}
			root.Doc.WrapMode = tt.fields.wrapMode
			root.Doc.LineNumMode = tt.fields.lineNumMode
			root.ViewSync(tt.args.ctx)
			root.prepareScreen()
			root.draw(tt.args.ctx)
			if root.Doc.bottomLN != tt.want.bottomLN {
				t.Logf("wdith %v heigh %v", root.scr.vWidth, root.scr.vHeight)
				t.Errorf("Root.draw() bottomLN = %v, want %v", root.Doc.bottomLN, tt.want.bottomLN)
			}
			if root.Doc.bottomLX != tt.want.bottomLX {
				t.Errorf("Root.draw() bottomLX = %v, want %v", root.Doc.bottomLX, tt.want.bottomLX)
			}
			if got := root.Doc.sectionHeaderHeight; got != tt.want.sectionHeaderHeight {
				t.Errorf("Root.draw() sectionHeaderHeight = %v, want %v", got, tt.want)
			}
		})
	}
}
