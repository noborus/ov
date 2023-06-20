package oviewer

import (
	"errors"
	"log"
	"regexp"
)

// columnMargin is the number of characters in the left and right margins of the screen.
const columnMargin = 2

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

// moveColumnLeft moves the cursor to the left by n amount.
func (root *Root) moveColumnLeft(n int) {
	m := root.Doc
	// left edge
	if m.columnCursor <= 0 && m.x <= 2 {
		if !root.Config.DisableColumnCycle {
			cursor := root.rightEndColumn()
			m.x = max(0, root.endRight()+columnMargin)
			m.columnCursor = cursor
			return
		}
	}

	x, cursor, err := root.columnX(-1)
	if err != nil {
		root.debugMessage(err.Error())
		m.x = x
		return
	}
	m.x = x
	m.columnCursor = cursor
}

// moveNormalLeft moves the screen left.
func (root *Root) moveNormalLeft(n int) {
	m := root.Doc
	m.x -= n
	if m.x < root.minStartX {
		m.x = root.minStartX
	}
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

// moveColumnRight moves the cursor to the right by n amount.
func (root *Root) moveColumnRight(n int) {
	m := root.Doc

	x, cursor, err := root.columnX(1)
	if err != nil {
		if errors.Is(err, ErrOverScreen) {
			if !root.Config.DisableColumnCycle {
				m.x = 0
				m.columnCursor = 0
				return
			}
		}
		root.debugMessage(err.Error())
		m.x = x
		return
	}
	m.x = x
	m.columnCursor = cursor
}

// moveNormalRight moves the screen right.
func (root *Root) moveNormalRight(n int) {
	m := root.Doc
	end := root.endRight()
	if end < m.x+n {
		m.x = end
	}
	m.x += n
}

// correctCursor returns the best cursor position from the current x position.
func (root *Root) correctCursor(cursor int) int {
	m := root.Doc
	if m.WrapMode {
		return cursor
	}

	width := root.scr.vWidth - root.scr.startX
	if m.ColumnWidth {
		line, valid := m.getLineC(m.topLN+m.firstLine(), m.TabWidth)
		if !valid {
			return cursor
		}
		return cursorFromPosition(line, m.columnWidths, cursor, m.x, m.x+width)
	}

	// delimiter
	for i := 0; i < m.firstLine()+TargetLineDelimiter; i++ {
		line, valid := m.getLineC(m.topLN+m.firstLine()+i, m.TabWidth)
		if !valid {
			continue
		}
		widths := splitByDelimiter(line.str, m.ColumnDelimiter, m.ColumnDelimiterReg)
		if len(widths) <= cursor {
			continue
		}
		return cursorFromPosition(line, widths, cursor, m.x, m.x+width)
	}
	return cursor
}

// cursorFromPosition returns the cursor x position.
// Returns the current cursor position
// if the cursor position is contained within the currently displayed screen.
// Returns the cursor position moved into the screen
// if the cursor position is not contained within the currently displayed screen.
func cursorFromPosition(line LineC, widths []int, cursor int, start int, end int) int {
	if len(widths)-1 < cursor {
		return len(widths) - 1
	}

	curPos := line.pos.x(widths[cursor])
	if curPos > start && curPos < end {
		return cursor
	}

	if curPos < start {
		for n := 0; n < len(widths); n++ {
			if widths[n] > start {
				return n
			}
		}
		return len(widths) - 1
	}

	if curPos > end {
		for n := len(widths) - 1; n >= 0; n-- {
			if widths[n] < end {
				return n
			}
		}
		return 0
	}
	return cursor
}

// correctX returns the best x position from the cursor.
func (root *Root) correctX(cursor int) (int, error) {
	m := root.Doc

	if cursor == 0 {
		return 0, nil
	}

	if m.ColumnWidth {
		if cursor < len(m.columnWidths) {
			return m.columnWidths[cursor-1], nil
		}
		return m.columnWidths[len(m.columnWidths)-1], nil
	}

	for i := 0; i < m.firstLine()+TargetLineDelimiter; i++ {
		line, valid := m.getLineC(m.topLN+m.firstLine()+i, m.TabWidth)
		if !valid {
			continue
		}
		widths := splitByDelimiter(line.str, m.ColumnDelimiter, m.ColumnDelimiterReg)
		if cursor > 0 && cursor < len(widths) {
			return line.pos.x(widths[cursor]) - columnMargin, nil
		}
	}
	return 0, ErrNoColumn
}

// columnX returns x and cursor when moving right and left.
// If the cursor is out of range, it returns an error.
// moveTo is positive for right(+1) and negative for left(-1).
func (root *Root) columnX(moveTo int) (int, int, error) {
	if root.Doc.ColumnWidth {
		return root.columnWidthX(moveTo)
	}
	return root.columnDelimiterX(moveTo)
}

// columnWidthX returns x and cursor from the orientation to move.
func (root *Root) columnWidthX(moveTo int) (int, int, error) {
	m := root.Doc

	width := root.scr.vWidth - root.scr.startX

	cursor := m.columnCursor + moveTo
	if cursor < 0 {
		return m.x, m.columnCursor, ErrNoColumn
	}

	var widths = make([]int, 0, len(m.columnWidths)+2)
	widths = append(widths, 0)
	widths = append(widths, m.columnWidths...)
	var cl, cr int
	if cursor < len(widths)-1 {
		cl = widths[cursor]
		cr = widths[cursor+1] - 1
	} else {
		cl = widths[len(widths)-1]
		cr = root.rightMost()
	}
	return screenAdjustX(m.x, m.x+width, cl, cr, widths, cursor)
}

// columnDelimiterX returns x and cursor from the orientation to move.
func (root *Root) columnDelimiterX(moveTo int) (int, int, error) {
	width := root.scr.vWidth - root.scr.startX
	if root.Doc.WrapMode {
		// dummy width
		width = root.scr.vWidth * 2
	}
	m := root.Doc
	maxColumn := 0
	cursor := max(0, m.columnCursor+moveTo)
	// m.firstLine()+TargetLineDelimiter = Maximum columnMode target.
	for i := 0; i < m.firstLine()+TargetLineDelimiter; i++ {
		line, valid := m.getLineC(m.topLN+m.firstLine()+i, m.TabWidth)
		if !valid {
			continue
		}
		widths := splitByDelimiter(line.str, m.ColumnDelimiter, m.ColumnDelimiterReg)
		maxColumn = max(maxColumn, len(widths)-1)
		if len(widths) <= 0 {
			continue
		}
		if cursor >= 0 && cursor < len(widths) {
			cl := line.pos.x(widths[cursor])
			cr := line.pos.x(len(line.str))
			if cursor < len(widths)-1 {
				cr = line.pos.x(widths[cursor+1])
			}
			return screenAdjustX(m.x, m.x+width, cl, cr, widths, cursor)
		} else {
			cl := line.pos.x(widths[len(widths)-1])
			cr := line.pos.x(len(line.str))
			return screenAdjustX(m.x, m.x+width, cl, cr, widths, cursor)
		}
	}

	if maxColumn > 0 {
		return m.x, m.columnCursor, ErrNoColumn
	}
	return 0, m.columnCursor, ErrNoDelimiter
}

// screenAdjustX returns x and cursor when moving left and right.
// If it moves too much, adjust the position and return.
// Returns an error when the end is reached.
func screenAdjustX(left int, right int, cl int, cr int, widths []int, cursor int) (int, int, error) {
	//       |left  cl            cr   right|
	// |<---width--->|<---width--->|<---width--->| widths
	//                cursor
	width := right - left
	// right edge + 1
	if cursor > len(widths)-1 {
		if cr > right {
			return cr - width + columnMargin, cursor - 1, nil
		}
		// don't scroll
		if left <= cl && cr < right {
			return left, cursor - 1, ErrOverScreen
		}
		return cl, cursor - 1, ErrOverScreen
	}

	// move cursor without scrolling
	if left <= cl && cr < right {
		return left, cursor, nil
	}
	// 0 ~ right edge
	// left scroll
	if cl < left {
		// Don't move the cursor.
		if cr < left {
			return max(0, left-width), cursor + 1, nil
		}

		// Move the cursor. Position at the beginning of the column.
		move := (left - cl)
		if move < right-left {
			return max(0, cl-columnMargin), cursor, nil
		}

		// Move the cursor. Position at the end of the column.
		if cl < left-width {
			return left - width, cursor, nil
		}

		log.Println("????")
		return left - width, cursor + 1, nil
	}

	// right scroll
	// cr >= right
	move := (cl - left)
	if move > width {
		return left + width - columnMargin, cursor - 1, nil
	}

	// little scroll.
	if left+(cr-cl) < cl {
		return left + (cr - cl) + columnMargin, cursor, nil
	}

	return cl - columnMargin, cursor, nil
}

// splitByDelimiter return a slice split by delimiter
func splitByDelimiter(str string, delimiter string, delimiterReg *regexp.Regexp) []int {
	indexes := allIndex(str, delimiter, delimiterReg)
	if len(indexes) == 0 {
		return nil
	}
	widths := make([]int, 0, len(indexes)+1)

	// The leftmost fence is not a delimiter.
	// If there is no fence on the left edge, the value starts from 0.
	if indexes[0][0] != 0 {
		widths = append(widths, 0)
	}
	for i := 0; i < len(indexes); i++ {
		widths = append(widths, indexes[i][1]+1)
	}
	// rightmost fence is not a delimiter.
	if len(str) == indexes[len(indexes)-1][1] {
		return widths[:len(widths)-1]
	}
	return widths
}

// Move to the left by half a screen.
// Called from a EventKey.
func (root *Root) moveHfLeft() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	m := root.Doc

	moveSize := (root.scr.vWidth / 2)
	if m.x > 0 && (m.x-moveSize) < 0 {
		m.x = 0
		return
	}

	m.x -= moveSize
	if m.x < root.minStartX {
		m.x = root.minStartX
	}
}

