package oviewer

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/term"
)

// toggleWrapMode toggles wrapMode each time it is called.
func (root *Root) toggleWrapMode(context.Context) {
	m := root.Doc
	m.WrapMode = !m.WrapMode
	root.resetSelect()

	// Move cursor to correct position
	x, err := m.optimalX(root.scr, m.columnCursor)
	if err != nil {
		root.setMessageLog(err.Error())
		return
	}
	m.scrollX = x
	root.setMessagef("Set WrapMode %t", m.WrapMode)
}

// toggleColumnMode toggles ColumnMode each time it is called.
func (root *Root) toggleColumnMode(context.Context) {
	root.Doc.ColumnMode = !root.Doc.ColumnMode

	if root.Doc.ColumnMode {
		root.prepareLines(root.scr.lines)
		root.Doc.columnCursor = root.Doc.optimalCursor(root.scr, root.Doc.columnCursor)
	}
	root.setMessagef("Set ColumnMode %t", root.Doc.ColumnMode)
}

// toggleColumnWidth toggles ColumnWidth each time it is called.
func (root *Root) toggleColumnWidth(context.Context) {
	if root.Doc.ColumnWidth {
		root.Doc.ColumnWidth = false
		root.Doc.ColumnMode = false
	} else {
		root.Doc.ColumnWidth = true
		root.Doc.ColumnMode = true
	}
	root.Doc.columnWidths = nil
	root.setMessagef("Set ColumnWidth %t", root.Doc.ColumnWidth)
}

// toggleAlternateRows toggles the AlternateRows each time it is called.
func (root *Root) toggleAlternateRows(context.Context) {
	root.Doc.AlternateRows = !root.Doc.AlternateRows
	root.setMessagef("Set AlternateRows %t", root.Doc.AlternateRows)
}

// toggleLineNumMode toggles LineNumMode every time it is called.
func (root *Root) toggleLineNumMode(ctx context.Context) {
	root.Doc.LineNumMode = !root.Doc.LineNumMode
	root.ViewSync(ctx)
	root.setMessagef("Set LineNumMode %t", root.Doc.LineNumMode)
}

// togglePlain toggles plain mode.
func (root *Root) togglePlain(context.Context) {
	root.Doc.PlainMode = !root.Doc.PlainMode
	root.setMessagef("Set PlainMode %t", root.Doc.PlainMode)
}

// toggleRainbow toggles the column rainbow mode.
// When enabled, each column will be displayed in a different color for better visual distinction.
func (root *Root) toggleRainbow(context.Context) {
	root.Doc.ColumnRainbow = !root.Doc.ColumnRainbow
	root.setMessagef("Set Column Rainbow Mode %t", root.Doc.ColumnRainbow)
}

// toggleFollowMode toggles follow mode.
func (root *Root) toggleFollowMode(context.Context) {
	root.Doc.FollowMode = !root.Doc.FollowMode
	root.Doc.pauseFollow = false
}

// toggleFollowAll toggles follow all mode.
func (root *Root) toggleFollowAll(context.Context) {
	root.FollowAll = !root.FollowAll
	root.Doc.pauseFollow = false
	root.mu.Lock()
	for _, doc := range root.DocList {
		doc.latestNum = doc.BufEndNum()
	}
	root.mu.Unlock()
}

// toggleFollowSection toggles follow section mode.
func (root *Root) toggleFollowSection(context.Context) {
	root.Doc.pauseFollow = false
	root.Doc.FollowSection = !root.Doc.FollowSection
}

// toggleHideOtherSection toggles hide other section mode.
func (root *Root) toggleHideOtherSection(context.Context) {
	root.Doc.HideOtherSection = !root.Doc.HideOtherSection
	root.setMessagef("Set HideOtherSection %t", root.Doc.HideOtherSection)
}

// toggleMouse toggles mouse control.
// When disabled, the mouse is controlled on the terminal side.
func (root *Root) toggleMouse(context.Context) {
	root.Config.DisableMouse = !root.Config.DisableMouse
	if root.Config.DisableMouse {
		root.Screen.DisableMouse()
		root.setMessage("Disable Mouse")
	} else {
		root.Screen.EnableMouse(MouseFlags)
		root.setMessage("Enable Mouse")
	}
}

