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

	mode        Mode
	input       string
	inputRegexp *regexp.Regexp
	EventInput  EventInput

	SearchCList    *Candidate
	GoCList        *Candidate
	DelimiterCList *Candidate
	TabWidthCList  *Candidate

	cursorX   int
	columnNum int
	message   string

	minStartPos int

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

type EventInput interface {
	Prompt() string
	Confirm(i string) tcell.Event
	Up(i string) string
	Down(i string) string
}

type Candidate struct {
	list []string
	p    int
}

func (root *Root) setMode(mode Mode) {
	root.mode = mode
	root.input = ""
	root.cursorX = 0
	switch mode {
	case search:
		root.EventInput = &SearchInput{
			clist: root.SearchCList,
		}
	case previous:
		root.EventInput = &PreviousInput{
			clist: root.SearchCList,
		}
	case goline:
		root.EventInput = &GotoInput{
			clist: root.GoCList,
		}
	case header:
		root.input = strconv.Itoa(root.Header)
		root.cursorX = runeWidth(root.input)
		root.EventInput = &HeaderInput{}
	case delimiter:
		root.input = root.ColumnDelimiter
		root.cursorX = runeWidth(root.input)
		root.EventInput = &DelimiterInput{
			clist: root.DelimiterCList,
		}
	case tabWidth:
		root.EventInput = &TABWidthInput{
			clist: root.TabWidthCList,
		}
	default:
		root.EventInput = &NormalInput{}
	}
	root.ShowCursor(root.cursorX, root.statusPos)
}

type NormalInput struct {
	tcell.EventTime
}

func (n *NormalInput) Prompt() string {
	return ""
}

func (n *NormalInput) Confirm(i string) tcell.Event {
	return nil
}

func (n *NormalInput) Up(i string) string {
	return ""
}

func (n *NormalInput) Down(i string) string {
	return ""
}

type SearchInput struct {
	clist *Candidate
	tcell.EventTime
}

func (s *SearchInput) Prompt() string {
	return "/"
}

func (s *SearchInput) Confirm(i string) tcell.Event {
	s.clist.list = append(s.clist.list, i)
	s.SetEventNow()
	return s
}

func (s *SearchInput) Up(i string) string {
	if len(s.clist.list) < 0 {
		return ""
	}
	if len(s.clist.list) > s.clist.p+1 {
		s.clist.p++
		return s.clist.list[s.clist.p]
	}
	s.clist.p = 0
	return s.clist.list[s.clist.p]
}

func (s *SearchInput) Down(i string) string {
	if len(s.clist.list) < 0 {
		return ""
	}
	if s.clist.p > 0 {
		s.clist.p--
		return s.clist.list[s.clist.p]
	}
	s.clist.p = len(s.clist.list) - 1
	return s.clist.list[s.clist.p]
}

type PreviousInput struct {
	clist *Candidate
	tcell.EventTime
}

func (p *PreviousInput) Prompt() string {
	return "?"
}

func (p *PreviousInput) Confirm(i string) tcell.Event {
	p.clist.list = append(p.clist.list, i)
	p.SetEventNow()
	return p
}

func (p *PreviousInput) Up(i string) string {
	if len(p.clist.list) < 0 {
		return ""
	}
	if len(p.clist.list) > p.clist.p+1 {
		p.clist.p++
		return p.clist.list[p.clist.p]
	}
	p.clist.p = 0
	return p.clist.list[p.clist.p]
}

func (p *PreviousInput) Down(i string) string {
	if len(p.clist.list) < 0 {
		return ""
	}
	if p.clist.p > 0 {
		p.clist.p--
		return p.clist.list[p.clist.p]
	}
	p.clist.p = len(p.clist.list) - 1
	return p.clist.list[p.clist.p]
}

type GotoInput struct {
	clist *Candidate
	tcell.EventTime
}

func (g *GotoInput) Prompt() string {
	return "Goto line:"
}

func (g *GotoInput) Confirm(i string) tcell.Event {
	g.clist.list = append(g.clist.list, i)
	g.SetEventNow()
	return g
}

func (g *GotoInput) Up(i string) string {
	if len(g.clist.list) > g.clist.p+1 {
		g.clist.p++
		return g.clist.list[g.clist.p]
	}
	g.clist.p = 0
	return g.clist.list[g.clist.p]
}

func (g *GotoInput) Down(i string) string {
	if g.clist.p > 0 {
		g.clist.p--
		return g.clist.list[g.clist.p]
	}
	g.clist.p = len(g.clist.list) - 1
	return g.clist.list[g.clist.p]
}

type HeaderInput struct {
	tcell.EventTime
}

func (h *HeaderInput) Prompt() string {
	return "Header length:"
}

func (h *HeaderInput) Confirm(i string) tcell.Event {
	h.SetEventNow()
	return h
}

func (h *HeaderInput) Up(i string) string {
	n, err := strconv.Atoi(i)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n + 1)
}

func (h *HeaderInput) Down(i string) string {
	n, err := strconv.Atoi(i)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n - 1)
}

