package oviewer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"

	"code.rocketnine.space/tslocum/cbind"
	"github.com/fsnotify/fsnotify"
	tcell "github.com/gdamore/tcell/v2"
)

// Root structure contains information about the drawing.
type Root struct {
	// tcell.Screen is the root screen.
	tcell.Screen
	// Doc contains the model of ov
	Doc *Document
	// help
	helpDoc *Document
	// log
	logDoc *Document

	// input contains the input mode.
	input *Input
	// cancelFunc saves the cancel function, which is a time-consuming process.
	cancelFunc context.CancelFunc

	// searcher is the searcher.
	searcher Searcher

	// keyConfig contains the binding settings for the key.
	keyConfig *cbind.Configuration
	// inputKeyConfig contains the binding settings for the key.
	inputKeyConfig *cbind.Configuration

	// Original string.
	OriginStr string

	// message is the message to display.
	message string
	// cancelKeys represents the cancellation key string.
	cancelKeys []string

	// DocList is the list of documents.
	DocList []*Document
	scr     SCR
	Config
	// screenMode represents the mode of screen.
	screenMode ScreenMode
	// Original position at the start of search.
	OriginPos int

	// CurrentDoc is the index of the current document.
	CurrentDoc int
	// minStartX is the minimum start position of x.
	minStartX int
	// x1, y1, x2, y2 are the coordinates selected by the mouse.
	x1 int
	y1 int
	x2 int
	y2 int

	// mu controls the RWMutex.
	mu sync.RWMutex

	// skipDraw is set to true when the mouse cursor just moves (no event occurs).
	skipDraw bool
	// mousePressed is a flag when the mouse selection button is pressed.
	mousePressed bool
	// mouseSelect is a flag with mouse selection.
	mouseSelect bool
	// mouseRectangle is a flag for rectangle selection.
	mouseRectangle bool
}

// SCR contains the screen information.
type SCR struct {
	// numbers is the line information of the currently displayed screen.
	// numbers (number of logical numbers and number of wrapping numbers) from y on the screen.
	numbers []LineNumber
	// vWidth represents the screen width.
	vWidth int
	// vHeight represents the screen height.
	vHeight int
	// startX is the start position of x.
	startX int
}

// LineNumber is Number of logical lines and number of wrapping lines on the screen.
type LineNumber struct {
	number int
	wrap   int
}

// general structure contains the general of the display.
// general contains values that determine the behavior of each document.
type general struct {
	// ColumnDelimiterReg is a compiled regular expression of ColumnDelimiter.
	ColumnDelimiterReg *regexp.Regexp
	// ColumnDelimiter is a column delimiter.
	ColumnDelimiter string
	// SectionDelimiterReg is a section delimiter.
	SectionDelimiterReg *regexp.Regexp
	// SectionDelimiter is a section delimiter.
	SectionDelimiter string
	// Specified string for jumpTarget.
	JumpTarget string
	// MultiColorWords specifies words to color separated by spaces.
	MultiColorWords []string

	// TabWidth is tab stop num.
	TabWidth int
	// HeaderLen is number of header rows to be fixed.
	Header int
	// SkipLines is the rows to skip.
	SkipLines int
	// WatchInterval is the watch interval (seconds).
	WatchInterval int
	// MarkStyleWidth is width to apply the style of the marked line.
	MarkStyleWidth int
	// SectionStartPosition is a section start position.
	SectionStartPosition int
	// AlternateRows alternately style rows.
	AlternateRows bool
	// ColumnMode is column mode.
	ColumnMode bool
	// ColumnWidth is column width mode.
	ColumnWidth bool
	// ColumnRainbow is column rainbow.
	ColumnRainbow bool
	// LineNumMode displays line numbers.
	LineNumMode bool
	// Wrap is Wrap mode.
	WrapMode bool
	// FollowMode is the follow mode.
	FollowMode bool
	// FollowAll is a follow mode for all documents.
	FollowAll bool
	// FollowSection is a follow mode that uses section instead of line.
	FollowSection bool
	// FollowName is the mode to follow files by name.
	FollowName bool
	// PlainMode is whether to enable the original character decoration.
	PlainMode bool
}

