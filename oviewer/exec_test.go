package oviewer

import (
	"os/exec"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestExecCommand(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type args struct {
		cmdStr []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				cmdStr: []string{"date"},
			},
			wantErr: false,
		},
		{
			name: "testNotFound",
			args: args{
				cmdStr: []string{"notFoundExec"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := exec.Command(tt.args.cmdStr[0], tt.args.cmdStr[1:]...)
			_, err := ExecCommand(command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
