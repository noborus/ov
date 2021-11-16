package oviewer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"code.rocketnine.space/tslocum/cbind"
	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v2"
)

// Root structure contains information about the drawing.
type Root struct {
	// tcell.Screen is the root screen.
	tcell.Screen
	// Config contains settings that determine the behavior of ov.
	Config

	// Doc contains the model of ov
	Doc *Document
	// help
	helpDoc *Document
	// log
	logDoc *Document

	// DocList
	DocList    []*Document
	CurrentDoc int
	// mu controls the RWMutex.
	mu sync.RWMutex

	// screenMode represents the mode of screen.
	screenMode ScreenMode
	// input contains the input mode.
	input *Input
	// Original position at the start of search.
	OriginPos int
	// Original string.
	OriginStr string
	// cancelFunc saves the cancel function, which is a time-consuming process.
	cancelFunc context.CancelFunc

	// searchWord for on-screen highlighting.
	searchWord string
	// searchReg for on-screen highlighting.
	searchReg *regexp.Regexp

	// keyConfig contains the binding settings for the key.
	keyConfig *cbind.Configuration
	// inputKeyConfig contains the binding settings for the key.
	inputKeyConfig *cbind.Configuration

	// message is the message to display.
	message string

	// vWidth represents the screen width.
	vWidth int
	// vHight represents the screen height.
	vHight int

	// startX is the start position of x.
	startX int

	// lnumber is an array that returns
	// lineNumber (number of logical lines and number of wrapping lines) from y on the screen.
	lnumber []lineNumber

	// skipDraw skips draw once when true.
	// skipDraw is set to true when the mouse cursor just moves (no event occurs).
	skipDraw bool

	// x1, y1, x2, y2 are the coordinates selected by the mouse.
	x1 int
	y1 int
	x2 int
	y2 int

	// mousePressed is a flag when the mouse selection button is pressed.
	mousePressed bool
	// mouseSelect is a flag with mouse selection.
	mouseSelect bool
	// mouseRectangle is a flag for rectangle selection.
	mouseRectangle bool

	// wrapHeaderLen is the actual header length when wrapped.
	wrapHeaderLen int

	// bottomLN is the last line number displayed.
	bottomLN int
	// bottomLX is the leftmost X position on the last line.
	bottomLX int

	// statusPos is the position of the status line.
	statusPos int
	// minStartX is the minimum start position of x.
	minStartX int

	// cancelKeys represents the cancellation key string.
	cancelKeys []string
}

// LineNumber is Number of logical lines and number of wrapping lines on the screen.
type lineNumber struct {
	line int
	wrap int
}

// general structure contains the general of the display.
// general contains values that determine the behavior of each document.
type general struct {
	// TabWidth is tab stop num.
	TabWidth int
	// HeaderLen is number of header rows to be fixed.
	Header int
	// SkipLines is the rows to skip.
	SkipLines int
	// Color to alternate rows
	AlternateRows bool
	// Column mode
	ColumnMode bool
	// Line Number
	LineNumMode bool
	// Wrap is Wrap mode.
	WrapMode bool
	// Column Delimiter
	ColumnDelimiter string
	// Follow mode.
	FollowMode bool
	// Follow all.
	FollowAll bool
	// MarkStyleWidth is width to apply the style of the marked line.
	MarkStyleWidth int
}

// Config represents the settings of ov.
type Config struct {
	// StyleAlternate is a style that applies line by line.
	StyleAlternate ovStyle
	// StyleHeader is the style that applies to the header.
	StyleHeader ovStyle
	// StyleHeader is the style that applies to the header.
	StyleBody ovStyle
	// StyleOverStrike is a style that applies to overstrikes.
	StyleOverStrike ovStyle
	// OverLineS is a style that applies to overstrike underlines.
	StyleOverLine ovStyle
	// StyleLineNumber is a style that applies line number.
	StyleLineNumber ovStyle
	// StyleSearchHighlight is the style that applies to the search highlight.
	StyleSearchHighlight ovStyle
	// StyleColumnHighlight is the style that applies to the column highlight.
	StyleColumnHighlight ovStyle
	// StyleMarkLine is a style that marked line.
	StyleMarkLine ovStyle

	// Old setting method.
	// Alternating background color.
	ColorAlternate string
	// Header color.
	ColorHeader string
	// OverStrike color.
	ColorOverStrike string
	// OverLine color.
	ColorOverLine string

	// General represents the general behavior.
	General general
	// Mode represents the operation of the customized mode.
	Mode map[string]general

	// Mouse support disable.
	DisableMouse bool
	// AfterWrite writes the current screen on exit.
	AfterWrite bool
	// QuiteSmall Quit if the output fits on one screen.
	QuitSmall bool
	// CaseSensitive is case-sensitive if true.
	CaseSensitive bool
	// RegexpSearch is Regular expression search if true.
	RegexpSearch bool
	// Incsearch is incremental server if true.
	Incsearch bool
	// Debug represents whether to enable the debug output.
	Debug bool

	// KeyBinding
	Keybind map[string][]string
}

