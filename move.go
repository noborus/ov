package oviewer

func (root *root) moveTop() {
	root.model.y = 0
}

func (root *root) moveEnd() {
	root.model.y = root.model.endY
}

func (root *root) movePgUp() {
	root.model.y -= root.model.vHight
}

func (root *root) movePgDn() {
	root.model.y += root.model.vHight
}

func (root *root) moveHfUp() {
	root.model.y -= (root.model.vHight / 2)
}

func (root *root) moveHfDn() {
	root.model.y += (root.model.vHight / 2)
}

func (root *root) moveUp() {
	root.model.y--
}

func (root *root) moveDown() {
	root.model.y++
}

func (root *root) moveLeft() {
	if root.model.WrapMode {
		return
	}
	root.model.x--
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

func (root *root) moveRight() {
	if root.model.WrapMode {
		return
	}
	root.model.x++
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
