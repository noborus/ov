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

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := ov.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(time.Second * 1)
	ov.Doc.Style.SearchHighlight = oviewer.OVStyle{
		Foreground: "gold",
		Reverse:    true,
		Blink:      true,
	}
	ov.MoveBottom()
	ov.BackSearch("main")

	time.Sleep(time.Second * 1)
	ov.MoveTop()
	ov.Search("import")

	time.Sleep(time.Second * 10)
	ov.Quit(context.Background())
	wg.Wait()
}
