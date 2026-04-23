package oviewer

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestReplaceEditorArgs(t *testing.T) {
	tests := []struct {
		name      string
		editorCmd string
		numStr    string
		fileName  string
		wantCmd   string
		wantArgs  []string
	}{
		{
			name:      "Simple vi with no args",
			editorCmd: "vi",
			numStr:    "10",
			fileName:  "file.txt",
			wantCmd:   "vi",
			wantArgs:  []string{"file.txt"},
		},
		{
			name:      "vi with +%d",
			editorCmd: "vi +%d",
			numStr:    "42",
			fileName:  "foo.txt",
			wantCmd:   "vi",
			wantArgs:  []string{"+42", "foo.txt"},
		},
		{
			name:      "vim with +%d %f",
			editorCmd: "vim +%d %f",
			numStr:    "7",
			fileName:  "bar.txt",
			wantCmd:   "vim",
			wantArgs:  []string{"+7", "bar.txt"},
		},
		{
			name:      "custom editor with %f in middle",
			editorCmd: "myeditor --file=%f --line=%d",
			numStr:    "99",
			fileName:  "baz.txt",
			wantCmd:   "myeditor",
			wantArgs:  []string{"--file=baz.txt", "--line=99"},
		},
		{
			name:      "editor with no %f",
			editorCmd: "nano --line=%d",
			numStr:    "5",
			fileName:  "abc.txt",
			wantCmd:   "nano",
			wantArgs:  []string{"--line=5", "abc.txt"},
		},
		{
			name:      "editor with %%d and %%f",
			editorCmd: "ed --show=%%d --file=%%f +%d %f",
			numStr:    "3",
			fileName:  "def.txt",
			wantCmd:   "ed",
			wantArgs:  []string{"--show=%d", "--file=%f", "+3", "def.txt"},
		},
		{
			name:      "empty editorCmd",
			editorCmd: "",
			numStr:    "1",
			fileName:  "empty.txt",
			wantCmd:   DefaultEditor,
			wantArgs:  []string{"empty.txt"},
		},
		{
			name:      "editorCmd with only spaces",
			editorCmd: "   ",
			numStr:    "1",
			fileName:  "space.txt",
			wantCmd:   DefaultEditor,
			wantArgs:  []string{"space.txt"},
		},
		{
			name:      "editorCmd with quoted args",
			editorCmd: `vim --cmd "set number" +%d %f`,
			numStr:    "12",
			fileName:  "quoted.txt",
			wantCmd:   "vim",
			wantArgs:  []string{"--cmd", "set number", "+12", "quoted.txt"},
		},
		{
			name:      "editorCmd with multiple %f",
			editorCmd: "multi %f --again=%f",
			numStr:    "2",
			fileName:  "multi.txt",
			wantCmd:   "multi",
			wantArgs:  []string{"multi.txt", "--again=multi.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs := replaceEditorArgs(tt.editorCmd, tt.numStr, tt.fileName)
			if gotCmd != tt.wantCmd {
				t.Errorf("replaceEditorArgs() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("replaceEditorArgs() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestRoot_identifyEditor(t *testing.T) {
	tests := []struct {
		name         string
		ovedit       string
		configEditor string
		editor       string
		want         string
	}{
		{
			name:         "OVEDIT takes precedence",
			ovedit:       "ovedit --wait",
			configEditor: "config-editor",
			editor:       "env-editor",
			want:         "ovedit --wait",
		},
		{
			name:         "config editor overrides EDITOR",
			configEditor: "config-editor",
			editor:       "env-editor",
			want:         "config-editor",
		},
		{
			name:   "EDITOR fallback",
			editor: "env-editor",
			want:   "env-editor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OVEDIT", tt.ovedit)
			t.Setenv("EDITOR", tt.editor)
			root := rootHelper(t)
			root.Config.Editor = tt.configEditor

			if got := root.identifyEditor(); got != tt.want {
				t.Errorf("identifyEditor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRoot_saveTempFile(t *testing.T) {
	root := rootFileReadHelper(t, filepath.Join(testdata, "test.txt"))

	fileName, err := root.saveTempFile()
	if err != nil {
		t.Fatalf("saveTempFile() error = %v", err)
	}
	defer os.Remove(fileName)

	got, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "test\n" {
		t.Errorf("saveTempFile() content = %q, want %q", string(got), "test\n")
	}

	info, err := os.Stat(fileName)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	wantMode := os.FileMode(0o400)
	if runtime.GOOS == "windows" {
		wantMode = 0o444
	}
	if info.Mode().Perm() != wantMode {
		t.Errorf("saveTempFile() mode = %#o, want %#o", info.Mode().Perm(), wantMode)
	}
}
