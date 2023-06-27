package oviewer

import (
	"errors"
	"log"
	"regexp"
)

// columnMargin is the number of characters in the left and right margins of the screen.
const columnMargin = 2

// TargetLineDelimiter covers from line 1 to this line,
// because there are headers and separators.
const TargetLineDelimiter = 10

// moveBeginLeft moves to the beginning of the screen.
func (m *Document) moveBeginLeft() {
	m.x = 0
}

// moveEndRight moves to the end of the screen.
func (m *Document) moveEndRight(scr SCR) {
	m.x = m.endRight(scr)
}

// correctCursor returns the best cursor position from the current x position.
func (m *Document) correctCursor(cursor int) int {
	if m.WrapMode {
		return cursor
	}

	if m.ColumnWidth {
		return m.correctCursorWidth(cursor)
	}
	return m.correctCursorDelimiter(cursor)
}

func (m *Document) correctCursorWidth(cursor int) int {
	line, valid := m.getLineC(m.topLN+m.firstLine(), m.TabWidth)
	if !valid {
		return cursor
	}
	return cursorFromPosition(line, m.columnWidths, cursor, m.x, m.x+m.width)
}

func (m *Document) correctCursorDelimiter(cursor int) int {
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
		return cursorFromPosition(line, widths, cursor, m.x, m.x+m.width)
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
func (m *Document) correctX(cursor int) (int, error) {
	if cursor == 0 {
		return 0, nil
	}

	if m.ColumnWidth {
		return m.correctXWidth(cursor)
	}
	return m.correctXDelimiter(cursor)
}

func (m *Document) correctXWidth(cursor int) (int, error) {
	if cursor < len(m.columnWidths) {
		return m.columnWidths[cursor-1], nil
	}
	return m.columnWidths[len(m.columnWidths)-1], nil
}

func (m *Document) correctXDelimiter(cursor int) (int, error) {
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
func (m *Document) columnX(scr SCR, moveTo int) (int, int, error) {
	if m.ColumnWidth {
		return m.columnXWidth(scr, moveTo)
	}
	return m.columnXDelimiter(moveTo)
}

// columnXWidth returns x and cursor from the orientation to move.
func (m *Document) columnXWidth(scr SCR, moveTo int) (int, int, error) {
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
		cr = m.rightMost(scr)
	}
	return screenAdjustX(m.x, m.x+m.width, cl, cr, widths, cursor)
}

// columnXDelimiter returns x and cursor from the orientation to move.
func (m *Document) columnXDelimiter(moveTo int) (int, int, error) {
	width := m.width
	if m.WrapMode {
		// dummy width
		width = width * 2
	}
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

// moveHfLeft moves to the left half screen.
func (m *Document) moveHfLeft() {
	moveSize := (m.width / 2)
	if m.x > 0 && (m.x-moveSize) < 0 {
		m.x = 0
		return
	}
	m.x -= moveSize
}

// moveHfRight moves to the right half screen.
func (m *Document) moveHfRight() {
	if m.x < 0 {
		m.x = 0
		return
	}
	m.x += (m.width / 2)
}

// moveNormalLeft moves to the left.
func (m *Document) moveNormalLeft(n int) {
	m.x -= n
}

// moveNormalRight moves to the right.
func (m *Document) moveNormalRight(n int) {
	m.x += n
}

// moveColumnLeft moves to the left column.
func (m *Document) moveColumnLeft(n int, scr SCR, cycle bool) error {
	if m.columnCursor <= 0 && m.x <= 2 {
		if !cycle {
			cursor := m.rightEndColumn()
			m.x = max(0, m.endRight(scr)+columnMargin)
			m.columnCursor = cursor
			return nil
		}
	}

	x, cursor, err := m.columnX(scr, -1)
	if err != nil {
		m.x = x
		return err
	}
	m.x = x
	m.columnCursor = cursor
	return nil
}

// moveColumnRight moves to the right column.
func (m *Document) moveColumnRight(n int, scr SCR, cycle bool) error {
	x, cursor, err := m.columnX(scr, 1)
	if err != nil {
		if errors.Is(err, ErrOverScreen) {
			if !cycle {
				m.x = 0
				m.columnCursor = 0
				return nil
			}
		}
		m.x = x
		return err
	}
	m.x = x
	m.columnCursor = cursor
	return nil
}

// endRight returns x when the content displayed on the current screen is right edge.
func (m *Document) endRight(scr SCR) int {
	x := m.rightMost(scr)
	x = max(0, x-(m.width-1))
	return x
}

// rightmost of the content displayed on the screen.
func (m *Document) rightMost(scr SCR) int {
	maxLen := 0
	for _, line := range scr.numbers {
		lc, err := m.contents(line.number, m.TabWidth)
		if err != nil {
			continue
		}
		maxLen = max(maxLen, len(lc)-1)
	}
	return maxLen
}

// rightEndColumn returns the number of columns of the rightmost line.
func (m *Document) rightEndColumn() int {
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