// OVPromptConfigNormal is the normal prompt setting.
type OVPromptConfigNormal struct {
	// ShowFilename controls whether to display filename.
	ShowFilename bool
	// InvertColor controls whether the text is colored and inverted.
	InvertColor bool
}

// OVPromptConfig is the prompt setting.
type OVPromptConfig struct {
	// Normal is the normal prompt setting.
	Normal OVPromptConfigNormal
}

// Config represents the settings of ov.
type Config struct {
	// KeyBinding
	Keybind map[string][]string
	// Mode represents the operation of the customized mode.
	Mode map[string]general
	// ViewMode represents the view mode.
	// ViewMode sets several settings together and can be easily switched.
	ViewMode string
	// Default keybindings. Disabled if the default keybinding is "disable".
	DefaultKeyBind string
	// StyleColumnRainbow  is the style that applies to the column rainbow color highlight.
	StyleColumnRainbow []OVStyle
	// StyleMultiColorHighlight is the style that applies to the multi color highlight.
	StyleMultiColorHighlight []OVStyle

	// Prompt is the prompt setting.
	Prompt OVPromptConfig

	// StyleHeader is the style that applies to the header.
	StyleHeader OVStyle
	// StyleBody is the style that applies to the body.
	StyleBody OVStyle
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
	// StyleJumpTargetLine is the line that displays the search results.
	StyleJumpTargetLine OVStyle
	// StyleAlternate is a style that applies line by line.
	StyleAlternate OVStyle
	// StyleOverStrike is a style that applies to overstrike.
	StyleOverStrike OVStyle
	// StyleOverLine is a style that applies to overstrike underlines.
	StyleOverLine OVStyle
	// General represents the general behavior.
	General general
	// BeforeWriteOriginal specifies the number of lines before the current position.
	// 0 is the top of the current screen
	BeforeWriteOriginal int
	// AfterWriteOriginal specifies the number of lines after the current position.
	// 0 specifies the bottom of the screen.
	AfterWriteOriginal int
	// MemoryLimit is a number that limits chunk loading.
	MemoryLimit int
	// MemoryLimitFile is a number that limits the chunks loading a file into memory.
	MemoryLimitFile int
	// Mouse support disable.
	DisableMouse bool
	// IsWriteOriginal is true, write the current screen on quit.
	IsWriteOriginal bool
	// QuitSmall Quit if the output fits on one screen.
	QuitSmall bool
	// CaseSensitive is case-sensitive if true.
	CaseSensitive bool
	// SmartCaseSensitive is lowercase search ignores case, if true.
	SmartCaseSensitive bool
	// RegexpSearch is Regular expression search if true.
	RegexpSearch bool
	// Incsearch is incremental search if true.
	Incsearch bool

	// DisableColumnCycle is disable column cycle.
	DisableColumnCycle bool
	// Debug represents whether to enable the debug output.
	Debug bool
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
	// If true, add strike through.
	StrikeThrough bool
	// If true, add overline (not yet supported).
	OverLine bool
	// If true, sub blink.
	UnBlink bool
	// If true, sub bold.
	UnBold bool
	// If true, sub dim.
	UnDim bool
	// If true, sub italic.
	UnItalic bool
	// If true, sub reverse.
	UnReverse bool
	// If true, sub underline.
	UnUnderline bool
	// If true, sub strike through.
	UnStrikeThrough bool
	// if true, sub underline (not yet supported).
	UnOverLine bool
}

