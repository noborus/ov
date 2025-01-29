package oviewer

import (
	"errors"
	"regexp"
)

// columnMargin is the number of characters in the left and right margins of the screen.
const columnMargin = 2

// TargetLineDelimiter covers from line 1 to this line,
// because there are headers and separators.
const TargetLineDelimiter = 10

// moveBeginLeft move to the left edge of the screen.
func (m *Document) moveBeginLeft() {
	m.x = 0
	m.columnCursor = 0
}

// moveEndRight move to the right edge of the screen.
func (m *Document) moveEndRight(scr SCR) {
	m.x = m.endRight(scr)
	m.columnCursor = m.rightmostColumn(scr)
}

// optimalCursor returns the optimal cursor position from the current x position.
func (m *Document) optimalCursor(scr SCR, cursor int) int {
	if m.ColumnWidth {
		return m.optimalCursorWidth(scr, cursor)
	}
	return m.optimalCursorDelimiter(scr, cursor)
}

// optimalCursorWidth returns the optimal cursor position when in columnWidth mode.
func (m *Document) optimalCursorWidth(scr SCR, cursor int) int {
	if m.WrapMode {
		return cursor
	}
	lineC, ok := scr.lines[m.topLN+m.firstLine()]
	if !ok || !lineC.valid {
		return cursor
	}
	return optimalCursor(lineC, m.columnWidths, cursor, m.x, m.x+m.width)
}

// optimalCursorDelimiter returns the optimal cursor position when in columnDelimiter mode.
func (m *Document) optimalCursorDelimiter(scr SCR, cursor int) int {
	if m.WrapMode && cursor != 0 {
		return cursor
	}

	for i := 0; i < m.firstLine()+TargetLineDelimiter; i++ {
		lineC, ok := scr.lines[m.topLN+m.firstLine()+i]
		if !ok || !lineC.valid {
			continue
		}
		widths := splitByDelimiter(lineC.str, m.ColumnDelimiter, m.ColumnDelimiterReg)
		m.setColumnStart(widths)
		if len(widths) <= cursor {
			continue
		}
		return optimalCursor(lineC, widths, cursor, m.x, m.x+m.width)
	}
	return cursor
}