// toggleStatusLine toggles the status line visibility.
func (root *Root) toggleStatusLine(context.Context) {
	root.Doc.StatusLine = !root.Doc.StatusLine
	if root.Doc.StatusLine {
		root.setMessage("Status Line visible")
	} else {
		root.setMessage("Status Line hidden")
	}
}

// toggleSidebar toggles the sidebar visibility.
func (root *Root) toggleSidebar(ctx context.Context, mode SidebarMode) {
	if mode == SidebarModeNone {
		root.closeSidebar(ctx)
		return
	}
	if root.sidebarVisible && root.sidebarMode == mode {
		root.closeSidebar(ctx)
		return
	}
	root.openSidebar(ctx, mode)
}

// openSidebar opens the sidebar with the specified mode.
func (root *Root) openSidebar(ctx context.Context, mode SidebarMode) {
	width, err := calcSideWidth(root.Config.SidebarWidth, root.scr.vWidth)
	if err != nil {
		root.setMessagef("Invalid sidebar width '%s': %s. Using default %s.", root.Config.SidebarWidth, err.Error(), defaultSidebarWidth)
		width, _ = calcSideWidth(defaultSidebarWidth, root.scr.vWidth)
	}
	root.sidebarMode = mode
	root.sidebarVisible = true
	root.sidebarWidth = width
	root.ViewSync(ctx)
	root.setMessagef("Sidebar %s visible", mode.String())
}

// calcSideWidth calculates the sidebar width based on the configuration string.
func calcSideWidth(sidewidth string, width int) (int, error) {
	w, err := calcPosition(sidewidth, width)
	if err != nil {
		return 0, err
	}
	w = min(max(minSidebarWidth, w), maxSidebarWidth)
	return int(w), nil
}

func (root *Root) closeSidebar(ctx context.Context) {
	root.sidebarVisible = false
	root.sidebarMode = SidebarModeNone
	root.sidebarWidth = 0
	root.ViewSync(ctx)
	root.setMessage("Sidebar hidden")
}

// toggleSidebarHelp toggles the help sidebar visibility.
func (root *Root) toggleSidebarHelp(ctx context.Context) {
	root.toggleSidebar(ctx, SidebarModeHelp)
}

// toggleSidebarMarks toggles the mark list sidebar visibility.
func (root *Root) toggleSidebarMarks(ctx context.Context) {
	root.toggleSidebar(ctx, SidebarModeMarks)
}

// toggleSidebarDocList toggles the document list sidebar visibility.
func (root *Root) toggleSidebarDocList(ctx context.Context) {
	root.toggleSidebar(ctx, SidebarModeDocList)
}

// toggleRuler cycles through the ruler types (None, Relative, Absolute) each time it is called.
func (root *Root) toggleRuler(ctx context.Context) {
	switch root.Doc.RulerType {
	case RulerNone:
		root.Doc.RulerType = RulerRelative
	case RulerRelative:
		root.Doc.RulerType = RulerAbsolute
	case RulerAbsolute:
		root.Doc.RulerType = RulerNone
	}
	root.setMessagef("Set Ruler %s", root.Doc.RulerType.String())
	root.resetSelect()
	root.ViewSync(ctx)
}

// closeFile requests the file to be closed.
func (root *Root) closeFile(context.Context) {
	if err := root.Doc.closeFile(); err != nil {
		root.setMessageLog(err.Error())
		return
	}
	root.setMessageLogf("close file %s", root.Doc.FileName)
}

// reload performs a reload of the current document.
func (root *Root) reload(m *Document) {
	if err := m.reload(); err != nil {
		root.setMessageLogf("cannot reload: %s", err)
		return
	}
	root.releaseEventBuffer()
	// Reserve time to read.
	time.Sleep(100 * time.Millisecond)
	root.setMessageLogf("reload file %s", root.Doc.FileName)
}

// toggleWatch toggles watch mode.
func (root *Root) toggleWatch(context.Context) {
	if root.Doc.WatchMode {
		root.Doc.unwatchMode()
	} else {
		root.Doc.watchMode()
	}
	atomic.StoreInt32(&root.Doc.watchRestart, 1)
}

