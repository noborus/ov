package main

import (
	"github.com/noborus/ov/oviewer"
)

func main() {
	ov, err := oviewer.Open("main.go")
	if err != nil {
		panic(err)
	}
	if err := ov.Run(); err != nil {
		panic(err)
	}
}
