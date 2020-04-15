package oviewer

func (root *root) moveTop() {
	root.Model.lineNum = 0
	root.Model.yy = 0
}

func (root *root) moveEnd() {
	root.moveBottomNum(root.Model.endNum)
}

func (root *root) moveNum(num int) {
	root.Model.lineNum = num
	root.Model.yy = 0
}

func (root *root) moveBottomNum(num int) {
	n := root.bottomLineNum(num) + 1
	root.moveNum(n)
}

func (root *root) realHightNum() int {
	return root.bottomPos - (root.Model.lineNum + root.Header)
}

func (root *root) movePgUp() {
	n := root.Model.lineNum - root.realHightNum()
	if n >= root.Model.lineNum {
		n = root.Model.lineNum - 1
	}
	root.moveNum(n)
}

func (root *root) movePgDn() {
	n := root.bottomPos - root.Header
	if n <= root.Model.lineNum {
		n = root.Model.lineNum + 1
	}
	root.moveNum(n)
}

func (root *root) moveHfUp() {
	root.moveNum(root.Model.lineNum - (root.realHightNum() / 2))
}

func (root *root) moveHfDn() {
	root.moveNum(root.Model.lineNum + (root.realHightNum() / 2))
}

func (root *root) moveUp() {
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

func (root *root) moveDown() {
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

func (root *root) moveLeft() {
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

func (root *root) moveRight() {
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

func (root *root) columnModeX() int {
	m := root.Model
	line := m.getLine(root.Header + 2)
	r := rangePosition(line, root.ColumnDelimiter, root.columnNum)
	if r.start < 0 || r.end < 0 {
		root.columnNum = 0
		r = rangePosition(line, root.ColumnDelimiter, root.columnNum)
	}
	lc, err := m.lineToContents(root.Header+2, root.TabWidth)
	if err != nil {
		return 0
	}
	return lc.cMap[r.start]
}

func (root *root) moveHfLeft() {
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

func (root *root) moveHfRight() {
	if root.WrapMode {
		return
	}
	if root.Model.x < 0 {
		root.Model.x = 0
	} else {
		root.Model.x += (root.Model.vWidth / 2)
	}
}
