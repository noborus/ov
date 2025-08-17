package main

import (
	"github.com/noborus/ov/oviewer"
)

func main() {
	ov, err := oviewer.Open("main.go")
	if err != nil {
		panic(err)
	}
	// Set header lines before Run().
	// Options must be set before Run(); do not use Config directly.
	ov.Config.General.SetHeader(1)
	ov.Config.General.SetStatusLine(false)
	ov.SetConfig(ov.Config)

	if err := ov.Run(); err != nil {
		panic(err)
	}
}
