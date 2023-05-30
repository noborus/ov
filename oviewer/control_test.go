package oviewer

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDocument_ControlFile(t *testing.T) {
	type fields struct {
		FileName   string
		FollowMode bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "testNoFile",
			fields: fields{
				FileName:   "noFile",
				FollowMode: false,
			},
			wantErr: false,
		},
		{
			name: "testTestFile",
			fields: fields{
				FileName:   filepath.Join(testdata, "test.txt"),
				FollowMode: false,
			},
			wantErr: false,
		},
		{
			name: "testLargeFile",
			fields: fields{
				FileName:   filepath.Join(testdata, "large.txt"),
				FollowMode: false,
			},
			wantErr: false,
		},
		{
			name: "testLargeFileFollow",
			fields: fields{
				FileName:   filepath.Join(testdata, "large.txt"),
				FollowMode: true,
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
			m.FollowMode = tt.fields.FollowMode
			f, err := open(tt.fields.FileName)
			if err != nil {
				// Ignore open errors and test *os.file for invalid values.
				f = nil
			}
			if err := m.ControlFile(f); (err != nil) != tt.wantErr {
				t.Errorf("Document.ControlFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDocument_requestLoad(t *testing.T) {
	type fields struct {
		FileName string
		chunkNum int
		lineNum  int
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "testLargeFile1",
			fields: fields{
				FileName: filepath.Join(testdata, "large.txt"),
				chunkNum: 1,
				lineNum:  10001,
			},
			want:    []byte("10002\n"),
			wantErr: false,
		},
		{
			name: "testLargeFile2",
			fields: fields{
				FileName: filepath.Join(testdata, "large.txt"),
				chunkNum: 99,
				lineNum:  999999,
			},
			want:    []byte("1000000\n"),
			wantErr: false,
		},
		{
			name: "testLargeFileFail",
			fields: fields{
				FileName: filepath.Join(testdata, "large.txt"),
				chunkNum: 1,
				lineNum:  888888,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "testLargeFileOverFail",
			fields: fields{
				FileName: filepath.Join(testdata, "large.txt"),
				chunkNum: 1,
				lineNum:  1000001,
			},
			want:    nil,
			wantErr: true,
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

			if err := m.ControlFile(f); err != nil {
				t.Fatalf("Document.ControlFile() fatal = %v, wantErr %v", err, tt.wantErr)
			}
			for !m.BufEOF() {
			}
			m.requestLoad(tt.fields.chunkNum)

			chunkNum, cn := chunkLineNum(tt.fields.lineNum)
			got, err := m.store.GetChunkLine(chunkNum, cn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.ControlFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Document.requestLoad() = %v, want %v", got, tt.want)
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

func CopyToTempFile(t *testing.T, fileName string) string {
	t.Helper()
	tmpdir := t.TempDir()
	fname := filepath.Join(tmpdir, filepath.Base(fileName))
	orig, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	defer orig.Close()

	dst, err := os.Create(fname)
	if err != nil {
		t.Fatal(err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, orig); err != nil {
		t.Fatal(err)
	}
	return fname
}

func TestDocument_requestFollow(t *testing.T) {
	type fields struct {
		FileName    string
		appendBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "testLargeFile1",
			fields: fields{
				FileName:    filepath.Join(testdata, "large.txt"),
				appendBytes: []byte("a\n"),
			},
			want:    []byte("a\n"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fname := CopyToTempFile(t, tt.fields.FileName)
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			f, err := open(fname)
			if err != nil {
				t.Fatal("open error", tt.fields.FileName)
			}
			defer f.Close()
			m.FollowMode = true
			if err := m.ControlFile(f); err != nil {
				t.Fatalf("Document.ControlFile() fatal = %v, wantErr %v", err, tt.wantErr)
			}

			for !m.BufEOF() {
			}

			af, err := os.OpenFile(fname, os.O_APPEND|os.O_RDWR, 0600)
			if err != nil {
				t.Fatal("open error", tt.fields.FileName)
			}
			defer af.Close()
			af.Write(tt.fields.appendBytes)

			done := make(chan bool)
			m.ctlCh <- controlSpecifier{
				request: requestFollow,
				done:    done,
			}
			<-done

			chunkNum, cn := chunkLineNum(m.BufEndNum() - 1)
			got, err := m.store.GetChunkLine(chunkNum, cn)

			if (err != nil) != tt.wantErr {
				t.Errorf("Document.ControlFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Document.requestLoad() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