// Move to the right by half a screen.
// Called from a EventKey.
func (root *Root) moveHfRight() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	m := root.Doc

	if m.x < 0 {
		m.x = 0
		return
	}

	m.x += (root.scr.vWidth / 2)
}

// moveBeginLeft moves to the beginning of the line.
func (root *Root) moveBeginLeft() {
	root.Doc.x = 0
}

// moveEndRight moves to the end of the line.
// Move so that the end of the currently displayed line is visible.
func (root *Root) moveEndRight() {
	root.Doc.x = root.endRight()
}

// endRight returns x when the content displayed on the current screen is right edge.
func (root *Root) endRight() int {
	x := root.rightMost()
	return x - ((root.scr.vWidth - root.scr.startX) - 1)
}

// Rightmost of the content displayed on the screen.
func (root *Root) rightMost() int {
	m := root.Doc
	maxLen := 0
	for _, line := range root.scr.numbers {
		lY := line.number
		lc, err := m.contents(lY, m.TabWidth)
		if err != nil {
			continue
		}
		maxLen = max(maxLen, len(lc)-1)
	}
	return maxLen
}

// rightEndColumn returns the number of columns of the rightmost line.
func (root *Root) rightEndColumn() int {
	m := root.Doc

	if m.ColumnWidth {
		return len(m.columnWidths)
	}

	maxColumn := 0
	for i := 0; i < m.firstLine()+TargetLineDelimiter; i++ {
		line, valid := m.getLineC(m.topLN+m.firstLine()+i, m.TabWidth)
		if !valid {
			continue
		}
		widths := splitByDelimiter(line.str, m.ColumnDelimiter, m.ColumnDelimiterReg)
		maxColumn = max(maxColumn, len(widths)-1)
	}
	return maxColumn
}
