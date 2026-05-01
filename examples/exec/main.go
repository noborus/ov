//go:build !windows && !plan9

package main

import (
	"os/exec"

	"github.com/noborus/ov/oviewer"
)

func main() {
	timeCmd, err := exec.LookPath("time")
	if err != nil {
		timeCmd = "/usr/bin/time"
	}

	// -p is a POSIX format and works on GNU/BSD time implementations.
	cmd := oviewer.NewCommand(timeCmd, "-p", "ls", "-alF")
	ov, err := cmd.Exec()
	if err != nil {
		panic(err)
	}
	defer func() {
		cmd.Wait()
	}()
	ov.FollowAll = true
	if err := ov.Run(); err != nil {
		panic(err)
	}
}
