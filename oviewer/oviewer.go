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

	// headerLen is the actual header length when wrapped.
	headerLen int

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
	// AlternateRows alternately style rows.
	AlternateRows bool
	// ColumnMode is column mode.
	ColumnMode bool
	// LineNumMode displays line numbers.
	LineNumMode bool
	// Wrap is Wrap mode.
	WrapMode bool
	// ColumnDelimiter is a column delimiter.
	ColumnDelimiter string
	// FollowMode is the follow mode.
	FollowMode bool
	// FollowAll is a follow mode for all documents.
	FollowAll bool
	// FollowSection is a follow mode that uses section instead of line.
	FollowSection bool
	// WatchInterval is the watch interval (seconds).
	WatchInterval int
	// MarkStyleWidth is width to apply the style of the marked line.
	MarkStyleWidth int
	// SectionDelimiter is a section delimiter.
	SectionDelimiter string
	// SectionDelimiterReg is a section delimiter.
	SectionDelimiterReg *regexp.Regexp
	// SectionStartPosition is a section start position.
	SectionStartPosition int
}

// Config represents the settings of ov.
type Config struct {
	// StyleAlternate is a style that applies line by line.
	StyleAlternate OVStyle
	// StyleHeader is the style that applies to the header.
	StyleHeader OVStyle
	// StyleHeader is the style that applies to the header.
	StyleBody OVStyle
	// StyleOverStrike is a style that applies to overstrikes.
	StyleOverStrike OVStyle
	// OverLineS is a style that applies to overstrike underlines.
	StyleOverLine OVStyle
	// StyleLineNumber is a style that applies line number.
	StyleLineNumber OVStyle
	// StyleSearchHighlight is the style that applies to the search highlight.
	StyleSearchHighlight OVStyle
	// StyleColumnHighlight is the style that applies to the column highlight.
	StyleColumnHighlight OVStyle
	// StyleMarkLine is a style that marked line.
	StyleMarkLine OVStyle
	// StyleSectionLine is a style that section delimiter line.
	StyleSectionLine OVStyle

	// General represents the general behavior.
	General general
	// Mode represents the operation of the customized mode.
	Mode map[string]general

	// Mouse support disable.
	DisableMouse bool
	// IsWriteOriginal is true, write the current screen on quit.
	IsWriteOriginal bool
	// BeforeWriteOriginal specifies the number of lines before the current position.
	// 0 is the top of the current screen
	BeforeWriteOriginal int
	// AfterWriteOriginal specifies the number of lines after the current position.
	// 0 specifies the bottom of the screen.
	AfterWriteOriginal int

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

	// Old setting.

	// Deprecated: Alternating background color.
	ColorAlternate string
	// Deprecated: Header color.
	ColorHeader string
	// Deprecated: OverStrike color.
	ColorOverStrike string
	// Deprecated: OverLine color.
	ColorOverLine string
}

// OVStyle represents a style in addition to the original style.
type OVStyle struct {
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

// ov output destination.
var (
	// Redirect to standard output.
	//     echo "t" | ov> out
	STDOUTPIPE *os.File
	// Redirects the error output of ov --exec.
	//     ov --exec -- command 2> out
	STDERRPIPE *os.File
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

const MaxWriteLog int = 10

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
	// ErrAlreadyClose indicates that it is already closed.
	ErrAlreadyClose = errors.New("already closed")
)

// This is a function of tcell.NewScreen but can be replaced with mock.
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

	logDoc, err := NewLogDoc()
	if err != nil {
		return nil, err
	}
	root.logDoc = logDoc

	return root, nil
}

// NewConfig return the structure of Config with default values.
func NewConfig() Config {
	return Config{
		StyleHeader: OVStyle{
			Bold: true,
		},
		StyleAlternate: OVStyle{
			Background: "gray",
		},
		StyleOverStrike: OVStyle{
			Bold: true,
		},
		StyleOverLine: OVStyle{
			Underline: true,
		},
		StyleLineNumber: OVStyle{
			Bold: true,
		},
		StyleSearchHighlight: OVStyle{
			Reverse: true,
		},
		StyleColumnHighlight: OVStyle{
			Reverse: true,
		},
		StyleMarkLine: OVStyle{
			Background: "darkgoldenrod",
		},
		StyleSectionLine: OVStyle{
			Background: "green",
		},
		General: general{
			TabWidth:             8,
			MarkStyleWidth:       1,
			SectionStartPosition: 0,
		},
	}
}

// NewRoot returns the structure of the oviewer.
// NewRoot is a simplified version that can be used externally.
func NewRoot(r io.Reader) (*Root, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}
	if err := m.ReadReader(r); err != nil {
		return nil, err
	}
	return NewOviewer(m)
}

// Open reads the file named of the argument and return the structure of oviewer.
// If there is no file name, create Root from standard input.
// If there is only one file name, create Root from that file,
// but return an error if the open is an error.
// If there is more than one file name, create Root from multiple files.
func Open(fileNames ...string) (*Root, error) {
	switch len(fileNames) {
	case 0:
		return openSTDIN()
	case 1:
		return openFile(fileNames[0])
	default:
		return openFiles(fileNames)
	}
}

// openSTDIN creates root with standard input.
func openSTDIN() (*Root, error) {
	m, err := STDINDocument()
	if err != nil {
		return nil, err
	}
	return NewOviewer(m)
}

