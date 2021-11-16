//go:build !windows
// +build !windows

package oviewer

import (
	"os"
	"os/signal"
	"syscall"
)

func registerSIGTSTP(sigSuspend chan os.Signal) {
	signal.Notify(sigSuspend, syscall.SIGTSTP)
}