// optimalCursor returns the cursor position.
// Returns the current cursor position
// if the cursor position is contained within the currently displayed screen.
// Returns the cursor position moved into the screen
// if the cursor position is not contained within the currently displayed screen.
func optimalCursor(lineC LineC, widths []int, cursor int, start int, end int) int {
	cursor = max(0, cursor)
	if len(widths)-1 < cursor {
		return len(widths) - 1
	}

	curPos := lineC.pos.x(widths[cursor])
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

// optimalX returns the optimal x position from the cursor.
func (m *Document) optimalX(scr SCR, cursor int) (int, error) {
	if cursor == 0 {
		return 0, nil
	}
	if m.ColumnWidth {
		return m.optimalXWidth(cursor)
	}
	return m.optimalXDelimiter(scr, cursor)
}

// optimalXWidth returns the optimal x position of the column at the specified cursor position.
func (m *Document) optimalXWidth(cursor int) (int, error) {
	if len(m.columnWidths) == 0 {
		return 0, ErrNoColumn
	}
	cursor = min(cursor, len(m.columnWidths)) - 1
	return m.columnWidths[cursor], nil
}

// optimalXDelimiter returns the best x position of the column at the specified cursor position.
func (m *Document) optimalXDelimiter(scr SCR, cursor int) (int, error) {
	for i := 0; i < m.firstLine()+TargetLineDelimiter; i++ {
		lineC, ok := scr.lines[m.topLN+m.firstLine()+i]
		if !ok || !lineC.valid {
			continue
		}
		widths := splitByDelimiter(lineC.str, m.ColumnDelimiter, m.ColumnDelimiterReg)
		m.setColumnStart(widths)
		if cursor > 0 && cursor < len(widths) {
			return lineC.pos.x(widths[cursor]) - columnMargin, nil
		}
	}
	return 0, ErrNoColumn
}

// moveTo returns x and cursor when moving right and left.
// If the cursor is out of range, it returns an error.
// moveTo is positive for right(+1) and negative for left(-1).
func (m *Document) moveTo(scr SCR, moveTo int) (int, int, error) {
	if m.ColumnWidth {
		return m.moveToWidth(scr, moveTo)
	}
	return m.moveToDelimiter(scr, moveTo)
}

// moveToWidth returns x and cursor from the orientation to move.
func (m *Document) moveToWidth(scr SCR, moveTo int) (int, int, error) {
	cursor := m.columnCursor + moveTo
	if cursor < 0 {
		return m.x, m.columnCursor, ErrNoColumn
	}

	widths := make([]int, 0, len(m.columnWidths)+2)
	widths = append(widths, 0)
	widths = append(widths, m.columnWidths...)
	var cl, cr int
	if cursor < len(widths)-1 {
		cl = widths[cursor]
		cr = widths[cursor+1] - 1
	} else {
		cl = widths[len(widths)-1]
		cr = m.rightmost(scr)
	}
	return screenAdjustX(m.x, m.x+m.width, cl, cr, widths, cursor)
}

// moveToDelimiter returns x and cursor from the orientation to move.
func (m *Document) moveToDelimiter(scr SCR, moveTo int) (int, int, error) {
	screenWidth := m.width
	if m.WrapMode {
		// dummy width
		screenWidth = screenWidth * 2
	}
	cursor := max(0, m.columnCursor+moveTo)
	for i := 0; i < m.firstLine()+TargetLineDelimiter; i++ {
		lineC, ok := scr.lines[m.topLN+m.firstLine()+i]
		if !ok || !lineC.valid {
			continue
		}
		widths := splitByDelimiter(lineC.str, m.ColumnDelimiter, m.ColumnDelimiterReg)
		if len(widths) > 0 {
			m.setColumnStart(widths)
			return m.adjuextXDelimiter(cursor, widths, lineC, screenWidth)
		}
	}
	return 0, m.columnCursor, ErrNoDelimiter
}

// adjuextXDelimiter returns x and cursor when moving left and right.
func (m *Document) adjuextXDelimiter(cursor int, widths []int, lineC LineC, screenWidth int) (int, int, error) {
	cl, cr := 0, 0
	if cursor > 0 && cursor < len(widths) {
		cl = lineC.pos.x(widths[cursor-1])
		cr = lineC.pos.x(widths[cursor] - 1)
	}
	return screenAdjustX(m.x, m.x+screenWidth, cl, cr, widths, cursor)
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

// splitByDelimiter return a slice split by delimiter.
func splitByDelimiter(str string, delimiter string, delimiterReg *regexp.Regexp) []int {
	indexes := allIndex(str, delimiter, delimiterReg)
	if len(indexes) == 0 {
		return nil
	}
	widths := make([]int, 0, len(indexes)+1)

	for i := 0; i < len(indexes); i++ {
		widths = append(widths, indexes[i][0])
	}
	// Add the last width.
	lastDelmEnd := indexes[len(indexes)-1][1]
	if lastDelmEnd < len(str) {
		widths = append(widths, len(str))
	}
	return widths
}

// moveHfLeft moves to the left half screen.
func (m *Document) moveHfLeft() {
	hfSize := (m.width / 2)
	if m.x > 0 && (m.x-hfSize) < 0 {
		m.x = 0
		return
	}
	m.x -= hfSize
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
	if m.columnCursor <= m.columnStart && m.x <= columnMargin {
		if cycle {
			m.columnCursor = m.rightmostColumn(scr)
			m.x = max(0, m.endRight(scr))
			return nil
		}
	}

	var err error
	m.x, m.columnCursor, err = m.moveTo(scr, -n)
	return err
}

// moveColumnRight moves to the right column.
func (m *Document) moveColumnRight(n int, scr SCR, cycle bool) error {
	x, cursor, err := m.moveTo(scr, n)
	if err != nil {
		if !errors.Is(err, ErrOverScreen) || !cycle {
			return err
		}
		m.x = 0
		m.columnCursor = m.columnStart
		return nil
	}

	m.x = x
	m.columnCursor = cursor
	return nil
}

// endRight returns x when the content displayed on the current screen is right edge.
func (m *Document) endRight(scr SCR) int {
	x := m.rightmost(scr)
	x = max(0, x-(m.width-1))
	return x
}

// rightmost of the content displayed on the screen.
func (m *Document) rightmost(scr SCR) int {
	maxLen := 0
	for _, line := range scr.lines {
		if !line.valid {
			continue
		}
		lc := line.lc
		maxLen = max(maxLen, len(lc)-1)
	}
	return maxLen
}

// rightmostColumn returns the number of rightmost columns.
func (m *Document) rightmostColumn(scr SCR) int {
	if m.ColumnWidth {
		return len(m.columnWidths)
	}

	maxColumn := 0
	for _, line := range scr.lines {
		if !line.valid {
			continue
		}
		widths := splitByDelimiter(line.str, m.ColumnDelimiter, m.ColumnDelimiterReg)
		maxColumn = max(maxColumn, len(widths)-1)
	}
	return maxColumn
}

func (m *Document) setColumnStart(widths []int) {
	if len(widths) > 0 && widths[0] == 0 {
		m.columnStart = 1
	}
}
