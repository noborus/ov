package oviewer

func (root *root) moveTop() {
	root.model.lineNum = 0
	root.model.yy = 0
}

func (root *root) moveEnd() {
	root.moveBottomNum(root.model.endY)
}

func (root *root) moveNum(num int) {
	root.model.lineNum = num - root.model.HeaderLen
	root.model.yy = 0
}

func (root *root) moveBottomNum(num int) {
	n := root.bottomLineNum(num) + 1
	root.moveNum(n)
}

func (root *root) movePgUp() {
	root.moveNum(root.model.lineNum - (root.bottomPos - root.model.lineNum))
}

func (root *root) movePgDn() {
	root.moveNum(root.bottomPos)
}

func (root *root) moveHfUp() {
	root.moveNum(root.model.lineNum - ((root.bottomPos - root.model.lineNum) / 2))
}

func (root *root) moveHfDn() {
	root.moveNum(root.model.lineNum + ((root.bottomPos - root.model.lineNum) / 2))
}

func (root *root) moveUp() {
	if !root.model.WrapMode {
		root.model.yy = 0
		root.model.lineNum--
		return
	}
	// WrapMode
	contents := root.model.getContents(root.model.lineNum + root.model.HeaderLen)
	if len(contents) < root.model.vWidth || root.model.yy <= 0 {
		if (root.model.lineNum) >= 1 {
			pre := root.model.getContents(root.model.lineNum + root.model.HeaderLen - 1)
			yyLen := len(pre) / (root.model.vWidth + 1)
			root.model.yy = yyLen
		}
		root.model.lineNum--
		return
	}
	root.model.yy--
}

func (root *root) moveDown() {
	if !root.model.WrapMode {
		root.model.yy = 0
		root.model.lineNum++
		return
	}
	// WrapMode
	contents := root.model.getContents(root.model.lineNum + root.model.HeaderLen)
	if len(contents) < (root.model.vWidth * (root.model.yy + 1)) {
		root.model.yy = 0
		root.model.lineNum++
		return
	}
	root.model.yy++
}

func (root *root) moveLeft() {
	if root.model.WrapMode {
		return
	}
	root.model.x--
}

func (root *root) moveRight() {
	if root.model.WrapMode {
		return
	}
	root.model.x++
}

func (root *root) moveHfLeft() {
	if root.model.WrapMode {
		return
	}
	moveSize := (root.model.vWidth / 2)
	if root.model.x > 0 && (root.model.x-moveSize) < 0 {
		root.model.x = 0
	} else {
		root.model.x -= moveSize
	}
}

func (root *root) moveHfRight() {
	if root.model.WrapMode {
		return
	}
	if root.model.x < 0 {
		root.model.x = 0
	} else {
		root.model.x += (root.model.vWidth / 2)
	}
}
