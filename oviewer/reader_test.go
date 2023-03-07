package oviewer

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestDocument_ReadFile(t *testing.T) {
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
				"nofile",
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
		t.Run(tt.name, func(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
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
				FileName: "../testdata/normal.txt",
			},
			wantNum: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				FileName:  "../testdata/normal.txt",
				WatchMode: false,
				seekable:  true,
			},
			wantErr: false,
		},
		{
			name: "testWatchReload",
			fields: fields{
				FileName:  "../testdata/normal.txt",
				WatchMode: true,
				seekable:  true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal("NewDocument error")
			}
			m.FileName = tt.fields.FileName
			m.WatchMode = tt.fields.WatchMode
			m.seekable = tt.fields.seekable
			if err := m.ControlFile(tt.fields.FileName); err != nil {
				t.Fatal("ControlFile error")
			}
			if err := m.reload(); (err != nil) != tt.wantErr {
				t.Errorf("Document.reload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
