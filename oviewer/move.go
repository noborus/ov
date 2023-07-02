package oviewer

import (
	"strconv"
)

// MoveLine fires an eventGoto event that moves to the specified line.
func (root *Root) MoveLine(num int) {
	if !root.checkScreen() {
		return
	}
	root.sendGoto(num)
}

// sendGoto fires an eventGoto event that moves to the specified line.
func (root *Root) sendGoto(num int) {
	ev := &eventGoto{}
	ev.value = strconv.Itoa(num)
	ev.SetEventNow()
	root.postEvent(ev)
}

// MoveTop fires the event of moving to top.
func (root *Root) MoveTop() {
	root.MoveLine(root.Doc.BufStartNum())
}

// MoveBottom fires the event of moving to bottom.
func (root *Root) MoveBottom() {
	root.MoveLine(root.Doc.BufEndNum())
}

// Go to the top line.
// Called from a EventKey.
func (root *Root) moveTop() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveTop()
}

// Go to the bottom line.
// Called from a EventKey.
func (root *Root) moveBottom() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveBottom()
}

// Move up one screen.
// Called from a EventKey.
func (root *Root) movePgUp() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.movePgUp()
}

// Moves down one screen.
// Called from a EventKey.
func (root *Root) movePgDn() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.movePgDn()
}

// Moves up half a screen.
// Called from a EventKey.
func (root *Root) moveHfUp() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveHfUp()
}

// Moves down half a screen.
// Called from a EventKey.
func (root *Root) moveHfDn() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveHfDn()
}

// Move up one line.
// Called from a EventKey.
func (root *Root) moveUpOne() {
	root.moveUp(1)
}

// Move down one line.
// Called from a EventKey.
func (root *Root) moveDownOne() {
	root.moveDown(1)
}

// Move up by n amount.
func (root *Root) moveUp(n int) {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveYUp(n)
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
// Called from a EventKey.
func (root *Root) moveLeftOne() {
	root.moveLeft(1)
}

// Move to the right.
// Called from a EventKey.
func (root *Root) moveRightOne() {
	root.moveRight(1)
}

// Move left by n amount.
func (root *Root) moveLeft(n int) {
	root.resetSelect()
	defer root.releaseEventBuffer()

	if root.Doc.ColumnMode {
		root.moveColumnLeft(n)
		return
	}
	root.moveNormalLeft(n)
}

// Move right by n amount.
func (root *Root) moveRight(n int) {
	root.resetSelect()
	defer root.releaseEventBuffer()

	if root.Doc.ColumnMode {
		root.moveColumnRight(n)
		return
	}
	root.moveNormalRight(n)
}

// Move to the left by half a screen.
// Called from a EventKey.
func (root *Root) moveHfLeft() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveHfLeft()
	root.Doc.x = max(root.Doc.x, root.minStartX)
}

// Move to the right by half a screen.
// Called from a EventKey.
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
	if err := root.Doc.moveColumnLeft(n, root.scr, root.Config.DisableColumnCycle); err != nil {
		root.debugMessage(err.Error())
	}
}

// moveColumnRight moves the cursor to the right by n amount.
func (root *Root) moveColumnRight(n int) {
	if err := root.Doc.moveColumnRight(n, root.scr, root.Config.DisableColumnCycle); err != nil {
		root.debugMessage(err.Error())
	}
}
