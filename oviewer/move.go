package oviewer

import (
	"log"
	"strconv"
)

// bottomMargin is the margin of the bottom line when specifying section
const bottomMargin = 2

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
// Called from a EventKey.
func (root *Root) moveLeftOne() {
	root.moveLeft(1)
}

// Move to the right.
// Called from a EventKey.
func (root *Root) moveRightOne() {
	root.moveRight(1)
}

// Move to the width of the screen to the left.
// Called from a EventKey.
func (root *Root) moveWidthLeft() {
	root.moveLeft(root.Doc.ScrollWidthNum)
}

// Move to the width of the screen to the right.
// Called from a EventKey.
func (root *Root) moveWidthRight() {
	log.Println("ScrollWidth", root.Doc.ScrollWidth)
	root.moveRight(root.Doc.ScrollWidthNum)
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

// searchGoTo moves to the specified line and position after searching.
// Go to the specified line +root.Doc.JumpTarget Go to.
// If the search term is off screen, move until the search term is visible.
func (m *Document) searchGoTo(lN int, x int) {
	m.searchGoX(x)

	m.topLN = lN - m.firstLine()
	m.moveYUp(m.jumpTargetNum)
}

// Bottom line of jumpTarget when specifying a section
func (m *Document) bottomJumpTarget() int {
	return m.statusPos - bottomMargin
}

// searchGoSection will go to the section with the matching term after searching.
// Move the JumpTarget so that it can be seen from the beginning of the section.
func (m *Document) searchGoSection(lN int, x int) {
	m.searchGoX(x)

	sN, err := m.prevSection(lN)
	if err != nil {
		sN = 0
	}
	if m.SectionHeader {
		sN = (sN - m.firstLine() + m.sectionHeaderNum) + m.SectionStartPosition
		sN = max(sN, m.BufStartNum())
	}
	y := 0
	if sN < m.firstLine() {
		// topLN is negative if the section is less than header + skip.
		m.topLN = sN - m.firstLine()
	} else {
		m.topLN = sN
		sN += m.firstLine()
	}

	if !m.WrapMode {
		y = lN - sN
	} else {
		for n := sN; n < lN; n++ {
			listX := m.leftMostX(n)
			y += len(listX)
		}
	}

	if m.bottomJumpTarget() > y+m.headerLen {
		m.jumpTargetNum = y
		return
	}

	m.jumpTargetNum = m.bottomJumpTarget() - (m.headerLen + 1)
	m.moveYDown(y - m.jumpTargetNum)
}

// searchGoX moves to the specified x position.
func (m *Document) searchGoX(x int) {
	m.topLX = 0
	// If the search term is outside the height of the screen when in WrapMode.
	if nTh := x / m.width; nTh > m.height {
		m.topLX = nTh * m.width
	}

	// If the search term is outside the width of the screen when in NoWrapMode.
	if x < m.x || x > m.x+m.width-1 {
		m.x = x
	}
}
