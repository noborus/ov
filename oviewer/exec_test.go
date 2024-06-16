//go:build !windows
// +build !windows

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
			//nolint:gosec
			command := exec.Command(tt.args.cmdStr[0], tt.args.cmdStr[1:]...)
			_, err := ExecCommand(command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCommand_Exec(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		args   []string
		reload bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "testLs",
			fields: fields{
				args:   []string{"ls"},
				reload: false,
			},
			wantErr: false,
		},
		{
			name: "testNotFoundError",
			fields: fields{
				args:   []string{"notfound"},
				reload: false,
			},
			wantErr: true,
		},
		{
			name: "testLsReload",
			fields: fields{
				args:   []string{"ls"},
				reload: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand(tt.fields.args...)
			_, err := cmd.Exec()
			if tt.fields.reload {
				cmd.Reload()
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Command.Exec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
