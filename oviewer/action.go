package oviewer

import (
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
)

// toggleWrapMode toggles wrapMode each time it is called.
func (root *Root) toggleWrapMode() {
	m := root.Doc
	m.WrapMode = !m.WrapMode

	// Move cursor to correct position
	x, err := m.optimalX(m.columnCursor)
	if err != nil {
		root.setMessageLog(err.Error())
		return
	}
	// Move if off screen
	if x < m.x || x > m.x+(root.scr.vWidth-root.scr.startX) {
		m.x = x
	}
	root.setMessagef("Set WrapMode %t", m.WrapMode)
}

// toggleColumnMode toggles ColumnMode each time it is called.
func (root *Root) toggleColumnMode() {
	root.Doc.ColumnMode = !root.Doc.ColumnMode

	if root.Doc.ColumnMode {
		root.Doc.columnCursor = root.Doc.optimalCursor(root.Doc.columnCursor)
	}
	root.setMessagef("Set ColumnMode %t", root.Doc.ColumnMode)
}

// toggleColumnWidth toggles ColumnWidth each time it is called.
func (root *Root) toggleColumnWidth() {
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
func (root *Root) toggleAlternateRows() {
	root.Doc.AlternateRows = !root.Doc.AlternateRows
	root.setMessagef("Set AlternateRows %t", root.Doc.AlternateRows)
}

// toggleLineNumMode toggles LineNumMode every time it is called.
func (root *Root) toggleLineNumMode() {
	root.Doc.LineNumMode = !root.Doc.LineNumMode
	root.ViewSync()
	root.setMessagef("Set LineNumMode %t", root.Doc.LineNumMode)
}

// togglePlain toggles plain mode.
func (root *Root) togglePlain() {
	root.Doc.PlainMode = !root.Doc.PlainMode
	root.setMessagef("Set PlainMode %t", root.Doc.PlainMode)
}

// togglePlain toggles column rainbow mode.
func (root *Root) toggleRainbow() {
	root.Doc.ColumnRainbow = !root.Doc.ColumnRainbow
	root.setMessagef("Set Column Rainbow Mode %t", root.Doc.ColumnRainbow)
}

// toggleFollowMode toggles follow mode.
func (root *Root) toggleFollowMode() {
	root.Doc.FollowMode = !root.Doc.FollowMode
}

// toggleFollowAll toggles follow all mode.
func (root *Root) toggleFollowAll() {
	root.General.FollowAll = !root.General.FollowAll
	root.mu.Lock()
	for _, doc := range root.DocList {
		doc.latestNum = doc.BufEndNum()
	}
	root.mu.Unlock()
}

// toggleFollowSection toggles follow section mode.
func (root *Root) toggleFollowSection() {
	root.Doc.FollowSection = !root.Doc.FollowSection
}

// closeFile close the file.
func (root *Root) closeFile() {
	if root.Doc.documentType == DocHelp || root.Doc.documentType == DocLog {
		return
	}

	if root.Doc.checkClose() {
		root.setMessage("already closed")
		return
	}
	if root.Doc.seekable {
		root.setMessage("cannot close")
		return
	}
	root.Doc.requestClose()
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
}

// toggleWatch toggles watch mode.
func (root *Root) toggleWatch() {
	if root.Doc.WatchMode {
		root.Doc.unWatchMode()
	} else {
		root.Doc.watchMode()
	}
	root.Doc.watchRestart.Store(true)
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
	log.Printf("watch start at interval %d", m.WatchInterval)
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
func (root *Root) searchGo(lN int) {
	root.resetSelect()
	x := root.searchXPos(lN)
	if root.Doc.jumpTargetSection {
		root.Doc.searchGoSection(lN, x)
		return
	}
	root.Doc.searchGoTo(lN, x)
}

// goLine will move to the specified line.
// decimal number > line number
// 10 -> line 10
// decimal number + "." + decimal number > line number + number of wrapping lines
// 10.5 -> line 10 + 5 wrapping lines
// "." + decimal is a percentage position
// .5 -> 50% of the way down the file
// decimal + "%" is a percentage position
// 50% -> 50% of the way down the file
func (root *Root) goLine(input string) {
	if len(input) == 0 {
		return
	}
	num := calculatePosition(root.Doc.BufEndNum(), input)
	str := strconv.FormatFloat(num, 'f', 1, 64)
	if strings.HasSuffix(str, ".0") {
		// Line number only.
		lN, err := strconv.Atoi(str[:len(str)-2])
		if err != nil {
			root.setMessage(ErrInvalidNumber.Error())
			return
		}
		lN = root.Doc.moveLine(lN - 1)
		root.setMessagef("Moved to line %d", lN+1)
		return
	}

	// Line number and number of wrapping lines.
	inputs := strings.Split(str, ".")
	lN, err := strconv.Atoi(inputs[0])
	if err != nil {
		root.setMessage(ErrInvalidNumber.Error())
		return
	}
	nTh, err := strconv.Atoi(inputs[1])
	if err != nil {
		root.setMessage(ErrInvalidNumber.Error())
		return
	}
	lN, nTh = root.Doc.moveLineNth(lN-1, nTh)
	root.setMessagef("Moved to line %d.%d", lN+1, nTh)
}

// goLineNumber moves to the specified line number.
func (root *Root) goLineNumber(ln int) {
	ln = root.Doc.moveLine(ln - root.Doc.firstLine())
	root.setMessagef("Moved to line %d", ln+1)
}

// markNext moves to the next mark.
func (root *Root) markNext() {
	if len(root.Doc.marked) == 0 {
		return
	}

	if len(root.Doc.marked) > root.Doc.markedPoint+1 {
		root.Doc.markedPoint++
	} else {
		root.Doc.markedPoint = 0
	}
	root.goLineNumber(root.Doc.marked[root.Doc.markedPoint])
}

// markPrev moves to the previous mark.
func (root *Root) markPrev() {
	if len(root.Doc.marked) == 0 {
		return
	}

	if root.Doc.markedPoint > 0 {
		root.Doc.markedPoint--
	} else {
		root.Doc.markedPoint = len(root.Doc.marked) - 1
	}
	root.goLineNumber(root.Doc.marked[root.Doc.markedPoint])
}

// addMark marks the current line number.
func (root *Root) addMark() {
	c := min(root.Doc.topLN+root.Doc.firstLine(), root.Doc.BufEndNum())
	root.Doc.marked = remove(root.Doc.marked, c)
	root.Doc.marked = append(root.Doc.marked, c)
	root.setMessagef("Marked to line %d", c-root.Doc.firstLine()+1)
}

// removeMark removes the current line number from the mark.
func (root *Root) removeMark() {
	c := root.Doc.topLN + root.Doc.firstLine()
	marked := remove(root.Doc.marked, c)
	if len(root.Doc.marked) == len(marked) {
		root.setMessagef("Not marked line %d", c-root.Doc.firstLine()+1)
		return
	}
	root.Doc.marked = marked
	root.setMessagef("Remove the mark at line %d", c-root.Doc.firstLine()+1)
}

// removeAllMark removes all marks.
func (root *Root) removeAllMark() {
	root.Doc.marked = nil
	root.Doc.markedPoint = 0
	root.setMessage("Remove all marks")
}

// setHeader sets the number of lines in the header.
func (root *Root) setHeader(input string) {
	num, err := strconv.Atoi(input)
	if err != nil {
		root.setMessagef("Set header: %s", ErrInvalidNumber.Error())
		return
	}
	if num < 0 || num > root.scr.vHeight-1 {
		root.setMessagef("Set header %d: %s", num, ErrOutOfRange.Error())
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
	num, err := strconv.Atoi(input)
	if err != nil {
		root.setMessagef("Set skip line: %s", ErrInvalidNumber.Error())
		return
	}
	if num < 0 {
		root.setMessagef("Set skip line: %s", ErrOutOfRange.Error())
		return
	}
	if root.Doc.SkipLines == num {
		return
	}

	root.Doc.SkipLines = num
	root.setMessagef("Set skip lines %d", num)
}

func (root *Root) setSectionNum(input string) {
	num, err := strconv.Atoi(input)
	if err != nil {
		root.setMessagef("Set section header num: %s", ErrInvalidNumber.Error())
		return
	}
	if num < 0 {
		root.setMessagef("Set section header num: %s", ErrOutOfRange.Error())
		return
	}
	if root.Doc.SectionHeaderNum == num {
		return
	}

	root.Doc.SectionHeaderNum = num
	root.setMessagef("Set section header num %d", num)
}

// suspend suspends the current screen display and runs the shell.
// It will return when you exit the shell.
func (root *Root) suspend() {
	log.Println("Suspend")
	if err := root.Screen.Suspend(); err != nil {
		log.Println(err)
		return
	}
	fmt.Println("suspended ov")
	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = "CMD.EXE"
		} else {
			shell = "/bin/sh"
		}
	}
	c := exec.Command(shell, "-l")
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		root.setMessageLog(err.Error())
	}
	fmt.Println("resume ov")
	if err := root.Screen.Resume(); err != nil {
		root.setMessageLog(err.Error())
	}
	log.Println("Resume")
}

