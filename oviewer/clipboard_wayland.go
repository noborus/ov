package oviewer

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// waylandClipboardTimeout bounds wl-copy/wl-paste so a frozen compositor or a
// stale WAYLAND_DISPLAY (e.g. left over after a session restart) can't hang
// the single UI event-loop goroutine that calls these.
const waylandClipboardTimeout = 2 * time.Second

var (
	waylandAvailableOnce sync.Once
	waylandAvailableVal  bool
)

// waylandClipboardAvailable reports whether the Wayland clipboard can be used,
// i.e. the session is Wayland and wl-clipboard (wl-copy/wl-paste) is installed.
// golang.design/x/clipboard only reaches the clipboard via XWayland, so on a
// Wayland session without XWayland it silently fails; wl-clipboard talks to
// the compositor directly. The result is cached: it can't change for the
// life of the process.
func waylandClipboardAvailable() bool {
	waylandAvailableOnce.Do(func() {
		waylandAvailableVal = checkWaylandClipboardAvailable()
	})
	return waylandAvailableVal
}

func checkWaylandClipboardAvailable() bool {
	if os.Getenv("WAYLAND_DISPLAY") == "" {
		return false
	}
	if _, err := exec.LookPath("wl-copy"); err != nil {
		return false
	}
	if _, err := exec.LookPath("wl-paste"); err != nil {
		return false
	}
	return true
}

func writeWaylandClipboard(str string) error {
	ctx, cancel := context.WithTimeout(context.Background(), waylandClipboardTimeout)
	defer cancel()

	c := exec.CommandContext(ctx, "wl-copy")
	c.Stdin = strings.NewReader(str)
	return c.Run()
}

func readWaylandClipboard() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), waylandClipboardTimeout)
	defer cancel()

	c := exec.CommandContext(ctx, "wl-paste", "-n")
	var out bytes.Buffer
	c.Stdout = &out
	if err := c.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}
