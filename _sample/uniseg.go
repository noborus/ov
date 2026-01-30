package main

import (
	"fmt"

	"github.com/rivo/uniseg"
)

func main() {
	s := "a\u007fa"
	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		fmt.Printf("%q\n", gr.Str())
	}
}
