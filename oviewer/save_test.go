package oviewer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func Test_saveConfirmKey(t *testing.T) {
	type args struct {
		ev *tcell.EventKey
	}
	tests := []struct {
		name string
		args args
		want saveSelection
	}{
		{
			name: "saveOverWrite",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyRune, "O", tcell.ModNone),
			},
			want: saveOverWrite,
		},
		{
			name: "saveAppend",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyRune, "A", tcell.ModNone),
			},
			want: saveAppend,
		},
		{
			name: "saveCancel",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyRune, "N", tcell.ModNone),
			},
			want: saveCancel,
		},
		{
			name: "saveCancel2",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone),
			},
			want: saveCancel,
		},
		{
			name: "saveIgnore",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyRune, "I", tcell.ModNone),
			},
			want: saveIgnore,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := saveConfirmKey(tt.args.ev); got != tt.want {
				t.Errorf("saveConfirmKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_promptSaveFlag(t *testing.T) {
	tests := []struct {
		name       string
		existing   bool
		event      *tcell.EventKey
		wantFlag   int
		wantErr    bool
		wantPrompt string
	}{
		{
			name:     "new file",
			existing: false,
			wantFlag: os.O_WRONLY | os.O_CREATE,
		},
		{
			name:       "overwrite existing file",
			existing:   true,
			event:      tcell.NewEventKey(tcell.KeyRune, "O", tcell.ModNone),
			wantFlag:   os.O_WRONLY | os.O_TRUNC,
			wantPrompt: "overwrite? (O)overwrite, (A)Append, (N)cancel:",
		},
		{
			name:       "append existing file",
			existing:   true,
			event:      tcell.NewEventKey(tcell.KeyRune, "A", tcell.ModNone),
			wantFlag:   os.O_WRONLY | os.O_CREATE | os.O_APPEND,
			wantPrompt: "overwrite? (O)overwrite, (A)Append, (N)cancel:",
		},
		{
			name:       "cancel existing file",
			existing:   true,
			event:      tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone),
			wantErr:    true,
			wantPrompt: "overwrite? (O)overwrite, (A)Append, (N)cancel:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			fileName := filepath.Join(t.TempDir(), "saved.txt")
			if tt.existing {
				if err := os.WriteFile(fileName, []byte("existing\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if tt.event != nil {
				go func() {
					root.Screen.EventQ() <- tt.event
				}()
			}

			got, err := root.promptSaveFlag(fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("promptSaveFlag() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantFlag {
				t.Errorf("promptSaveFlag() flag = %v, want %v", got, tt.wantFlag)
			}
			if tt.wantPrompt != "" && root.message != tt.wantPrompt {
				t.Errorf("promptSaveFlag() prompt = %q, want %q", root.message, tt.wantPrompt)
			}
		})
	}
}

func TestRoot_saveBuffer(t *testing.T) {
	tests := []struct {
		name         string
		initial      string
		input        string
		event        *tcell.EventKey
		wantContains string
		wantMessage  string
	}{
		{
			name:         "save new file",
			input:        " saved.txt ",
			wantContains: "test\n",
			wantMessage:  "saved ",
		},
		{
			name:         "overwrite existing file",
			initial:      "old content\n",
			input:        "saved.txt",
			event:        tcell.NewEventKey(tcell.KeyRune, "O", tcell.ModNone),
			wantContains: "test\n",
			wantMessage:  "saved ",
		},
		{
			name:         "append existing file",
			initial:      "prefix\n",
			input:        "saved.txt",
			event:        tcell.NewEventKey(tcell.KeyRune, "A", tcell.ModNone),
			wantContains: "prefix\ntest\n",
			wantMessage:  "saved ",
		},
		{
			name:        "cancel existing file",
			initial:     "keep\n",
			input:       "saved.txt",
			event:       tcell.NewEventKey(tcell.KeyRune, "N", tcell.ModNone),
			wantMessage: "save cancel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, filepath.Join(testdata, "test.txt"))
			root.Doc.seekable = false
			fileName := filepath.Join(t.TempDir(), "saved.txt")
			if tt.initial != "" {
				if err := os.WriteFile(fileName, []byte(tt.initial), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if tt.event != nil {
				go func() {
					root.Screen.EventQ() <- tt.event
				}()
			}

			root.saveBuffer(strings.Replace(tt.input, "saved.txt", fileName, 1))

			if !strings.Contains(root.message, tt.wantMessage) {
				t.Fatalf("saveBuffer() message = %q, want substring %q", root.message, tt.wantMessage)
			}

			got, err := os.ReadFile(fileName)
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantContains != "" && !strings.Contains(string(got), tt.wantContains) {
				t.Errorf("saveBuffer() wrote %q, want substring %q", string(got), tt.wantContains)
			}
			if tt.wantMessage == "save cancel" && string(got) != tt.initial {
				t.Errorf("saveBuffer() modified file on cancel: got %q, want %q", string(got), tt.initial)
			}
		})
	}
}