// toggleMouse toggles mouse control.
// When disabled, the mouse is controlled on the terminal side.
func (root *Root) toggleMouse() {
	root.Config.DisableMouse = !root.Config.DisableMouse
	if root.Config.DisableMouse {
		root.Screen.DisableMouse()
		root.setMessage("Disable Mouse")
	} else {
		root.Screen.EnableMouse(MouseFlags)
		root.setMessage("Enable Mouse")
	}
}

// setViewMode switches to the preset display mode.
// Set header lines and columnMode together.
func (root *Root) setViewMode(modeName string) {
	c, ok := root.Config.Mode[modeName]
	if !ok {
		if modeName != "general" {
			root.setMessagef("%s mode not found", modeName)
			return
		}
		c = root.General
	}

	root.Doc.general = mergeGeneral(root.Doc.general, c)
	root.Doc.regexpCompile()
	root.Doc.ClearCache()
	root.ViewSync()
	root.setMessagef("Set mode %s", modeName)
}

// setDelimiter sets the delimiter string.
func (root *Root) setDelimiter(input string) {
	root.Doc.setDelimiter(input)
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
		root.Doc.unWatchMode()
	} else {
		root.Doc.watchMode()
	}
	root.Doc.watchRestart.Store(true)
	root.setMessageLogf("Set watch interval %d", interval)
}

