package oviewer

import (
	"testing"
)

func TestDocument_Write(t *testing.T) {
	type args struct {
		log []byte
	}
	tests := []struct {
		name    string
		repeat  int
		args    args
		want    int
		wantErr bool
	}{
		{
			name:   "test1",
			repeat: 1,
			args: args{
				log: []byte("test"),
			},
			want:    4,
			wantErr: false,
		},
		{
			name:   "testChunkSize",
			repeat: 30002,
			args: args{
				log: []byte("test"),
			},
			want:    4,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewDocument()
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < tt.repeat; i++ {
				got, err := m.Write(tt.args.log)
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
