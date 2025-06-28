//go:build !unix

package oviewer

import (
	"os"
)

// Dummy function because there is no SIGTSTP in non-Unix systems.
func registerSIGTSTP() chan os.Signal {
	sigSuspend := make(chan os.Signal, 1)
	return sigSuspend
}

// suspendProcess is a dummy function.
func suspendProcess() error {
	return nil
}
