package oviewer

import (
	"bytes"
	"reflect"
	"testing"
)

func TestOpenDocument(t *testing.T) {
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
			args:    args{fileName: "../testdata/nofile"},
			wantErr: true,
		},
		{
			name:    "normal.txt",
			args:    args{fileName: "../testdata/normal.txt"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		want    lineContents
		wantErr bool
	}{
		{
			name: "testEmpty",
			str:  ``,
			args: args{
				lN:       0,
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
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			if err := m.ReadAll(bytes.NewBufferString(tt.str)); err != nil {
				t.Fatal(err)
			}
			<-m.eofCh
			t.Logf("num:%d", m.BufEndNum())
			got, err := m.contentsLN(tt.args.lN, tt.args.tabWidth)
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
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			if err := m.ReadAll(bytes.NewBufferString(tt.str)); err != nil {
				t.Fatal(err)
			}
			w := &bytes.Buffer{}
			<-m.eofCh
			m.bottomLN = m.BufEndNum()
			m.Export(w, tt.args.start, tt.args.end)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Document.Export() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestDocument_getContents(t *testing.T) {
	type fields struct {
		FileName string
	}
	type args struct {
		lN       int
		tabWidth int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   lineContents
	}{
		{
			name: "test normal",
			fields: fields{
				FileName: "../testdata/normal.txt",
			},
			args: args{
				lN:       0,
				tabWidth: 8,
			},
			want: parseString("khaki	mediumseagreen	steelblue	forestgreen	royalblue	mediumseagreen", 8),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := OpenDocument(tt.fields.FileName)
			if err != nil {
				t.Fatalf("OpenDocument %s", err)
			}

			for m.eof != 1 {
			}
			if got := m.getContents(tt.args.lN, tt.args.tabWidth); !reflect.DeepEqual(got, tt.want) {
				g, _ := ContentsToStr(got)
				w, _ := ContentsToStr(tt.want)
				t.Errorf("Document.getContents() = [%v], want [%v]", g, w)
			}
		})
	}
}
