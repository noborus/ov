package oviewer

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cbind"
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

	// input contains the input mode.
	input *Input
	// keyConfig contains the binding settings for the key.
	keyConfig *cbind.Configuration

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
type general struct {
	// TabWidth is tab stop num.
	TabWidth int
	// HeaderLen is number of header rows to be fixed.
	Header int
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
}

// Config represents the settings of ov.
type Config struct {
	// StyleAlternate is a style that applies line by line.
	StyleAlternate ovStyle
	// StyleHeader is the style that applies to the header.
	StyleHeader ovStyle
	// StyleOverStrike is a style that applies to overstrikes.
	StyleOverStrike ovStyle
	// OverLineS is a style that applies to overstrike underlines.
	StyleOverLine ovStyle
	// StyleLineNumber is a style that applies line number.
	StyleLineNumber ovStyle

	// Old setting method.
	// Alternating background color.
	ColorAlternate string
	// Header color.
	ColorHeader string
	// OverStrike color.
	ColorOverStrike string
	// OverLine color.
	ColorOverLine string

	General general

	// Mouse support disable.
	DisableMouse bool
	// AfterWrite writes the current screen on exit.
	AfterWrite bool
	// QuiteSmall Quit if the output fits on one screen.
	QuitSmall bool
	// CaseSensitive is case-sensitive if true
	CaseSensitive bool
	// Debug represents whether to enable the debug output.
	Debug bool
	// KeyBinding
	Keybind map[string][]string
}

type ovStyle struct {
	Background    string
	Foreground    string
	Blink         bool
	Bold          bool
	Dim           bool
	Italic        bool
	Reverse       bool
	Underline     bool
	StrikeThrough bool
}

var (
	// OverStrikeStyle represents the overstrike style.
	OverStrikeStyle tcell.Style
	// OverLineStyle represents the overline underline style.
	OverLineStyle tcell.Style
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
	// ErrCancel indicates cancel.
	ErrCancel = errors.New("cancel")
	// ErrInvalidNumber indicates an invalid number.
	ErrInvalidNumber = errors.New("invalid number")
	// ErrFailedKeyBind indicates keybinding failed.
	ErrFailedKeyBind = errors.New("failed to set keybind")
	// ErrSignalCatch indicates that the signal has been caught.
	ErrSignalCatch = errors.New("signal catch")
)

// NewOviewer return the structure of oviewer.
func NewOviewer(docs ...*Document) (*Root, error) {
	root := &Root{
		minStartX: -10,
	}
	root.Config = NewConfig()
	root.keyConfig = cbind.NewConfiguration()
	root.DocList = append(root.DocList, docs...)
	root.Doc = root.DocList[0]
	root.input = NewInput()

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
		General: general{
			TabWidth: 8,
		},
	}
}

func (root *Root) screenInit() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err = screen.Init(); err != nil {
		return err
	}
	root.Screen = screen
	return nil
}

// Open reads the file named of the argument and return the structure of oviewer.
func Open(fileNames ...string) (*Root, error) {
	if len(fileNames) == 0 {
		return openSTDIN()
	}
	return openFiles(fileNames)
}

func openSTDIN() (*Root, error) {
	docList := make([]*Document, 0, 1)
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}
	err = m.ReadFile("")
	if err != nil {
		return nil, err
	}
	docList = append(docList, m)
	return NewOviewer(docList...)
}