type DelimiterInput struct {
	clist *Candidate
	tcell.EventTime
}

func (d *DelimiterInput) Prompt() string {
	return "Delimiter:"
}

func (d *DelimiterInput) Confirm(i string) tcell.Event {
	d.clist.list = append(d.clist.list, i)
	d.SetEventNow()
	return d
}

func (d *DelimiterInput) Up(i string) string {
	if d.clist.p > 0 {
		d.clist.p--
		return d.clist.list[d.clist.p]
	}
	d.clist.p = len(d.clist.list) - 1
	return d.clist.list[d.clist.p]
}

func (d *DelimiterInput) Down(i string) string {
	if len(d.clist.list) > d.clist.p+1 {
		d.clist.p++
		return d.clist.list[d.clist.p]
	}
	d.clist.p = 0
	return d.clist.list[d.clist.p]
}

type TABWidthInput struct {
	clist *Candidate
	tcell.EventTime
}

func (t *TABWidthInput) Prompt() string {
	return "TAB width:"
}

func (t *TABWidthInput) Confirm(i string) tcell.Event {
	t.clist.list = append(t.clist.list, i)
	t.SetEventNow()
	return t
}

func (t *TABWidthInput) Up(i string) string {
	if t.clist.p > 0 {
		t.clist.p--
		return t.clist.list[t.clist.p]
	}
	t.clist.p = len(t.clist.list) - 1
	return t.clist.list[t.clist.p]
}

func (t *TABWidthInput) Down(i string) string {
	if len(t.clist.list) > t.clist.p+1 {
		t.clist.p++
		return t.clist.list[t.clist.p]
	}
	t.clist.p = 0
	return t.clist.list[t.clist.p]
}

//Debug represents whether to enable the debug output.
var Debug bool

var (
	HeaderStyle     = tcell.StyleDefault.Bold(true)
	ColorAlternate  = tcell.ColorGray
	OverStrikeStyle = tcell.StyleDefault.Bold(true)
	OverLineStyle   = tcell.StyleDefault.Underline(true)
)

// New returns the entire structure of oviewer.
func New() *Root {
	root := &Root{}
	root.Model = NewModel()

	root.TabWidth = 8
	root.minStartPos = -10
	root.ColumnDelimiter = ""
	root.columnNum = 0
	root.startPos = 0
	root.EventInput = &NormalInput{}
	root.GoCList = &Candidate{
		list: []string{
			"1000",
			"100",
			"10",
			"0",
		},
	}
	root.DelimiterCList = &Candidate{
		list: []string{
			"\t",
			"|",
			",",
		},
	}
	root.TabWidthCList = &Candidate{
		list: []string{
			"3",
			"2",
			"4",
			"8",
		},
	}
	root.SearchCList = &Candidate{
		list: []string{},
	}
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
	root.preparelineNum()
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

func (root *Root) preparelineNum() {
	root.startPos = 0
	if root.LineNumMode {
		root.startPos = len(fmt.Sprintf("%d", root.Model.BufEndNum())) + 1
	}
}

func (root *Root) updateEndNum() {
	root.preparelineNum()
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
			root.Search()
		case *PreviousInput:
			root.BackSearch()
		case *GotoInput:
			root.GoLine()
		case *HeaderInput:
			root.SetHeader()
		case *DelimiterInput:
			root.SetDelimiter()
		case *TABWidthInput:
			root.SetTabWidth()
		case *tcell.EventKey:
			root.HandleEvent(ev)
		case *tcell.EventResize:
			root.Resize()
		}
	}
}

// Search is a Up search.
func (root *Root) Search() {
	root.inputRegexp = regexpComple(root.input, root.CaseSensitive)
	root.postSearch(root.search(root.Model.lineNum))
}

// BackSearch reverse search.
func (root *Root) BackSearch() {
	root.inputRegexp = regexpComple(root.input, root.CaseSensitive)
	root.postSearch(root.backSearch(root.Model.lineNum))
}

// GoLine will move to the specified line.
func (root *Root) GoLine() {
	lineNum, err := strconv.Atoi(root.input)
	root.input = ""
	if err != nil {
		return
	}
	root.moveNum(lineNum - root.Header)
}

// SetHeader sets the number of lines in the header.
func (root *Root) SetHeader() {
	line, err := strconv.Atoi(root.input)
	root.input = ""
	if err != nil {
		return
	}
	if line < 0 || line > root.Model.vHight-1 {
		return
	}
	if root.Header == line {
		return
	}
	root.Header = line
	root.setWrapHeaderLen()
	root.Model.ClearCache()
}

// SetDelimiter sets the delimiter string.
func (root *Root) SetDelimiter() {
	root.ColumnDelimiter = root.input
	root.input = ""
}

// SetTabWidth sets the tab width.
func (root *Root) SetTabWidth() {
	width, err := strconv.Atoi(root.input)
	root.input = ""
	if err != nil {
		return
	}
	if root.TabWidth == width {
		return
	}
	root.TabWidth = width
	root.Model.ClearCache()
}
