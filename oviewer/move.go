package oviewer

import (
	"log"
)

// Go to the top line.
func (root *Root) moveTop() {
	root.moveLine(0)
}

// Go to the bottom line.
func (root *Root) moveBottom() {
	root.moveLine(root.Doc.endNum + 1)
}

// Move to the specified line.
func (root *Root) moveLine(num int) {
	root.resetSelect()
	root.Doc.lineNum = num
	root.Doc.firstStartX = 0
}

// Move up one screen.
func (root *Root) movePgUp() {
	root.resetSelect()
	root.moveNumUp(root.statusPos - root.headerLen())
}

// Moves down one screen.
func (root *Root) movePgDn() {
	root.resetSelect()
	root.Doc.lineNum = root.bottomPos - root.Doc.Header
	root.Doc.firstStartX = root.bottomEndX
}

// Moves up half a screen.
func (root *Root) moveHfUp() {
	root.moveNumUp((root.statusPos - root.headerLen()) / 2)
}

// Moves down half a screen.
func (root *Root) moveHfDn() {
	root.moveNumDown((root.statusPos - root.headerLen()) / 2)
}

func numOfSlice(listX []int, x int) int {
	for n, v := range listX {
		if v >= x {
			return n
		}
	}
	return len(listX) - 1
}

func numOfReverseSlice(listX []int, x int) int {
	for n := len(listX) - 1; n >= 0; n-- {
		if listX[n] <= x {
			return n
		}
	}
	return 0
}

// Moves up by the specified number of y.
func (root *Root) moveNumUp(moveY int) {
	if !root.Doc.WrapMode {
		root.Doc.lineNum -= moveY
		return
	}

	// WrapMode
	num := root.Doc.lineNum + root.Doc.Header
	x := root.Doc.firstStartX

	listX, err := root.leftMostX(num)
	if err != nil {
		log.Println(num, err)
		return
	}
	n := numOfSlice(listX, x)

	for y := moveY; y > 0; y-- {
		if n <= 0 {
			num--
			if num < root.Doc.Header {
				num = 0
				x = 0
				break
			}
			listX, err = root.leftMostX(num)
			if err != nil {
				log.Println(num, err)
				return
			}
			n = len(listX)
		}
		if n > 0 {
			x = listX[n-1]
		} else {
			x = 0
		}
		n--
	}
	root.Doc.lineNum = num - root.Doc.Header
	root.Doc.firstStartX = x
}

// Moves down by the specified number of y.
func (root *Root) moveNumDown(moveY int) {
	if !root.Doc.WrapMode {
		root.Doc.lineNum += moveY
		return
	}

	// WrapMode
	num := root.Doc.lineNum + root.Doc.Header
	x := root.Doc.firstStartX

	listX, err := root.leftMostX(num)
	if err != nil {
		log.Println(num, err)
		return
	}
	n := numOfReverseSlice(listX, x)

	for y := 0; y < moveY; y++ {
		if n >= len(listX) {
			num++
			if num > root.Doc.endNum {
				break
			}
			listX, err = root.leftMostX(num)
			if err != nil {
				log.Println(num, err)
				return
			}
			n = 0
		}
		x = 0
		if len(listX) > 0 && n < len(listX) {
			x = listX[n]
		}
		n++
	}
	root.Doc.lineNum = num - root.Doc.Header
	root.Doc.firstStartX = x
}

// Move up one line.
func (root *Root) moveUp() {
	root.resetSelect()

	if root.Doc.lineNum == 0 && root.Doc.firstStartX == 0 {
		return
	}

	if !root.Doc.WrapMode {
		root.Doc.firstStartX = 0
		root.Doc.lineNum--
		return
	}

	// WrapMode.
	// Same line.
	if root.Doc.firstStartX > 0 {
		listX, err := root.leftMostX(root.Doc.lineNum + root.Doc.Header)
		if err != nil {
			log.Println(err)
			return
		}
		for n, x := range listX {
			if x >= root.Doc.firstStartX {
				root.Doc.firstStartX = listX[n-1]
				return
			}
		}
	}

	// Previous line.
	root.Doc.lineNum--
	if root.Doc.lineNum < 0 {
		root.Doc.lineNum = 0
		root.Doc.firstStartX = 0
		return
	}
	listX, err := root.leftMostX(root.Doc.lineNum + root.Doc.Header)
	if err != nil {
		log.Println(err)
		return
	}
	if len(listX) <= 0 {
		root.Doc.firstStartX = 0
		return
	}
	root.Doc.firstStartX = listX[len(listX)-1]
}

// Move down one line.
func (root *Root) moveDown() {
	root.resetSelect()

	if !root.Doc.WrapMode {
		root.Doc.firstStartX = 0
		root.Doc.lineNum++
		return
	}

	// WrapMode
	listX, err := root.leftMostX(root.Doc.lineNum + root.Doc.Header)
	if err != nil {
		log.Println(err)
		return
	}
	for _, x := range listX {
		if x > root.Doc.firstStartX {
			root.Doc.firstStartX = x
			return
		}
	}

	root.Doc.firstStartX = 0
	root.Doc.lineNum++
}

// Move to the left.
func (root *Root) moveLeft() {
	root.resetSelect()
	if root.Doc.ColumnMode {
		if root.Doc.columnNum > 0 {
			root.Doc.columnNum--
			root.Doc.x = root.columnModeX()
		}
		return
	}
	if root.Doc.WrapMode {
		return
	}
	root.Doc.x--
	if root.Doc.x < root.minStartX {
		root.Doc.x = root.minStartX
	}
}

// Move to the right.
func (root *Root) moveRight() {
	root.resetSelect()
	if root.Doc.ColumnMode {
		root.Doc.columnNum++
		root.Doc.x = root.columnModeX()
		return
	}
	if root.Doc.WrapMode {
		return
	}
	root.Doc.x++
}

// columnModeX returns the actual x from root.Doc.columnNum.
func (root *Root) columnModeX() int {
	lc, err := root.Doc.lineToContents(root.Doc.lineNum+root.Doc.Header, root.Doc.TabWidth)
	if err != nil {
		return 0
	}
	lineStr, byteMap := contentsToStr(lc)
	start, end := rangePosition(lineStr, root.Doc.ColumnDelimiter, root.Doc.columnNum)
	if start < 0 || end < 0 {
		root.Doc.columnNum = 0
		start, _ = rangePosition(lineStr, root.Doc.ColumnDelimiter, root.Doc.columnNum)
	}
	return byteMap[start]
}

// Move to the left by half a screen.
func (root *Root) moveHfLeft() {
	if root.Doc.WrapMode {
		return
	}
	root.resetSelect()
	moveSize := (root.vWidth / 2)
	if root.Doc.x > 0 && (root.Doc.x-moveSize) < 0 {
		root.Doc.x = 0
	} else {
		root.Doc.x -= moveSize
		if root.Doc.x < root.minStartX {
			root.Doc.x = root.minStartX
		}
	}
}

// Move to the right by half a screen.
func (root *Root) moveHfRight() {
	if root.Doc.WrapMode {
		return
	}
	root.resetSelect()
	if root.Doc.x < 0 {
		root.Doc.x = 0
	} else {
		root.Doc.x += (root.vWidth / 2)
	}
}
