package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/noborus/ov/oviewer"
)

func Test_initConfig(t *testing.T) {
	tests := []struct {
		name    string
		xdg     string
		cfgFile string
		wantErr bool
	}{
		{
			name:    "test-ov.yaml",
			xdg:     "",
			cfgFile: "ov.yaml",
			wantErr: false,
		},
		{
			name:    "test-ov-less.yaml",
			xdg:     "",
			cfgFile: "ov-less.yaml",
			wantErr: false,
		},
		{
			name:    "no-file.yaml",
			xdg:     "",
			cfgFile: "no-file.yaml",
			wantErr: true,
		},
		{
			name:    "not found",
			xdg:     "dummy",
			cfgFile: "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.xdg != "" {
				t.Setenv("XDG_CONFIG_HOME", tt.xdg)
			}

			cfgFile = tt.cfgFile
			// Backup original stderr
			origStderr := os.Stderr

			// Create a buffer to capture stderr output
			r, w, _ := os.Pipe()
			os.Stderr = w

			initConfig()
			w.Close()
			// Restore original stderr
			os.Stderr = origStderr

			// Read captured stderr output
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatal(err)
			}

			capturedStderr := buf.String()

			// Now you can assert capturedStderr
			// For example, check if it contains a specific error message
			got := len(capturedStderr) > 0
			if got != tt.wantErr {
				t.Errorf("initConfig() error = %v, wantErr %v", capturedStderr, tt.wantErr)
			}
		})
	}
}

func TestRunOviewer_QuitIfOneScreenWithFilter(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "ov-filter-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer tmp.Close()

	if _, err := tmp.WriteString("ok\nng\n"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}

	origConfig := config
	origFilter := filter
	origNonMatchFilter := nonMatchFilter
	origPattern := pattern
	t.Cleanup(func() {
		config = origConfig
		filter = origFilter
		nonMatchFilter = origNonMatchFilter
		pattern = origPattern
	})

	config = oviewer.NewConfig()
	config.QuitSmall = true
	filter = "ok"
	nonMatchFilter = ""
	pattern = ""

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = origStdout
	})

	runErr := RunOviewer([]string{tmp.Name()})
	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if runErr != nil {
		t.Fatalf("RunOviewer() error = %v", runErr)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "ok") {
		t.Errorf("output = %q, want to contain %q", got, "ok")
	}
	if strings.Contains(got, "ng") {
		t.Errorf("output = %q, want not to contain %q", got, "ng")
	}
}

func TestRootCmd_FlagFAndFilter(t *testing.T) {
	origConfig := config
	origFilter := filter
	origCfgFile := cfgFile
	t.Cleanup(func() {
		config = origConfig
		filter = origFilter
		cfgFile = origCfgFile
	})

	config = oviewer.NewConfig()
	filter = ""
	cfgFile = ""

	rootCmd.SetArgs([]string{"-F", "--filter", "ok", "--version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	if !config.QuitSmall {
		t.Errorf("config.QuitSmall = %v, want true", config.QuitSmall)
	}
	if filter != "ok" {
		t.Errorf("filter = %q, want %q", filter, "ok")
	}
}
