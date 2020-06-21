// Package oviewer provides a pager for terminals.
package oviewer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gdamore/tcell"
	"golang.org/x/crypto/ssh/terminal"
)

// The Root structure contains information about the drawing.
type Root struct {
	// tcell.Screen is the root screen.
	tcell.Screen
	// Config contains settings that determine the behavior of ov.
	Config
	// Model contains the model of ov
	Model *Model
	// Input contains the input mode.
	Input *Input
	// fileName is the file name to display.
	fileName string
	// message is the message to display.
	message string
	// wrapHeaderLen is the actual header length when wrapped.
	wrapHeaderLen int
	// bottomPos is the position of the last line displayed.
	bottomPos int
	// statusPos is the position of the status line.
	statusPos int
	// columnNum is the number of columns.
	columnNum int
	// startX is the start position of x.
	startX int
	// minStartX is the minimum start position of x.
	minStartX int
}

// Config represents the settings of ov.
type Config struct {
	// Alternating background color.
	ColorAlternate string
	// Header color.
	ColorHeader string
	// OverStrike color.
	ColorOverStrike string
	// OverLine color.
	ColorOverLine string

	// Column Delimiter
	ColumnDelimiter string
	// Wrap is Wrap mode.
	WrapMode bool
	// TabWidth is tab stop num.
	TabWidth int
	// HeaderLen is number of header rows to be fixed.
	Header int
	// AfterWrite writes the current screen on exit.
	AfterWrite bool
	// QuiteSmall Quit if the output fits on one screen.
	QuitSmall bool
	// CaseSensitive is case-sensitive if true
	CaseSensitive bool
	// Color to alternate rows
	AlternateRows bool
	// Column mode
	ColumnMode bool
	// Line Number
	LineNumMode bool
	//Debug represents whether to enable the debug output.
	Debug bool
}

var (
	// HeaderStyle represents the style of the header.
	HeaderStyle = tcell.StyleDefault.Bold(true)
	// ColorAlternate represents alternating colors.
	ColorAlternate = tcell.ColorGray
	// OverStrikeStyle represents the overstrike style.
	OverStrikeStyle = tcell.StyleDefault.Bold(true)
	// OverLineStyle represents the overline underline style.
	OverLineStyle = tcell.StyleDefault.Underline(true)
)

var (
	// ErrOutOfRange indicates that value is out of range.
	ErrOutOfRange = errors.New("out of range")
	// ErrFatalCache indicates that the cache value had a fatal error.
	ErrFatalCache = errors.New("fatal error in cache value")
	// ErrMissingFile indicates that the file does not exist.
	ErrMissingFile = errors.New("missing filename")
	// ErrNotFound indicates not found.
	ErrNotFound = errors.New("not found")
	// ErrInvalidNumber indicates an invalid number.
	ErrInvalidNumber = errors.New("invalid number")
)

