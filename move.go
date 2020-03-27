package oviewer

func (root *root) moveTop() {
	root.model.y = 0
	root.model.yy = 0
}

func (root *root) moveEnd() {
	root.model.y = root.model.endY - 1
	root.model.yy = 0
}

func (root *root) moveNum(num int) {
	root.model.y = num
	root.model.yy = 0
}

func (root *root) movePgUp() {
	root.model.y -= root.model.vHight
	root.model.yy = 0
}

func (root *root) movePgDn() {
	root.model.y += root.model.vHight
	root.model.yy = 0
}

func (root *root) moveHfUp() {
	root.model.y -= (root.model.vHight / 2)
	root.model.yy = 0
}

func (root *root) moveHfDn() {
	root.model.y += (root.model.vHight / 2)
	root.model.yy = 0
}

func (root *root) moveUp() {
	if !root.model.WrapMode {
		root.model.yy = 0
		root.model.y--
		return
	}
	// WrapMode
	contents := root.model.getContents(root.model.y)
	if len(contents) < root.model.vWidth || root.model.yy <= 0 {
		if (root.model.y) >= 1 {
			pre := root.model.getContents(root.model.y - 1)
			yyLen := len(pre) / (root.model.vWidth + 1)
			root.model.yy = yyLen
		}
		root.model.y--
		return
	}
	root.model.yy--
}

func (root *root) moveDown() {
	if !root.model.WrapMode {
		root.model.yy = 0
		root.model.y++
		return
	}
	// WrapMode
	contents := root.model.getContents(root.model.y)
	if len(contents) < (root.model.vWidth * (root.model.yy + 1)) {
		root.model.yy = 0
		root.model.y++
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
