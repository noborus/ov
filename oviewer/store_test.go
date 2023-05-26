package oviewer

import (
	"sync"
	"testing"

	lru "github.com/hashicorp/golang-lru/v2"
)

func Test_newLoadChunks(t *testing.T) {
	type global struct {
		MemoryLimit     int
		MemoryLimitFile int
	}
	type args struct {
		isFile bool
	}
	tests := []struct {
		name   string
		global global
		args   args
		want   int
	}{
		{
			name: "testMemLimit",
			global: global{
				MemoryLimit:     3,
				MemoryLimitFile: 100,
			},
			args: args{
				isFile: false,
			},
			want: 4,
		},
		{
			name: "testFileLimit",
			global: global{
				MemoryLimit:     -1,
				MemoryLimitFile: 10,
			},
			args: args{
				isFile: true,
			},
			want: 11,
		},
		{
			name: "testDefault",
			global: global{
				MemoryLimit:     -1,
				MemoryLimitFile: 100,
			},
			args: args{
				isFile: false,
			},
			want: 101,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MemoryLimit = tt.global.MemoryLimit
			MemoryLimitFile = tt.global.MemoryLimitFile
			if got := newLoadChunks(tt.args.isFile); got != tt.want {
				t.Errorf("newLoadChunks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_store_chunkRange(t *testing.T) {
	type fields struct {
		chunks   []*chunk
		startNum int
		endNum   int
	}
	type args struct {
		chunkNum int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		want1  int
	}{
		{
			name: "test0",
			fields: fields{
				chunks: []*chunk{
					NewChunk(0),
				},
				startNum: 0,
				endNum:   10,
			},
			args:  args{chunkNum: 0},
			want:  0,
			want1: 10,
		},
		{
			name: "test1Chunk",
			fields: fields{
				chunks: []*chunk{
					NewChunk(0),
					NewChunk(0),
					NewChunk(0),
				},
				startNum: 0,
				endNum:   20010,
			},
			args:  args{chunkNum: 1},
			want:  0,
			want1: 10000,
		},
		{
			name: "test2Chunk",
			fields: fields{
				chunks: []*chunk{
					NewChunk(0),
					NewChunk(0),
					NewChunk(0),
				},
				startNum: 0,
				endNum:   20010,
			},
			args:  args{chunkNum: 2},
			want:  0,
			want1: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &store{
				chunks:   tt.fields.chunks,
				startNum: tt.fields.startNum,
				endNum:   tt.fields.endNum,
			}
			got, got1 := s.chunkRange(tt.args.chunkNum)
			if got != tt.want {
				t.Errorf("store.chunkRange() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("store.chunkRange() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_store_swapChunksFile(t *testing.T) {
	type fields struct {
		loadedChunks *lru.Cache[int, struct{}]
		chunks       []*chunk
		mu           sync.Mutex
		startNum     int
		endNum       int
		noNewlineEOF int32
		eof          int32
		size         int64
		offset       int64
	}
	type args struct {
		chunkNum int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &store{
				loadedChunks: tt.fields.loadedChunks,
				chunks:       tt.fields.chunks,
				mu:           tt.fields.mu,
				startNum:     tt.fields.startNum,
				endNum:       tt.fields.endNum,
				noNewlineEOF: tt.fields.noNewlineEOF,
				eof:          tt.fields.eof,
				size:         tt.fields.size,
				offset:       tt.fields.offset,
			}
			s.swapLoadedFile(tt.args.chunkNum)
		})
	}
}
