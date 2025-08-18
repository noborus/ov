package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/noborus/ov/oviewer"
)

func main() {
	ov, err := oviewer.Open("main.go")
	if err != nil {
		log.Fatal(err)
	}
	// All Documents.
	/*
		ov.Config.General.Style = oviewer.StyleConfig{
			SearchHighlight: &oviewer.OVStyle{
				Foreground: "gold",
				Reverse:    true,
				Blink:      true,
			},
		}
		ov.SetConfig(ov.Config)
	*/
	// Only this Document.
	ov.Doc.General.Style = oviewer.StyleConfig{
		SearchHighlight: &oviewer.OVStyle{
			Foreground: "gold",
			Reverse:    true,
			Blink:      true,
		},
	}
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := ov.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(time.Second * 1)
	ov.MoveBottom()
	ov.BackSearch("main")

	time.Sleep(time.Second * 1)
	ov.MoveTop()
	ov.Search("import")

	time.Sleep(time.Second * 10)
	ov.Quit(context.Background())
	wg.Wait()
}
