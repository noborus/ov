package oviewer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestRoot_Filter(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName string
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
				fileName: filepath.Join(testdata, "test.txt"),
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
			root := rootFileReadHelper(t, tt.fields.fileName)
			root.Filter(tt.args.str, tt.args.nonMatch)
		})
	}
}

func TestRoot_filter2(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileNames []string
		nonMatch  bool
	}
	type args struct {
		str string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "filter-nil",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
				nonMatch:  false,
			},
			args: args{
				str: "",
			},
			want: 1,
		},
		{
			name: "filter1",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
				nonMatch:  false,
			},
			args: args{
				str: "test",
			},
			want: 2,
		},
		{
			name: "filter non-match",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
				nonMatch:  true,
			},
			args: args{
				str: "non",
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileNames...)
			root.Doc.jumpTargetSection = true
			root.Doc.nonMatch = tt.fields.nonMatch
			root.filter(context.Background(), tt.args.str)
			if root.DocumentLen() != tt.want {
				t.Errorf("filter() = %v, want %v", root.DocumentLen(), tt.want)
			}
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
		skipLines int
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
			name: "test3.txtMatch",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test3.txt")},
				skipLines: 0,
				header:    0,
			},
			args: args{
				searcher: NewSearcher("3", nil, false, false),
			},
			want: "3",
		},
		{
			name: "test3.txtHeader",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test3.txt")},
				skipLines: 0,
				header:    1,
			},
			args: args{
				searcher: NewSearcher("2", nil, false, false),
			},
			want: "1",
		},
		{
			name: "test3.txtSkipLines",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test3.txt")},
				skipLines: 1,
				header:    1,
			},
			args: args{
				searcher: NewSearcher("4", nil, false, false),
			},
			want: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileNames...)
			root.Doc.SkipLines = tt.fields.skipLines
			root.Doc.Header = tt.fields.header
			root.filterDocument(context.Background(), tt.args.searcher)
			filterDoc := root.DocList[len(root.DocList)-1]
			filterDoc.cond.L.Lock()
			filterDoc.cond.Wait()
			filterDoc.cond.L.Unlock()
			line := filterDoc.getLineC(0)
			if line.str != tt.want {
				t.Errorf("filterDocument() = %v, want %v", line.str, tt.want)
			}
		})
	}
}

func TestRoot_filterCancel(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root := rootFileReadHelper(t, filepath.Join(testdata, "test3.txt"))
	searcher := NewSearcher("test", nil, false, false)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	root.filterDocument(ctx, searcher)
	if root.DocumentLen() != 2 {
		t.Errorf("filterDocument() = %v, want %v", root.DocumentLen(), 2)
	}
	filterDoc := root.getDocument(1)
	if filterDoc.BufEndNum() != 0 {
		t.Errorf("filterDocument() = %v, want %v", filterDoc.BufEndNum(), 0)
	}
}

func TestRoot_filterLink(t *testing.T) {
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
		want   int
	}{
		{
			name: "test3.txt",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test3.txt")},
				header:    0,
			},
			args: args{
				searcher: NewSearcher("1000", nil, false, false),
			},
			want: 10000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileNames...)
			root.Doc.Header = tt.fields.header
			root.input.value = "1000"
			ctx := context.Background()
			root.filterDocument(ctx, tt.args.searcher)
			filterDoc := root.DocList[len(root.DocList)-1]
			filterDoc.WaitEOF()
			filterDoc.LineNumMode = true
			root.prepareStartX()
			filterDoc.topLN = 2
			root.previousDoc(ctx)
			t.Log(root.Doc.topLN)
			if root.Doc.topLN != tt.want {
				t.Errorf("filterDocument() = %v, want %v", root.Doc.topLN, tt.want)
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
		{
			name: "test.txt-2",
			fields: fields{
				fileNames: []string{filepath.Join(testdata, "test.txt"), filepath.Join(testdata, "test2.txt")},
			},
			args: args{
				searcher: NewSearcher("test", nil, false, false),
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileNames...)
			root.filterDocument(context.Background(), tt.args.searcher)
			filterDoc := root.DocList[len(root.DocList)-1]
			filterDoc.WaitEOF()
			ctx := context.Background()
			root.closeAllFilter(ctx)
			if len(root.DocList) != tt.want {
				t.Errorf("closeAllFilter() = %v, want %v", len(root.DocList), tt.want)
			}
			// twice makes no change.
			root.closeAllFilter(ctx)
			if len(root.DocList) != tt.want {
				t.Errorf("closeAllFilter() = %v, want %v", len(root.DocList), tt.want)
			}
		})
	}
}
