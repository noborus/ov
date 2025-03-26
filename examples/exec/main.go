package main

import (
	"github.com/noborus/ov/oviewer"
)

func main() {
	cmd := oviewer.NewCommand("/usr/bin/time", "--verbose", "ls", "-alF")
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