// openFile creates root in one file.
// If there is only one file, an error will occur if the file fails to open.
func openFile(fileName string) (*Root, error) {
	m, err := OpenDocument(fileName)
	if err != nil {
		return nil, err
	}
	return NewOviewer(m)
}

// openFiles opens multiple files and creates root.
// It will continue even if there are files that fail to open.
func openFiles(fileNames []string) (*Root, error) {
	errors := make([]string, 0)
	docList := make([]*Document, 0)
	for _, fileName := range fileNames {
		m, err := OpenDocument(fileName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("open error: %s", err))
			continue
		}
		docList = append(docList, m)
	}

	if len(docList) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrMissingFile, fileNames[0])
	}
	root, err := NewOviewer(docList...)
	if err != nil {
		return nil, err
	}

	for _, e := range errors {
		log.Println(e)
	}
	return root, err
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
				root.mu.Lock()
				for _, doc := range root.DocList {
					if doc.FileName == event.Name {
						select {
						case doc.changCh <- struct{}{}:
						default:
						}
					}
				}
				root.mu.Unlock()
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

// SetKeyHandler assigns a new key handler.
func (root *Root) SetKeyHandler(name string, keys []string, handler func()) error {
	return setHandler(root.keyConfig, name, keys, handler)
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

	// Do not set the key bindings in NewOviewer
	// because it is done after loading the config.
	keyBind, err := root.setKeyConfig()
	if err != nil {
		return err
	}
	help, err := NewHelp(keyBind)
	if err != nil {
		return err
	}
	root.helpDoc = help

	if !root.Config.DisableMouse {
		root.Screen.EnableMouse()
	}
	root.Config.General.SectionDelimiterReg = regexpCompile(root.Config.General.SectionDelimiter, true)

	root.optimizedMan()

	for n, doc := range root.DocList {
		doc.general = root.Config.General
		w := ""
		if doc.general.WatchInterval > 0 {
			doc.watchMode()
			w = "(watch)"
		}
		log.Printf("open [%d]%s%s", n, doc.FileName, w)
	}

	root.setModeConfig()

	root.ViewSync()
	// Exit if fits on screen
	if root.QuitSmall && root.docSmall() {
		root.IsWriteOriginal = true
		return nil
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)

	sigSuspend := registerSIGTSTP()

	quitChan := make(chan struct{})

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		// Undo screen when goroutine panic.
		defer func() {
			root.Close()
		}()
		root.main(ctx, quitChan)
	}()

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

func (root *Root) optimizedMan() {
	// Call from man command.
	manPN := os.Getenv("MAN_PN")
	if len(manPN) == 0 {
		return
	}

	root.Doc.Caption = manPN
	// Bug?? Clipboard fails when called by man.
	root.Screen.DisableMouse()
}

func (root *Root) setModeConfig() {
	list := make([]string, 0, len(root.Config.Mode)+1)
	list = append(list, "general")
	for name := range root.Config.Mode {
		list = append(list, name)
	}
	root.input.ModeCandidate.list = list
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
	root.drawStatus()
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

func ToTcellStyle(s OVStyle) tcell.Style {
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

func applyStyle(style tcell.Style, s OVStyle) tcell.Style {
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

// overwriteGeneral overwrites a general structure with a struct.
func overwriteGeneral(a general, b general) general {
	if b.TabWidth != 0 {
		a.TabWidth = b.TabWidth
	}
	if b.Header != 0 {
		a.Header = b.Header
	}
	if b.SkipLines != 0 {
		a.SkipLines = b.SkipLines
	}
	a.AlternateRows = b.AlternateRows
	a.ColumnMode = b.ColumnMode
	a.LineNumMode = b.LineNumMode
	a.WrapMode = b.WrapMode
	a.FollowMode = b.FollowMode
	a.FollowAll = b.FollowAll
	a.FollowSection = b.FollowSection
	if b.ColumnDelimiter != "" {
		a.ColumnDelimiter = b.ColumnDelimiter
	}
	if b.WatchInterval != 0 {
		a.WatchInterval = b.WatchInterval
	}
	if b.MarkStyleWidth != 0 {
		a.MarkStyleWidth = b.MarkStyleWidth
	}
	if b.SectionDelimiter != "" {
		a.SectionDelimiter = b.SectionDelimiter
	}
	if b.SectionStartPosition != 0 {
		a.SectionStartPosition = b.SectionStartPosition
	}
	return a
}

// prepareView prepares when the screen size is changed.
func (root *Root) prepareView() {
	screen := root.Screen
	root.vWidth, root.vHight = screen.Size()

	// Do not allow size 0.
	root.vWidth = max(root.vWidth, 1)
	root.vHight = max(root.vHight, 1)

	root.lnumber = make([]lineNumber, root.vHight+1)
	root.statusPos = root.vHight - statusLine
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
		lc, err := m.contentsLN(y, root.Doc.TabWidth)
		if err != nil {
			log.Printf("docSmall %d: %s", y, err)
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
	if m.bottomLN == 0 {
		m.bottomLN = m.BufEndNum()
	}

	start := max(0, m.topLN-root.BeforeWriteOriginal)
	end := m.bottomLN - 1
	if root.AfterWriteOriginal != 0 {
		end = m.topLN + root.AfterWriteOriginal - 1
	}

	m.Export(os.Stdout, start, end)
}

// WriteLog write to the log terminal.
func (root *Root) WriteLog() {
	m := root.logDoc
	start := max(0, m.BufEndNum()-MaxWriteLog)
	end := m.BufEndNum()
	m.Export(os.Stdout, start, end)
}
