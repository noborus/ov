//go:build windows
// +build windows

package oviewer

import (
	"os"
	"os/signal"
)

// Dummy function because there is no sigtstp in windows.
func registerSIGTSTP() chan os.Signal {
	sigSuspend := make(chan os.Signal, 1)
	return sigSuspend
}