var (
	// MemoryLimit is a number that limits the chunks to load into memory.
	MemoryLimit int
	// MemoryLimitFile is a number that limits the chunks loading a file into memory.
	MemoryLimitFile int

	// OverStrikeStyle represents the overstrike style.
	OverStrikeStyle tcell.Style
	// OverLineStyle represents the overline underline style.
	OverLineStyle tcell.Style
	// SkipExtract is a flag to skip extracting compressed files.
	SkipExtract bool
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

// MouseFlags represents which events of the mouse should be captured.
// Set the mode to MouseDragEvents when the mouse is enabled in oviewer.
// Does not track mouse movements except when dragging.
const MouseFlags = tcell.MouseDragEvents

// MaxWriteLog is the maximum number of lines to output to the log
// when the debug flag is enabled.
const MaxWriteLog int = 10

var (
	// ErrOutOfRange indicates that value is out of range.
	ErrOutOfRange = errors.New("out of range")
	// ErrNotInMemory indicates that value is not in memory.
	ErrNotInMemory = errors.New("not in memory")
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
	// ErrNoColumn indicates that cursor specified a nonexistent column.
	ErrNoColumn = errors.New("no column")
	// ErrNoDelimiter indicates that the line containing the delimiter could not be found.
	ErrNoDelimiter = errors.New("no delimiter")
	// ErrOverScreen indicates that the specified screen is out of range.
	ErrOverScreen = errors.New("over screen")
	// ErrOutOfChunk indicates that the specified Chunk is out of range.
	ErrOutOfChunk = errors.New("out of chunk")
	// ErrNotLoaded indicates that it cannot be loaded.
	ErrNotLoaded = errors.New("not loaded")
	// ErrEOFreached indicates that EOF has been reached.
	ErrEOFreached = errors.New("EOF reached")
	// ErrPreventReload indicates that reload is prevented.
	ErrPreventReload = errors.New("prevent reload")
	// ErrOverChunkLimit indicates that the chunk limit has been exceeded.
	ErrOverChunkLimit = errors.New("over chunk limit")
	// ErrAlreadyLoaded indicates that the chunk already loaded.
	ErrAlreadyLoaded = errors.New("chunk already loaded")
	// ErrEvictedMemory indicates that it has been evicted from memory.
	ErrEvictedMemory = errors.New("evicted memory")
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

// NewRoot returns the structure of the oviewer.
// NewRoot is a simplified version that can be used externally.
func NewRoot(r io.Reader) (*Root, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}
	if err := m.ControlReader(r, nil); err != nil {
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
	viewMode, overwrite := config.Mode[config.ViewMode]
	if overwrite {
		config.General = mergeGeneral(config.General, viewMode)
	}
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
				root.watchEvent(event)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	for _, doc := range root.DocList {
		fileName, err := filepath.Abs(doc.FileName)
		if err != nil {
			log.Println(err)
			continue
		}
		doc.filepath = fileName

		path := filepath.Dir(fileName)
		if err := watcher.Add(path); err != nil {
			root.debugMessage(fmt.Sprintf("watcher %s:%s", doc.FileName, err))
		}
	}
}

// watchEvent sends a notification to the document.
func (root *Root) watchEvent(event fsnotify.Event) {
	root.mu.Lock()
	defer root.mu.Unlock()

	for _, m := range root.DocList {
		if m.filepath == event.Name {
			switch event.Op {
			case fsnotify.Write:
				root.sendRequest(m, requestFollow)
			case fsnotify.Remove, fsnotify.Create:
				if m.FollowName {
					root.sendRequest(m, requestReload)
				}
			}
		}
	}
}

func (root *Root) sendRequest(m *Document, request request) {
	select {
	case m.ctlCh <- controlSpecifier{request: request}:
		root.debugMessage(fmt.Sprintf("notify send %v", request))
	default:
		root.debugMessage(fmt.Sprintf("notify send fail %s", request))
	}
}

// setKeyConfig sets key bindings.
func (root *Root) setKeyConfig() (map[string][]string, error) {
	keyBind := GetKeyBinds(root.Config)
	if err := root.setHandlers(keyBind); err != nil {
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
		return fmt.Errorf("failed to create watcher: %w", err)
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
		root.Screen.EnableMouse(MouseFlags)
	}

	root.optimizedMan()
	root.setModeConfig()
	for n, doc := range root.DocList {
		doc.general = root.Config.General
		doc.regexpCompile()

		if doc.FollowName {
			doc.FollowMode = true
		}
		if doc.ColumnWidth {
			doc.ColumnMode = true
		}
		w := ""
		if doc.general.WatchInterval > 0 {
			doc.watchMode()
			w = "(watch)"
		}
		log.Printf("open [%d]%s%s", n, doc.FileName, w)
	}

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
		root.eventLoop(ctx, quitChan)
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

// optimizedMan optimizes execution with the Man command.
func (root *Root) optimizedMan() {
	// Call from man command.
	manPN := os.Getenv("MAN_PN")
	if len(manPN) == 0 {
		return
	}

	root.Doc.Caption = manPN
}

// setModeConfig sets mode config.
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

// setMessagef displays a formatted message in status.
func (root *Root) setMessagef(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	root.setMessage(msg)
}

// setMessage displays a message in status.
func (root *Root) setMessage(msg string) {
	if root.message == msg {
		return
	}
	root.message = msg
	root.debugMessage(msg)
	root.drawStatus()
	root.Show()
}

// setMessageLogf displays a formatted message in the status and outputs it to the log.
func (root *Root) setMessageLogf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	root.setMessageLog(msg)
}

// setMessageLog displays a message in the status and outputs it to the log.
func (root *Root) setMessageLog(msg string) {
	if root.message == msg {
		return
	}
	root.message = msg
	log.Print(msg)
	root.drawStatus()
	root.Show()
}

// debugMessage outputs a debug message.
func (root *Root) debugMessage(msg string) {
	if !root.Debug {
		return
	}
	if root.Doc == root.logDoc {
		return
	}

	if len(msg) == 0 {
		return
	}
	log.Printf("%s:%s", root.Doc.FileName, msg)
}

// ToTcellStyle convert from ovStyle to tcell style.
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
		style = style.Blink(true)
	}
	if s.Bold {
		style = style.Bold(true)
	}
	if s.Dim {
		style = style.Dim(true)
	}
	if s.Italic {
		style = style.Italic(true)
	}
	if s.Reverse {
		style = style.Reverse(true)
	}
	if s.Underline {
		style = style.Underline(true)
	}
	if s.StrikeThrough {
		style = style.StrikeThrough(true)
	}
	if s.UnBlink {
		style = style.Blink(false)
	}
	if s.UnBold {
		style = style.Bold(false)
	}
	if s.UnDim {
		style = style.Dim(false)
	}
	if s.UnItalic {
		style = style.Italic(false)
	}
	if s.UnReverse {
		style = style.Reverse(false)
	}
	if s.UnUnderline {
		style = style.Underline(false)
	}
	if s.UnStrikeThrough {
		style = style.StrikeThrough(false)
	}
	return style
}

