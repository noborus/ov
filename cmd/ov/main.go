package main

import "github.com/noborus/oviewer/cmd"

// Version set "git describe --tags --abbrev=0"
var Version string

// Revision set "git rev-parse --short HEAD"
var Revision string

func main() {
	cmd.Execute(Version, Revision)
}
