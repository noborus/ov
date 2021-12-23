package oviewer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

// toggleWrapMode toggles wrapMode each time it is called.
func (root *Root) toggleWrapMode() {
	root.Doc.WrapMode = !root.Doc.WrapMode
	root.Doc.x = 0
	root.setMessagef("Set WrapMode %t", root.Doc.WrapMode)
}

//  toggleColumnMode toggles ColumnMode each time it is called.
func (root *Root) toggleColumnMode() {
	root.Doc.ColumnMode = !root.Doc.ColumnMode
	root.setMessagef("Set ColumnMode %t", root.Doc.ColumnMode)
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

// toggleFollowMode toggles follow mode.
func (root *Root) toggleFollowMode() {
	root.Doc.FollowMode = !root.Doc.FollowMode
}

// toggleFollowAll toggles follow all mode.
func (root *Root) toggleFollowAll() {
	root.General.FollowAll = !root.General.FollowAll
}

// setDocument sets the Document.
func (root *Root) setDocument(m *Document) {
	root.Doc = m
	root.ViewSync()
}

// helpDisplay is to switch between helpDisplay screen and normal screen.
func (root *Root) helpDisplay() {
	if root.screenMode == Help {
		root.toNormal()
		return
	}
	root.setDocument(root.helpDoc)
	root.screenMode = Help
}

// LogDisplay is to switch between Log screen and normal screen.
func (root *Root) logDisplay() {
	if root.screenMode == LogDoc {
		root.toNormal()
		return
	}
	root.setDocument(root.logDoc)
	root.screenMode = LogDoc
}

// toNormal displays a normal document.
func (root *Root) toNormal() {
	root.mu.RLock()
	defer root.mu.RUnlock()

	root.setDocument(root.DocList[root.CurrentDoc])
	root.screenMode = Docs
}

// goLine will move to the specified line.
func (root *Root) goLine(input string) {
	lN, err := strconv.Atoi(input)
	if err != nil {
		root.setMessage(ErrInvalidNumber.Error())
		return
	}
	lN = root.moveLine(lN - 1)
	root.setMessagef("Moved to line %d", lN+1)
}

// goLineNumber moves to the specified line number.
func (root *Root) goLineNumber(ln int) {
	ln = root.moveLine(ln - root.Doc.firstLine())
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
	c := min(root.Doc.topLN+root.Doc.firstLine(), root.Doc.endNum)
	root.Doc.marked = removeInt(root.Doc.marked, c)
	root.Doc.marked = append(root.Doc.marked, c)
	root.setMessagef("Marked to line %d", c-root.Doc.firstLine()+1)
}

// removeMark removes the current line number from the mark.
func (root *Root) removeMark() {
	c := root.Doc.topLN + root.Doc.firstLine()
	marked := removeInt(root.Doc.marked, c)
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
	if num < 0 || num > root.vHight-1 {
		root.setMessagef("Set header %d: %s", num, ErrOutOfRange.Error())
		return
	}
	if root.Doc.Header == num {
		return
	}

	root.Doc.Header = num
	root.setMessagef("Set header lines %d", num)
	root.Doc.ClearCache()
}

// setSkipLines sets the number of lines to skip.
func (root *Root) setSkipLines(input string) {
	num, err := strconv.Atoi(input)
	if err != nil {
		root.setMessagef("Set skip line: %s", ErrInvalidNumber.Error())
		return
	}
	if num < 0 || num > root.vHight-1 {
		root.setMessagef("Set skip line: %s", ErrOutOfRange.Error())
		return
	}
	if root.Doc.SkipLines == num {
		return
	}

	root.Doc.SkipLines = num
	root.setMessagef("Set skip lines %d", num)
	root.Doc.ClearCache()
}

// nextDoc displays the next document.
func (root *Root) nextDoc() {
	root.setDocumentNum(root.CurrentDoc + 1)
	root.input.mode = Normal
}

// previouseDoc displays the previous document.
func (root *Root) previousDoc() {
	root.setDocumentNum(root.CurrentDoc - 1)
	root.input.mode = Normal
}

// switchDocument displays the document of the specified docNum.
func (root *Root) switchDocument(docNum int) {
	root.setDocumentNum(docNum)
	root.debugMessage(fmt.Sprintf("switch document %s", root.Doc.FileName))
}

// addDocument adds a document and displays it.
func (root *Root) addDocument(m *Document) {
	root.mu.Lock()
	defer root.mu.Unlock()
	root.setMessagef("add %s", m.FileName)
	m.general = root.Config.General

	root.DocList = append(root.DocList, m)
	root.CurrentDoc = len(root.DocList) - 1

	root.setDocument(m)
}

// closeDocument closes the document.
// If there is only one document, do nothing.
func (root *Root) closeDocument() {
	if root.DocumentLen() == 1 {
		return
	}

	root.mu.Lock()
	defer root.mu.Unlock()

	root.setMessagef("close [%d]%s", root.CurrentDoc, root.Doc.FileName)
	root.DocList = append(root.DocList[:root.CurrentDoc], root.DocList[root.CurrentDoc+1:]...)
	if root.CurrentDoc > 0 {
		root.CurrentDoc--
	}
	doc := root.DocList[root.CurrentDoc]

	root.setDocument(doc)
}

// setDocumentNum actually specifies docNum to display the document.
// This function is called internally from next / previous / switch / add.
func (root *Root) setDocumentNum(docNum int) {
	root.mu.Lock()
	defer root.mu.Unlock()

	if docNum >= len(root.DocList) {
		docNum = len(root.DocList) - 1
	}
	if docNum < 0 {
		docNum = 0
	}
	root.CurrentDoc = docNum
	m := root.DocList[root.CurrentDoc]
	root.setDocument(m)
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
		log.Println(err)
	}
	fmt.Println("resume ov")
	if err := root.Screen.Resume(); err != nil {
		log.Println(err)
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
		root.Screen.EnableMouse()
		root.setMessage("Enable Mouse")
	}
}

// setViewMode switches to the preset display mode.
// Set header lines and columMode together.
func (root *Root) setViewMode(input string) {
	c, ok := root.Config.Mode[input]
	if !ok {
		if input != "general" {
			root.setMessagef("%s mode not found", input)
			return
		}
		c = root.General
	}

	root.Doc.general = c
	root.Doc.ClearCache()
	root.ViewSync()
	root.setMessagef("Set mode %s", input)
}

// setDelimiter sets the delimiter string.
func (root *Root) setDelimiter(input string) {
	root.Doc.ColumnDelimiter = input
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
	root.debugMessage(fmt.Sprintf("Update EndNum:%d", root.Doc.BufEndNum()))
	root.prepareStartX()
	root.statusDraw()
	root.Screen.Sync()
}
