//go:build solaris || illumos

package oviewer

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

// registerSIGTSTP registers SIGTSTP signal.
func registerSIGTSTP() chan os.Signal {
	sigSuspend := make(chan os.Signal, 1)
	signal.Notify(sigSuspend, syscall.SIGTSTP)
	return sigSuspend
}

// suspendProcess sends SIGSTOP signal to the process group.
func suspendProcess() error {
	pid, err := unix.Getpgrp()
	if err != nil {
		return err
	}

	return unix.Kill(-pid, syscall.SIGSTOP)
}
