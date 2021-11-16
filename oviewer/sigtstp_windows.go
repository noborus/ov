//go:build windows
// +build windows

package oviewer

import (
	"os"
	"os/signal"
	"syscall"
)

// Dummy function because there is no sigtstp in windows.
func registerSIGTSTP(sigSuspend chan os.Signal) {
}
