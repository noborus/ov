package main

import (
	"os/exec"

	"github.com/noborus/ov/oviewer"
)

func main() {
	command := exec.Command("/usr/bin/time", "--verbose", "ls", "-alF")
	ov, err := oviewer.ExecCommand(command)
	if err != nil {
		panic(err)
	}
	ov.General.FollowAll = true
	if err := ov.Run(); err != nil {
		panic(err)
	}
}
