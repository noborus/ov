package oviewer

import (
	"fmt"
	"log"
	"strconv"
)

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

func (root *Root) toggleFollowMode() {
	root.Doc.FollowMode = !root.Doc.FollowMode
}

func (root *Root) toggleFollowAll() {
	root.General.FollowAll = !root.General.FollowAll
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
	root.mu.RLock()
	defer root.mu.RUnlock()

	root.setDocument(root.DocList[root.CurrentDoc])
	root.input.mode = Normal
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

func (root *Root) markNext() {
	root.goLine(newGotoInput(root.input.GoCandidate).Up(""))
}

func (root *Root) markPrev() {
	root.goLine(newGotoInput(root.input.GoCandidate).Down(""))
}

func (root *Root) nextDoc() {
	root.mu.RLock()
	defer root.mu.RUnlock()

	if root.CurrentDoc+1 >= len(root.DocList) {
		root.setMessage("No next doc")
		return
	}
	root.CurrentDoc++
	root.setDocument(root.DocList[root.CurrentDoc])
	root.input.mode = Normal
}

func (root *Root) previousDoc() {
	root.mu.RLock()
	defer root.mu.RUnlock()

	if root.CurrentDoc <= 0 {
		root.setMessage("No previous doc")
		return
	}
	root.CurrentDoc--
	root.setDocument(root.DocList[root.CurrentDoc])
	root.input.mode = Normal
}

func (root *Root) addDocument(m *Document) {
	log.Printf("add: %s", m.FileName)
	m.general = root.Config.General

	root.mu.Lock()
	root.DocList = append(root.DocList, m)
	root.CurrentDoc = len(root.DocList) - 1
	root.mu.Unlock()

	root.setDocument(m)
}

func (root *Root) closeDocument() {
	if root.DocumentLen() == 1 {
		return
	}

	m := root.Doc
	log.Printf("close : %s", m.FileName)

	root.mu.Lock()
	root.DocList = append(root.DocList[:root.CurrentDoc], root.DocList[root.CurrentDoc+1:]...)
	if root.CurrentDoc > 0 {
		root.CurrentDoc--
	}
	doc := root.DocList[root.CurrentDoc]
	root.mu.Unlock()

	root.setDocument(doc)
	m.Close()
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
}
