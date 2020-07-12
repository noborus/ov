package main

import (
	"fmt"

	"github.com/noborus/ov/oviewer"
)

func main() {
	ov, err := oviewer.Open("main.go")
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := ov.Run(); err != nil {
		fmt.Println(err)
		return
	}
}