// New returns the entire structure of oviewer.
func New() *Root {
	root := &Root{}

	root.Model = NewModel()
	root.Input = NewInput()

	root.TabWidth = 8
	root.minStartX = -10
	root.ColumnDelimiter = ""
	root.columnNum = 0
	root.startX = 0
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

// Run is reads the file(or stdin) and starts
// the terminal pager.
func (root *Root) Run(args []string) error {
	err := root.Model.NewCache()
	if err != nil {
		return err
	}

	root.setGlobalStyle()

	var reader io.Reader
	fileName := ""
	switch len(args) {
	case 0:
		if terminal.IsTerminal(0) {
			return ErrMissingFile
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
	err = root.Model.ReadAll(reader)
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

// setGlobalStyle sets some styles that are determined by the settings.
func (root *Root) setGlobalStyle() {
	if root.ColorAlternate != "" {
		ColorAlternate = tcell.GetColor(root.ColorAlternate)
	}
	if root.ColorHeader != "" {
		HeaderStyle = HeaderStyle.Foreground(tcell.GetColor(root.ColorHeader))
	}
	if root.ColorOverStrike != "" {
		OverStrikeStyle = OverStrikeStyle.Foreground(tcell.GetColor(root.ColorOverStrike))
	}
	if root.ColorOverLine != "" {
		OverLineStyle = OverLineStyle.Foreground(tcell.GetColor(root.ColorOverLine))
	}
}

// contentsSmall returns with bool whether the file to display fits on the screen.
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

// HeaderLen returns the actual number of lines in the header.
func (root *Root) HeaderLen() int {
	if root.WrapMode {
		return root.wrapHeaderLen
	}
	return root.Header
}

// setWrapHeaderLen sets the value in wrapHeaderLen.
func (root *Root) setWrapHeaderLen() {
	m := root.Model
	root.wrapHeaderLen = 0
	for y := 0; y < root.Header; y++ {
		root.wrapHeaderLen += 1 + (len(m.GetContents(y, root.TabWidth)) / m.vWidth)
	}
}

// bottomLineNum returns the display start line
// when the last line number as an argument.
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

// toggleWrapMode toggles wrapMode each time it is called.
func (root *Root) toggleWrapMode() {
	root.WrapMode = !root.WrapMode
	root.Model.x = 0
	root.setWrapHeaderLen()
}

//  toggleColumnMode toggles ColumnMode each time it is called.
func (root *Root) toggleColumnMode() {
	root.ColumnMode = !root.ColumnMode
}

// toggleAlternateRows toggles the AlternateRows each time it is called.
func (root *Root) toggleAlternateRows() {
	root.Model.ClearCache()
	root.AlternateRows = !root.AlternateRows
}

// toggleLineNumMode toggles LineNumMode every time it is called.
func (root *Root) toggleLineNumMode() {
	root.LineNumMode = !root.LineNumMode
	root.updateEndNum()
}

// eventAppQuit represents a quit event.
type eventAppQuit struct {
	tcell.EventTime
}

// Quit executes a quit event.
func (root *Root) Quit() {
	ev := &eventAppQuit{}
	ev.SetEventNow()
	go func() { root.Screen.PostEventWait(ev) }()
}

// eventTimer represents a timer event.
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
	root.prepareStartX()
	root.PrepareView()
	root.Draw()
}

// countTimer fires events periodically until it reaches EOF.
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

// prepareStartX prepares startX.
func (root *Root) prepareStartX() {
	root.startX = 0
	if root.LineNumMode {
		root.startX = len(fmt.Sprintf("%d", root.Model.BufEndNum())) + 1
	}
}

// updateEndNum updates the last line number.
func (root *Root) updateEndNum() {
	root.prepareStartX()
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
		case *SearchInput:
			root.Search(root.Input.value)
		case *BackSearchInput:
			root.BackSearch(root.Input.value)
		case *GotoInput:
			root.GoLine(root.Input.value)
		case *HeaderInput:
			root.SetHeader(root.Input.value)
		case *DelimiterInput:
			root.SetDelimiter(root.Input.value)
		case *TABWidthInput:
			root.SetTabWidth(root.Input.value)
		case *tcell.EventKey:
			root.message = ""
			if root.Input.mode == normal {
				root.DefaultKeyEvent(ev)
			} else {
				root.InputKeyEvent(ev)
			}
		case *tcell.EventResize:
			root.Resize()
		}
	}
}

// GoLine will move to the specified line.
func (root *Root) GoLine(input string) {
	lineNum, err := strconv.Atoi(input)
	root.Input.value = ""
	if err != nil {
		root.message = ErrInvalidNumber.Error()
		return
	}
	root.moveNum(lineNum - root.Header - 1)
	root.message = fmt.Sprintf("Moved to line %d", lineNum)
}

// MarkLineNum stores the specified number of lines.
func (root *Root) MarkLineNum(lineNum int) {
	s := strconv.Itoa(lineNum + 1)
	root.Input.GoCList.list = toLast(root.Input.GoCList.list, s)
	root.Input.GoCList.p = 0
	root.message = fmt.Sprintf("Marked to line %d", lineNum)
}

// SetHeader sets the number of lines in the header.
func (root *Root) SetHeader(input string) {
	lineNum, err := strconv.Atoi(input)
	root.Input.value = ""
	if err != nil {
		root.message = ErrInvalidNumber.Error()
		return
	}
	if lineNum < 0 || lineNum > root.Model.vHight-1 {
		root.message = ErrOutOfRange.Error()
		return
	}
	if root.Header == lineNum {
		return
	}
	root.Header = lineNum
	root.message = fmt.Sprintf("Set Header %d", lineNum)
	root.setWrapHeaderLen()
	root.Model.ClearCache()
}

// SetDelimiter sets the delimiter string.
func (root *Root) SetDelimiter(input string) {
	root.ColumnDelimiter = input
	root.Input.value = ""
}

// SetTabWidth sets the tab width.
func (root *Root) SetTabWidth(input string) {
	width, err := strconv.Atoi(input)
	root.Input.value = ""
	if err != nil {
		root.message = ErrInvalidNumber.Error()
		return
	}
	if root.TabWidth == width {
		return
	}
	root.TabWidth = width
	root.message = fmt.Sprintf("Set tab width %d", width)
	root.Model.ClearCache()
}
