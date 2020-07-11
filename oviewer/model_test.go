package oviewer

import (
	"testing"
)

func TestModel_ReadFile(t *testing.T) {
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
			m, err := NewModel()
			if err != nil {
				t.Fatal(err)
			}
			if err := m.ReadFile(tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("Model.ReadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