// ovStyle represents a style in addition to the original style.
type ovStyle struct {
	// Background is a color name string.
	Background string
	// Foreground is a color name string.
	Foreground string
	// If true, add blink.
	Blink bool
	// If true, add bold.
	Bold bool
	// If true, add dim.
	Dim bool
	// If true, add italic.
	Italic bool
	// If true, add reverse.
	Reverse bool
	// If true, add underline.
	Underline bool
	// If true, add strikethrough.
	StrikeThrough bool
}

var (
	// OverStrikeStyle represents the overstrike style.
	OverStrikeStyle tcell.Style
	// OverLineStyle represents the overline underline style.
	OverLineStyle tcell.Style
)

// ScreenMode represents the state of the screen.
type ScreenMode int

const (
	// Docs is a normal document screen mode.
	Docs ScreenMode = iota
	// Help is Help screen mode.
	Help
	// LogDoc is Error screen mode.
	LogDoc
)

var (
	// ErrOutOfRange indicates that value is out of range.
	ErrOutOfRange = errors.New("out of range")
	// ErrFatalCache indicates that the cache value had a fatal error.
	ErrFatalCache = errors.New("fatal error in cache value")
	// ErrMissingFile indicates that the file does not exist.
	ErrMissingFile = errors.New("missing filename")
	// ErrIsDirectory indicates that specify a directory instead of a file.
	ErrIsDirectory = errors.New("is a directory")
	// ErrNotFound indicates not found.
	ErrNotFound = errors.New("not found")
	// ErrCancel indicates cancel.
	ErrCancel = errors.New("cancel")
	// ErrInvalidNumber indicates an invalid number.
	ErrInvalidNumber = errors.New("invalid number")
	// ErrFailedKeyBind indicates keybinding failed.
	ErrFailedKeyBind = errors.New("failed to set keybind")
	// ErrSignalCatch indicates that the signal has been caught.
	ErrSignalCatch = errors.New("signal catch")
)

var tcellNewScreen = tcell.NewScreen

// NewOviewer return the structure of oviewer.
// NewOviewer requires one or more documents.
func NewOviewer(docs ...*Document) (*Root, error) {
	if len(docs) == 0 {
		return nil, ErrNotFound
	}
	root := &Root{
		minStartX: -10,
	}
	root.Config = NewConfig()
	root.keyConfig = cbind.NewConfiguration()
	root.inputKeyConfig = cbind.NewConfiguration()
	root.DocList = append(root.DocList, docs...)
	root.Doc = root.DocList[0]
	root.input = NewInput()
	root.screenMode = Docs

	screen, err := tcellNewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("Screen.Init(): %w", err)
	}
	root.Screen = screen

	return root, nil
}

// NewConfig return the structure of Config with default values.
func NewConfig() Config {
	return Config{
		StyleHeader: ovStyle{
			Bold: true,
		},
		StyleAlternate: ovStyle{
			Background: "gray",
		},
		StyleOverStrike: ovStyle{
			Bold: true,
		},
		StyleOverLine: ovStyle{
			Underline: true,
		},
		StyleLineNumber: ovStyle{
			Bold: true,
		},
		StyleSearchHighlight: ovStyle{
			Reverse: true,
		},
		StyleColumnHighlight: ovStyle{
			Reverse: true,
		},
		StyleMarkLine: ovStyle{
			Background: "darkgoldenrod",
		},
		General: general{
			TabWidth:       8,
			MarkStyleWidth: 1,
		},
	}
}

// Open reads the file named of the argument and return the structure of oviewer.
func Open(fileNames ...string) (*Root, error) {
	switch len(fileNames) {
	case 0:
		return openSTDIN()
	case 1:
		m, err := openFile(fileNames[0])
		if err != nil {
			return nil, err
		}
		return NewOviewer(m)
	default:
		return openFiles(fileNames)
	}
}

// NewRoot returns the structure of the oviewer.
// NewRoot is a simplified version that can be used externally.
func NewRoot(read io.Reader) (*Root, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}

	if err := m.ReadAll(read); err != nil {
		return nil, err
	}
	return NewOviewer(m)
}

