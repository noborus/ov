package oviewer

import (
	"os/exec"
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestNewOviewer(t *testing.T) {
	type args struct {
		docs []*Document
	}
	tests := []struct {
		name    string
		args    args
		want    *Root
		wantErr bool
	}{
		{
			name:    "testEmpty",
			args:    args{},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOviewer(tt.args.docs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOviewer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOviewer() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
				command: exec.Command("sh", "-c", "echo \"test\""),
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

func TestRoot_Run(t *testing.T) {
	tests := []struct {
		name    string
		ovArgs  []string
		wantErr bool
	}{
		{
			name:    "test1",
			ovArgs:  []string{"../testdata/test.txt"},
			wantErr: false,
		},
		{
			name:    "testEmpty",
			ovArgs:  []string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.ovArgs...)
			if err != nil {
				t.Errorf("NewOviewer error = %v", err)
			}
			root.Screen = tcell.NewSimulationScreen("")
			go func() {
				if err := root.Run(); (err != nil) != tt.wantErr {
					t.Errorf("Root.Run() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()
			root.Quit()
		})
	}
}
