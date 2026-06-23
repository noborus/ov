package oviewer

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

// waylandClipboardAvailable reports whether the Wayland clipboard can be used,
// i.e. the session is Wayland and wl-clipboard (wl-copy/wl-paste) is installed.
// golang.design/x/clipboard only reaches the clipboard via XWayland, so on a
// Wayland session without XWayland it silently fails; wl-clipboard talks to
// the compositor directly.
func waylandClipboardAvailable() bool {
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

// writeWaylandClipboard writes str to the Wayland clipboard via wl-copy.
func writeWaylandClipboard(str string) error {
	c := exec.Command("wl-copy")
	c.Stdin = strings.NewReader(str)
	return c.Run()
}

// readWaylandClipboard reads the Wayland clipboard via wl-paste.
func readWaylandClipboard() (string, error) {
	c := exec.Command("wl-paste", "-n")
	var out bytes.Buffer
	c.Stdout = &out
	if err := c.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}
