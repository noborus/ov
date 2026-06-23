//go:build !windows

package oviewer

import (
	"os"
	"path/filepath"
	"testing"
)

// withFakeWaylandTools creates fake wl-copy/wl-paste executables backed by a
// temp file, prepends their directory to PATH, and restores both on cleanup.
func withFakeWaylandTools(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	clipFile := filepath.Join(dir, "clip.txt")

	scripts := map[string]string{
		"wl-copy":  "#!/bin/sh\ncat > \"" + clipFile + "\"\n",
		"wl-paste": "#!/bin/sh\ncat \"" + clipFile + "\"\n",
	}
	for name, content := range scripts {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
			t.Fatalf("failed to write fake %s: %v", name, err)
		}
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
}

func Test_waylandClipboardAvailable(t *testing.T) {
	t.Run("no WAYLAND_DISPLAY", func(t *testing.T) {
		t.Setenv("WAYLAND_DISPLAY", "")
		if waylandClipboardAvailable() {
			t.Error("waylandClipboardAvailable() = true, want false without WAYLAND_DISPLAY")
		}
	})

	t.Run("WAYLAND_DISPLAY set but wl-clipboard missing", func(t *testing.T) {
		t.Setenv("WAYLAND_DISPLAY", "wayland-0")
		t.Setenv("PATH", t.TempDir())
		if waylandClipboardAvailable() {
			t.Error("waylandClipboardAvailable() = true, want false without wl-copy/wl-paste in PATH")
		}
	})

	t.Run("WAYLAND_DISPLAY set and wl-clipboard installed", func(t *testing.T) {
		withFakeWaylandTools(t)
		if !waylandClipboardAvailable() {
			t.Error("waylandClipboardAvailable() = false, want true with WAYLAND_DISPLAY and wl-copy/wl-paste in PATH")
		}
	})
}

func Test_waylandClipboard_writeRead(t *testing.T) {
	withFakeWaylandTools(t)

	want := "ov wayland clipboard test"
	if err := writeWaylandClipboard(want); err != nil {
		t.Fatalf("writeWaylandClipboard() error = %v", err)
	}

	got, err := readWaylandClipboard()
	if err != nil {
		t.Fatalf("readWaylandClipboard() error = %v", err)
	}
	if got != want {
		t.Errorf("readWaylandClipboard() = %q, want %q", got, want)
	}
}