func openSTDIN() (*Root, error) {
	docList := make([]*Document, 0, 1)
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}

	if err := m.ReadFile(""); err != nil {
		return nil, err
	}
	docList = append(docList, m)
	return NewOviewer(docList...)
}

func openFile(fileName string) (*Document, error) {
	fi, err := os.Stat(fileName)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("%s %w", fileName, ErrIsDirectory)
	}
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}
	if err := m.ReadFile(fileName); err != nil {
		return nil, err
	}
	return m, nil
}

func openFiles(fileNames []string) (*Root, error) {
	docList := make([]*Document, 0)
	for _, fileName := range fileNames {
		m, err := openFile(fileName)
		if err != nil {
			log.Println(err)
			continue
		}
		docList = append(docList, m)
	}

	if len(docList) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrMissingFile, fileNames[0])
	}

	return NewOviewer(docList...)
}

// SetConfig sets config.
func (root *Root) SetConfig(config Config) {
	root.Config = config
}

// SetWatcher sets file monitoring.
func (root *Root) SetWatcher(watcher *fsnotify.Watcher) {
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				root.mu.RLock()
				for _, doc := range root.DocList {
					if doc.FileName == event.Name {
						select {
						case doc.changCh <- struct{}{}:
						default:
						}
					}
				}
				root.mu.RUnlock()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	for _, doc := range root.DocList {
		err := watcher.Add(doc.FileName)
		if err != nil {
			root.debugMessage(fmt.Sprintf("watcher %s:%s", doc.FileName, err))
		}
	}
}

func (root *Root) setKeyConfig() (map[string][]string, error) {
	keyBind := GetKeyBinds(root.Config.Keybind)
	if err := root.setKeyBind(keyBind); err != nil {
		return nil, err
	}

	keys, ok := keyBind[actionCancel]
	if !ok {
		log.Printf("no cancel key")
	} else {
		root.cancelKeys = keys
	}
	return keyBind, nil
}

// Run starts the terminal pager.
func (root *Root) Run() error {
	defer root.Close()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	root.SetWatcher(watcher)

	keyBind, err := root.setKeyConfig()
	if err != nil {
		return err
	}

	help, err := NewHelp(keyBind)
	if err != nil {
		return err
	}
	root.helpDoc = help

	logDoc, err := NewLogDoc()
	if err != nil {
		return err
	}
	root.logDoc = logDoc

	if !root.Config.DisableMouse {
		root.Screen.EnableMouse()
	}

	// Call from man command.
	manPN := os.Getenv("MAN_PN")
	if len(manPN) > 0 {
		root.Doc.FileName = manPN
		// Bug?? Clipboard fails when called by man.
		root.Screen.DisableMouse()
	}

	for n, doc := range root.DocList {
		log.Printf("open [%d]%s", n, doc.FileName)
		doc.general = root.Config.General
	}
	root.setGlobalStyle()
	root.Screen.Clear()

	list := make([]string, 0, len(root.Config.Mode)+1)
	list = append(list, "general")
	for name := range root.Config.Mode {
		list = append(list, name)
	}
	root.input.ModeCandidate.list = list

	root.ViewSync()
	// Exit if fits on screen
	if root.QuitSmall && root.docSmall() {
		root.AfterWrite = true
		return nil
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)

	sigSuspend := make(chan os.Signal, 1)
	registerSIGTSTP(sigSuspend)

	quitChan := make(chan struct{})

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go root.main(ctx, quitChan)

	for {
		select {
		case <-quitChan:
			return nil
		case <-sigSuspend:
			root.Suspend()
		case sig := <-sigs:
			return fmt.Errorf("%w [%s]", ErrSignalCatch, sig)
		}
	}
}

// Close closes the oviewer.
func (root *Root) Close() {
	root.Screen.Fini()
}

func (root *Root) setMessage(msg string) {
	if root.message == msg {
		return
	}
	root.message = msg
	root.debugMessage(msg)
	root.statusDraw()
	root.Show()
}

func (root *Root) setMessagef(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	root.setMessage(msg)
}

func (root *Root) debugMessage(msg string) {
	if !root.Debug {
		return
	}
	root.message = msg
	if len(msg) == 0 {
		return
	}
	log.Printf("%s:%s", root.Doc.FileName, msg)
}

