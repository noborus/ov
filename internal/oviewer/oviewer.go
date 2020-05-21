// Package oviewer provides a pager for terminals.
package oviewer

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/gdamore/tcell"
	"golang.org/x/crypto/ssh/terminal"
)

// The Root structure contains information about the drawing.
type Root struct {
	Header          int
	TabWidth        int
	AfterWrite      bool
	WrapMode        bool
	QuitSmall       bool
	CaseSensitive   bool
	AlternateRows   bool
	ColumnMode      bool
	ColumnDelimiter string
	LineNumMode     bool

	Model         *Model
	wrapHeaderLen int
	startPos      int
	bottomPos     int
	statusPos     int
	fileName      string
	mode          Mode
	input         string
	inputRegexp   *regexp.Regexp
	cursorX       int
	columnNum     int
	message       string

	minStartPos int

	HeaderStyle    tcell.Style
	ColorAlternate tcell.Color

	tcell.Screen
}

// Mode represents the state of the input.
type Mode int

const (
	normal Mode = iota
	search
	previous
	goline
	header
	delimiter
	tabWidth
)

//Debug represents whether to enable the debug output.
var Debug bool

// New returns the entire structure of oviewer.
func New() *Root {
	root := &Root{}
	root.Model = NewModel()

	root.minStartPos = -10
	root.HeaderStyle = tcell.StyleDefault.Bold(true)
	root.ColorAlternate = tcell.ColorGray
	root.ColumnDelimiter = ""
	root.columnNum = 0
	root.startPos = 0
	return root
}

// PrepareView prepares when the screen size is changed.
func (root *Root) PrepareView() {
	m := root.Model
	screen := root.Screen
	m.vWidth, m.vHight = screen.Size()
	root.setWrapHeaderLen()
	root.statusPos = m.vHight - 1
}

// HeaderLen returns the actual number of lines in the header.
func (root *Root) HeaderLen() int {
	if root.WrapMode {
		return root.wrapHeaderLen
	}
	return root.Header
}

func (root *Root) setWrapHeaderLen() {
	m := root.Model
	root.wrapHeaderLen = 0
	for y := 0; y < root.Header; y++ {
		root.wrapHeaderLen += 1 + (len(m.GetContents(y, root.TabWidth)) / m.vWidth)
	}
}

func (root *Root) bottomLineNum(num int) int {
	m := root.Model
	if !root.WrapMode {
		if num <= m.vHight {
			return 0
		}
		return num - (m.vHight - root.Header) + 1
	}

	for y := m.vHight - root.wrapHeaderLen; y > 0; {
		y -= 1 + (len(m.GetContents(num, root.TabWidth)) / m.vWidth)
		num--
	}
	num++
	return num
}

func (root *Root) toggleWrap() {
	if root.WrapMode {
		root.WrapMode = false
	} else {
		root.WrapMode = true
		root.Model.x = 0
	}
	root.setWrapHeaderLen()
}

func (root *Root) toggleColumnMode() {
	if root.ColumnMode {
		root.ColumnMode = false
	} else {
		root.ColumnMode = true
	}
}

func (root *Root) toggleAlternateRows() {
	root.Model.ClearCache()
	if root.AlternateRows {
		root.AlternateRows = false
	} else {
		root.AlternateRows = true
	}
}

func (root *Root) toggleLineNumMode() {
	if root.LineNumMode {
		root.LineNumMode = false
	} else {
		root.LineNumMode = true
	}
	root.updateEndNum()
}

func (root *Root) setMode(mode Mode) {
	root.mode = mode
	root.input = ""
	root.cursorX = 0
	root.ShowCursor(root.cursorX, root.statusPos)
}

// GoLine will move to the specified line.
func (root *Root) GoLine() {
	lineNum, err := strconv.Atoi(root.input)
	if err != nil {
		return
	}
	root.input = ""
	root.moveNum(lineNum - root.Header)
}

// SetHeader sets the number of lines in the header.
func (root *Root) SetHeader() {
	line, _ := strconv.Atoi(root.input)
	if line >= 0 && line <= root.Model.vHight-1 {
		if root.Header != line {
			root.Header = line
			root.setWrapHeaderLen()
			root.Model.ClearCache()
		}
	}
	root.input = ""
}

// SetDelimiter sets the delimiter string.
func (root *Root) SetDelimiter() {
	root.ColumnDelimiter = root.input
	root.input = ""
}

// SetTabWidth sets the tab width.
func (root *Root) SetTabWidth() {
	width, _ := strconv.Atoi(root.input)
	root.TabWidth = width
	root.input = ""
}

type eventAppQuit struct {
	tcell.EventTime
}

// Quit executes a quit event.
func (root *Root) Quit() {
	ev := &eventAppQuit{}
	ev.SetEventNow()
	go func() { root.Screen.PostEventWait(ev) }()
}

type eventTimer struct {
	tcell.EventTime
}

// runOnTime runs at time.
func (root *Root) runOnTime() {
	ev := &eventTimer{}
	ev.SetEventNow()
	go func() { root.Screen.PostEventWait(ev) }()
}

// Resize is a wrapper function that calls sync.
func (root *Root) Resize() {
	root.Sync()
}

// Sync redraws the whole thing.
func (root *Root) Sync() {
	root.PreparelineNum()
	root.PrepareView()
	root.Draw()
}

func (root *Root) countTimer() {
	timer := time.NewTicker(time.Millisecond * 500)
loop:
	for {
		<-timer.C
		root.runOnTime()
		if root.Model.eof {
			break loop
		}
	}
	timer.Stop()
}

func (root *Root) PreparelineNum() {
	root.startPos = 0
	if root.LineNumMode {
		root.startPos = len(fmt.Sprintf("%d", root.Model.BufEndNum())) + 1
	}
}

func (root *Root) updateEndNum() {
	root.PreparelineNum()
	root.statusDraw()
}

// main is manages and executes events in the main routine.
func (root *Root) main() {
	screen := root.Screen
	go root.countTimer()
loop:
	for {
		root.Draw()
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *eventTimer:
			root.updateEndNum()
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
func (root *Root) Run(args []string) error {
	m := root.Model
	err := m.NewCache()
	if err != nil {
		return err
	}

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
	defer root.Screen.Fini()

	screen.Clear()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		<-c
		root.Screen.Fini()
		os.Exit(1)
	}()

	root.Sync()
	// Exit if fits on screen
	if root.QuitSmall && root.contentsSmall() {
		root.AfterWrite = true
		return nil
	}

	root.main()
	return nil
}

func (root *Root) contentsSmall() bool {
	root.PrepareView()
	m := root.Model
	hight := 0
	for y := 0; y < m.BufEndNum(); y++ {
		hight += 1 + (len(m.GetContents(y, root.TabWidth)) / m.vWidth)
		if hight > m.vHight {
			return false
		}
	}
	return true
}

// WriteOriginal writes to the original terminal.
func (root *Root) WriteOriginal() {
	m := root.Model
	for i := 0; i < m.vHight-1; i++ {
		if m.lineNum+i >= m.BufLen() {
			break
		}
		fmt.Println(m.GetLine(m.lineNum + i))
	}
}