// mergeGeneral overwrites a general structure with a struct.
func mergeGeneral(src general, dst general) general {
	if dst.TabWidth != 0 {
		src.TabWidth = dst.TabWidth
	}
	if dst.Header != 0 {
		src.Header = dst.Header
	}
	if dst.SkipLines != 0 {
		src.SkipLines = dst.SkipLines
	}
	if dst.AlternateRows {
		src.AlternateRows = dst.AlternateRows
	}
	if dst.ColumnMode {
		src.ColumnMode = dst.ColumnMode
	}
	if dst.ColumnWidth {
		src.ColumnWidth = dst.ColumnWidth
	}
	if dst.ColumnRainbow {
		src.ColumnRainbow = dst.ColumnRainbow
	}
	if dst.LineNumMode {
		src.LineNumMode = dst.LineNumMode
	}
	if !dst.WrapMode { // Because wrap mode defaults to true.
		src.WrapMode = dst.WrapMode
	}
	if dst.FollowMode {
		src.FollowMode = dst.FollowMode
	}
	if dst.FollowAll {
		src.FollowAll = dst.FollowAll
	}
	if dst.FollowSection {
		src.FollowSection = dst.FollowSection
	}
	if dst.FollowName {
		src.FollowName = dst.FollowName
	}
	if dst.ColumnDelimiter != "" {
		src.ColumnDelimiter = dst.ColumnDelimiter
	}
	if dst.WatchInterval != 0 {
		src.WatchInterval = dst.WatchInterval
	}
	if dst.MarkStyleWidth != 0 {
		src.MarkStyleWidth = dst.MarkStyleWidth
	}
	if dst.SectionDelimiter != "" {
		src.SectionDelimiter = dst.SectionDelimiter
	}
	if dst.SectionStartPosition != 0 {
		src.SectionStartPosition = dst.SectionStartPosition
	}
	if dst.JumpTarget != "" {
		src.JumpTarget = dst.JumpTarget
	}
	if len(dst.MultiColorWords) > 0 {
		src.MultiColorWords = dst.MultiColorWords
	}
	return src
}