func setStyle(s ovStyle) tcell.Style {
	style := tcell.StyleDefault
	style = style.Background(tcell.GetColor(s.Background))
	style = style.Foreground(tcell.GetColor(s.Foreground))
	style = style.Blink(s.Blink)
	style = style.Bold(s.Bold)
	style = style.Dim(s.Dim)
	style = style.Italic(s.Italic)
	style = style.Reverse(s.Reverse)
	style = style.Underline(s.Underline)
	style = style.StrikeThrough(s.StrikeThrough)

	return style
}

func applyStyle(style tcell.Style, s ovStyle) tcell.Style {
	if s.Background != "" {
		style = style.Background(tcell.GetColor(s.Background))
	}
	if s.Foreground != "" {
		style = style.Foreground(tcell.GetColor(s.Foreground))
	}
	if s.Blink {
		style = style.Blink(s.Blink)
	}
	if s.Bold {
		style = style.Bold(s.Bold)
	}
	if s.Dim {
		style = style.Dim(s.Dim)
	}
	if s.Italic {
		style = style.Italic(s.Italic)
	}
	if s.Reverse {
		style = style.Reverse(s.Reverse)
	}
	if s.Underline {
		style = style.Underline(s.Underline)
	}
	if s.StrikeThrough {
		style = style.StrikeThrough(s.StrikeThrough)
	}
	return style
}

func (root *Root) setGlobalStyle() {
	OverStrikeStyle = setStyle(root.Config.StyleOverStrike)
	OverLineStyle = setStyle(root.Config.StyleOverLine)
	root.setOldGlobalStyle()
}

// setGlobalStyle sets some styles that are determined by the settings.
func (root *Root) setOldGlobalStyle() {
	if root.ColorAlternate != "" {
		root.StyleAlternate = ovStyle{Background: root.ColorAlternate}
	}
	if root.ColorHeader != "" {
		root.StyleHeader = ovStyle{Foreground: root.ColorHeader}
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

	// Do not allow size 0.
	root.vWidth = max(root.vWidth, 1)
	root.vHight = max(root.vHight, 1)

	root.lnumber = make([]lineNumber, root.vHight+1)
	root.setWrapHeaderLen()
	root.statusPos = root.vHight - 1
}

// docSmall returns with bool whether the file to display fits on the screen.
func (root *Root) docSmall() bool {
	if len(root.DocList) > 1 {
		return false
	}
	root.prepareView()
	m := root.Doc
	if !m.BufEOF() {
		return false
	}
	hight := 0
	for y := 0; y < m.BufEndNum(); y++ {
		lc, err := m.lnToContents(y, root.Doc.TabWidth)
		if err != nil {
			log.Println(err, y)
			continue
		}
		hight += 1 + (len(lc) / root.vWidth)
		if hight > root.vHight {
			return false
		}
	}
	return true
}

// WriteOriginal writes to the original terminal.
func (root *Root) WriteOriginal() {
	m := root.Doc
	p := 0
	for n := m.topLN; ; n++ {
		if n >= m.BufEndNum() {
			break
		}

		lc, err := m.lnToContents(n, root.Doc.TabWidth)
		if err != nil {
			log.Println(err, n)
			continue
		}
		p += 1 + (len(lc) / root.vWidth)
		if p >= root.vHight {
			break
		}

		fmt.Println(m.GetLine(n))
	}
}

// WriteLog write to the log terminal.
func (root *Root) WriteLog() {
	maxWriteLog := 10
	m := root.logDoc

	n := m.BufEndNum() - maxWriteLog
	for i := 0; i < maxWriteLog; i++ {
		str := strings.ReplaceAll(m.GetLine(n+i), "\n", "")
		if len(str) > 0 {
			fmt.Fprintln(os.Stderr, str)
		}
	}
}

// headerLen returns the actual number of lines in the header.
func (root *Root) headerLen() int {
	if root.Doc.WrapMode {
		return root.wrapHeaderLen
	}
	return root.Doc.Header
}

// leftMostX returns a list of left - most x positions when wrapping.
func (root *Root) leftMostX(lN int) ([]int, error) {
	lc, err := root.Doc.lnToContents(lN, root.Doc.TabWidth)
	if err != nil {
		return nil, err
	}

	listX := make([]int, 0, (len(lc)/root.vWidth)+1)
	width := (root.vWidth - root.startX)

	listX = append(listX, 0)
	for n := width; n < len(lc); n += width {
		if lc[n-1].width == 2 {
			n--
		}
		listX = append(listX, n)
	}
	return listX, nil
}

// DocumentLen returns the number of Docs.
func (root *Root) DocumentLen() int {
	root.mu.RLock()
	defer root.mu.RUnlock()
	return len(root.DocList)
}
