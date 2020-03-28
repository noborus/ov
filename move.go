package oviewer

func (root *root) moveTop() {
	root.Model.lineNum = 0
	root.Model.yy = 0
}

func (root *root) moveEnd() {
	root.moveBottomNum(root.Model.endY)
}

func (root *root) moveNum(num int) {
	root.Model.lineNum = num - root.Model.HeaderLen
	root.Model.yy = 0
}

func (root *root) moveBottomNum(num int) {
	n := root.bottomLineNum(num) + 1
	root.moveNum(n)
}

func (root *root) movePgUp() {
	root.moveNum(root.Model.lineNum - (root.bottomPos - root.Model.lineNum))
}

func (root *root) movePgDn() {
	root.moveNum(root.bottomPos)
}

func (root *root) moveHfUp() {
	root.moveNum(root.Model.lineNum - ((root.bottomPos - root.Model.lineNum) / 2))
}

func (root *root) moveHfDn() {
	root.moveNum(root.Model.lineNum + ((root.bottomPos - root.Model.lineNum) / 2))
}

func (root *root) moveUp() {
	if !root.Model.WrapMode {
		root.Model.yy = 0
		root.Model.lineNum--
		return
	}
	// WrapMode
	contents := root.Model.getContents(root.Model.lineNum + root.Model.HeaderLen)
	if len(contents) < root.Model.vWidth || root.Model.yy <= 0 {
		if (root.Model.lineNum) >= 1 {
			pre := root.Model.getContents(root.Model.lineNum + root.Model.HeaderLen - 1)
			yyLen := len(pre) / (root.Model.vWidth + 1)
			root.Model.yy = yyLen
		}
		root.Model.lineNum--
		return
	}
	root.Model.yy--
}

func (root *root) moveDown() {
	if !root.Model.WrapMode {
		root.Model.yy = 0
		root.Model.lineNum++
		return
	}
	// WrapMode
	contents := root.Model.getContents(root.Model.lineNum + root.Model.HeaderLen)
	if len(contents) < (root.Model.vWidth * (root.Model.yy + 1)) {
		root.Model.yy = 0
		root.Model.lineNum++
		return
	}
	root.Model.yy++
}

func (root *root) moveLeft() {
	if root.Model.WrapMode {
		return
	}
	root.Model.x--
}

func (root *root) moveRight() {
	if root.Model.WrapMode {
		return
	}
	root.Model.x++
}

func (root *root) moveHfLeft() {
	if root.Model.WrapMode {
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
	if root.Model.WrapMode {
		return
	}
	if root.Model.x < 0 {
		root.Model.x = 0
	} else {
		root.Model.x += (root.Model.vWidth / 2)
	}
}