// prepareView prepares when the screen size is changed.
func (root *Root) prepareView() {
	screen := root.Screen
	root.scr.vWidth, root.scr.vHeight = screen.Size()

	// Do not allow size 0.
	root.scr.vWidth = max(root.scr.vWidth, 1)
	root.scr.vHeight = max(root.scr.vHeight, 1)
	root.scr.numbers = make([]LineNumber, root.scr.vHeight+1)
	root.Doc.statusPos = root.scr.vHeight - statusLine
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
	height := 0
	for y := 0; y < m.BufEndNum(); y++ {
		lc, err := m.contents(y, root.Doc.TabWidth)
		if err != nil {
			log.Printf("docSmall %d: %s", y, err)
			continue
		}
		height += 1 + (len(lc) / root.scr.vWidth)
		if height > root.scr.vHeight {
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

	if err := m.Export(os.Stdout, start, end); err != nil {
		log.Println(err)
	}
}

// WriteOriginalTillQuit writes doc to the original terminal till the
// doc was viewed (Document.BufEndNum).
func (root *Root) WriteOriginalTillQuit() {
	m := root.Doc
	for n := 0; n < m.topLN; n++ {
		fmt.Println(m.GetLine(n))
	}
	root.WriteOriginal()
}

// WriteLog write to the log terminal.
func (root *Root) WriteLog() {
	m := root.logDoc
	start := max(0, m.BufEndNum()-MaxWriteLog)
	end := m.BufEndNum()
	if err := m.Export(os.Stdout, start, end); err != nil {
		log.Println(err)
	}
}

// lineNumber returns the line information from y on the screen.
func (scr SCR) lineNumber(y int) LineNumber {
	if y >= 0 && y <= len(scr.numbers) {
		return scr.numbers[y]
	}
	return scr.numbers[0]
}

// debugNumOfChunk outputs the number of chunks.
func (root *Root) debugNumOfChunk() {
	if !root.Debug {
		return
	}
	log.Println("MemoryLimit:", root.MemoryLimit)
	log.Println("MemoryLimitFile:", root.MemoryLimitFile)
	for _, doc := range root.DocList {
		if !doc.seekable {
			if MemoryLimit > 0 {
				log.Printf("%s: The number of chunks is %d, of which %d(%v) are loaded", doc.FileName, len(doc.store.chunks), doc.store.loadedChunks.Len(), doc.store.loadedChunks.Keys())
			}
			continue
		}
		for n, chunk := range doc.store.chunks {
			if n != 0 && len(chunk.lines) != 0 {
				if !doc.store.loadedChunks.Contains(n) {
					log.Printf("chunk %d is not under control %d", n, len(chunk.lines))
				}
			}
		}
		log.Printf("%s(seekable): The number of chunks is %d, of which %d(%v) are loaded", doc.FileName, len(doc.store.chunks), doc.store.loadedChunks.Len(), doc.store.loadedChunks.Keys())
	}
}
