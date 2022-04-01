package main

import (
	"log"

	"github.com/noborus/ov/oviewer"
)

func main() {
	doc, err := oviewer.NewDocument()
	if err != nil {
		log.Fatal(err)
	}
	// Use ReadFile to pass the file name to work in follow mode.
	if err := doc.ReadFile("/var/log/syslog"); err != nil {
		log.Fatal(err)
	}

	ov, err := oviewer.NewOviewer(doc)
	if err != nil {
		log.Fatal(err)
	}
	// Set in general as individual modes will be overwritten.
	ov.General.FollowMode = true

	if err := ov.Run(); err != nil {
		log.Fatal(err)
	}
}
