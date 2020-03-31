// Package oviewer provides a pager for terminals.
package oviewer

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgraph-io/ristretto"
	"github.com/gdamore/tcell"
	"golang.org/x/crypto/ssh/terminal"
)

type root struct {
	Model         *Model
	wrapHeaderLen int
	bottomPos     int
	statusPos     int
	fileName      string
	mode          Mode
	input         string
	message       string

	minStartPos int

	tcell.Screen
}

type Mode int

const (
	normal Mode = iota
	search
	previous
	goline
	header
)

var Debug bool

func (root *root) PrepareView() {
	m := root.Model
	screen := root.Screen
	m.vWidth, m.vHight = screen.Size()
	root.setWrapHeaderLen()
	root.statusPos = m.vHight - 1
}

func (root *root) HeaderLen() int {
	if root.Model.WrapMode {
		return root.wrapHeaderLen
	}
	return root.Model.HeaderLen
}

func (root *root) setWrapHeaderLen() {
	m := root.Model
	root.wrapHeaderLen = 0
	for y := 0; y < m.HeaderLen; y++ {
		root.wrapHeaderLen += 1 + (len(m.getContents(y)) / m.vWidth)
	}
}

func (root *root) bottomLineNum(num int) int {
	m := root.Model
	if !m.WrapMode {
		if num <= m.vHight {
			return 0
		}
		return num - (m.vHight - root.Model.HeaderLen) + 1
	}

	for y := m.vHight - root.wrapHeaderLen; y > 0; {
		y -= 1 + (len(m.getContents(num)) / m.vWidth)
		num--
	}
	num++
	return num
}

type eventAppQuit struct {
	tcell.EventTime
}

func (root *root) Quit() {
	ev := &eventAppQuit{}
	ev.SetEventNow()
	go func() { root.Screen.PostEventWait(ev) }()
}

func (root *root) Resize() {
	root.Sync()
}

func (root *root) Sync() {
	root.PrepareView()
	root.Draw()
}

// main is manages and executes events in the main routine.
func (root *root) main() {
	screen := root.Screen
loop:
	for {
		root.Draw()
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *eventAppQuit:
			break loop
		case *tcell.EventKey:
			root.HandleEvent(ev)
		case *tcell.EventResize:
			root.Resize()
		}
	}
}

// Run is reads the file(or stdin) and starts
// the terminal pager.
func (root *root) Run(args []string) error {
	m := root.Model
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return err
	}
	m.cache = cache

	var reader io.Reader
	fileName := ""
	switch len(args) {
	case 0:
		if terminal.IsTerminal(0) {
			return fmt.Errorf("missing filename")
		}
		fileName = "(STDIN)"
		reader = uncompressedReader(os.Stdin)
	case 1:
		fileName = args[0]
		r, err := os.Open(fileName)
		if err != nil {
			return err
		}
		reader = uncompressedReader(r)
		defer r.Close()
	default:
		readers := make([]io.Reader, 0)
		for _, fileName := range args {
			r, err := os.Open(fileName)
			if err != nil {
				return err
			}
			readers = append(readers, uncompressedReader(r))
			reader = io.MultiReader(readers...)
			defer r.Close()
		}
	}
	err = m.ReadAll(reader)
	if err != nil {
		return err
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	err = screen.Init()
	if err != nil {
		return err
	}

	root.Screen = screen
	root.fileName = fileName
	root.Model = m
	defer root.Screen.Fini()

	screen.Clear()
	root.Sync()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		<-c
		root.Screen.Fini()
		os.Exit(1)
	}()

	root.main()

	return nil
}

func (root *root) PostWrite() {
	m := root.Model
	for i := 0; i < m.vHight; i++ {
		if m.lineNum+i >= len(m.buffer) {
			break
		}
		fmt.Println(m.buffer[m.lineNum+i])
	}
}

// New returns the entire structure of oviewer.
func New() *root {
	root := &root{}
	root.Model = NewModel()

	root.minStartPos = -10

	return root
}
