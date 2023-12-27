package oviewer

import (
	"bytes"
	"io"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/spf13/viper"
)

const cwd = ".."

var testdata = filepath.Join(cwd, "testdata")

func fakeScreen() (tcell.Screen, error) {
	return tcell.NewSimulationScreen(""), nil
}

func TestNewOviewer(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
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

func TestOpen(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type args struct {
		fileNames []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				fileNames: []string{filepath.Join(testdata, "/test.txt")},
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				fileNames: []string{
					filepath.Join(testdata, "test.txt"),
					filepath.Join(testdata, "test2.txt"),
				},
			},
			wantErr: false,
		},
		{
			name: "testErr",
			args: args{
				fileNames: []string{filepath.Join(testdata, "err.txt")},
			},
			wantErr: true,
		},
		{
			name: "testDir",
			args: args{
				fileNames: []string{testdata},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.args.fileNames...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			root.Quit()
		})
	}
}

func TestNewRoot(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type args struct {
		read io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *Root
		wantErr bool
	}{
		{
			name:    "test1",
			args:    args{read: bytes.NewBufferString("test")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRoot(tt.args.read)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRoot_Run(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	tests := []struct {
		name    string
		general general
		ovArgs  []string
		wantErr bool
	}{
		{
			name: "testEmpty",
			general: general{
				TabWidth:       8,
				MarkStyleWidth: 1,
			},
			ovArgs:  []string{},
			wantErr: false,
		},
		{
			name: "test1",
			general: general{
				TabWidth:       8,
				MarkStyleWidth: 1,
			},
			ovArgs:  []string{filepath.Join(testdata, "test.txt")},
			wantErr: false,
		},
		{
			name: "testHeader",
			general: general{
				TabWidth:       8,
				MarkStyleWidth: 1,
				Header:         1,
			},
			ovArgs:  []string{filepath.Join(testdata, "test.txt")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.ovArgs...)
			if err != nil {
				t.Fatalf("NewOviewer error = %v", err)
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

func Test_applyStyle(t *testing.T) {
	t.Parallel()
	type args struct {
		style tcell.Style
		s     OVStyle
	}
	tests := []struct {
		name string
		args args
		want tcell.Style
	}{
		{
			name: "test1",
			args: args{
				style: tcell.StyleDefault,
				s: OVStyle{
					Background: "red",
					Foreground: "white",
				},
			},
			want: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := applyStyle(tt.args.style, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_setKeyConfig(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	tests := []struct {
		name    string
		cfgFile string
		want    []string
		wantErr bool
	}{
		{
			name:    "test-ov.yaml",
			cfgFile: filepath.Join(cwd, "ov.yaml"),
			want:    []string{"Enter", "Down", "ctrl+N"},
			wantErr: false,
		},
		{
			name:    "test-ov-less.yaml",
			cfgFile: filepath.Join(cwd, "ov-less.yaml"),
			want:    []string{"e", "ctrl+e", "j", "J", "ctrl+j", "Enter", "Down"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(filepath.Join(testdata, "test.txt"))
			if err != nil {
				t.Fatalf("NewOviewer error = %v", err)
			}
			root.Screen = tcell.NewSimulationScreen("")
			viper.SetConfigFile(tt.cfgFile)
			var config Config
			viper.AutomaticEnv() // read in environment variables that match
			if err := viper.ReadInConfig(); err != nil {
				t.Fatal("failed to read config file:", err)
			}
			viper.Unmarshal(&config)
			root.SetConfig(config)
			got, err := root.setKeyConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("Root.setKeyConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			action := "down"
			if !reflect.DeepEqual(got[action], tt.want) {
				t.Errorf("Root.setKeyConfig() = %v, want %v", got[action], tt.want)
			}
		})
	}
}
