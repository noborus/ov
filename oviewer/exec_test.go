package oviewer

import (
	"os/exec"
	"testing"
)

func TestExecCommand(t *testing.T) {
	type args struct {
		command *exec.Cmd
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				command: exec.Command("date"),
			},
			wantErr: false,
		},
		{
			name: "testNotFound",
			args: args{
				command: exec.Command("notFoundExec"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ExecCommand(tt.args.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
