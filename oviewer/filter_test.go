package oviewer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestRoot_Filter(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames []string
	}
	type args struct {
		str      string
		nonMatch bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Filter",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
			},
			args: args{
				str:      "test",
				nonMatch: false,
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.fields.fileNames...)
			if err != nil {
				t.Fatalf("NewOviewer error = %v", err)
			}
			for !root.Doc.BufEOF() {
			}
			root.Filter(tt.args.str, tt.args.nonMatch)
		})
	}
}

func TestRoot_filter(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames []string
	}
	type args struct {
		str      string
		nonMatch bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "filter-nil",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
			},
			args: args{
				str:      "test",
				nonMatch: false,
			},
			want: "test",
		},
		{
			name: "filter1",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
			},
			args: args{
				str:      "test",
				nonMatch: false,
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.fields.fileNames...)
			if err != nil {
				t.Fatalf("NewOviewer error = %v", err)
			}
			for !root.Doc.BufEOF() {
			}
			root.input.value = tt.args.str
			root.filter(context.Background())
		})
	}
}

func TestRoot_filterDocument(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames []string
		header    int
	}
	type args struct {
		searcher Searcher
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "test.txt",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
				header:    0,
			},
			args: args{
				searcher: NewSearcher("test", nil, false, false),
			},
			want: "test",
		},
		{
			name: "test3.txt",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test3.txt")},
				header:    1,
			},
			args: args{
				searcher: NewSearcher("2", nil, false, false),
			},
			want: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.fields.fileNames...)
			if err != nil {
				t.Fatalf("NewOviewer error = %v", err)
			}
			for !root.Doc.BufEOF() {
			}
			root.Doc.Header = tt.fields.header
			root.filterDocument(context.Background(), tt.args.searcher)
			filterDoc := root.DocList[len(root.DocList)-1]
			for !filterDoc.BufEOF() {
			}
			line := filterDoc.getLineC(0, 0)
			if string(line.str) != tt.want {
				t.Errorf("filterDocument() = %v, want %v", string(line.str), tt.want)
			}
		})
	}
}

func TestRoot_closeAllFilter(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames []string
	}
	type args struct {
		searcher Searcher
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "test.txt",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
			},
			args: args{
				searcher: NewSearcher("test", nil, false, false),
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.fields.fileNames...)
			if err != nil {
				t.Fatalf("NewOviewer error = %v", err)
			}
			for !root.Doc.BufEOF() {
			}
			root.filterDocument(context.Background(), tt.args.searcher)
			filterDoc := root.DocList[len(root.DocList)-1]
			for !filterDoc.BufEOF() {
			}
			ctx := context.Background()
			root.closeAllFilter(ctx)
			if len(root.DocList) != tt.want {
				t.Errorf("closeAllFilter() = %v, want %v", len(root.DocList), tt.want)
			}
		})
	}
}
