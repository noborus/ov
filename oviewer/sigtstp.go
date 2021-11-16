//go:build !windows
// +build !windows

package oviewer

import (
	"os"
	"os/signal"
	"syscall"
)

func registerSIGTSTP() chan os.Signal {
	sigSuspend := make(chan os.Signal, 1)
	signal.Notify(sigSuspend, syscall.SIGTSTP)
	return sigSuspend
}
