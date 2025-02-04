package oviewer

// columnMargin is the number of characters in the left and right margins of the screen.
const columnMargin = 2

// moveHfLeft moves to the left half screen.
func (m *Document) moveHfLeft() {
	hfSize := (m.width / 2)
	if m.x > 0 && (m.x-hfSize) <= 0 {
		m.x = 0
		return
	}
	m.x -= hfSize
}

// moveHfRight moves to the right half screen.
func (m *Document) moveHfRight() {
	if m.x <= 0 {
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

// moveBeginLeft moves the document view to the left edge of the screen.
// It sets the horizontal position to 0 and determines the starting column.
func (m *Document) moveBeginLeft(scr SCR) {
	m.x = 0
	m.columnCursor = determineColumnStart(scr.lines)
}

// moveEndRight moves to the right edge of the screen.
func (m *Document) moveEndRight(scr SCR) {
	m.x, m.columnCursor = rightEdge(scr.lines, m.width)
}

// rightEdge returns the right edge of the screen.
func rightEdge(lines map[int]LineC, width int) (int, int) {
	maxX := 0
	maxColumn := 0
	for _, lineC := range lines {
		if !lineC.valid {
			continue
		}
		maxX = max(maxX, len(lineC.lc)-width)
		maxColumn = max(maxColumn, len(lineC.columnRanges)-1)
	}
	return maxX, maxColumn
}

// optimalCursor returns the optimal cursor position from the current x position.
// This function is used to set the column cursor to the column displayed on the screen when in columnMode.
func (m *Document) optimalCursor(scr SCR, cursor int) int {
	lineC, err := targetLineWithColumns(scr)
	if err != nil {
		return cursor
	}

	cursor = max(cursor, determineColumnStart(scr.lines))
	columns := lineC.columnRanges
	if len(columns)-1 < cursor {
		// If the cursor is out of range, set it to the last column.
		return len(columns) - 1
	}

	left := m.x
	right := m.x + (scr.vWidth - scr.startX)
	curPos := columns[cursor].start
	if curPos > left && curPos < right {
		// If the cursor is within the range,
		return cursor
	}

	// If the cursor is out of range, move to the nearest column.
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

	lineC, err := targetLineWithColumns(scr)
	if err != nil {
		return 0, err
	}

	columns := lineC.columnRanges
	if cursor >= len(columns) {
		return m.x, ErrOverScreen
	}

	// Move if on screen.
	if columns[cursor].end < m.x+(scr.vWidth-scr.startX) {
		return m.x, nil
	}

	// Move if off screen.
	vh := m.vHeaderWidth(lineC)
	x := (columns[cursor].start - vh) - columnMargin
	rightEdge := len(lineC.lc)
	if rightEdge-x < scr.vWidth {
		x = rightEdge - (scr.vWidth - columnMargin)
	}
	return x, nil
}

// moveColumnLeft moves to the left column.
func (m *Document) moveColumnLeft(n int, scr SCR, cycle bool) error {
	lineC, err := targetLineWithColumns(scr)
	if err != nil {
		return err
	}

	vh := m.vHeaderWidth(lineC)
	cursor := m.columnCursor
	if !m.WrapMode && isValidCursor(lineC, cursor) {
		cl := lineC.columnRanges[cursor].start
		if m.x > cl {
			// Movement of x only.
			move := ((m.x - vh) - (scr.vWidth + columnMargin)) * n
			m.x = max(move, cl)
			return nil
		}
	}

	columnStart := determineColumnStart(scr.lines)
	if m.columnCursor <= columnStart {
		if cycle {
			m.moveEndRight(scr)
			return nil
		}
		return ErrOverScreen
	}

	// Move to the previous column.
	cursor = m.columnCursor - n
	if cursor < m.HeaderColumn {
		m.x = 0
		m.columnCursor = cursor
		return nil
	}
	m.columnCursor = cursor

	// Movement x.

	cl := lineC.columnRanges[cursor].start
	cr := lineC.columnRanges[cursor].end
	x := max(0, m.x)
	if cl > x+vh && cr < x+scr.vWidth {
		// Movement of cursor only (x is within the range).
		return nil
	}

	if cr < scr.vWidth-columnMargin {
		// Set m.x to 0 because the right edge of the cursor fits within the screen
		// even when displayed from the beginning.
		m.x = 0
		return nil
	}
	m.x = max(cl-columnMargin, cr-(scr.vWidth-columnMargin)) - vh
	return nil
}

// moveColumnRight moves to the right column.
func (m *Document) moveColumnRight(n int, scr SCR, cycle bool) error {
	lineC, err := targetLineWithColumns(scr)
	if err != nil {
		return err
	}

	vh := m.vHeaderWidth(lineC)
	cursor := m.columnCursor
	if !m.WrapMode && isValidCursor(lineC, cursor) {
		cr := lineC.columnRanges[cursor].end
		if m.x+scr.vWidth < cr {
			// Movement of x only.
			move := ((m.x - vh) + (scr.vWidth - columnMargin)) * n
			m.x = min(move, cr-scr.vWidth)
			return nil
		}
	}

	// Move to the next column.
	cursor = m.columnCursor + n
	if !isValidCursor(lineC, cursor) {
		if cycle {
			m.moveBeginLeft(scr)
			return nil
		}
		return ErrOverScreen
	}
	m.columnCursor = cursor
	rightEdge := (m.x + scr.vWidth) - columnMargin

	cl := lineC.columnRanges[cursor].start
	cr := lineC.columnRanges[cursor].end
	if cr < rightEdge {
		// Movement of cursor only.
		return nil
	}

	// Movement x.

	if len(lineC.lc) < rightEdge {
		// Set m.x to the right edge of the screen because the line is shorter than the screen width.
		m.x = len(lineC.lc) - (scr.vWidth + columnMargin)
		return nil
	}
	m.x = min(cr+columnMargin, cl+(scr.vWidth-columnMargin)) - scr.vWidth
	return nil
}

// targetLineWithColumns returns the target line that contains columns.
// It iterates through the lines in the screen range and returns the first valid line with columns.
func targetLineWithColumns(scr SCR) (LineC, error) {
	for lN := scr.bodyLN; lN < scr.bodyEnd; lN++ {
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

// isValidCursor checks if the given cursor position is within the valid range of column ranges in the provided LineC.
// It returns true if the cursor is within the range, otherwise false.
func isValidCursor(lineC LineC, cursor int) bool {
	return cursor >= 0 && cursor < len(lineC.columnRanges)
}

// determineColumnStart determines the start index of the column.
// If the column starts with a delimiter, it returns 1. Otherwise, it returns 0.
// This function checks all lines to ensure that a CSV file starting with a comma
// is correctly interpreted as having an empty first value.
func determineColumnStart(lines map[int]LineC) int {
	for _, lineC := range lines {
		if !lineC.valid {
			continue
		}
		columns := lineC.columnRanges
		if len(columns) == 0 {
			continue
		}
		if columns[0].end > 0 {
			return 0
		}
	}
	return 1
}