// watchControl start/stop watch mode.
func (root *Root) watchControl() {
	m := root.Doc
	m.WatchInterval = max(m.WatchInterval, 1)
	if atomic.LoadInt32(&m.tickerState) == 1 {
		m.tickerDone <- struct{}{}
		<-m.tickerDone
	}
	if !root.Doc.WatchMode {
		return
	}
	log.Printf("watch start at interval: %d\n", m.WatchInterval)
	m.ticker = time.NewTicker(time.Duration(m.WatchInterval) * time.Second)
	atomic.StoreInt32(&m.tickerState, 1)
	go func() {
		for {
			select {
			case <-m.tickerDone:
				log.Println("watch stop")
				m.ticker.Stop()
				atomic.StoreInt32(&m.tickerState, 0)
				m.tickerDone <- struct{}{}
				return
			case <-m.ticker.C:
				root.sendReload(m)
			}
		}
	}()
}

// searchGo will go to the line with the matching term after searching.
// Jump by section if JumpTargetSection is true.
func (root *Root) searchGo(ctx context.Context, lN int, searcher Searcher) {
	root.setPauseFollow()
	root.resetSelect()
	root.Doc.lastSearchLN = lN
	start, end := root.searchXPos(lN, searcher)
	if root.Doc.jumpTargetSection {
		root.Doc.searchGoSection(ctx, lN, start, end)
		return
	}
	root.debugMessage(fmt.Sprintf("searchGo:%d->%d", root.Doc.topLN, lN))
	root.Doc.searchGoTo(lN, start, end)
}

// goLine will move to the specified line.
// decimal number > line number
// 10 -> line 10.
// decimal number + "." + decimal number > line number + number of wrapping lines
// 10.5 -> line 10 + 5 wrapping lines.
// "." + decimal is a percentage position
// .5 -> 50% of the way down the file.
// decimal + "%" is a percentage position
// 50% -> 50% of the way down the file.
func (root *Root) goLine(input string) {
	if len(input) == 0 {
		return
	}
	num, err := calcPosition(input, root.Doc.BufEndNum())
	if err != nil {
		root.setMessage(ErrInvalidNumber.Error())
		return
	}

	integerPart, fractionalPart := math.Modf(num)
	lN := int(integerPart)
	nTh := int(fractionalPart * 10)
	if nTh == 0 {
		lN = root.Doc.moveLine(lN - 1)
		root.Doc.showGotoF = true
		root.setMessagef("Moved to line %d", lN+1)
		return
	}
	lN, nTh = root.Doc.moveLineNth(lN-1, nTh)
	root.setMessagef("Moved to line %d.%d", lN+1, nTh)
}

// goLineNumber moves to the specified line number.
func (root *Root) goLineNumber(lN int) {
	lN = root.Doc.moveLine(lN - root.Doc.firstLine())
	root.setMessagef("Moved to line %d", lN+1)
}

// setHeader sets the number of lines in the header.
func (root *Root) setHeader(input string) {
	num, err := specifyOnScreen(input, root.scr.vHeight-1)
	if err != nil {
		root.setMessagef("Set header lines: %s", err.Error())
		return
	}
	if root.Doc.Header == num {
		return
	}

	root.Doc.Header = num
	root.Doc.columnWidths = nil
	root.setMessagef("Set header lines %d", num)
}

// setSkipLines sets the number of lines to skip.
func (root *Root) setSkipLines(input string) {
	num, err := specifyOnScreen(input, root.scr.vHeight-1)
	if err != nil {
		root.setMessagef("Set skip lines: %s", err.Error())
		return
	}
	if root.Doc.SkipLines == num {
		return
	}

	root.Doc.SkipLines = num
	root.Doc.columnWidths = nil
	root.setMessagef("Set skip lines %d", num)
}

// setSectionNum sets the number of section headers.
func (root *Root) setSectionNum(input string) {
	num, err := specifyOnScreen(input, root.scr.vHeight-1)
	if err != nil {
		root.setMessagef("Set section header num: %s", err.Error())
		return
	}
	if root.Doc.SectionHeaderNum == num {
		return
	}

	root.Doc.SectionHeaderNum = num
	root.setMessagef("Set section header num %d", num)
}

// specifyOnScreen checks if the input is a valid number.
func specifyOnScreen(input string, height int) (int, error) {
	num, err := strconv.Atoi(input)
	if err != nil {
		return 0, ErrInvalidNumber
	}
	if num < 0 || num > height {
		return 0, ErrOutOfRange
	}
	return num, nil
}

