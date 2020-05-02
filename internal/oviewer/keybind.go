package oviewer

import (
	"strconv"

	"github.com/gdamore/tcell"
)

// HandleEvent handles all events.
func (root *Root) HandleEvent(ev *tcell.EventKey) bool {
	root.message = ""
	switch root.mode {
	case search:
		return root.inputEvent(ev, root.Search)
	case previous:
		return root.inputEvent(ev, root.BackSearch)
	case goline:
		return root.inputEvent(ev, root.GoLine)
	case header:
		return root.inputEvent(ev, root.SetHeader)
	case delimiter:
		return root.inputEvent(ev, root.SetDelimiter)
	default:
		return root.defaultEvent(ev)
	}
}

func (root *Root) defaultEvent(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		root.Quit()
		return true
	case tcell.KeyEnter:
		root.moveDown()
		return true
	case tcell.KeyHome:
		root.moveTop()
		return true
	case tcell.KeyEnd:
		root.moveEnd()
		return true
	case tcell.KeyLeft:
		if ev.Modifiers()&tcell.ModCtrl > 0 {
			root.moveHfLeft()
			return true
		}
		root.moveLeft()
		return true
	case tcell.KeyRight:
		if ev.Modifiers()&tcell.ModCtrl > 0 {
			root.moveHfRight()
			return true
		}
		root.moveRight()
		return true
	case tcell.KeyDown, tcell.KeyCtrlN:
		root.moveDown()
		return true
	case tcell.KeyUp, tcell.KeyCtrlP:
		root.moveUp()
		return true
	case tcell.KeyPgDn, tcell.KeyCtrlV:
		root.movePgDn()
		return true
	case tcell.KeyPgUp, tcell.KeyCtrlB:
		root.movePgUp()
		return true
	case tcell.KeyCtrlU:
		root.moveHfUp()
		return true
	case tcell.KeyCtrlD:
		root.moveHfDn()
		return true
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q':
			root.Quit()
			return true
		case 'Q':
			root.AfterWrite = true
			root.Quit()
			return true
		case 'W', 'w':
			root.keyWrap()
			return true
		case '?':
			root.input = ""
			root.keyPrevious()
			return true
		case 'c':
			root.keyColumnMode()
			return true
		case 'd':
			root.input = ""
			root.keyDelimiter()
			return true
		case '/':
			root.input = ""
			root.keySearch()
			return true
		case 'n':
			root.NextSearch()
			return true
		case 'N':
			root.NextBackSearch()
			return true
		case 'g':
			root.input = ""
			root.keyGoLine()
			return true
		case 'H':
			root.input = ""
			root.keyHeader()
			return true
		case 'C':
			root.keyAlternateRows()
			return true
		}
	}
	return true
}

func (root *Root) keyWrap() {
	if root.WrapMode {
		root.WrapMode = false
	} else {
		root.WrapMode = true
		root.Model.x = 0
	}
	root.setWrapHeaderLen()
}

func (root *Root) keyColumnMode() {
	if root.ColumnMode {
		root.ColumnMode = false
	} else {
		root.ColumnMode = true
	}
}

func (root *Root) keyAlternateRows() {
	root.Model.ClearCache()
	if root.AlternateRows {
		root.AlternateRows = false
	} else {
		root.AlternateRows = true
	}
}

func (root *Root) keySearch() {
	root.mode = search
}

func (root *Root) keyDelimiter() {
	root.mode = delimiter
}

func (root *Root) keyPrevious() {
	root.mode = previous
}

func (root *Root) keyGoLine() {
	root.mode = goline
}

func (root *Root) keyHeader() {
	root.mode = header
}

func (root *Root) inputEvent(ev *tcell.EventKey, fn func()) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		root.mode = normal
	case tcell.KeyEnter:
		fn()
		root.mode = normal
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(root.input) > 0 {
			r := []rune(root.input)
			root.input = string(r[:len(r)-1])
		}
	case tcell.KeyTAB:
		root.input += "\t"
	case tcell.KeyCtrlA:
		root.CaseSensitive = !root.CaseSensitive
	case tcell.KeyRune:
		root.input += string(ev.Rune())
	}
	return true
}

// GoLine will move to the specified line.
func (root *Root) GoLine() {
	lineNum, err := strconv.Atoi(root.input)
	if err != nil {
		return
	}
	root.input = ""
	root.moveNum(lineNum - root.Header)
}

// SetHeader sets the number of lines in the header.
func (root *Root) SetHeader() {
	line, _ := strconv.Atoi(root.input)
	if line >= 0 && line <= root.Model.vHight-1 {
		if root.Header != line {
			root.Header = line
			root.setWrapHeaderLen()
			root.Model.ClearCache()
		}
	}
	root.input = ""
}

// SetDelimiter sets the delimiter string.
func (root *Root) SetDelimiter() {
	root.ColumnDelimiter = root.input
	root.input = ""
}

// Search is a forward search.
func (root *Root) Search() {
	root.postSearch(root.search(root.Model.lineNum))
}

// NextSearch will re-run the forward search.
func (root *Root) NextSearch() {
	root.postSearch(root.search(root.Model.lineNum + root.Header + 1))
}

// BackSearch reverse search.
func (root *Root) BackSearch() {
	root.postSearch(root.backSearch(root.Model.lineNum))
}

// NextBackSearch will re-run the reverse search.
func (root *Root) NextBackSearch() {
	root.postSearch(root.backSearch(root.Model.lineNum + root.Header - 1))
}

func (root *Root) postSearch(lineNum int, err error) {
	if err != nil {
		root.message = err.Error()
		return
	}
	root.moveNum(lineNum - root.Header)
}
