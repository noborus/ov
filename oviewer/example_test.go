package oviewer_test

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/noborus/ov/oviewer"
)

func ExampleOpen() {
	ov, err := oviewer.Open("example_test.go")
	if err != nil {
		panic(err)
	}
	if err := ov.Run(); err != nil {
		panic(err)
	}
}

func ExampleNewRoot() {
	r := strings.NewReader(strings.Repeat("north\n", 99))
	ov, err := oviewer.NewRoot(r)
	if err != nil {
		panic(err)
	}
	if err := ov.Run(); err != nil {
		panic(err)
	}
}

func ExampleNewOviewer() {
	doc, err := oviewer.NewDocument()
	if err != nil {
		panic(err)
	}
	s := "Hello, World!"
	if err := doc.ReadAll(bytes.NewBufferString(s)); err != nil {
		panic(err)
	}

	ov, err := oviewer.NewOviewer(doc)
	if err != nil {
		panic(err)
	}
	if err := ov.Run(); err != nil {
		panic(err)
	}
}

func ExampleExecCommand() {
	command := exec.Command("ls", "-alF")
	ov, err := oviewer.ExecCommand(command)
	if err != nil {
		panic(err)
	}
	if err := ov.Run(); err != nil {
		panic(err)
	}
}

func ExampleSearch() {
	doc, err := oviewer.NewDocument()
	if err != nil {
		panic(err)
	}
	s := "Hello, World!"
	if err := doc.ReadAll(bytes.NewBufferString(s)); err != nil {
		panic(err)
	}

	ov, err := oviewer.NewOviewer(doc)
	if err != nil {
		panic(err)
	}
	ov.Search("H")
	if err := ov.Run(); err != nil {
		panic(err)
	}
}
