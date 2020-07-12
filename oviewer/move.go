package oviewer

import "fmt"

// Go to the top line.
func (root *Root) moveTop() {
	root.lineNum = 0
	root.yy = 0
}

// Go to the bottom line.
func (root *Root) moveBottom() {
	root.message = fmt.Sprintf("endnum:%d", root.Model.endNum)
	n := root.bottomLineNum(root.Model.endNum) + 1
	root.moveLine(n)
}

// Move to the specified line.
func (root *Root) moveLine(num int) {
	root.lineNum = num
	root.yy = 0
}

// Move up one screen.
func (root *Root) movePgUp() {
	n := root.lineNum - root.realHightNum()
	if n >= root.lineNum {
		n = root.lineNum - 1
	}
	root.moveLine(n)
}

// Moves down one screen.
func (root *Root) movePgDn() {
	n := root.bottomPos - root.Header
	if n <= root.lineNum {
		n = root.lineNum + 1
	}
	root.moveLine(n)
}

// realHightNum returns the actual number of line on the screen.
func (root *Root) realHightNum() int {
	return root.bottomPos - (root.lineNum + root.Header)
}

// Moves up half a screen.
func (root *Root) moveHfUp() {
	root.moveLine(root.lineNum - (root.realHightNum() / 2))
}

// Moves down half a screen.
func (root *Root) moveHfDn() {
	root.moveLine(root.lineNum + (root.realHightNum() / 2))
}

// Move up one line.
func (root *Root) moveUp() {
	if !root.WrapMode {
		root.yy = 0
		root.lineNum--
		return
	}
	// WrapMode
	contents := root.Model.getContents(root.lineNum+root.Header, root.TabWidth)
	if len(contents) < root.vWidth || root.yy <= 0 {
		if (root.lineNum) >= 1 {
			pre := root.Model.getContents(root.lineNum+root.Header-1, root.TabWidth)
			yyLen := len(pre) / (root.vWidth + 1)
			root.yy = yyLen
		}
		root.lineNum--
		return
	}
	root.yy--
}

// Move down one line.
func (root *Root) moveDown() {
	if root.lineNum > root.bottomLineNum(root.Model.endNum) {
		if root.Model.BufEOF() {
			root.message = "EOF"
		}
		return
	}

	if !root.WrapMode {
		root.yy = 0
		root.lineNum++
		return
	}
	// WrapMode
	contents := root.Model.getContents(root.lineNum+root.Header, root.TabWidth)
	if len(contents) < (root.vWidth * (root.yy + 1)) {
		root.yy = 0
		root.lineNum++
		return
	}
	root.yy++
}

// Move to the left.
func (root *Root) moveLeft() {
	if root.ColumnMode {
		if root.columnNum > 0 {
			root.columnNum--
			root.x = root.columnModeX()
		}
		return
	}
	if root.WrapMode {
		return
	}
	root.x--
}

// Move to the right.
func (root *Root) moveRight() {
	if root.ColumnMode {
		root.columnNum++
		root.x = root.columnModeX()
		return
	}
	if root.WrapMode {
		return
	}
	root.x++
}

// columnModeX returns the actual x from root.columnNum.
func (root *Root) columnModeX() int {
	m := root.Model
	line := m.GetLine(root.Header + 2)
	start, end := rangePosition(line, root.ColumnDelimiter, root.columnNum)
	if start < 0 || end < 0 {
		root.columnNum = 0
		start, _ = rangePosition(line, root.ColumnDelimiter, root.columnNum)
	}
	lc, err := m.lineToContents(root.Header+2, root.TabWidth)
	if err != nil {
		return 0
	}
	return lc.byteMap[start]
}

// Move to the left by half a screen.
func (root *Root) moveHfLeft() {
	if root.WrapMode {
		return
	}
	moveSize := (root.vWidth / 2)
	if root.x > 0 && (root.x-moveSize) < 0 {
		root.x = 0
	} else {
		root.x -= moveSize
	}
}

// Move to the right by half a screen.
func (root *Root) moveHfRight() {
	if root.WrapMode {
		return
	}
	if root.x < 0 {
		root.x = 0
	} else {
		root.x += (root.vWidth / 2)
	}
}
