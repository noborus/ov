//go:build windows
// +build windows

package oviewer

import (
	"os"
)

// Dummy function because there is no sigtstp in windows.
func registerSIGTSTP() chan os.Signal {
	sigSuspend := make(chan os.Signal, 1)
	return sigSuspend
}
