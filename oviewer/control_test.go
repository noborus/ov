package oviewer

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
		{
			name: "testLargeFileMem",
			fields: fields{
				FileName:   filepath.Join(testdata, "large.txt.zst"),
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
			name: "testLargeFile3",
			fields: fields{
				FileName: filepath.Join(testdata, "test3.txt"),
				chunkNum: 1,
				lineNum:  12344,
			},
			want:    []byte("12345\n"),
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
				t.Errorf("Document.requestLoad() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestDocument_requestLoadMem(t *testing.T) {
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
			name: "testLargeFileZstd",
			fields: fields{
				FileName: filepath.Join(testdata, "large.txt.zst"),
				chunkNum: 99,
				lineNum:  999999,
			},
			want:    []byte("1000000\n"),
			wantErr: false,
		},
		{
			name: "testLargeFileZstd2",
			fields: fields{
				FileName: filepath.Join(testdata, "large.txt.zst"),
				chunkNum: 10,
				lineNum:  10001,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MemoryLimit = 2
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			f, err := open(tt.fields.FileName)
			if err != nil {
				t.Fatal("open error", tt.fields.FileName)
			}
			m.FollowMode = true
			if err := m.ControlFile(f); err != nil {
				t.Fatalf("Document.ControlFile() fatal = %v, wantErr %v", err, tt.wantErr)
			}
			for line := 0; !m.BufEOF(); line++ {
				m.requestLoad(line)
			}

			chunkNum, cn := chunkLineNum(tt.fields.lineNum)
			got, err := m.store.GetChunkLine(chunkNum, cn)
			if (err != nil) != tt.wantErr {
				t.Log(MemoryLimit)
				t.Errorf("Document.ControlFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Document.requestLoad() = %v, want %v", string(got), string(tt.want))
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
			if _, err := af.Write(tt.fields.appendBytes); err != nil {
				t.Fatal("open error", tt.fields.FileName)
			}

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

func TestDocument_requestClose(t *testing.T) {
	type fields struct {
		FileName string
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{
			name: "testLargeClose",
			fields: fields{
				FileName: filepath.Join(testdata, "large.txt"),
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
			f, err := open(tt.fields.FileName)
			if err != nil {
				t.Fatal("open error", tt.fields.FileName)
			}
			if err := m.ControlFile(f); (err != nil) != tt.wantErr {
				t.Errorf("Document.ControlFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			for !m.BufEOF() {
			}
			if got := m.requestClose(); got != tt.want {
				t.Errorf("Document.requestSearch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_ControlReader(t *testing.T) {
	type args struct {
		r      io.Reader
		reload func() *bufio.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				r:      bufio.NewReader(strings.NewReader("test")),
				reload: nil,
			},
			wantErr: false,
		},
		{
			name: "testLarge",
			args: args{
				r:      bufio.NewReader(strings.NewReader(strings.Repeat("test\n", 10000))),
				reload: nil,
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
			if err := m.ControlReader(tt.args.r, tt.args.reload); (err != nil) != tt.wantErr {
				t.Errorf("Document.ControlReader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDocument_ControlReaderCtl(t *testing.T) {
	type args struct {
		r      io.Reader
		reload func() *bufio.Reader
	}
	tests := []struct {
		name    string
		args    args
		sc      controlSpecifier
		wantErr bool
		want    bool
	}{
		{
			name: "testLoad",
			args: args{
				r:      bufio.NewReader(strings.NewReader("test")),
				reload: nil,
			},
			sc: controlSpecifier{
				request:  requestLoad,
				done:     make(chan bool),
				chunkNum: 1,
			},
			wantErr: false,
			want:    true,
		},
		{
			name: "testReloadNil",
			args: args{
				r:      bufio.NewReader(strings.NewReader("test")),
				reload: nil,
			},
			sc: controlSpecifier{
				request: requestReload,
				done:    make(chan bool),
			},
			wantErr: false,
			want:    true,
		},
		{
			name: "testReload",
			args: args{
				r: bufio.NewReader(strings.NewReader("test")),
				reload: func() *bufio.Reader {
					return bufio.NewReader(strings.NewReader("test reload"))
				},
			},
			sc: controlSpecifier{
				request: requestReload,
				done:    make(chan bool),
			},
			wantErr: false,
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			if err := m.ControlReader(tt.args.r, tt.args.reload); (err != nil) != tt.wantErr {
				t.Errorf("Document.ControlReader() error = %v, wantErr %v", err, tt.wantErr)
			}
			m.ctlCh <- tt.sc
			got1 := <-tt.sc.done
			if got1 != tt.want {
				t.Errorf("Document.ControlReader() = %v, want %v", got1, tt.want)
			}
		})
	}
}

func TestDocument_ControlLog(t *testing.T) {
	tests := []struct {
		name    string
		sc      controlSpecifier
		wantErr bool
		want    bool
	}{
		{
			name: "test",
			sc: controlSpecifier{
				request: requestReload,
				done:    make(chan bool),
			},
			wantErr: false,
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			if err := m.ControlLog(); (err != nil) != tt.wantErr {
				t.Errorf("Document.ControlLog() error = %v, wantErr %v", err, tt.wantErr)
			}
			m.ctlCh <- tt.sc
			got1 := <-tt.sc.done
			if got1 != tt.want {
				t.Errorf("Document.ControlLog() = %v, want %v", got1, tt.want)
			}
		})
	}
}
