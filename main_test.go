package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_initConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		cfgFile string
		errStr  string
	}{
		{
			name:    "test-ov.yaml",
			cfgFile: "ov.yaml",
			errStr:  "",
		},
		{
			name:    "test-ov-less.yaml",
			cfgFile: "ov-less.yaml",
			errStr:  "",
		},
		{
			name:    "no-file.yaml",
			cfgFile: "no-file.yaml",
			errStr:  "no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			io.Copy(&buf, r)
			capturedStderr := buf.String()

			// Now you can assert capturedStderr
			// For example, check if it contains a specific error message
			if !strings.Contains(capturedStderr, tt.errStr) {
				t.Errorf("initConfig() error = %v, wantErr %v", capturedStderr, tt.errStr)
			}
			//keyBind := oviewer.GetKeyBinds(config)
			//log.Println(oviewer.KeyBindString(keyBind))
		})
	}
}
