package oviewer

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_uncompressedReader(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test.gz",
			args: args{
				fileName: filepath.Join(testdata, "test.txt.gz"),
			},
			want: "GZIP",
		},
		{
			name: "test.lz4",
			args: args{
				fileName: filepath.Join(testdata, "test.txt.lz4"),
			},
			want: "LZ4",
		},
		{
			name: "test.xz",
			args: args{
				fileName: filepath.Join(testdata, "test.txt.xz"),
			},
			want: "XZ",
		},
		{
			name: "test.zst",
			args: args{
				fileName: filepath.Join(testdata, "test.txt.zst"),
			},
			want: "ZSTD",
		},
		{
			name: "test.bz2",
			args: args{
				fileName: filepath.Join(testdata, "test.txt.bz2"),
			},
			want: "BZIP2",
		},
		{
			name: "test.txt",
			args: args{
				fileName: filepath.Join(testdata, "test.txt"),
			},
			want: "UNCOMPRESSED",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.args.fileName)
			if err != nil {
				t.Fatal(err)
			}
			got, _ := uncompressedReader(f, true)
			if got.String() != tt.want {
				t.Errorf("uncompressedReader() got = %v, want %v", got, tt.want)
			}
		})
	}
}
