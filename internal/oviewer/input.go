package oviewer

import (
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

// HandleInputEvent handles input events.
func (root *Root) HandleInputEvent(ev *tcell.EventKey) bool {
	input := root.inputEvent(ev)
	// Input not confirmed or canceled.
	if input == "" {
		return true
	}

	// Input is confirmed.
	switch root.mode {
	case search:
		root.Search(input)
	case previous:
		root.BackSearch(input)
	case goline:
		root.GoLine(input)
	case header:
		root.SetHeader(input)
	case delimiter:
		root.SetDelimiter(input)
	case tabWidth:
		root.SetTabWidth(input)
	}
	root.mode = normal
	return true
}

func (root *Root) inputEvent(ev *tcell.EventKey) string {
	switch ev.Key() {
	case tcell.KeyEscape:
		root.mode = normal
		return ""
	case tcell.KeyEnter:
		return root.input
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if root.cursorX == 0 {
			return ""
		}
		pos := stringWidth(root.input, root.cursorX)
		runes := []rune(root.input)
		root.input = string(runes[:pos])
		root.cursorX = runeWidth(root.input)
		root.input += string(runes[pos+1:])
	case tcell.KeyDelete:
		pos := stringWidth(root.input, root.cursorX)
		runes := []rune(root.input)
		root.input = string(runes[:pos+1])
		if len(runes) > pos+1 {
			root.input += string(runes[pos+2:])
		}
	case tcell.KeyLeft:
		if root.cursorX > 0 {
			pos := stringWidth(root.input, root.cursorX)
			runes := []rune(root.input)
			if pos >= 0 {
				root.cursorX = runeWidth(string(runes[:pos]))
				if pos > 0 && runes[pos-1] == '\t' {
					root.cursorX--
				}
			}
		}
	case tcell.KeyRight:
		pos := stringWidth(root.input, root.cursorX+1)
		runes := []rune(root.input)
		root.cursorX = runeWidth(string(runes[:pos+1]))
	case tcell.KeyTAB:
		pos := stringWidth(root.input, root.cursorX+1)
		runes := []rune(root.input)
		root.input = string(runes[:pos])
		root.input += "\t"
		root.cursorX += 2
		root.input += string(runes[pos:])
	case tcell.KeyCtrlA:
		root.CaseSensitive = !root.CaseSensitive
	case tcell.KeyRune:
		pos := stringWidth(root.input, root.cursorX+1)
		runes := []rune(root.input)
		root.input = string(runes[:pos])
		r := ev.Rune()
		root.input += string(r)
		root.input += string(runes[pos:])
		root.cursorX += runewidth.RuneWidth(r)
	}
	return ""
}

func stringWidth(str string, cursor int) int {
	width := 0
	i := 0
	for _, r := range string(str) {
		width += runewidth.RuneWidth(r)
		if r == '\t' {
			width += 2
		}
		if width >= cursor {
			return i
		}
		i++
	}
	return i
}

func runeWidth(str string) int {
	width := 0
	for _, r := range string(str) {
		width += runewidth.RuneWidth(r)
		if r == '\t' {
			width += 2
		}
	}
	return width
}

// Search is a forward search.
func (root *Root) Search(input string) {
	root.inputRegexp = regexpComple(input, root.CaseSensitive)
	root.postSearch(root.search(root.Model.lineNum))
}

// BackSearch reverse search.
func (root *Root) BackSearch(input string) {
	root.inputRegexp = regexpComple(input, root.CaseSensitive)
	root.postSearch(root.backSearch(root.Model.lineNum))
}

// GoLine will move to the specified line.
func (root *Root) GoLine(input string) {
	lineNum, err := strconv.Atoi(input)
	if err != nil {
		return
	}
	root.moveNum(lineNum - root.Header)
	root.input = ""
}

// SetHeader sets the number of lines in the header.
func (root *Root) SetHeader(input string) {
	line, err := strconv.Atoi(input)
	if err != nil {
		return
	}
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
func (root *Root) SetDelimiter(input string) {
	root.ColumnDelimiter = input
	root.input = ""
}

// SetTabWidth sets the tab width.
func (root *Root) SetTabWidth(input string) {
	width, err := strconv.Atoi(input)
	if err != nil {
		return
	}
	if root.TabWidth != width {
		root.TabWidth = width
		root.Model.ClearCache()
	}
	root.input = ""
}
