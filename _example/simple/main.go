package main

import (
	"fmt"
	"time"

	"github.com/noborus/ov/oviewer"
)

func main() {
	ov, err := oviewer.Open([]string{"main.go"})
	if err != nil {
		fmt.Println(err)
		return
	}
	ov.ColorHeader = "red"
	ov.Header = 1
	done := make(chan struct{})
	go func() {
		if err := ov.Run(); err != nil {
			fmt.Println(err)
			return
		}
		close(done)
	}()
	time.Sleep(time.Second * 1)
	m, err := oviewer.NewModel()
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := m.ReadFile([]string{"worldcitiespop.csv"}); err != nil {
		fmt.Println(err)
		return
	}
	/*
		ov.Model = m
	*/
	ov.Model.ClearCache()
	ov.MoveBottom()
	<-done
}
