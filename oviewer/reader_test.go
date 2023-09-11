package oviewer

import (
	"bufio"
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDocument_ReadFile(t *testing.T) {
	t.Parallel()
	type args struct {
		args string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "testNoFile",
			args: args{
				"noFile",
			},
			wantErr: true,
		},
		{
			name: "testSTDIN",
			args: args{
				"",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			if err := m.ReadFile(tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("Document.ReadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDocument_ReadAll(t *testing.T) {
	t.Parallel()
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				r: bytes.NewBufferString("foo\nbar\n"),
			},
			wantErr: false,
		},
		{
			name: "testOverChunkSize",
			args: args{
				r: bytes.NewBufferString(strings.Repeat("a\n", ChunkSize+1)),
			},
			wantErr: false,
		},
		{
			name: "testLongLine",
			args: args{
				r: bytes.NewBufferString(strings.Repeat("a", 4097)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			if err := m.ReadAll(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("Document.ReadAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDocument_reset(t *testing.T) {
	t.Parallel()
	type fields struct {
		FileName string
	}
	tests := []struct {
		name    string
		fields  fields
		wantNum int
	}{
		{
			name: "testReset",
			fields: fields{
				FileName: filepath.Join(testdata, "normal.txt"),
			},
			wantNum: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := OpenDocument(tt.fields.FileName)
			if err != nil {
				t.Fatalf("OpenDocument %s", err)
			}
			for !m.BufEOF() {
			}
			m.reset()
			if m.BufEndNum() != tt.wantNum {
				t.Errorf("Document.reset() %v != %v", m.BufEndNum(), tt.wantNum)
			}
		})
	}
}

func TestDocument_reload(t *testing.T) {
	t.Parallel()
	type fields struct {
		FileName  string
		WatchMode bool
		seekable  bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "testReload",
			fields: fields{
				FileName:  filepath.Join(testdata, "normal.txt"),
				WatchMode: false,
				seekable:  true,
			},
			wantErr: false,
		},
		{
			name: "testWatchReload",
			fields: fields{
				FileName:  filepath.Join(testdata, "normal.txt"),
				WatchMode: true,
				seekable:  true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := NewDocument()
			if err != nil {
				t.Fatal("NewDocument error")
			}
			m.FileName = tt.fields.FileName
			m.WatchMode = tt.fields.WatchMode
			m.seekable = tt.fields.seekable
			f, err := open(tt.fields.FileName)
			if err != nil {
				t.Fatal("open error", tt.fields.FileName)
			}
			if err := m.ControlFile(f); err != nil {
				t.Fatal("ControlFile error", tt.fields.FileName)
			}
			if err := m.reload(); (err != nil) != tt.wantErr {
				t.Errorf("Document.reload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDocument_loadReadMem(t *testing.T) {
	type fields struct {
		seekable bool
		eof      int32
		chunks   int
	}
	type args struct {
		reader   *bufio.Reader
		chunkNum int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "testLoadReadMem1",
			fields: fields{
				chunks:   0,
				seekable: false,
				eof:      0,
			},
			args: args{
				reader:   bufio.NewReader(bytes.NewBufferString("foo\nbar\n")),
				chunkNum: 0,
			},
			wantErr: false,
		},
		{
			name: "testLoadReadMem1",
			fields: fields{
				chunks:   3,
				seekable: false,
				eof:      0,
			},
			args: args{
				reader:   bufio.NewReader(bytes.NewBufferString("foo\nbar\n")),
				chunkNum: 2,
			},
			wantErr: false,
		},
		{
			name: "testLoadReadMemEOF",
			fields: fields{
				chunks:   0,
				seekable: false,
				eof:      1,
			},
			args: args{
				reader:   bufio.NewReader(bytes.NewBufferString("foo\nbar\n")),
				chunkNum: 0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			m.seekable = tt.fields.seekable
			for i := 0; i < tt.fields.chunks; i++ {
				m.store.chunks = append(m.store.chunks, &chunk{})
			}
			atomic.StoreInt32(&m.store.eof, tt.fields.eof)
			if _, err := m.loadReadMem(tt.args.reader, tt.args.chunkNum); (err != nil) != tt.wantErr {
				t.Errorf("Document.loadReadMem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
