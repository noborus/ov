package oviewer

import (
	"path/filepath"
	"testing"
)

func TestDocument_ControlFile(t *testing.T) {
	type fields struct {
		FileName string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "testNoFile",
			fields:  fields{FileName: ""},
			wantErr: false,
		},
		{
			name: "testTestFile",
			fields: fields{
				FileName: filepath.Join(testdata, "test.txt"),
			},
			wantErr: false,
		},
		{
			name: "testLargeFile",
			fields: fields{
				FileName: filepath.Join(testdata, "large.txt"),
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
			f, err := open(tt.fields.FileName)
			if err != nil {
				t.Fatal("open error", tt.fields.FileName)
			}
			if err := m.ControlFile(f); (err != nil) != tt.wantErr {
				t.Errorf("Document.ControlFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDocument_requestSearch(t *testing.T) {
	type fields struct {
		FileName   string
		searchWord string
		chunkNum   int
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{
			name: "testLargeFileFalse",
			fields: fields{
				FileName:   filepath.Join(testdata, "large.txt"),
				chunkNum:   1,
				searchWord: "999999",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "testLargeFileTrue",
			fields: fields{
				FileName:   filepath.Join(testdata, "large.txt"),
				chunkNum:   99,
				searchWord: "999999",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			searcher := NewSearcher(tt.fields.searchWord, nil, false, false)
			f, err := open(tt.fields.FileName)
			if err != nil {
				t.Fatal("open error", tt.fields.FileName)
			}
			if err := m.ControlFile(f); (err != nil) != tt.wantErr {
				t.Errorf("Document.ControlFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			for !m.BufEOF() {
			}
			if got := m.requestSearch(tt.fields.chunkNum, searcher); got != tt.want {
				t.Errorf("Document.requestSearch() = %v, want %v", got, tt.want)
			}
		})
	}
}