// suspend suspends the current screen display and runs the shell.
// It will return when you exit the shell.
func (root *Root) suspend(context.Context) {
	log.Println("Suspend")
	if err := root.Screen.Suspend(); err != nil {
		root.setMessageLog(err.Error())
		return
	}
	defer func() {
		log.Println("Resume")
		if err := root.Screen.Resume(); err != nil {
			log.Println(err)
		}
	}()

	subshell := os.Getenv("OV_SUBSHELL")
	// If the OS is something other than Windows
	// or if the environment variable Subshell is not set,
	// suspend with sigstop.
	if runtime.GOOS != "windows" && subshell != "1" {
		fmt.Println("suspended ov (use 'fg' to resume)")
		if err := suspendProcess(); err != nil {
			root.setMessageLog(err.Error())
		}
		return
	}

	// If the OS is Windows,
	// or if the environment variable Subshell is set,
	// start the subshell.
	fmt.Println("suspended ov (use 'exit' to resume)")
	shell := os.Getenv("SHELL")
	if err := root.subShell(shell); err != nil {
		root.setMessageLog(err.Error())
	}
}

func (root *Root) subShell(shell string) error {
	if shell == "" {
		shell = getShell()
	}
	stdin := os.Stdin
	if !term.IsTerminal(int(stdin.Fd())) {
		tty, err := getTTY()
		if err != nil {
			return fmt.Errorf("failed to get stdin: %w", err)
		}
		defer tty.Close()
		stdin = tty
	}
	c := exec.Command(shell, "-l")
	c.Stdin = stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return err
	}
	fmt.Println("resume ov")
	return nil
}

// setViewMode switches to the preset display mode.
func (root *Root) setViewMode(ctx context.Context, modeName string) {
	if modeName == "" {
		return
	}

	settings, err := root.modeConfig(modeName)
	if err != nil {
		root.setMessageLog(err.Error())
		return
	}
	m := root.Doc
	m.RunTimeSettings = settings
	m.conv = m.converterType(m.Converter)
	// Set caption.
	if settings.Caption != "" {
		m.Caption = settings.Caption
	}
	m.regexpCompile()
	m.ClearCache()
	root.ViewSync(ctx)
	root.setMessageLogf("Set mode %s", modeName)
}

// modeConfig returns the configuration of the specified mode.
func (root *Root) modeConfig(modeName string) (RunTimeSettings, error) {
	if modeName == nameGeneral {
		settings := updateRunTimeSettings(root.Doc.RunTimeSettings, root.Config.General)
		return settings, nil
	}

	viewMode, ok := root.Config.Mode[modeName]
	if !ok {
		return RunTimeSettings{}, fmt.Errorf("%w: %s", ErrInvalidModeName, modeName)
	}
	settings := updateRunTimeSettings(root.Doc.RunTimeSettings, viewMode)
	return settings, nil
}

// setConverter sets the converter type.
func (root *Root) setConverter(ctx context.Context, name string) {
	m := root.Doc
	if m.Converter == name {
		return
	}
	m.Converter = name
	m.conv = m.converterType(name)
	m.ClearCache()
	root.ViewSync(ctx)
	root.setMessagef("Set %s converter", name)
}

// alignFormat sets converter type to align.
func (root *Root) alignFormat(ctx context.Context) {
	if root.Doc.Converter == convAlign {
		root.esFormat(ctx)
		return
	}
	root.setConverter(ctx, convAlign)
}

// rawFormat sets converter type to raw.
func (root *Root) rawFormat(ctx context.Context) {
	if root.Doc.Converter == convRaw {
		root.esFormat(ctx)
		return
	}
	root.setConverter(ctx, convRaw)
}

// esFormat sets converter type to es.
func (root *Root) esFormat(ctx context.Context) {
	root.setConverter(ctx, convEscaped)
}

// setDelimiter sets the delimiter string.
func (root *Root) setDelimiter(input string) {
	root.Doc.setDelimiter(input)
	root.prepareLines(root.scr.lines)
	root.Doc.columnCursor = root.Doc.optimalCursor(root.scr, root.Doc.columnCursor)
	root.setMessagef("Set delimiter %s", input)
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
	root.setMessagef("Set tab width %d", width)
	root.Doc.ClearCache()
}

// setWatchInterval sets the Watch interval.
func (root *Root) setWatchInterval(input string) {
	interval, err := strconv.Atoi(input)
	if err != nil {
		root.setMessage(ErrInvalidNumber.Error())
		return
	}
	if root.Doc.WatchInterval == interval {
		return
	}

	root.Doc.WatchInterval = interval
	if root.Doc.WatchInterval == 0 {
		root.Doc.unwatchMode()
	} else {
		root.Doc.watchMode()
	}
	atomic.StoreInt32(&root.Doc.watchRestart, 1)
	root.setMessageLogf("Set watch interval %d", interval)
}

