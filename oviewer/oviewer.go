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
	"github.com/gdamore/tcell/v2"
)

// Root is the root structure of the oviewer.
type Root struct {
	// tcell.Screen is the root screen.
	tcell.Screen
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
	Config
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

	// skipDraw is set to true when the mouse cursor just moves (no event occurs).
	skipDraw bool
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
	// startX is the start position of x.
	startX int
	// startY is the start position of y.
	startY int
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
	mouseSelect bool
	// mousePressed is a flag when the mouse selection button is pressed.
	mousePressed bool
	// mouseRectangle is a flag for rectangle selection.
	mouseRectangle bool
}

// LineNumber is Number of logical lines and number of wrapping lines on the screen.
type LineNumber struct {
	number int
	wrap   int
}

func newLineNumber(number, wrap int) LineNumber {
	return LineNumber{number: number, wrap: wrap}
}

const MinStartX = -10

// RulerType is the type of ruler.
type RulerType int

const (
	RulerNone RulerType = iota
	RulerRelative
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

// general structure contains the general of the display.
// general contains values that determine the behavior of each document.
type general struct {
	// Converter is the converter name.
	Converter string
	// Caption is an additional caption to display after the file name.
	Caption string
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
	// Header is number of header lines to be fixed.
	Header int
	// VerticalHeader is the number of vertical header lines.
	VerticalHeader int
	// HeaderColumn is the number of columns from the left to be fixed.
	// If 0 is specified, no columns are fixed.
	HeaderColumn int
	// SkipLines is the rows to skip.
	SkipLines int
	// WatchInterval is the watch interval (seconds).
	WatchInterval int
	// MarkStyleWidth is width to apply the style of the marked line.
	MarkStyleWidth int
	// SectionStartPosition is a section start position.
	SectionStartPosition int
	// SectionHeaderNum is the number of lines in the section header.
	SectionHeaderNum int
	// HScrollWidth is the horizontal scroll width.
	HScrollWidth string
	// HScrollWidthNum is the horizontal scroll width.
	HScrollWidthNum int
	// RulerType is the ruler type (0: none, 1: relative, 2: absolute).
	RulerType RulerType
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
	// SectionHeader is whether to display the section header.
	SectionHeader bool
	// HideOtherSection is whether to hide other sections.
	HideOtherSection bool
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

// The name of the converter that can be specified.
const (
	convEscaped string = "es"    // convEscaped processes escape sequence(default).
	convRaw     string = "raw"   // convRaw is displayed without processing escape sequences as they are.
	convAlign   string = "align" // convAlign is aligned in each column.
)

const (
	nameGeneral string = "general"
)

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
	// ErrNotTerminal indicates that it is not a terminal.
	ErrNotTerminal = errors.New("not a terminal")
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
	// ErrCannotClose indicates that it cannot be closed.
	ErrCannotClose = errors.New("cannot close")
	// ErrRequestClose indicates that the request is to close.
	ErrRequestClose = errors.New("request close")
	// ErrNoColumn indicates that cursor specified a nonexistent column.
	ErrNoColumn = errors.New("no column")
	// ErrNoDelimiter indicates that the line containing the delimiter could not be found.
	ErrNoDelimiter = errors.New("no delimiter")
	// ErrNoMoreSection indicates that the section could not be found.
	ErrNoMoreSection = errors.New("no more section")
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
	// ErrNotAlignMode indicates that it is not an align mode.
	ErrNotAlignMode = errors.New("not an align mode")
	// ErrNoColumnSelected indicates that no column is selected.
	ErrNoColumnSelected = errors.New("no column selected")
	// ErrInvalidSGR indicates that the SGR is invalid.
	ErrInvalidSGR = errors.New("invalid SGR")
	// ErrNotSupport indicates that it is not supported.
	ErrNotSupport = errors.New("not support")
	// ErrInvalidDocumentNum indicates that the document number is invalid.
	ErrInvalidDocumentNum = errors.New("invalid document number")
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
		minStartX: MinStartX,
	}
	root.Config = NewConfig()
	root.keyConfig = cbind.NewConfiguration()
	root.inputKeyConfig = cbind.NewConfiguration()
	root.DocList = append(root.DocList, docs...)
	root.Doc = root.DocList[0]
	root.input = NewInput()

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

	if len(openErrs) > 1 {
		openErrs = append([]error{ErrMissingFile}, openErrs...)
	}
	errs := errors.Join(openErrs...)
	// If none of the documents are present, the program exits with an error.
	if len(docList) == 0 {
		return nil, errs
	}

	root, err := NewOviewer(docList...)
	if err != nil {
		return nil, err
	}
	// Errors that could not be OpenDocument are output to the log.
	if errs, ok := errs.(interface{ Unwrap() []error }); ok {
		for _, err := range errs.Unwrap() {
			log.Println(err)
		}
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
		log.Printf("no cancel key")
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

	// Quit if fits on screen.
	if root.QuitSmall && root.DocumentLen() == 1 && root.docSmall() {
		root.IsWriteOriginal = true
		return nil
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

	if !root.Config.DisableMouse {
		root.Screen.EnableMouse(MouseFlags)
	}

	if root.Config.ShrinkChar != "" {
		Shrink = []rune(root.Config.ShrinkChar)[0]
	}
	SetShrinkContent(Shrink)

	root.setCaption()

	root.setViewModeConfig()
	root.prepareAllDocuments()
	// Quit by filter result. This is evaluated lazily.
	if root.QuitSmallFilter {
		root.quitSmallCountDown = QuitSmallCountDown
		root.QuitSmall = false
	}
	root.ViewSync(ctx)
	return nil
}

// setCaption sets the caption.
// optimizes execution with the Man command.
func (root *Root) setCaption() {
	if root.General.Caption != "" {
		root.Doc.Caption = root.General.Caption
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
	list := make([]string, 0, len(root.Config.Mode)+1)
	list = append(list, nameGeneral)
	for name := range root.Config.Mode {
		list = append(list, name)
	}
	root.input.Candidate[ViewMode].list = list
}

// prepareAllDocuments prepares all documents.
func (root *Root) prepareAllDocuments() {
	for n, doc := range root.DocList {
		doc.general = root.Config.General
		doc.regexpCompile()

		if doc.FollowName {
			doc.FollowMode = true
		}
		if doc.ColumnWidth {
			doc.ColumnMode = true
		}
		if doc.general.Converter != "" {
			doc.conv = doc.converterType(doc.general.Converter)
		}
		w := ""
		if doc.general.WatchInterval > 0 {
			doc.watchMode()
			w = "(watch)"
		}
		log.Printf("open [%d]%s%s", n, doc.FileName, w)
	}
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
	if root.Doc == root.logDoc.Document {
		return
	}

	if len(msg) == 0 {
		return
	}
	log.Printf("%s:%s", root.Doc.FileName, msg)
}

// mergeGeneral overwrites a general structure with a struct.
func mergeGeneral(src general, dst general) general {
	if dst.Converter != "" {
		src.Converter = dst.Converter
	}
	if dst.ColumnDelimiter != "" {
		src.ColumnDelimiter = dst.ColumnDelimiter
	}
	if dst.SectionDelimiter != "" {
		src.SectionDelimiter = dst.SectionDelimiter
	}
	if dst.JumpTarget != "" {
		src.JumpTarget = dst.JumpTarget
	}
	if len(dst.MultiColorWords) > 0 {
		src.MultiColorWords = dst.MultiColorWords
	}
	if dst.TabWidth != 0 {
		src.TabWidth = dst.TabWidth
	}
	if dst.Header != 0 {
		src.Header = dst.Header
	}
	if dst.VerticalHeader != 0 {
		src.VerticalHeader = dst.VerticalHeader
	}
	if dst.HeaderColumn != 0 {
		src.HeaderColumn = dst.HeaderColumn
	}
	if dst.SkipLines != 0 {
		src.SkipLines = dst.SkipLines
	}
	if dst.WatchInterval != 0 {
		src.WatchInterval = dst.WatchInterval
	}
	if dst.MarkStyleWidth != 0 {
		src.MarkStyleWidth = dst.MarkStyleWidth
	}
	if dst.SectionStartPosition != 0 {
		src.SectionStartPosition = dst.SectionStartPosition
	}
	if dst.SectionHeaderNum != 0 {
		src.SectionHeaderNum = dst.SectionHeaderNum
	}
	if dst.HScrollWidth != "" {
		src.HScrollWidth = dst.HScrollWidth
	}
	if dst.HScrollWidthNum != 0 {
		src.HScrollWidthNum = dst.HScrollWidthNum
	}
	if dst.RulerType != RulerNone {
		src.RulerType = dst.RulerType
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
	if dst.SectionHeader {
		src.SectionHeader = dst.SectionHeader
	}
	if dst.HideOtherSection {
		src.HideOtherSection = dst.HideOtherSection
	}
	if dst.PlainMode {
		src.PlainMode = dst.PlainMode
	}
	return src
}

// docSmall returns with bool whether the file to display fits on the screen.
func (root *Root) docSmall() bool {
	root.prepareScreen()
	m := root.Doc
	if !m.BufEOF() {
		return false
	}
	height := 0
	for y := 0; y < m.BufEndNum(); y++ {
		lc, err := m.contents(y)
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
	output := os.Stdout
	root.writeOriginal(output)
}

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
	start := max(m.topLN+m.firstLine()+secAdd, root.scr.sectionHeaderEnd) - root.BeforeWriteOriginal
	end := m.bottomLN - 1
	if root.AfterWriteOriginal != 0 {
		end = m.topLN + root.AfterWriteOriginal - 1
	}
	if err := m.Export(output, start, end); err != nil {
		log.Println(err)
	}
}

// WriteLog write to the log terminal.
func (root *Root) WriteLog() {
	root.writeLog(os.Stdout)
}

func (root *Root) writeLog(w io.Writer) {
	m := root.logDoc
	start := max(0, m.BufEndNum()-MaxWriteLog)
	end := m.BufEndNum()
	if err := m.Export(w, start, end); err != nil {
		log.Println(err)
	}
}

// lineNumber returns the line information from y on the screen.
func (scr SCR) lineNumber(y int) LineNumber {
	if len(scr.numbers) == 0 {
		return LineNumber{}
	}
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
