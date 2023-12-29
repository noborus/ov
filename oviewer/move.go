package oviewer

// Called by event key to change document position.

// bottomMargin is the margin of the bottom line when specifying section.
const bottomMargin = 2

// Go to the top line.
func (root *Root) moveTop() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.startTopFlag = true
	root.Doc.moveTop()
}

// Go to the bottom line.
func (root *Root) moveBottom() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveBottom()
}

// Move up one screen.
func (root *Root) movePgUp() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.movePgUp()
}

// Moves down one screen.
func (root *Root) movePgDn() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.movePgDn()
}

// Moves up half a screen.
func (root *Root) moveHfUp() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveHfUp()
}

// Moves down half a screen.
func (root *Root) moveHfDn() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveHfDn()
}

// Move up one line.
func (root *Root) moveUpOne() {
	root.moveUp(1)
}

// Move down one line.
func (root *Root) moveDownOne() {
	root.moveDown(1)
}

// Move up by n amount.
func (root *Root) moveUp(n int) {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveLimitYUp(n)
}

// Move down by n amount.
func (root *Root) moveDown(n int) {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveYDown(n)
}

// nextSection moves down to the next section's delimiter.
func (root *Root) nextSection() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	if err := root.Doc.moveNextSection(); err != nil {
		// Last section or no section.
		root.setMessage("No more next sections")
	}
}

// prevSection moves up to the delimiter of the previous section.
func (root *Root) prevSection() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	if err := root.Doc.movePrevSection(); err != nil {
		root.setMessage("No more previous sections")
		return
	}
}

// lastSection moves to the last section.
func (root *Root) lastSection() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveLastSection()
}

// Move to the left.
func (root *Root) moveLeftOne() {
	root.moveLeft(1)
}

// Move to the right.
func (root *Root) moveRightOne() {
	root.moveRight(1)
}

// Move to the width of the screen to the left.
func (root *Root) moveWidthLeft() {
	root.moveLeft(root.Doc.HScrollWidthNum)
}

// Move to the width of the screen to the right.
func (root *Root) moveWidthRight() {
	root.moveRight(root.Doc.HScrollWidthNum)
}

// Move left by n amount.
func (root *Root) moveLeft(n int) {
	root.resetSelect()
	defer root.releaseEventBuffer()

	if root.Doc.ColumnMode {
		root.moveColumnLeft(1)
		return
	}
	root.moveNormalLeft(n)
}

// Move right by n amount.
func (root *Root) moveRight(n int) {
	root.resetSelect()
	defer root.releaseEventBuffer()

	if root.Doc.ColumnMode {
		root.moveColumnRight(1)
		return
	}
	root.moveNormalRight(n)
}

// Move to the left by half a screen.
func (root *Root) moveHfLeft() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveHfLeft()
	root.Doc.x = max(root.Doc.x, root.minStartX)
}

// Move to the right by half a screen.
func (root *Root) moveHfRight() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveHfRight()
}

// moveBeginLeft moves to the beginning of the line.
func (root *Root) moveBeginLeft() {
	root.Doc.moveBeginLeft()
}

// moveEndRight moves to the end of the line.
// Move so that the end of the currently displayed line is visible.
func (root *Root) moveEndRight() {
	root.Doc.moveEndRight(root.scr)
}

// moveNormalLeft moves the screen left.
func (root *Root) moveNormalLeft(n int) {
	root.Doc.moveNormalLeft(n)
	root.Doc.x = max(root.Doc.x, root.minStartX)
}

// moveNormalRight moves the screen right.
func (root *Root) moveNormalRight(n int) {
	root.Doc.moveNormalRight(n)
}

// moveColumnLeft moves the cursor to the left by n amount.
func (root *Root) moveColumnLeft(n int) {
	if err := root.Doc.moveColumnLeft(n, root.scr, !root.Config.DisableColumnCycle); err != nil {
		root.debugMessage(err.Error())
	}
}

// moveColumnRight moves the cursor to the right by n amount.
func (root *Root) moveColumnRight(n int) {
	if err := root.Doc.moveColumnRight(n, root.scr, !root.Config.DisableColumnCycle); err != nil {
		root.debugMessage(err.Error())
	}
}
