package oviewer

// columnMargin is the number of characters in the left and right margins of the screen.
const columnMargin = 2

// moveHfLeft moves to the left half screen.
func (m *Document) moveHfLeft() {
	hfSize := (m.width / 2)
	if m.scrollX > 0 && (m.scrollX-hfSize) <= 0 {
		m.scrollX = 0
		return
	}
	m.scrollX -= hfSize
}

// moveHfRight moves to the right half screen.
func (m *Document) moveHfRight() {
	if m.scrollX < 0 {
		m.scrollX = 0
		return
	}
	m.scrollX += (m.width / 2)
}

// moveNormalLeft moves to the left.
func (m *Document) moveNormalLeft(n int) {
	m.scrollX -= n
}

// moveNormalRight moves to the right.
func (m *Document) moveNormalRight(n int) {
	m.scrollX += n
}

// moveBeginLeft moves the document view to the left edge of the screen.
// It sets the horizontal position to 0 and determines the starting column.
func (m *Document) moveBeginLeft(_ SCR) {
	m.scrollX = 0
	m.columnCursor = m.columnStart
}

// moveEndRight moves to the right edge of the screen.
func (m *Document) moveEndRight(scr SCR) {
	x, cursor := maxLineSize(scr.lines)
	width := m.width - columnMargin
	m.scrollX = max(0, x-width)
	m.columnCursor = cursor
}

// optimalCursor returns the optimal cursor position from the current x position.
// This function is used to set the column cursor to the column displayed on the screen when in columnMode.
func (m *Document) optimalCursor(scr SCR, cursor int) int {
	lineC, err := targetLineWithColumns(scr)
	if err != nil {
		return cursor
	}

	cursor = max(cursor, m.columnStart)
	columns := lineC.columnRanges
	if len(columns)-1 < cursor {
		// If the cursor is out of range, set it to the last column.
		return len(columns) - 1
	}

	leftLimit := m.scrollX + m.vHeaderWidth(lineC)
	rightLimit := (m.scrollX + m.width) - columnMargin
	cl := columns[cursor].start
	cr := columns[cursor].end
	// No need to move if on screen.
	if cl >= leftLimit && cr <= rightLimit {
		return cursor
	}

	// If the cursor is out of range, move to the nearest column.
	if cl < leftLimit {
		for n := range columns {
			if columns[n].start > leftLimit {
				return n
			}
		}
		return len(columns) - 1
	}
	if cl > rightLimit {
		for n := len(columns) - 1; n >= 0; n-- {
			if columns[n].end < rightLimit {
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
	if cursor < m.HeaderColumn+m.columnStart {
		return 0, nil
	}

	lineC, err := targetLineWithColumns(scr)
	if err != nil {
		return 0, err
	}

	columns := lineC.columnRanges
	if cursor >= len(columns) {
		return m.scrollX, ErrOverScreen
	}

	width := m.width - columnMargin
	vh := m.vHeaderWidth(lineC) + 1
	leftLimit := m.scrollX + vh
	rightLimit := m.scrollX + width
	cl := columns[cursor].start + 1
	cr := columns[cursor].end

	// No need to move if on screen.
	if cl > leftLimit && cr < rightLimit {
		return m.scrollX, nil
	}

	// Move if off screen.
	x := (cl - vh) - columnMargin
	lineWidth := len(lineC.lc)
	if lineWidth-x < width {
		x = lineWidth - width
	}
	x = max(0, x)
	return x, nil
}

// moveColumnLeft moves to the left column.
func (m *Document) moveColumnLeft(n int, scr SCR, cycle bool) error {
	lineC, err := targetLineWithColumns(scr)
	if err != nil {
		return err
	}

	width := m.width - columnMargin
	vh := m.vHeaderWidth(lineC) + 1
	// Check if only scrolling is needed without moving the cursor.
	if !m.WrapMode && isValidCursor(lineC, m.columnCursor) {
		cl := lineC.columnRanges[m.columnCursor].start
		if m.scrollX > 0 && m.scrollX+vh > cl {
			// Movement of x only.
			m.scrollX = max(0, cl-vh)
			return nil
		}
	}

	// Check if the cursor is at the beginning of the column.
	cursor := m.columnCursor - n
	if cursor < m.columnStart {
		if cycle {
			m.moveEndRight(scr)
			return nil
		}
		return ErrOverScreen
	}

	// Move to the previous column.
	m.columnCursor = cursor
	if cursor < m.HeaderColumn+m.columnStart {
		m.scrollX = 0
		return nil
	}

	columns := lineC.columnRanges
	if len(columns) <= cursor {
		// If the cursor is out of range (column count is inconsistent across lines),
		// only update the cursor position without changing the display position (m.x).
		// This allows the cursor to move to a position that may exist on other lines.
		return nil
	}

	leftLimit := max(0, m.scrollX) + vh
	cl := columns[cursor].start
	cr := columns[cursor].end

	// No need to move if on screen.
	if cl > leftLimit {
		return nil
	}

	if cr < width {
		// Set m.x to 0 because the right edge of the cursor fits within the screen
		// even when displayed from the beginning.
		m.scrollX = 0
		return nil
	}
	m.scrollX = max(cl-columnMargin, cr-(width-vh)) - vh
	return nil
}

// moveColumnRight moves to the right column.
func (m *Document) moveColumnRight(n int, scr SCR, cycle bool) error {
	lineC, err := targetLineWithColumns(scr)
	if err != nil {
		return err
	}

	width := m.width - columnMargin
	vh := m.vHeaderWidth(lineC) + 1
	// Check if only scrolling is needed without moving the cursor.
	if !m.WrapMode && isValidCursor(lineC, m.columnCursor) {
		cr := lineC.columnRanges[m.columnCursor].end
		if m.scrollX+width < cr {
			// Movement of x only.
			move := ((m.scrollX - vh) + width) * n
			m.scrollX = min(move, cr-width)
			return nil
		}
	}
	// Move to the next column.
	cursor := m.columnCursor + n
	if !isValidCursor(lineC, cursor) {
		if cycle {
			m.moveBeginLeft(scr)
			return nil
		}
		return ErrOverScreen
	}
	m.columnCursor = cursor
	rightLimit := m.scrollX + width
	cl := lineC.columnRanges[cursor].start
	cr := lineC.columnRanges[cursor].end

	// No need to move if on screen.
	if cr < rightLimit {
		return nil
	}

	// Move if off screen.
	if len(lineC.lc) < rightLimit {
		// Set m.x to the right edge of the screen because the line is shorter than the screen width.
		m.scrollX = max(0, len(lineC.lc)-width)
		return nil
	}
	m.scrollX = max(0, min(cr-width, cl-vh))
	return nil
}

// maxLineSize returns the maximum x position and the maximum column position.
func maxLineSize(lines map[int]LineC) (int, int) {
	maxX := 0
	maxColumn := 0
	for _, lineC := range lines {
		if !lineC.valid {
			continue
		}
		maxX = max(maxX, len(lineC.lc))
		maxColumn = max(maxColumn, len(lineC.columnRanges)-1)
	}
	return maxX, maxColumn
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
