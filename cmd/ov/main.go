package main

import "github.com/noborus/oviewer/cmd"

var Version = "v0.0.1"

// Revision set "git rev-parse --short HEAD"
var Revision = "HEAD"

func main() {
	cmd.Execute(Version, Revision)
}
