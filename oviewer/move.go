package oviewer

import "fmt"

// Go to the top line.
func (root *Root) MoveTop() {
	root.Model.lineNum = 0
	root.Model.yy = 0
}

// Go to the bottom line.
func (root *Root) MoveEnd() {
	root.message = fmt.Sprintf("endnum:%d", root.Model.endNum)
	n := root.bottomLineNum(root.Model.endNum) + 1
	root.MoveNum(n)
}

// Move to the specified line.
func (root *Root) MoveNum(num int) {
	root.Model.lineNum = num
	root.Model.yy = 0
}

// Move up one screen.
func (root *Root) movePgUp() {
	n := root.Model.lineNum - root.realHightNum()
	if n >= root.Model.lineNum {
		n = root.Model.lineNum - 1
	}
	root.MoveNum(n)
}

// Moves down one screen.
func (root *Root) movePgDn() {
	n := root.bottomPos - root.Header
	if n <= root.Model.lineNum {
		n = root.Model.lineNum + 1
	}
	root.MoveNum(n)
}

// realHightNum returns the actual number of line on the screen.
func (root *Root) realHightNum() int {
	return root.bottomPos - (root.Model.lineNum + root.Header)
}

// Moves up half a screen.
func (root *Root) moveHfUp() {
	root.MoveNum(root.Model.lineNum - (root.realHightNum() / 2))
}

// Moves down half a screen.
func (root *Root) moveHfDn() {
	root.MoveNum(root.Model.lineNum + (root.realHightNum() / 2))
}

// Move up one line.
func (root *Root) moveUp() {
	if !root.WrapMode {
		root.Model.yy = 0
		root.Model.lineNum--
		return
	}
	// WrapMode
	contents := root.Model.getContents(root.Model.lineNum+root.Header, root.TabWidth)
	if len(contents) < root.Model.vWidth || root.Model.yy <= 0 {
		if (root.Model.lineNum) >= 1 {
			pre := root.Model.getContents(root.Model.lineNum+root.Header-1, root.TabWidth)
			yyLen := len(pre) / (root.Model.vWidth + 1)
			root.Model.yy = yyLen
		}
		root.Model.lineNum--
		return
	}
	root.Model.yy--
}

// Move down one line.
func (root *Root) moveDown() {
	if root.Model.lineNum > root.bottomLineNum(root.Model.endNum) {
		root.message = "EOF"
		return
	}

	if !root.WrapMode {
		root.Model.yy = 0
		root.Model.lineNum++
		return
	}
	// WrapMode
	contents := root.Model.getContents(root.Model.lineNum+root.Header, root.TabWidth)
	if len(contents) < (root.Model.vWidth * (root.Model.yy + 1)) {
		root.Model.yy = 0
		root.Model.lineNum++
		return
	}
	root.Model.yy++
}

// Move to the left.
func (root *Root) moveLeft() {
	if root.ColumnMode {
		if root.columnNum > 0 {
			root.columnNum--
			root.Model.x = root.columnModeX()
		}
		return
	}
	if root.WrapMode {
		return
	}
	root.Model.x--
}

// Move to the right.
func (root *Root) moveRight() {
	if root.ColumnMode {
		root.columnNum++
		root.Model.x = root.columnModeX()
		return
	}
	if root.WrapMode {
		return
	}
	root.Model.x++
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
	moveSize := (root.Model.vWidth / 2)
	if root.Model.x > 0 && (root.Model.x-moveSize) < 0 {
		root.Model.x = 0
	} else {
		root.Model.x -= moveSize
	}
}

// Move to the right by half a screen.
func (root *Root) moveHfRight() {
	if root.WrapMode {
		return
	}
	if root.Model.x < 0 {
		root.Model.x = 0
	} else {
		root.Model.x += (root.Model.vWidth / 2)
	}
}
