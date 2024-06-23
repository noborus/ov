package oviewer

import (
	"bytes"
	"context"
	"path/filepath"
	"reflect"
	"testing"
)

func docHelper(t *testing.T, str string) *Document {
	t.Helper()
	m, err := NewDocument()
	if err != nil {
		t.Fatal(err)
	}
	if err := m.ReadAll(bytes.NewBufferString(str)); err != nil {
		t.Fatal(err)
	}
	for !m.BufEOF() {
	}
	return m
}

func docFileReadHelper(t *testing.T, fileName string) *Document {
	t.Helper()
	m, err := OpenDocument(fileName)
	if err != nil {
		t.Fatal(err)
	}
	for !m.BufEOF() {
	}

	return m
}

func TestOpenDocument(t *testing.T) {
	t.Parallel()
	type args struct {
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		want    *Document
		wantErr bool
	}{
		{
			name:    "no file",
			args:    args{fileName: filepath.Join(testdata, "nofile")},
			wantErr: true,
		},
		{
			name:    "normal.txt",
			args:    args{fileName: filepath.Join(testdata, "normal.txt")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := OpenDocument(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDocument_lineToContents(t *testing.T) {
	t.Parallel()
	type args struct {
		lN       int
		tabWidth int
	}
	tests := []struct {
		name    string
		str     string
		args    args
		want    contents
		wantErr bool
	}{
		{
			name: "testEmpty",
			str:  ``,
			args: args{
				lN:       1,
				tabWidth: 4,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test1",
			str:  "test\n",
			args: args{
				lN:       0,
				tabWidth: 4,
			},
			want:    StrToContents("test", 4),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := docHelper(t, tt.str)
			t.Logf("num:%d", m.BufEndNum())
			got, err := m.contents(tt.args.lN, tt.args.tabWidth)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.lineToContents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Document.lineToContents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_Export(t *testing.T) {
	t.Parallel()
	type args struct {
		start int
		end   int
	}
	tests := []struct {
		name  string
		str   string
		args  args
		wantW string
	}{
		{
			name: "test1",
			str:  "a\nb\nc\nd\ne\n\f\n",
			args: args{
				start: 0,
				end:   2,
			},
			wantW: "a\nb\nc\n",
		},
		{
			name: "test2",
			str:  "a\nb\nc\nd\ne\n\f\n",
			args: args{
				start: 1,
				end:   2,
			},
			wantW: "b\nc\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := docHelper(t, tt.str)
			w := &bytes.Buffer{}
			for !m.BufEOF() {
			}
			m.bottomLN = m.BufEndNum()
			if err := m.Export(w, tt.args.start, tt.args.end); err != nil {
				t.Fatal(err)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Document.Export() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestDocument_searchLine(t *testing.T) {
	t.Parallel()
	type fields struct {
		FileName string
	}
	type args struct {
		ctx      context.Context
		searcher Searcher
		num      int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "test forward search",
			fields: fields{
				FileName: filepath.Join(testdata, "normal.txt"),
			},
			args: args{
				ctx:      context.Background(),
				searcher: NewSearcher("a", nil, false, false),
				num:      0,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "test forward search not found",
			fields: fields{
				FileName: filepath.Join(testdata, "normal.txt"),
			},
			args: args{
				ctx:      context.Background(),
				searcher: NewSearcher("notfoundnotfound", nil, false, false),
				num:      0,
			},
			want:    40,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := docFileReadHelper(t, tt.fields.FileName)
			got, err := m.SearchLine(tt.args.ctx, tt.args.searcher, tt.args.num)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.searchLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.searchLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_backSearchLine(t *testing.T) {
	t.Parallel()
	type fields struct {
		FileName string
	}
	type args struct {
		ctx      context.Context
		searcher Searcher
		num      int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "test backward search",
			fields: fields{
				FileName: filepath.Join(testdata, "normal.txt"),
			},
			args: args{
				ctx:      context.Background(),
				searcher: NewSearcher("a", nil, false, false),
				num:      100,
			},
			want:    39,
			wantErr: false,
		},
		{
			name: "test backward search not found",
			fields: fields{
				FileName: filepath.Join(testdata, "normal.txt"),
			},
			args: args{
				ctx:      context.Background(),
				searcher: NewSearcher("notfoundnotfound", nil, false, false),
				num:      100,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := docFileReadHelper(t, tt.fields.FileName)
			got, err := m.BackSearchLine(tt.args.ctx, tt.args.searcher, tt.args.num)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.backSearchLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.backSearchLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
