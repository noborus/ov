package oviewer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"codeberg.org/tslocum/cbind"
	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/noborus/tcellansi"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// Root is the root structure of the oviewer.
type Root struct {
	// tcell.Screen is the root screen.
	Screen tcell.Screen

	// settings contains the runtime settings template.
	// These settings serve as the base configuration for each document.
	// Individual documents can override these settings as needed.
	settings RunTimeSettings

	// Doc is the current document.
	Doc *Document
	// helpDoc is the help document.
	helpDoc *Document
	// logDoc is the log document.
	logDoc *LogDocument

	// input contains the input mode.
	input *Input
	// cancelFunc saves the cancel function, which is a time-consuming process.
	cancelFunc context.CancelFunc

	// searcher is the search structure.
	searcher Searcher

	// keyConfig contains the binding settings for the key.
	keyConfig *cbind.Configuration
	// inputKeyConfig contains the binding settings for the key.
	inputKeyConfig *cbind.Configuration

	// Pattern is the search pattern.
	Pattern string
	// OnExit is output on exit.
	OnExit []string

	// Original string.
	OriginStr string

	// searchOpt is the search option.
	searchOpt string

	// message is the message to display.
	message string
	// cancelKeys represents the cancellation key string.
	cancelKeys []string
	// DocList is the list of documents.
	DocList []*Document
	// scr contains the screen information.
	scr SCR
	// Config is the configuration of ov.
	Config Config
	// Original position at the start of search.
	OriginPos int

	// CurrentDoc is the index of the current document.
	CurrentDoc int
	// minStartX is the minimum start position of x.
	minStartX int

	// quitSmallCountDown is the countdown to quit if the output fits on one screen.
	quitSmallCountDown int

	// mu controls the RWMutex.
	mu sync.RWMutex
	// isClosed indicates whether it is closed.
	isClosed atomic.Bool

	// FollowAll is a follow mode for all documents.
	FollowAll bool
	// skipDraw is set to true when the mouse cursor just moves (no event occurs).
	skipDraw bool
	// clickState is the state of mouse click.
	clickState ClickState

	// sidebarVisible is whether the sidebar is visible.
	sidebarVisible bool
	// sidebarMode is the mode of the sidebar.
	sidebarMode SidebarMode
	// previousSidebarMode is the previous mode of the sidebar before it was hidden.
	previousSidebarMode SidebarMode
	// SidebarItems is the items to display in the sidebar.
	SidebarItems []SidebarItem
	// SidebarHelpItems is the help items to display in the sidebar.
	SidebarHelpItems []SidebarItem
	// sidebarWidth is the width of the sidebar.
	sidebarWidth int
	// sidebarScrolls holds scroll positions for each sidebarMode.
	sidebarScrolls map[SidebarMode]sidebarScroll
}

// SCR contains the screen information.
type SCR struct {
	// lines is the lines of the screen.
	lines map[int]LineC
	// numbers is the line information of the currently displayed screen.
	// numbers (number of logical numbers and number of wrapping numbers) from y on the screen.
	numbers []LineNumber
	// vWidth represents the screen width.
	vWidth int
	// vHeight represents the screen height.
	vHeight int
	// statusLineHeight is the height of the status line.
	statusLineHeight int

	// rulerHeight is the height of the ruler.
	rulerHeight int
	// HeaderLN is the number of header lines.
	headerLN int
	// headerEnd is the end of the header.
	headerEnd int
	// sectionHeaderLN is the number of section headers.
	sectionHeaderLN int
	// sectionHeaderEnd is the end of the section header.
	sectionHeaderEnd int
	// bodyLN is the number of body lines.
	bodyLN int
	// bodyEnd is the end of the body.
	bodyEnd int

	// x1, y1, x2, y2 are the coordinates selected by the mouse.
	x1 int
	y1 int
	x2 int
	y2 int
	// mouseSelect is a flag with mouse selection.
	mouseSelect MouseSelectState
	// mousePressed is a flag when the mouse selection button is pressed.
	mousePressed bool
	// mouseRectangle is a flag for rectangle selection.
	mouseRectangle bool
	// hasAnchorPoint is a flag whether anchor point exists.
	hasAnchorPoint bool
	// forceDisplaySync forces synchronous display.
	forceDisplaySync bool
}

