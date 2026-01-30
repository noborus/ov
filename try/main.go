package main

import (
	"fmt"

	"github.com/rivo/uniseg"
)

func main() {
	str := "🇩🇪🏳️‍🌈!"
	state := -1
	var c string
	for len(str) > 0 {
		var width int
		c, str, width, state = uniseg.FirstGraphemeClusterInString(str, state)
		fmt.Println(c, width, state)
	}
}
