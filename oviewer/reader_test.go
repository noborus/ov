package oviewer

import (
	"bytes"
	"io"
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
