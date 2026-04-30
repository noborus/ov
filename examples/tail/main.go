//go:build !windows && !plan9

package main

import (
	"log"
	"os"
	"runtime"

	"github.com/noborus/ov/oviewer"
)

func main() {
	fileName := "/var/log/syslog"
	if runtime.GOOS == "darwin" {
		fileName = "/var/log/system.log"
	}
	doc, err := oviewer.NewDocument()
	if err != nil {
		log.Fatal(err)
	}
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
	ov.Doc.General.SetMultiColorWords([]string{"error:", "info:", "warn:", "debug:"})
	ov.Doc.General.SetFollowMode(true)
	if err := ov.Run(); err != nil {
		log.Fatal(err)
	}
}
