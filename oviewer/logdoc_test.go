package oviewer

import (
	"testing"
)

func TestDocument_Write(t *testing.T) {
	type args struct {
		logs [][]byte
	}
	tests := []struct {
		name      string
		ChunkSize int
		args      args
		want      int
		wantErr   bool
	}{
		{
			name:      "test1",
			ChunkSize: 10000,
			args: args{
				logs: [][]byte{
					[]byte("test"),
				},
			},
			want:    4,
			wantErr: false,
		},
		{
			name:      "testChunkSize",
			ChunkSize: 2,
			args: args{
				logs: [][]byte{
					[]byte("test1"),
					[]byte("test2"),
					[]byte("test3"),
					[]byte("test4"),
					[]byte("test5"),
					[]byte("test6"),
				},
			},
			want:    5,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := ChunkSize
			ChunkSize = tt.ChunkSize
			defer func() {
				ChunkSize = tmp
			}()
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			for _, log := range tt.args.logs {
				got, err := m.Write(log)
				if (err != nil) != tt.wantErr {
					t.Errorf("Document.Write() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("Document.Write() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
