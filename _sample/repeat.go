package main

import (
	"fmt"
	"strings"
)

func main() {
	tenRep := max(1, 80/100)
	oneRep := max(1, 80/10)
	str := strings.Repeat(
		"         1         2         3         4         5         6         7         8         9         0",
		tenRep,
	)
	fmt.Println(str)
	str2 := strings.Repeat("1234567890", oneRep)
	fmt.Println(str2)

}