// setWriteBA sets the number before and after the line
// to be written at the end.
func (root *Root) setWriteBA(ctx context.Context, input string) {
	before, after, err := rangeBA(input)
	if err != nil {
		root.setMessage(ErrInvalidNumber.Error())
		return
	}
	root.Config.BeforeWriteOriginal = before
	root.Config.AfterWriteOriginal = after
	root.debugMessage(fmt.Sprintf("Before:After:%d:%d", root.Config.BeforeWriteOriginal, root.Config.AfterWriteOriginal))
	root.Config.IsWriteOnExit = true
	root.Quit(ctx)
}

// rangeBA returns the before after number from a string.
func rangeBA(str string) (int, int, error) {
	ba := strings.Split(str, ":")
	bstr := ba[0]
	if bstr == "" {
		bstr = "0"
	}
	before, err := strconv.Atoi(bstr)
	if err != nil {
		return 0, 0, err
	}

	if len(ba) == 1 {
		return before, 0, nil
	}

	astr := ba[1]
	if astr == "" {
		astr = "0"
	}
	after, err := strconv.Atoi(astr)
	if err != nil {
		return before, 0, err
	}
	return before, after, nil
}

// setSectionDelimiter sets the delimiter string.
func (root *Root) setSectionDelimiter(input string) {
	root.Doc.setSectionDelimiter(input)
	root.setMessagef("Set section delimiter %s", input)
}

// setSectionStart sets the section start position.
func (root *Root) setSectionStart(input string) {
	num, err := strconv.Atoi(input)
	if err != nil {
		root.setMessagef("Set section start position: %s", ErrInvalidNumber.Error())
		return
	}
	if num < -root.scr.vHeight || num > root.scr.vHeight-1 {
		root.setMessagef("Set section start position: %s", ErrOutOfRange.Error())
		return
	}
	root.Doc.SectionStartPosition = num
	root.setMessagef("Set section start position %s", input)
}

// setMultiColor set multiple strings to highlight with multiple colors.
func (root *Root) setMultiColor(input string) {
	quoted := false
	f := strings.FieldsFunc(input, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})

	root.Doc.setMultiColorWords(f)
	root.setMessagef("Set multicolor strings [%s]", input)
}

// setJumpTarget sets the position of the search result.
func (root *Root) setJumpTarget(input string) {
	if input == "" {
		return // no change
	}

	root.Doc.JumpTarget = input
	num, section := jumpPosition(input, root.scr.vHeight)
	root.Doc.jumpTargetSection = section
	if root.Doc.jumpTargetSection {
		root.setMessagef("Set JumpTarget section start")
		return
	}
	if num < 0 || num > root.scr.vHeight-1 {
		root.setMessagef("Set JumpTarget %d: %s", num, ErrOutOfRange.Error())
		return
	}
	if root.Doc.jumpTargetHeight == num {
		return
	}
	root.Doc.jumpTargetHeight = num
	root.setMessagef("Set JumpTarget %d", num)
}

// setVerticalHeader sets the vertical header position.
func (root *Root) setVerticalHeader(input string) {
	num, err := specifyOnScreen(input, root.scr.vWidth-1)
	if err != nil {
		root.setMessagef("Set vertical header: %s", err.Error())
		return
	}
	root.Doc.HeaderColumn = 0
	if root.Doc.VerticalHeader == num {
		return
	}

	root.Doc.VerticalHeader = num
	root.setMessagef("Set vertical header %d", num)
}

// setHeaderColumn sets the vertical header column position.
func (root *Root) setHeaderColumn(input string) {
	num, err := strconv.Atoi(input)
	if err != nil {
		root.setMessagef("Set vertical header column: %s", ErrInvalidNumber)
		return
	}
	root.Doc.VerticalHeader = 0
	if root.Doc.HeaderColumn == num {
		return
	}

	root.Doc.HeaderColumn = num
	root.setMessagef("Set vertical header column %d", num)
}

// resize is a wrapper function that calls viewSync.
func (root *Root) resize(ctx context.Context) {
	root.ViewSync(ctx)
}

