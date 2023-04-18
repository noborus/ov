package main

import (
	"log"
	"os"

	"github.com/noborus/ov/oviewer"
)

func main() {
	doc, err := oviewer.NewDocument()
	if err != nil {
		log.Fatal(err)
	}
	fileName := "/var/log/syslog"
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	doc.FileName = fileName
	// Use ControlFile to pass the file name to work in follow mode.
	if err := doc.ControlFile(f); err != nil {
		log.Fatal(err)
	}

	ov, err := oviewer.NewOviewer(doc)
	if err != nil {
		log.Fatal(err)
	}
	// Set in general as individual modes will be overwritten.
	ov.General.FollowMode = true
	ov.General.MultiColorWords = []string{"error:", "info:", "warn:", "debug:"}
	if err := ov.Run(); err != nil {
		log.Fatal(err)
	}
}
