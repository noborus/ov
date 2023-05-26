package oviewer

import (
	"testing"
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

func testNewStoreFile(t *testing.T, chunkNum int) *store {
	t.Helper()
	s := NewStore()
	s.chunks = make([]*chunk, chunkNum)
	s.setNewLoadChunks(true)
	for i := 0; i < chunkNum; i++ {
		chunk := NewChunk(0)
		chunk.lines = make([][]byte, ChunkSize)
		for j := 0; j < ChunkSize; j++ {
			chunk.lines[j] = []byte("a")
		}
		s.chunks[i] = chunk
	}
	return s
}

func Test_store_swapChunksFile(t *testing.T) {
	type global struct {
		MemoryLimit     int
		MemoryLimitFile int
	}
	type fields struct {
		maxChunks int
	}
	type args struct {
		chunkNums []int
		contains  int
	}
	tests := []struct {
		name   string
		global global
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "test0",
			fields: fields{maxChunks: 100},
			global: global{
				MemoryLimit:     3,
				MemoryLimitFile: 100,
			},
			args: args{
				chunkNums: []int{0},
				contains:  0,
			},
			want: false,
		},
		{
			name:   "test1",
			fields: fields{maxChunks: 100},
			global: global{
				MemoryLimit:     3,
				MemoryLimitFile: 100,
			},
			args: args{
				chunkNums: []int{1},
				contains:  1,
			},
			want: true,
		},
		{
			name:   "testFalse",
			fields: fields{maxChunks: 100},
			global: global{
				MemoryLimit:     3,
				MemoryLimitFile: 100,
			},
			args: args{
				chunkNums: []int{99},
				contains:  1,
			},
			want: false,
		},
		{
			name:   "testEvict",
			fields: fields{maxChunks: 100},
			global: global{
				MemoryLimit:     3,
				MemoryLimitFile: 3,
			},
			args: args{
				chunkNums: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
				contains:  2,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MemoryLimit = tt.global.MemoryLimit
			MemoryLimitFile = tt.global.MemoryLimitFile
			s := testNewStoreFile(t, tt.fields.maxChunks)
			for _, num := range tt.args.chunkNums {
				s.swapLoadedFile(num)
			}

			if got := s.loadedChunks.Contains(tt.args.contains); got != tt.want {
				t.Logf("loadedChunks: %v", s.loadedChunks.Len())
				t.Errorf("store.swapChunksFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