// jumpPosition determines the position of the jump.
func jumpPosition(str string, height int) (int, bool) {
	s := strings.ToLower(strings.Trim(str, " "))
	if len(s) == 0 {
		return 0, false
	}
	if s[0] == 's' {
		return 0, true
	}
	n, err := calcPosition(s, height)
	if err != nil {
		return 0, false
	}
	num := int(math.Round(n))
	if num < 0 {
		return (height - 1) + num, false
	}
	return num, false
}

// calcPosition returns the number from the length for positive
// numbers (1), returns dot.number for percentages (.5) = 50%,
// and returns the % after the number for percentages (50%). return.
func calcPosition(str string, length int) (float64, error) {
	if len(str) == 0 || str == "0" {
		return 0, nil
	}

	var p float64 = 0
	if strings.HasPrefix(str, ".") {
		str = strings.TrimLeft(str, ".")
		i, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return 0, err
		}
		p = i / 10
	}
	if strings.HasSuffix(str, "%") {
		str = strings.TrimRight(str, "%")
		i, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return 0, err
		}
		p = i / 100
	}

	if p > 0 {
		return float64(length) * p, nil
	}

	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}

// TailSync move to tail and sync.
func (root *Root) TailSync(ctx context.Context) {
	if !root.Doc.pauseFollow {
		root.Doc.moveBottom()
	}
	root.ViewSync(ctx)
}

// tailSection moves to the last section
// and adjusts to its original position.
func (root *Root) tailSection(ctx context.Context) {
	if root.Doc.pauseFollow {
		return
	}
	moved := root.Doc.topLN - root.Doc.lastSectionPosNum
	root.Doc.moveLastSection(ctx)
	if moved > 0 && (root.Doc.topLN+moved) < root.Doc.BufEndNum() {
		root.Doc.moveLine(root.Doc.topLN + moved)
	}
	root.Doc.lastSectionPosNum = root.Doc.topLN
}

// updateEndNum updates the last line number.
func (root *Root) updateEndNum() {
	root.debugMessage(fmt.Sprintf("Update EndNum:%d", root.Doc.BufEndNum()))
	root.prepareStartX()
	root.drawStatus()
	root.Screen.Sync()
}

// updateLatestNum updates the last line number in follow mode.
func (root *Root) updateLatestNum() bool {
	num := root.Doc.BufEndNum()
	if root.Doc.latestNum == num {
		return false
	}
	root.skipDraw = false
	root.Doc.latestNum = num
	return true
}

// follow monitors and switches the document update
// in follow mode.
func (root *Root) follow(ctx context.Context) {
	if root.updateLatestNum() {
		root.TailSync(ctx)
	}
}

// followAll monitors and switches all document updates
// in follow all mode.
func (root *Root) followAll(ctx context.Context) {
	if root.Doc.documentType != DocNormal {
		return
	}

	current := root.CurrentDoc
	root.mu.RLock()
	for n, doc := range root.DocList {
		if doc.latestNum != doc.BufEndNum() {
			current = n
		}
	}
	root.mu.RUnlock()

	if root.CurrentDoc != current {
		root.switchDocument(ctx, current)
	}
	if root.updateLatestNum() {
		root.TailSync(ctx)
	}
}

// followSection monitors and switches the document update
// in follow section mode.
func (root *Root) followSection(ctx context.Context) {
	if root.updateLatestNum() {
		root.tailSection(ctx)
	}
}

// Cancel follow mode and follow all mode.
func (root *Root) Cancel(context.Context) {
	root.FollowAll = false
	root.Doc.FollowMode = false
}

// toggleWriteOriginal toggles the write flag.
func (root *Root) toggleWriteOriginal(context.Context) {
	root.Config.IsWriteOriginal = !root.Config.IsWriteOriginal
	root.setMessagef("Set WriteOriginal %t", root.Config.IsWriteOriginal)
}

// WriteQuit sets the write flag and executes a quit event.
func (root *Root) WriteQuit(ctx context.Context) {
	if root.Doc.documentType == DocHelp || root.Doc.documentType == DocLog {
		log.Println("WriteQuit: not a normal document")
		root.Quit(ctx)
		return
	}

	root.Config.IsWriteOnExit = true
	if root.Doc.HideOtherSection && root.Config.AfterWriteOriginal == 0 {
		// hide other section.
		root.Config.AfterWriteOriginal = root.bottomSectionLN(ctx)
	}

	// Do not write if BeforeWriteOriginal is set (greater than 0).
	if root.Config.BeforeWriteOriginal > 0 {
		root.Quit(ctx)
		return
	}

	root.OnExit = root.ScreenContent()

	root.Quit(ctx)
}