// LineNumber is Number of logical lines and number of wrapping lines on the screen.
type LineNumber struct {
	number int
	wrap   int
}

func newLineNumber(number, wrap int) LineNumber {
	return LineNumber{number: number, wrap: wrap}
}

// MinStartX is the minimum start position of x.
const MinStartX = -10

// RulerType is the type of ruler.
type RulerType int

const (
	// RulerNone is no ruler.
	RulerNone RulerType = iota
	// RulerRelative is a relative ruler.
	RulerRelative
	// RulerAbsolute is an absolute ruler.
	RulerAbsolute
)

// String returns the string representation of the ruler type.
func (r RulerType) String() string {
	switch r {
	case RulerRelative:
		return "relative"
	case RulerAbsolute:
		return "absolute"
	default:
		return "none"
	}
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

// MouseFlags represents which events of the mouse should be captured.
// Set the mode to MouseDragEvents when the mouse is enabled in oviewer.
// Does not track mouse movements except when dragging.
const MouseFlags = tcell.MouseDragEvents

// QuitSmallCountDown is the countdown to quit if the output fits on one screen.
// UpdateInterval(50 * time.Millisecond) * 10 = 500ms.
const QuitSmallCountDown = 10

// MaxWriteLog is the maximum number of lines to output to the log
// when the debug flag is enabled.
const MaxWriteLog int = 10

var (
	// ErrOutOfRange indicates that value is out of range.
	ErrOutOfRange = errors.New("value out of range")
	// ErrNotInMemory indicates that value is not in memory.
	ErrNotInMemory = errors.New("value not in memory")
	// ErrFatalCache indicates that the cache value had a fatal error.
	ErrFatalCache = errors.New("fatal cache error")
	// ErrMissingFile indicates that the file does not exist.
	ErrMissingFile = errors.New("missing file")
	// ErrIsDirectory indicates that specify a directory instead of a file.
	ErrIsDirectory = errors.New("is a directory")
	// ErrNotFound indicates not found.
	ErrNotFound = errors.New("not found")
	// ErrNotTerminal indicates that it is not a terminal.
	ErrNotTerminal = errors.New("not a terminal")
	// ErrCancel indicates cancel.
	ErrCancel = errors.New("cancel")
	// ErrInvalidNumber indicates an invalid number.
	ErrInvalidNumber = errors.New("invalid number")
	// ErrFailedKeyBind indicates keybinding failed.
	ErrFailedKeyBind = errors.New("failed to set key bindings")
	// ErrSignalCatch indicates that the signal has been caught.
	ErrSignalCatch = errors.New("signal caught")
	// ErrAlreadyClose indicates that it is already closed.
	ErrAlreadyClose = errors.New("already closed")
	// ErrCannotClose indicates that it cannot be closed.
	ErrCannotClose = errors.New("cannot be closed")
	// ErrRequestClose indicates that the request is to close.
	ErrRequestClose = errors.New("close requested")
	// ErrNoColumn indicates that cursor specified a nonexistent column.
	ErrNoColumn = errors.New("column not found")
	// ErrNoDelimiter indicates that the line containing the delimiter could not be found.
	ErrNoDelimiter = errors.New("delimiter not found")
	// ErrNoMoreSection indicates that the section could not be found.
	ErrNoMoreSection = errors.New("no more sections")
	// ErrOverScreen indicates that the specified screen is out of range.
	ErrOverScreen = errors.New("screen position out of range")
	// ErrOutOfChunk indicates that the specified Chunk is out of range.
	ErrOutOfChunk = errors.New("chunk out of range")
	// ErrNotLoaded indicates that it cannot be loaded.
	ErrNotLoaded = errors.New("not loaded")
	// ErrEOFreached indicates that EOF has been reached.
	ErrEOFreached = errors.New("EOF reached")
	// ErrPreventReload indicates that reload is prevented.
	ErrPreventReload = errors.New("reload prevented")
	// ErrOverChunkLimit indicates that the chunk limit has been exceeded.
	ErrOverChunkLimit = errors.New("chunk limit exceeded")
	// ErrAlreadyLoaded indicates that the chunk already loaded.
	ErrAlreadyLoaded = errors.New("chunk already loaded")
	// ErrEvictedMemory indicates that it has been evicted from memory.
	ErrEvictedMemory = errors.New("evicted from memory")
	// ErrNotAlignMode indicates that it is not an align mode.
	ErrNotAlignMode = errors.New("not in align mode")
	// ErrNoColumnSelected indicates that no column is selected.
	ErrNoColumnSelected = errors.New("no column selected")
	// ErrInvalidSGR indicates that the SGR is invalid.
	ErrInvalidSGR = errors.New("invalid SGR")
	// ErrNotSupport indicates that it is not supported.
	ErrNotSupport = errors.New("not supported")
	// ErrInvalidDocumentNum indicates that the document number is invalid.
	ErrInvalidDocumentNum = errors.New("invalid document number")
	// ErrInvalidModeName indicates that the specified view mode was not found.
	ErrInvalidModeName = errors.New("view mode not found")
	// ErrInvalidRGBColor indicates that the RGB color is invalid.
	ErrInvalidRGBColor = errors.New("invalid RGB color")
	// ErrInvalidKey indicates that the key format is invalid.
	ErrInvalidKey = errors.New("invalid key format")
)

// This is a function of tcell.NewScreen but can be replaced with mock.
var tcellNewScreen = tcell.NewScreen

// SetTcellNewScreen sets the function to create a new tcell screen.
// This is used for testing purposes to replace the tcell.NewScreen function.
// It allows for mocking the screen creation in tests.
func SetTcellNewScreen(f func() (tcell.Screen, error)) {
	tcellNewScreen = f
}

// NewOviewer return the structure of oviewer.
// NewOviewer requires one or more documents.
func NewOviewer(docs ...*Document) (*Root, error) {
	if len(docs) == 0 {
		return nil, ErrNotFound
	}
	root := &Root{
		minStartX:      MinStartX,
		settings:       NewRunTimeSettings(),
		Config:         NewConfig(),
		keyConfig:      cbind.NewConfiguration(),
		inputKeyConfig: cbind.NewConfiguration(),
		input:          NewInput(),
		sidebarScrolls: make(map[SidebarMode]sidebarScroll),
	}
	root.DocList = append(root.DocList, docs...)
	root.Doc = root.DocList[0]
	w, h := terminalSize()
	screen, err := virtualScreen(w, h)
	if err != nil {
		return nil, err
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
		root, err := openSTDIN()
		if err != nil {
			return nil, ErrMissingFile
		}
		return root, nil
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
	var openErrs []error
	docList := make([]*Document, 0)
	for _, fileName := range fileNames {
		m, err := OpenDocument(fileName)
		if err != nil {
			openErrs = append(openErrs, err)
			continue
		}
		docList = append(docList, m)
	}

	// If none of the documents are present, the program exits with an error.
	if len(docList) == 0 {
		openErrs = append([]error{ErrMissingFile}, openErrs...)
		errs := errors.Join(openErrs...)
		return nil, errs
	}

	root, err := NewOviewer(docList...)
	if err != nil {
		return nil, err
	}
	// Errors that could not be OpenDocument are output to the log.
	if len(openErrs) > 0 {
		for _, e := range openErrs {
			log.Println(e)
		}
	}

	return root, nil
}

// SetConfig sets config.
func (root *Root) SetConfig(config Config) {
	// Old Style* settings are loaded with lower priority.
	root.settings = setOldStyle(root.settings, config)
	// Old Prompt settings are loaded with lower priority.
	root.settings = setOldPrompt(root.settings, config)
	// General settings.
	root.settings = updateRunTimeSettings(root.settings, config.General)

	// view mode.
	if config.ViewMode != "" {
		viewMode, ok := config.Mode[config.ViewMode]
		if ok {
			root.settings = updateRunTimeSettings(root.settings, viewMode)
		} else {
			root.setMessageLogf("view mode not found: %s", config.ViewMode)
		}
	}

	if config.SidebarMode != "" {
		for s := SidebarModeHelp; s < SidebarModeEnd; s++ {
			if strings.EqualFold(config.SidebarMode, s.String()) {
				root.sidebarMode = s
				root.previousSidebarMode = s
				root.sidebarVisible = true
				break
			}
		}
	}

	// Set the follow mode for all documents.
	root.FollowAll = root.settings.FollowAll
	// Set the minimum start position of x.
	root.minStartX = config.MinStartX
	// Set the caption from the environment variable.
	if root.settings.Caption == "" {
		root.settings.Caption = viper.GetString("CAPTION")
	}

	// Actually tabs when "\t" is specified as an option.
	if root.settings.ColumnDelimiter == "\\t" {
		root.settings.ColumnDelimiter = "\t"
	}

	// SectionHeader is enabled if SectionHeaderNum is greater than 0.
	if root.settings.SectionHeaderNum > 0 {
		root.settings.SectionHeader = true
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
	root.mu.RLock()
	defer root.mu.RUnlock()

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
		root.debugMessage(fmt.Sprintf("notify send fail %v", request))
	}
}

// setKeyConfig sets key bindings.
func (root *Root) setKeyConfig(ctx context.Context) (map[string][]string, error) {
	keyBind := GetKeyBinds(root.Config)
	if err := root.setHandlers(ctx, keyBind); err != nil {
		return nil, err
	}

	keys, ok := keyBind[actionCancel]
	if !ok {
		log.Println("no cancel key")
	} else {
		root.cancelKeys = keys
	}
	return keyBind, nil
}

// SetKeyHandler assigns a new key handler.
func (root *Root) SetKeyHandler(ctx context.Context, name string, keys []string, handler func(context.Context)) error {
	return setHandler(ctx, root.keyConfig, name, keys, handler)
}

// Run starts the terminal pager.
func (root *Root) Run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer root.Close()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()
	root.SetWatcher(watcher)
	if err := root.prepareRun(ctx); err != nil {
		return err
	}
	root.prepareScreen()
	if root.quitCheck() {
		return nil
	}
	if err := root.switchToRealScreen(); err != nil {
		return err
	}

	if !root.Config.DisableMouse {
		root.Screen.EnableMouse(MouseFlags)
	}
	if root.Config.ClipboardMethod != "OSC52" && root.Config.ClipboardMethod != "system" {
		if root.Screen.HasClipboard() {
			root.Config.ClipboardMethod = "OSC52"
		} else {
			root.Config.ClipboardMethod = "system"
		}
		root.debugMessage("clipboard method: " + root.Config.ClipboardMethod)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	sigSuspend := registerSIGTSTP()
	quitChan := make(chan struct{})

	root.monitorEOF()
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

// prepareRun prepares to run the oviewer.
func (root *Root) prepareRun(ctx context.Context) error {
	// Do not set the key bindings in NewOviewer
	// because it is done after loading the config.
	keyBind, err := root.setKeyConfig(ctx)
	if err != nil {
		return err
	}
	help, err := NewHelp(keyBind)
	if err != nil {
		return err
	}
	root.helpDoc = help

	if root.Config.ShrinkChar != "" {
		Shrink = root.Config.ShrinkChar
		SetShrinkContent(Shrink)
	}

	root.setCaption()

	// Validate sidebar width.
	if _, err = calcSideWidth(root.Config.SidebarWidth, 100); err != nil {
		log.Printf("invalid sidebar width: %s, set to default %s\n", root.Config.SidebarWidth, defaultSidebarWidth)
		root.Config.SidebarWidth = defaultSidebarWidth
	}

	root.setViewModeConfig()
	root.prepareAllDocuments()
	// follow mode or follow all disables quit if the output fits on one screen.
	if root.Doc.FollowMode || root.FollowAll {
		root.Config.QuitSmall = false
		root.Config.QuitSmallFilter = false
	}
	// Quit by filter result. This is evaluated lazily.
	if root.Config.QuitSmallFilter {
		root.quitSmallCountDown = QuitSmallCountDown
		root.Config.QuitSmall = false
	}
	root.ViewSync(ctx)
	return nil
}

// setCaption sets the caption.
// optimizes execution with the Man command.
func (root *Root) setCaption() {
	if root.settings.Caption != "" {
		root.Doc.Caption = root.settings.Caption
		return
	}

	// Call from man command.
	manPN := os.Getenv("MAN_PN")
	if len(manPN) == 0 {
		return
	}

	root.Doc.Caption = manPN
}

// setViewModeConfig sets view mode config.
func (root *Root) setViewModeConfig() {
	list := ListViewMode(root.Config)
	root.input.Candidate[ViewMode].list = list
}

// prepareAllDocuments prepares all documents.
func (root *Root) prepareAllDocuments() {
	for n, doc := range root.DocList {
		doc.RunTimeSettings = root.settings
		doc.RunTimeSettings = updateRunTimeSettings(doc.RunTimeSettings, doc.General)
		doc.regexpCompile()

		if doc.FollowName {
			doc.FollowMode = true
		}
		if doc.ColumnWidth {
			doc.ColumnMode = true
		}
		w := ""
		if doc.WatchInterval > 0 {
			doc.watchMode()
			w = "(watch)"
		}
		log.Printf("open [%d]%s%s\n", n, doc.FileName, w)
	}

	root.helpDoc.RunTimeSettings = updateRunTimeSettings(root.helpDoc.RunTimeSettings, root.Config.HelpDoc)
	root.logDoc.RunTimeSettings = updateRunTimeSettings(root.logDoc.RunTimeSettings, root.Config.LogDoc)
}

// Close closes the oviewer.
func (root *Root) Close() {
	root.isClosed.Store(true)
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
	root.Screen.Show()
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
	root.Screen.Show()
}

// debugMessage outputs a debug message.
func (root *Root) debugMessage(msg string) {
	if !root.Config.Debug {
		return
	}
	if root.Doc == root.logDoc.Document {
		return
	}

	if len(msg) == 0 {
		return
	}
	log.Printf("%s:%s\n", root.Doc.FileName, msg)
}

// docSmall returns with bool whether the file to display fits on the screen.
func (root *Root) docSmall() bool {
	// The screen is switched to the virtual screen for writing to the original terminal,
	// so save the original screen and restore it later.
	screen := root.Screen
	defer func() {
		root.Screen = screen
	}()

	m := root.Doc
	if !m.BufEOF() {
		return false
	}
	w, h := root.Screen.Size()
	bodyHeight := h - root.scr.statusLineHeight
	if m.BufEndNum() > bodyHeight {
		return false
	}
	if err := root.createVirtualScreen(w, bodyHeight+1); err != nil {
		log.Println(err)
		return false
	}
	defer root.Screen.Fini()
	start, end, err := root.drawVirtualScreen()
	if err != nil {
		log.Println(err)
		return false
	}
	if (end - start) >= bodyHeight {
		return false
	}

	strs := tcellansi.ScreenContentToStrings(root.Screen, 0, root.Doc.bodyWidth, start, end)
	strs = tcellansi.TrimRightSpaces(strs)
	root.OnExit = strs
	return true
}

// OutputOnExit outputs to the terminal when exiting.
func (root *Root) OutputOnExit() {
	output := os.Stdout
	root.outputOnExit(output)
}

// WriteOriginal writes to the original terminal.
func (root *Root) WriteOriginal() {
	output := os.Stdout
	root.writeOriginal(output)
}

// WriteCurrentScreen writes to the current screen.
func (root *Root) WriteCurrentScreen() {
	output := os.Stdout
	root.writeCurrentScreen(output)
}

// outputOnExit outputs to the terminal when exiting.
func (root *Root) outputOnExit(output io.Writer) {
	if root.Config.IsWriteOnExit {
		if root.Config.IsWriteOriginal {
			root.writeOriginal(output)
		} else {
			root.writeCurrentScreen(output)
		}
	}
	if root.Config.Debug {
		root.writeLog(output)
	}
}

// writeCurrentScreen writes to the current screen.
func (root *Root) writeCurrentScreen(output io.Writer) {
	strs := root.OnExit
	if strs == nil {
		root.Screen.Fini()
		w, h := terminalSize()
		if err := root.createVirtualScreen(w, h); err != nil {
			log.Println(err)
			return
		}
		start, end, err := root.drawVirtualScreen()
		if err != nil {
			log.Println(err)
			return
		}
		strs = tcellansi.ScreenContentToStrings(root.Screen, 0, root.Doc.bodyWidth, start, end)
		strs = tcellansi.TrimRightSpaces(strs)
	}
	for _, str := range strs {
		if _, err := output.Write([]byte(str)); err != nil {
			log.Println(err)
			return
		}
	}
}

// switchToRealScreen switches to the real screen for drawing to the actual terminal.
func (root *Root) switchToRealScreen() error {
	virtual := root.Screen
	real, err := realScreen()
	if err != nil {
		return err
	}
	root.Screen = real

Loop:
	for {
		select {
		case ev := <-virtual.EventQ():
			real.EventQ() <- ev
		default:
			break Loop
		}
	}
	virtual.Fini()
	return nil
}

// realScreen creates a real screen for drawing to the actual terminal.
func realScreen() (tcell.Screen, error) {
	screen, err := tcellNewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}
	screen.Sync() // Because the size cannot be obtained unless Sync is performed.
	w, h := screen.Size()
	if w <= 0 || h <= 0 {
		screen.Fini()
		return nil, ErrNotTerminal
	}
	return screen, nil
}

// terminalSize returns the size of the terminal.
func terminalSize() (int, int) {
	width := 80
	height := 25
	fd := int(os.Stdout.Fd())
	w, h, err := term.GetSize(fd)
	if err == nil && w > 0 && h > 0 {
		width = w
		height = h
	}
	return width, height
}

// virtualScreen creates a virtual screen for drawing without affecting the actual terminal.
func virtualScreen(width int, height int) (tcell.Screen, error) {
	mt := vt.NewMockTerm(vt.MockOptSize{X: vt.Col(width), Y: vt.Row(height)})
	screen, err := tcell.NewTerminfoScreenFromTty(mt)
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}
	return screen, nil
}

// createVirtualScreen creates a virtual screen based on the configuration
// for writing to the original terminal.
func (root *Root) createVirtualScreen(width int, height int) error {
	if root.Config.BeforeWriteOriginal != 0 || root.Config.AfterWriteOriginal != 0 {
		root.Doc.topLN = max(root.Doc.topLN+root.Doc.firstLine()-root.Config.BeforeWriteOriginal, 0)
		end := root.Doc.bottomLN
		if root.Config.AfterWriteOriginal != 0 {
			end = root.Doc.topLN + root.Config.AfterWriteOriginal + 1
		}
		height = max(end-root.Doc.topLN, 0)
	}
	screen, err := virtualScreen(width, height)
	if err != nil {
		return err
	}
	root.Screen = screen
	return nil
}

// drawVirtualScreen draws to the virtual screen and returns the valid lines.
func (root *Root) drawVirtualScreen() (int, int, error) {
	if root.Pattern != "" {
		root.setSearcher(root.Pattern, root.Config.CaseSensitive)
	}
	ctx := context.Background()
	root.prepareScreen()
	root.prepareDraw(ctx)
	root.draw(ctx)
	start, end := contentRange(root.scr)
	return start, end, nil
}

// contentRange returns the range of valid content on the screen.
func contentRange(scr SCR) (int, int) {
	start := 0
	end := 0
	valid := false
	for y := 0; y < scr.vHeight-scr.statusLineHeight; y++ {
		if y >= len(scr.numbers) {
			break
		}
		if !scr.lines[scr.numbers[y].number].valid {
			continue
		}
		if !valid {
			start = y
			valid = true
		}
		end = y + 1
	}
	return start, end
}

// ScreenContent returns the screen content.
func (root *Root) ScreenContent() []string {
	root.Screen.Sync()
	m := root.Doc
	start, end := contentRange(root.scr)
	end += root.Config.AfterWriteOriginal
	start -= root.Config.BeforeWriteOriginal
	end = max(end, start)
	strs := tcellansi.ScreenContentToStrings(root.Screen, 0, m.bodyWidth, start, end)
	strs = tcellansi.TrimRightSpaces(strs)
	return strs
}

// writeOriginal writes to the original terminal.
func (root *Root) writeOriginal(output io.Writer) {
	m := root.Doc
	if m.bottomLN == 0 {
		m.bottomLN = m.BufEndNum()
	}

	ctx := context.Background()
	root.prepareDraw(ctx)

	// header
	header := max(0, root.scr.headerLN)
	headerEnd := max(0, root.scr.headerEnd)
	if root.Doc.headerHeight > 0 {
		if err := m.Export(output, header, headerEnd-1); err != nil {
			log.Println(err)
		}
	}
	// section header
	secAdd := 0
	if m.sectionHeaderHeight > 0 {
		if err := m.Export(output, root.scr.sectionHeaderLN, root.scr.sectionHeaderEnd-1); err != nil {
			log.Println(err)
		}
		secAdd = m.SectionHeaderNum
	}
	// body
	start := max(m.topLN+m.firstLine()+secAdd, root.scr.sectionHeaderEnd) - root.Config.BeforeWriteOriginal
	end := m.bottomLN - 1
	if root.Config.AfterWriteOriginal != 0 {
		end = m.topLN + root.Config.AfterWriteOriginal - 1
	}
	if err := m.Export(output, start, end); err != nil {
		log.Println(err)
	}
}

// WriteLog write to the log terminal.
func (root *Root) WriteLog() {
	root.writeLog(os.Stdout)
}

func (root *Root) writeLog(output io.Writer) {
	m := root.logDoc
	start := max(0, m.BufEndNum()-MaxWriteLog)
	end := m.BufEndNum()
	if err := m.Export(output, start, end); err != nil {
		log.Println(err)
	}
}

// lineNumber returns the line information from y on the screen.
func (scr SCR) lineNumber(y int) LineNumber {
	if len(scr.numbers) == 0 {
		return LineNumber{}
	}
	if y >= 0 && y < len(scr.numbers) {
		return scr.numbers[y]
	}
	return scr.numbers[0]
}

// lineY returns the screen y position for a given LineNumber.
// If not found, returns -1.
func (scr SCR) lineY(ln LineNumber) int {
	for y, num := range scr.numbers {
		if num == ln {
			return y
		}
	}
	return -1
}

// debugNumOfChunk outputs the number of chunks.
func (root *Root) debugNumOfChunk() {
	if !root.Config.Debug {
		return
	}
	log.Println("MemoryLimit:", root.Config.MemoryLimit)
	log.Println("MemoryLimitFile:", root.Config.MemoryLimitFile)
	for _, doc := range root.DocList {
		if !doc.seekable {
			if MemoryLimit > 0 {
				log.Printf("%s: The number of chunks is %d, of which %d(%v) are loaded\n", doc.FileName, len(doc.store.chunks), doc.store.loadedChunks.Len(), doc.store.loadedChunks.Keys())
			}
			continue
		}
		for n, chunk := range doc.store.chunks {
			if n != 0 && len(chunk.lines) != 0 {
				if !doc.store.loadedChunks.Contains(n) {
					log.Printf("chunk %d is not under control %d\n", n, len(chunk.lines))
				}
			}
		}
		log.Printf("%s(seekable): The number of chunks is %d, of which %d(%v) are loaded\n", doc.FileName, len(doc.store.chunks), doc.store.loadedChunks.Len(), doc.store.loadedChunks.Keys())
	}
}

// ListViewMode returns the list of view modes.
func ListViewMode(config Config) []string {
	list := slices.Sorted(maps.Keys(config.Mode))
	list = append([]string{nameGeneral}, list...)
	return list
}