func openFiles(fileNames []string) (*Root, error) {
	docList := make([]*Document, 0)
	for _, fileName := range fileNames {
		fi, err := os.Stat(fileName)
		if err != nil {
			log.Println(err, fileName)
			continue
		}
		if fi.IsDir() {
			continue
		}
		m, err := NewDocument()
		if err != nil {
			return nil, err
		}
		err = m.ReadFile(fileName)
		if err != nil {
			log.Println(err, fileName)
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

func (root *Root) setKeyConfig() error {
	keyBind := GetKeyBinds(root.Config.Keybind)
	if err := root.setKeyBind(keyBind); err != nil {
		return err
	}

	keys, ok := keyBind[actionCancel]
	if !ok {
		log.Printf("no cancel key")
	} else {
		root.cancelKeys = keys
	}

	help, err := NewHelp(keyBind)
	if err != nil {
		return err
	}
	root.helpDoc = help
	return nil
}

// NewHelp generates a document for help.
func NewHelp(k KeyBind) (*Document, error) {
	help, err := NewDocument()
	if err != nil {
		return nil, err
	}
	help.FileName = "Help"
	str := KeyBindString(k)
	help.lines = append(help.lines, "\t\t\tov help\n")
	help.lines = append(help.lines, strings.Split(str, "\n")...)
	help.eof = true
	help.endNum = len(help.lines)
	return help, err
}

// NewLogDoc generates a document for log.
func NewLogDoc() (*Document, error) {
	logDoc, err := NewDocument()
	if err != nil {
		return nil, err
	}
	logDoc.FileName = "Log"
	log.SetOutput(logDoc)
	return logDoc, nil
}

// Write matches the interface of io.Writer.
// Therefore, the log.Print output is displayed by logDoc.
func (logDoc *Document) Write(p []byte) (int, error) {
	str := string(p)
	logDoc.lines = append(logDoc.lines, str)
	logDoc.endNum = len(logDoc.lines)
	return len(str), nil
}

// Run starts the terminal pager.
func (root *Root) Run() error {
	for _, doc := range root.DocList {
		doc.general = root.Config.General
	}
	if err := root.setKeyConfig(); err != nil {
		return err
	}
	logDoc, err := NewLogDoc()
	if err != nil {
		return err
	}
	root.logDoc = logDoc

	if err := root.screenInit(); err != nil {
		return err
	}
	defer root.Screen.Fini()

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

	for _, d := range root.DocList {
		log.Printf("open %s", d.FileName)
	}
	root.setGlobalStyle()
	root.Screen.Clear()

	root.ViewSync()
	// Exit if fits on screen
	if root.QuitSmall && root.docSmall() {
		root.AfterWrite = true
		return nil
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)

	quitChan := make(chan struct{})

	go root.main(quitChan)

	for {
		select {
		case <-quitChan:
			return nil
		case sig := <-sigs:
			return fmt.Errorf("%w [%s]", ErrSignalCatch, sig)
		}
	}
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

// setDocument sets the Document.
func (root *Root) setDocument(m *Document) {
	root.Doc = m
	root.Clear()
	root.ViewSync()
}

// Help is to switch between Help screen and normal screen.
func (root *Root) Help() {
	if root.input.mode == Help {
		root.toNormal()
		return
	}
	root.toHelp()
}

func (root *Root) toHelp() {
	root.setDocument(root.helpDoc)
	root.input.mode = Help
}

// LogDisplay is to switch between Log screen and normal screen.
func (root *Root) logDisplay() {
	if root.input.mode == LogDoc {
		root.toNormal()
		return
	}
	root.toLogDoc()
}

func (root *Root) toLogDoc() {
	root.setDocument(root.logDoc)
	root.input.mode = LogDoc
}

func (root *Root) toNormal() {
	root.setDocument(root.DocList[root.CurrentDoc])
	root.input.mode = Normal
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
	hight := 0
	for y := 0; y < m.BufEndNum(); y++ {
		lc, err := m.lineToContents(y, root.Doc.TabWidth)
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
	for i := 0; i < root.vHight-1; i++ {
		n := m.topLN + i
		if n >= m.BufEndNum() {
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

// setWrapHeaderLen sets the value in wrapHeaderLen.
func (root *Root) setWrapHeaderLen() {
	m := root.Doc
	root.wrapHeaderLen = 0
	for y := 0; y < root.Doc.Header; y++ {
		lc, err := m.lineToContents(y, root.Doc.TabWidth)
		if err != nil {
			log.Println(err, "WrapHeaderLen", y)
			continue
		}
		root.wrapHeaderLen++
		root.wrapHeaderLen += ((len(lc) - 1) / (root.vWidth - root.startX))
	}
}

// bottomLineNum returns the display start line
// when the last line number as an argument.
func (root *Root) bottomLineNum(lN int) (int, int) {
	hight := (root.vHight - root.headerLen()) - 2
	if lN < root.headerLen() {
		return 0, 0
	}

	if !root.Doc.WrapMode {
		return 0, lN - (hight + root.headerLen())
	}

	// WrapMode
	lX, lN := root.findNumUp(0, lN, hight)
	return lX, lN - root.Doc.Header
}

// leftMostX returns a list of left - most x positions when wrapping.
func (root *Root) leftMostX(lN int) ([]int, error) {
	lc, err := root.Doc.lineToContents(lN, root.Doc.TabWidth)
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

// findNumUp finds lX, lN when the number of lines is moved up from lX, lN.
func (root *Root) findNumUp(lX int, lN int, upY int) (int, int) {
	listX, err := root.leftMostX(lN)
	n := 0
	if err != nil {
		// lN has no lines.
		root.debugMessage(fmt.Sprintf("%s:%d", err.Error(), lN))
	} else {
		n = numOfSlice(listX, lX)
	}

	for y := upY; y > 0; y-- {
		if n <= 0 {
			lN--
			if lN < root.Doc.Header {
				lN = 0
				lX = 0
				break
			}
			listX, err = root.leftMostX(lN)
			if err != nil {
				log.Println(err, "findNumUp", lN)
				return 0, 0
			}
			n = len(listX)
		}
		if n > 0 {
			lX = listX[n-1]
		} else {
			lX = 0
		}
		n--
	}
	return lX, lN
}

// toggleWrapMode toggles wrapMode each time it is called.
func (root *Root) toggleWrapMode() {
	root.Doc.WrapMode = !root.Doc.WrapMode
	root.Doc.x = 0
	root.setWrapHeaderLen()
	root.setMessage(fmt.Sprintf("Set WrapMode %t", root.Doc.WrapMode))
}

//  toggleColumnMode toggles ColumnMode each time it is called.
func (root *Root) toggleColumnMode() {
	root.Doc.ColumnMode = !root.Doc.ColumnMode
	root.setMessage(fmt.Sprintf("Set ColumnMode %t", root.Doc.ColumnMode))
}

// toggleAlternateRows toggles the AlternateRows each time it is called.
func (root *Root) toggleAlternateRows() {
	root.Doc.ClearCache()
	root.Doc.AlternateRows = !root.Doc.AlternateRows
	root.setMessage(fmt.Sprintf("Set AlternateRows %t", root.Doc.AlternateRows))
}

// toggleLineNumMode toggles LineNumMode every time it is called.
func (root *Root) toggleLineNumMode() {
	root.Doc.LineNumMode = !root.Doc.LineNumMode
	root.ViewSync()
	root.setMessage(fmt.Sprintf("Set LineNumMode %t", root.Doc.LineNumMode))
}

// resize is a wrapper function that calls viewSync.
func (root *Root) resize() {
	root.ViewSync()
}

// ViewSync redraws the whole thing.
func (root *Root) ViewSync() {
	root.resetSelect()
	root.prepareStartX()
	root.prepareView()
	root.Screen.Sync()
}

// TailSync move to tail and sync.
func (root *Root) TailSync() {
	root.moveBottom()
	root.ViewSync()
}

// prepareStartX prepares startX.
func (root *Root) prepareStartX() {
	root.startX = 0
	if root.Doc.LineNumMode {
		root.startX = len(fmt.Sprintf("%d", root.Doc.BufEndNum())) + 1
	}
}

// updateEndNum updates the last line number.
func (root *Root) updateEndNum() {
	root.debugMessage(fmt.Sprintf("Update EndNum:%d", root.Doc.endNum))
	root.prepareStartX()
	root.statusDraw()
}

// goLine will move to the specified line.
func (root *Root) goLine(input string) {
	lN, err := strconv.Atoi(input)
	if err != nil {
		root.setMessage(ErrInvalidNumber.Error())
		return
	}

	root.moveLine(lN - root.Doc.Header - 1)
	root.setMessage(fmt.Sprintf("Moved to line %d", lN))
}

// markLineNum stores the specified number of lines.
func (root *Root) markLineNum() {
	s := strconv.Itoa(root.Doc.topLN + 1)
	root.input.GoCandidate.list = toLast(root.input.GoCandidate.list, s)
	root.input.GoCandidate.p = 0
	root.setMessage(fmt.Sprintf("Marked to line %d", root.Doc.topLN))
}

// setHeader sets the number of lines in the header.
func (root *Root) setHeader(input string) {
	num, err := strconv.Atoi(input)
	if err != nil {
		root.setMessage(ErrInvalidNumber.Error())
		return
	}
	if num < 0 || num > root.vHight-1 {
		root.setMessage(ErrOutOfRange.Error())
		return
	}
	if root.Doc.Header == num {
		return
	}

	root.Doc.Header = num
	root.setMessage(fmt.Sprintf("Set Header %d", num))
	root.setWrapHeaderLen()
	root.Doc.ClearCache()
}

// setDelimiter sets the delimiter string.
func (root *Root) setDelimiter(input string) {
	root.Doc.ColumnDelimiter = input
	root.setMessage(fmt.Sprintf("Set delimiter %s", input))
}

// setTabWidth sets the tab width.
func (root *Root) setTabWidth(input string) {
	width, err := strconv.Atoi(input)
	if err != nil {
		root.setMessage(ErrInvalidNumber.Error())
		return
	}
	if root.Doc.TabWidth == width {
		return
	}

	root.Doc.TabWidth = width
	root.setMessage(fmt.Sprintf("Set tab width %d", width))
	root.Doc.ClearCache()
}

func (root *Root) markNext() {
	root.goLine(newGotoInput(root.input.GoCandidate).Up(""))
}

func (root *Root) markPrev() {
	root.goLine(newGotoInput(root.input.GoCandidate).Down(""))
}

func (root *Root) nextDoc() {
	root.CurrentDoc++
	root.CurrentDoc = min(root.CurrentDoc, len(root.DocList)-1)
	root.setDocument(root.DocList[root.CurrentDoc])
	root.input.mode = Normal
}

func (root *Root) previousDoc() {
	root.CurrentDoc--
	root.CurrentDoc = max(root.CurrentDoc, 0)
	root.setDocument(root.DocList[root.CurrentDoc])
	root.input.mode = Normal
}

func (root *Root) addDocument(m *Document) {
	log.Printf("add: %s", m.FileName)
	m.general = root.Config.General
	root.DocList = append(root.DocList, m)
	root.CurrentDoc = len(root.DocList) - 1
	root.setDocument(m)
}

func (root *Root) closeDocument() {
	if len(root.DocList) == 1 {
		return
	}
	s := root.DocList
	root.DocList = append(s[:root.CurrentDoc], s[root.CurrentDoc+1:]...)
	log.Printf("close : %s", root.Doc.FileName)
	if root.CurrentDoc > 0 {
		root.CurrentDoc--
	}
	root.setDocument(root.DocList[root.CurrentDoc])
}

func (root *Root) toggleMouse() {
	root.Config.DisableMouse = !root.Config.DisableMouse
	if root.Config.DisableMouse {
		root.Screen.DisableMouse()
		root.setMessage("Disable Mouse")
	} else {
		root.Screen.EnableMouse()
		root.setMessage("Enable Mouse")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
