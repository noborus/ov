// Package oviewer provides a pager for terminals.
package oviewer

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gdamore/tcell"
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

	// lineNum is the starting position of the current y.
	lineNum int
	// yy represents the number of wrapped lines.
	yy int
	// x is the starting position of the current x.
	x int
	// startX is the start position of x.
	startX int
	// columnNum is the number of columns.
	columnNum int

	// message is the message to display.
	message string

	// vWidth represents the screen width.
	vWidth int
	// vHight represents the screen height.
	vHight int

	// wrapHeaderLen is the actual header length when wrapped.
	wrapHeaderLen int
	// bottomPos is the position of the last line displayed.
	bottomPos int
	// statusPos is the position of the status line.
	statusPos int
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

// New return the structure of oviewer.
func New(m *Model) (*Root, error) {
	root := &Root{
		Config: Config{
			ColumnDelimiter: "",
			TabWidth:        8,
		},
		minStartX: -10,
		columnNum: 0,
		startX:    0,
	}

	root.Model = m
	root.Input = NewInput()

	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err = screen.Init(); err != nil {
		return nil, err
	}
	root.Screen = screen

	return root, nil
}

// Open reads the file named of the argument and return the structure of oviewer.
func Open(fileNames []string) (*Root, error) {
	m, err := NewModel()
	if err != nil {
		return nil, err
	}
	err = m.ReadFile(fileNames)
	if err != nil {
		return nil, err
	}

	return New(m)
}

// Run starts the terminal pager.
func (root *Root) Run() error {
	defer root.Screen.Fini()
	root.setGlobalStyle()
	root.Screen.Clear()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		<-c
		root.Screen.Fini()
		os.Exit(1)
	}()

	root.viewSync()
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

// prepareView prepares when the screen size is changed.
func (root *Root) prepareView() {
	screen := root.Screen
	root.vWidth, root.vHight = screen.Size()
	root.setWrapHeaderLen()
	root.statusPos = root.vHight - 1
}

// contentsSmall returns with bool whether the file to display fits on the screen.
func (root *Root) contentsSmall() bool {
	root.prepareView()
	m := root.Model
	hight := 0
	for y := 0; y < m.BufEndNum(); y++ {
		hight += 1 + (len(m.getContents(y, root.TabWidth)) / root.vWidth)
		if hight > root.vHight {
			return false
		}
	}
	return true
}

// WriteOriginal writes to the original terminal.
func (root *Root) WriteOriginal() {
	m := root.Model
	for i := 0; i < root.vHight-1; i++ {
		n := root.lineNum + i
		if n >= m.BufEndNum() {
			break
		}
		fmt.Println(m.GetLine(n))
	}
}

// headerLen returns the actual number of lines in the header.
func (root *Root) headerLen() int {
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
		root.wrapHeaderLen += 1 + (len(m.getContents(y, root.TabWidth)) / root.vWidth)
	}
}

// bottomLineNum returns the display start line
// when the last line number as an argument.
func (root *Root) bottomLineNum(num int) int {
	m := root.Model
	if !root.WrapMode {
		if num <= root.vHight {
			return 0
		}
		return num - (root.vHight - root.Header) + 1
	}

	for y := root.vHight - root.wrapHeaderLen; y > 0; {
		y -= 1 + (len(m.getContents(num, root.TabWidth)) / root.vWidth)
		num--
	}
	num++
	return num
}

// toggleWrapMode toggles wrapMode each time it is called.
func (root *Root) toggleWrapMode() {
	root.WrapMode = !root.WrapMode
	root.x = 0
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

// resize is a wrapper function that calls sync.
func (root *Root) resize() {
	root.viewSync()
}

// Sync redraws the whole thing.
func (root *Root) viewSync() {
	root.prepareStartX()
	root.prepareView()
	root.draw()
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

// goLine will move to the specified line.
func (root *Root) goLine(input string) {
	lineNum, err := strconv.Atoi(input)
	if err != nil {
		root.message = ErrInvalidNumber.Error()
		return
	}

	root.moveLine(lineNum - root.Header - 1)
	root.message = fmt.Sprintf("Moved to line %d", lineNum)
}

// markLineNum stores the specified number of lines.
func (root *Root) markLineNum(lineNum int) {
	s := strconv.Itoa(lineNum + 1)
	root.Input.GoCandidate.list = toLast(root.Input.GoCandidate.list, s)
	root.Input.GoCandidate.p = 0
	root.message = fmt.Sprintf("Marked to line %d", lineNum)
}

// setHeader sets the number of lines in the header.
func (root *Root) setHeader(input string) {
	lineNum, err := strconv.Atoi(input)
	if err != nil {
		root.message = ErrInvalidNumber.Error()
		return
	}
	if lineNum < 0 || lineNum > root.vHight-1 {
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

// setDelimiter sets the delimiter string.
func (root *Root) setDelimiter(input string) {
	root.ColumnDelimiter = input
	root.message = fmt.Sprintf("Set delimiter %s", input)
}

// setTabWidth sets the tab width.
func (root *Root) setTabWidth(input string) {
	width, err := strconv.Atoi(input)
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