// setWriteBA sets the number before and after the line
// to be written at the end.
func (root *Root) setWriteBA(input string) {
	before, after, err := rangeBA(input)
	if err != nil {
		root.setMessage(ErrInvalidNumber.Error())
		return
	}
	root.BeforeWriteOriginal = before
	root.AfterWriteOriginal = after
	root.debugMessage(fmt.Sprintf("Before:After:%d:%d", root.BeforeWriteOriginal, root.AfterWriteOriginal))
	root.IsWriteOriginal = true
	root.Quit()
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
	num, section := jumpPosition(root.scr.vHeight, input)
	root.Doc.jumpTargetSection = section
	if root.Doc.jumpTargetSection {
		root.setMessagef("Set JumpTarget section start")
		return
	}
	if num < 0 || num > root.scr.vHeight-1 {
		root.setMessagef("Set JumpTarget %d: %s", num, ErrOutOfRange.Error())
		return
	}
	if root.Doc.jumpTargetNum == num {
		return
	}
	root.Doc.JumpTarget = input
	root.Doc.jumpTargetNum = num
	root.setMessagef("Set JumpTarget %d", num)
}

// resize is a wrapper function that calls viewSync.
func (root *Root) resize() {
	root.ViewSync()
}

// jumpPosition determines the position of the jump.
func jumpPosition(height int, str string) (int, bool) {
	if len(str) == 0 {
		return 0, false
	}
	s := strings.ToLower(strings.Trim(str, " "))
	if len(s) > 0 && s[0] == 's' {
		return 0, true
	}
	num := int(math.Round(calculatePosition(height, s)))
	if num < 0 {
		return (height - 1) + num, false
	}
	return num, false
}

// CalculatePosition returns the number from the length for positive
// numbers (1), returns dot.number for percentages (.5) = 50%,
// and returns the % after the number for percentages (50%). return.
func calculatePosition(length int, str string) float64 {
	if len(str) == 0 || str == "0" {
		return 0
	}

	var p float64 = 0
	if strings.HasPrefix(str, ".") {
		str = strings.TrimLeft(str, ".")
		i, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return 0
		}
		p = i / 10
	}
	if strings.HasSuffix(str, "%") {
		str = strings.TrimRight(str, "%")
		i, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return 0
		}
		p = i / 100
	}

	if p != 0 {
		return float64(length) * p
	}

	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return num
}

// ViewSync redraws the whole thing.
func (root *Root) ViewSync() {
	root.resetSelect()
	root.prepareStartX()
	root.prepareView()
	root.Screen.Sync()
	root.Doc.jumpTargetNum, root.Doc.jumpTargetSection = jumpPosition(root.scr.vHeight, root.Doc.JumpTarget)
}

// TailSync move to tail and sync.
func (root *Root) TailSync() {
	root.Doc.moveBottom()
	root.ViewSync()
}

// tailSection moves to the last section
// and adjusts to its original position.
func (root *Root) tailSection() {
	moved := root.Doc.topLN - root.Doc.lastSectionPosNum
	root.Doc.moveLastSection()
	if moved > 0 && (root.Doc.topLN+moved) < root.Doc.BufEndNum() {
		root.Doc.moveLine(root.Doc.topLN + moved)
	}
	root.Doc.lastSectionPosNum = root.Doc.topLN
}

// prepareStartX prepares startX.
func (root *Root) prepareStartX() {
	root.scr.startX = 0
	if root.Doc.LineNumMode {
		if root.Doc.parent != nil {
			root.scr.startX = len(fmt.Sprintf("%d", root.Doc.parent.BufEndNum())) + 1
			return
		}
		root.scr.startX = len(fmt.Sprintf("%d", root.Doc.BufEndNum())) + 1
	}
}

// updateEndNum updates the last line number.
func (root *Root) updateEndNum() {
	root.debugMessage(fmt.Sprintf("Update EndNum:%d", root.Doc.BufEndNum()))
	root.prepareStartX()
	root.drawStatus()
	root.Screen.Sync()
}

// follow updates the document in follow mode.
func (root *Root) follow() {
	if root.General.FollowAll {
		root.followAll()
	}
	num := root.Doc.BufEndNum()
	if root.Doc.latestNum == num {
		return
	}

	root.skipDraw = false
	if root.Doc.FollowSection {
		root.tailSection()
	} else {
		root.TailSync()
	}
	root.Doc.latestNum = num
}

// followAll monitors and switches all document updates
// in follow all mode.
func (root *Root) followAll() {
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
		root.switchDocument(current)
	}
}

// Cancel follow mode and follow all mode.
func (root *Root) Cancel() {
	root.General.FollowAll = false
	root.Doc.FollowMode = false
}

// WriteQuit sets the write flag and executes a quit event.
func (root *Root) WriteQuit() {
	root.IsWriteOriginal = true
	root.Quit()
}
