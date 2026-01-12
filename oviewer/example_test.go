package oviewer_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/noborus/ov/oviewer"
)

const cwd = ".."

var testdata = filepath.Join(cwd, "testdata")

// fakeScreen returns a fake screen.
func fakeScreen() (tcell.Screen, error) {
	// width, height := 80, 25
	mt := vt.NewMockTerm(vt.MockOptSize{X: 80, Y: 25})
	return tcell.NewTerminfoScreenFromTty(mt)
}

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
	if err := doc.ControlReader(bytes.NewBufferString(s), nil); err != nil {
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
	if err := doc.ControlReader(bytes.NewBufferString(s), nil); err != nil {
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

func TestExample_HeaderOption(t *testing.T) {
	oviewer.SetTcellNewScreen(fakeScreen)
	defer func() {
		oviewer.SetTcellNewScreen(tcell.NewScreen)
	}()
	ov, err := oviewer.Open(filepath.Join(testdata, "test.txt"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	ov.Config.General.SetHeader(2)
	if ov.Config.General.Header == nil || *ov.Config.General.Header != 2 {
		t.Errorf("Config.General.SetHeader did not set header correctly: got %v", ov.Config.General.Header)
	}
}

func TestExample_DocumentGeneral(t *testing.T) {
	oviewer.SetTcellNewScreen(fakeScreen)
	defer func() {
		oviewer.SetTcellNewScreen(tcell.NewScreen)
	}()
	ov, err := oviewer.Open(filepath.Join(testdata, "test.txt"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	ov.Doc.General.SetHeader(3)
	if ov.Doc.General.Header == nil || *ov.Doc.General.Header != 3 {
		t.Errorf("Doc.General.SetHeader did not set header correctly: got %v", ov.Doc.General.Header)
	}
}

func TestExample_FollowMode(t *testing.T) {
	oviewer.SetTcellNewScreen(fakeScreen)
	defer func() {
		oviewer.SetTcellNewScreen(tcell.NewScreen)
	}()
	ov, err := oviewer.Open(filepath.Join(testdata, "test.txt"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	ov.Config.General.SetFollowMode(true)
	if ov.Config.General.FollowMode == nil || *ov.Config.General.FollowMode != true {
		t.Errorf("Options.SetFollowMode did not set follow mode correctly: got %v", ov.Config.General.FollowMode)
	}
	ov.Doc.General.SetFollowMode(false)
	if ov.Doc.General.FollowMode == nil || *ov.Doc.General.FollowMode != false {
		t.Errorf("Doc.General.SetFollowMode did not set follow mode correctly: got %v", ov.Doc.General.FollowMode)
	}
}