// bottomSectionLN returns the number of lines to write.
func (root *Root) bottomSectionLN(ctx context.Context) int {
	if root.Doc.SectionDelimiter == "" {
		return root.Config.AfterWriteOriginal
	}
	lN, err := root.Doc.nextSection(ctx, root.Doc.topLN+root.Doc.firstLine()-root.Doc.SectionStartPosition)
	if err != nil {
		return root.Config.AfterWriteOriginal
	}
	return lN - (root.Doc.topLN + root.Doc.firstLine() - root.Doc.SectionStartPosition)
}

// toggleFixedColumn toggles the fixed column.
func (root *Root) toggleFixedColumn(context.Context) {
	cursor := root.Doc.columnCursor - root.Doc.columnStart + 1
	if root.Doc.HeaderColumn == cursor {
		cursor = 0
	}
	root.Doc.HeaderColumn = cursor
	root.Doc.VerticalHeader = 0
	root.setMessagef("Set Header Column %d", cursor)
}

// ShrinkColumn shrinks the specified column.
func (root *Root) ShrinkColumn(_ context.Context, cursor int) error {
	return root.Doc.shrinkColumn(cursor, true)
}

// ExpandColumn expands the specified column.
func (root *Root) ExpandColumn(_ context.Context, cursor int) error {
	return root.Doc.shrinkColumn(cursor, false)
}

// toggleShrinkColumn shrinks or expands the current cursor column.
func (root *Root) toggleShrinkColumn(_ context.Context) {
	cursor := root.Doc.columnCursor
	shrink, err := root.Doc.isColumnShrink(cursor)
	if err != nil {
		root.setMessage(err.Error())
	}
	if err := root.Doc.shrinkColumn(cursor, !shrink); err != nil {
		root.setMessage(err.Error())
	}
}

// shrinkColumn shrinks or expands the specified column.
func (m *Document) shrinkColumn(cursor int, shrink bool) error {
	if err := m.isValidColumn(cursor); err != nil {
		return err
	}
	m.alignConv.columnAttrs[cursor].shrink = shrink
	m.ClearCache()
	return nil
}

// isColumnShrink returns whether the specified column is shrink.
func (m *Document) isColumnShrink(cursor int) (bool, error) {
	if err := m.isValidColumn(cursor); err != nil {
		return false, err
	}
	return m.alignConv.columnAttrs[cursor].shrink, nil
}

// toggleRightAlign toggles the right align of the current cursor column.
func (root *Root) toggleRightAlign(_ context.Context) {
	m := root.Doc
	align, err := m.toggleRightAlign(m.columnCursor)
	if err != nil {
		root.setMessage(err.Error())
		return
	}
	root.setMessagef("Set %s", align)
}

// toggleRightAlign toggles the right align of the specified column.
func (m *Document) toggleRightAlign(cursor int) (specifiedAlign, error) {
	if err := m.isValidColumn(cursor); err != nil {
		return Unspecified, err
	}
	align := m.alignConv.columnAttrs[cursor].specifiedAlign + 1
	if align > LeftAlign {
		align = RightAlign
	}
	m.alignConv.columnAttrs[cursor].specifiedAlign = align
	m.ClearCache()
	return align, nil
}

// isValidColumn checks if the specified column is valid.
func (m *Document) isValidColumn(cursor int) error {
	if m.Converter != convAlign {
		return ErrNotAlignMode
	}
	if cursor < 0 || cursor >= len(m.alignConv.columnAttrs) {
		return ErrNoColumnSelected
	}
	return nil
}

// setPauseFollow sets pauseFollow to true if sticky follow is not disabled.
func (root *Root) setPauseFollow() {
	// If sticky follow is disabled, do nothing.
	if root.Config.DisableStickyFollow {
		return
	}
	// If already paused, do nothing.
	if root.Doc.pauseFollow {
		return
	}
	if root.FollowAll || root.Doc.FollowMode || root.Doc.FollowSection {
		root.Doc.pauseFollow = true
		root.Doc.pauseLastNum = max(root.Doc.BufEndNum()-1, 0)
	}
}
