package oviewer

import (
	"testing"
)

func TestDocument_ReadFile(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "testNoFile",
			args: args{
				[]string{
					"nofile",
				},
			},
			wantErr: true,
		},
		{
			name: "testSTDIN",
			args: args{
				[]string{},
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
