//go:build !windows

package oviewer

import (
	"os"
	"path/filepath"
	"sync"
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

// checkWaylandClipboardAvailable holds the actual detection logic;
// waylandClipboardAvailable just memoizes it for the process lifetime, which
// would make these env/PATH-toggling subtests order-dependent if they called
// the memoized wrapper instead.
func Test_checkWaylandClipboardAvailable(t *testing.T) {
	t.Run("no WAYLAND_DISPLAY", func(t *testing.T) {
		t.Setenv("WAYLAND_DISPLAY", "")
		if checkWaylandClipboardAvailable() {
			t.Error("checkWaylandClipboardAvailable() = true, want false without WAYLAND_DISPLAY")
		}
	})

	t.Run("WAYLAND_DISPLAY set but wl-clipboard missing", func(t *testing.T) {
		t.Setenv("WAYLAND_DISPLAY", "wayland-0")
		t.Setenv("PATH", t.TempDir())
		if checkWaylandClipboardAvailable() {
			t.Error("checkWaylandClipboardAvailable() = true, want false without wl-copy/wl-paste in PATH")
		}
	})

	t.Run("WAYLAND_DISPLAY set and wl-clipboard installed", func(t *testing.T) {
		withFakeWaylandTools(t)
		if !checkWaylandClipboardAvailable() {
			t.Error("checkWaylandClipboardAvailable() = false, want true with WAYLAND_DISPLAY and wl-copy/wl-paste in PATH")
		}
	})
}

func Test_waylandClipboardAvailable_caches(t *testing.T) {
	withFakeWaylandTools(t)
	waylandAvailableOnce = sync.Once{}
	t.Cleanup(func() { waylandAvailableOnce = sync.Once{} })

	if !waylandClipboardAvailable() {
		t.Fatal("waylandClipboardAvailable() = false, want true with WAYLAND_DISPLAY and wl-copy/wl-paste in PATH")
	}

	// Even after the environment no longer satisfies checkWaylandClipboardAvailable,
	// the memoized result from the first call should stick for the process lifetime.
	t.Setenv("WAYLAND_DISPLAY", "")
	if !waylandClipboardAvailable() {
		t.Error("waylandClipboardAvailable() = false, want cached true despite WAYLAND_DISPLAY being unset afterward")
	}
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
