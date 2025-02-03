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

// targetLine returns the target line that contains columns.
func (m *Document) targetLine(scr SCR) (LineC, error) {
	for lN := m.topLN + m.firstLine(); lN < m.topLN+m.firstLine()+TargetLineDelimiter; lN++ {
		lineC, ok := scr.lines[lN]
		if !ok || !lineC.valid {
			continue
		}
		if len(lineC.columnRanges) == 0 {
			continue
		}
		return lineC, nil
	}
	return LineC{}, ErrNoColumn
}

// optimalCursor returns the optimal cursor position from the current x position.
// This function is used to set the column cursor to the column displayed on the screen when in columnMode.
func (m *Document) optimalCursor(scr SCR, cursor int) int {
	lineC, err := m.targetLine(scr)
	if err != nil {
		return cursor
	}

	left := m.x
	right := m.x + (scr.vWidth - scr.startX)
	cursor = max(0, cursor)
	columns := lineC.columnRanges
	if len(columns)-1 < cursor {
		return len(columns) - 1
	}

	curPos := columns[cursor].start
	if curPos > left && curPos < right {
		return cursor
	}

	if curPos < left {
		for n := 0; n < len(columns); n++ {
			if columns[n].start > left {
				return n
			}
		}
		return len(columns) - 1
	}

	if curPos > right {
		for n := len(columns) - 1; n >= 0; n-- {
			if columns[n].end < right {
				return n
			}
		}
		return 0
	}
	return cursor
}

// optimalX returns the optimal x position from the cursor.
// This function is used to set the x position to the column displayed on the screen when in columnMode.
func (m *Document) optimalX(scr SCR, cursor int) (int, error) {
	if cursor == 0 {
		return 0, nil
	}
	if cursor < m.HeaderColumn {
		return 0, nil
	}
	lineC, err := m.targetLine(scr)
	if err != nil {
		return 0, err
	}

	right := m.x + (scr.vWidth - scr.startX)
	columns := lineC.columnRanges
	m.setColumnStart(columns)
	if cursor >= len(columns) {
		return m.x, nil
	}
	if columns[cursor].end < right {
		return m.x, nil
	}

	// Move if off screen
	vh := m.vHeaderWidth(lineC)
	x := (columns[cursor].start - vh) - columnMargin
	end := len(lineC.lc)
	if end-x < scr.vWidth {
		x = end - (scr.vWidth - columnMargin)
	}
	return x, nil
}

// moveTo returns x and cursor when moving right and left.
// If the cursor is out of range, it returns an error.
// moveTo is positive for right(+1) and negative for left(-1).
func (m *Document) moveTo(scr SCR, moveTo int) (int, int, error) {
	cursor := max(0, m.columnCursor+moveTo)
	if (cursor == 0) || (cursor < m.HeaderColumn) {
		return 0, cursor, nil
	}

	lineC, err := m.targetLine(scr)
	if err != nil {
		return 0, m.columnCursor, err
	}
	columns := lineC.columnRanges
	m.setColumnStart(columns)
	if cursor >= len(columns) {
		return m.x, cursor, ErrOverScreen
	}
	cl, cr := 0, 0
	if cursor > 0 && cursor < len(columns) {
		cl = columns[cursor].start
		cr = columns[cursor].end - 1
	}
	vh := 0
	if m.HeaderColumn > 0 && m.HeaderColumn < len(columns) {
		vh = columns[m.HeaderColumn-1].end
	}
	mx, cursor, err := screenAdjustX(m.x+vh, m.x+scr.vWidth, cl, cr, columns, cursor)
	return mx - vh, cursor, err
}

// screenAdjustX returns x and cursor when moving left and right.
// If it moves too much, adjust the position and return.
// Returns an error when the end is reached.
func screenAdjustX(left int, right int, cl int, cr int, columns []columnRange, cursor int) (int, int, error) {
	//       |left  cl            cr   right|
	// |<---width--->|<---width--->|<---width--->| widths
	//                cursor
	width := (right - left) - columnMargin
	// move cursor without scrolling
	if left <= cl && cr < right {
		return left, cursor, nil
	}

	// right edge + 1
	if cursor > len(columns)-1 {
		if cr > right {
			return cr - width + columnMargin, cursor - 1, nil
		}
		// don't scroll
		if left <= cl && cr < right {
			return left, cursor - 1, ErrOverScreen
		}
		if columns[len(columns)-1].end > right {
			return right, cursor - 1, nil
		}
		return left, cursor - 1, ErrOverScreen
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
		if columns[len(columns)-1].end < right {
			return cr - (right - columnMargin), cursor - 1, nil
		}
		return right - width, cursor - 1, nil
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

func (m *Document) setColumnStart(columns []columnRange) {
	if len(columns) > 0 && columns[0].end == 0 {
		m.columnStart = 1
	}
}
