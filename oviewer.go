// Package oviewer provides a pager for terminals.
package oviewer

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
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

func (root *root) setWrapHeaderLen() {
	m := root.Model
	root.wrapHeaderLen = 0
	for y := 0; y < m.HeaderLen; y++ {
		root.wrapHeaderLen += 1 + (len(m.getContents(y)) / m.vWidth)
	}
}

func (root *root) HeaderLen() int {
	if root.Model.WrapMode {
		return root.wrapHeaderLen
	}
	return root.Model.HeaderLen
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

func contentsToStr(contents []content) (string, map[int]int) {
	buf := make([]rune, 0)
	cIndex := make(map[int]int)
	byteLen := 0
	for n := range contents {
		if contents[n].width > 0 {
			cIndex[byteLen] = n
			buf = append(buf, contents[n].mainc)
			b := string(contents[n].mainc)
			byteLen += len(b)
		}
	}
	s := string(buf)
	cIndex[len(s)] = len(contents)
	return s, cIndex
}

func doReverse(cp *[]content, searchWord string) {
	contents := *cp
	for n := range contents {
		contents[n].style = contents[n].style.Reverse(false)
	}
	if searchWord == "" {
		return
	}
	s, cIndex := contentsToStr(contents)
	for i := strings.Index(s, searchWord); i >= 0; {
		start := cIndex[i]
		end := cIndex[i+len(searchWord)]
		for ci := start; ci < end; ci++ {
			contents[ci].style = contents[ci].style.Reverse(true)
		}
		j := strings.Index(s[i+1:], searchWord)
		if j >= 0 {
			i += j + 1
		} else {
			break
		}
	}
}

func (root *root) setContentString(vx int, vy int, contents []content, style tcell.Style) {
	screen := root.Screen
	for x, content := range contents {
		screen.SetContent(vx+x, vy, content.mainc, content.combc, style)
	}
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

func (root *root) main() {
	screen := root.Screen

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		<-c
		root.Quit()
	}()

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
	default:
		readers := make([]io.Reader, 0)
		for _, fileName := range args {
			r, err := os.Open(fileName)
			if err != nil {
				return err
			}
			readers = append(readers, uncompressedReader(r))
			reader = io.MultiReader(readers...)
		}
	}
	m.ReadAll(reader)

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
	root.minStartPos = -10
	defer func() {
		root.Screen.Fini()
		if root.Model.PostWrite {
			for i := 0; i < m.vHight; i++ {
				if m.lineNum+i >= len(m.buffer) {
					break
				}
				fmt.Println(m.buffer[m.lineNum+i])
			}
		}
	}()

	screen.Clear()
	root.Sync()
	root.main()

	return nil
}

func New() *root {
	root := &root{}
	root.Model = NewModel()

	return root
}
