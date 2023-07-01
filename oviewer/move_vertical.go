package oviewer

import (
	"context"
	"log"
	"sync/atomic"
)

// lastLineMargin is the margin of the last line
const lastLineMargin = 1

// moveLine moves to the specified line.
func (m *Document) moveLine(lN int) int {
	if m.BufEndNum() == 0 {
		return 0
	}
	if lN > m.BufEndNum() {
		return m.BufEndNum() - 1
	}
	m.topLN = lN
	m.topLX = 0
	return lN
}

// moveTop moves to the top.
func (m *Document) moveTop() {
	m.moveLine(m.BufStartNum())
}

// moveBottom moves to the bottom.
func (m *Document) moveBottom() {
	// If the file is seekable, move to the end of the file.
	if m.seekable && atomic.LoadInt32(&m.store.eof) == 0 && atomic.LoadInt32(&m.tmpFollow) == 0 {
		m.requestEnd()
	}

	m.topLX, m.topLN = m.bottomLineNum(m.BufEndNum()-1, m.height-lastLineMargin)
}

// Move to the nth wrapping line of the specified line.
func (m *Document) moveLineNth(lN int, nTh int) (int, int) {
	lN = m.moveLine(lN)
	if !m.WrapMode {
		return lN, 0
	}

	listX := m.leftMostX(lN + m.firstLine())
	if len(listX) == 0 {
		return lN, 0
	}

	if nTh >= len(listX) {
		nTh = len(listX) - 1
	}

	m.topLN = lN
	m.topLX = listX[nTh]

	return lN, nTh
}

// movePgUp moves up one screen.
func (m *Document) movePgUp() {
	m.moveYUp(m.height)
	if m.topLN < m.BufStartNum() {
		m.moveTop()
	}
}

// movePgDn moves down one screen.
func (m *Document) movePgDn() {
	y := m.bottomLN - m.firstLine()
	x := m.bottomLX
	m.limitMoveDown(x, y)
}

// moveHfUp moves up half a screen.
func (m *Document) moveHfUp() {
	m.moveYUp(m.height / 2)
	if m.topLN < m.BufStartNum() {
		m.moveTop()
	}
}

// moveHfDn moves down half a screen.
func (m *Document) moveHfDn() {
	m.moveYDown(m.height / 2)
}

// limitMoveDown limits the movement of the cursor when moving down.
func (m *Document) limitMoveDown(x int, y int) {
	if y+m.height < m.BufEndNum()-m.SkipLines {
		m.topLN = y
		m.topLX = x
		return
	}

	tx, tn := m.bottomLineNum(m.BufEndNum()-1, m.height-lastLineMargin)
	if y < tn || (y == tn && x < tx) {
		m.topLN = y
		m.topLX = x
		return
	}

	if m.topLN < tn || (m.topLN == tn && m.topLX < tx) {
		m.topLN = tn
		m.topLX = tx
	}
}

// numOfWrap returns the number of wrap from lX and lY.
func (m *Document) numOfWrap(lX int, lY int) int {
	listX := m.leftMostX(lY)
	if len(listX) == 0 {
		return 0
	}
	return numOfSlice(listX, lX)
}

// numOfSlice returns what number x is in slice.
func numOfSlice(listX []int, x int) int {
	for n, v := range listX {
		if v >= x {
			return n
		}
	}
	return len(listX) - 1
}

// numOfReverseSlice returns what number x is from the back of slice.
func numOfReverseSlice(listX []int, x int) int {
	for n := len(listX) - 1; n >= 0; n-- {
		if listX[n] <= x {
			return n
		}
	}
	return 0
}

// MoveYUp moves up by the specified number of y.
func (m *Document) moveYUp(moveY int) {
	if !m.WrapMode {
		m.topLN -= moveY
		return
	}

	// WrapMode
	LX, lN := m.numUp(m.topLX, m.topLN+m.firstLine(), moveY)
	m.topLX = LX
	m.topLN = lN - m.firstLine()
}

// bottomLineNum returns the display start line
// when the last line number as an argument.
func (m *Document) bottomLineNum(lN int, height int) (int, int) {
	if lN < m.headerLen {
		return 0, 0
	}
	if !m.WrapMode {
		return 0, (lN + 1) - (height + m.firstLine())
	}
	// WrapMode
	listX := m.leftMostX(lN)
	topLX, topLN := m.numUp(listX[len(listX)-1], lN, height)
	return topLX, topLN - m.firstLine()
}

// numUp finds lX, lN when the number of lines is moved up from lX, lN.
func (m *Document) numUp(lX int, lN int, upY int) (int, int) {
	listX := m.leftMostX(lN)
	n := len(listX)
	for y := upY; y > 0; y-- {
		if n <= 0 {
			lN--
			listX = m.leftMostX(lN)
			n = len(listX)
		}
		if n > 0 {
			lX = listX[n-1]
		} else {
			lX = 0
		}
		n--
	}
	return lX, lN
}

// MoveYDown moves down by the specified number of y.
func (m *Document) moveYDown(moveY int) {
	lN := m.topLN + m.firstLine()
	if !m.WrapMode {
		m.limitMoveDown(0, lN+moveY)
		return
	}

	// WrapMode
	x := m.topLX
	listX := m.leftMostX(lN)
	n := numOfReverseSlice(listX, x)

	for y := 0; y < moveY; y++ {
		if n >= len(listX) {
			lN++
			if lN > m.BufEndNum() {
				break
			}
			listX = m.leftMostX(lN)
			n = 0
		}
		x = 0
		if len(listX) > 0 && n < len(listX) {
			x = listX[n]
		}
		n++
	}

	m.limitMoveDown(x, lN-m.Header)
}

// moveUp moves up by the specified number of lines.
func (m *Document) moveUp(n int) {
	if (m.topLN < m.BufStartNum()) || ((m.topLN == m.BufStartNum()) && (m.topLX == 0)) {
		return
	}

	if !m.WrapMode {
		m.topLN -= n
		m.topLX = 0
		return
	}

	// WrapMode.
	// Same line.
	if m.topLX > 0 {
		listX := m.leftMostX(m.topLN + m.firstLine())
		for n, x := range listX {
			if x >= m.topLX {
				m.topLX = listX[n-1]
				return
			}
		}
	}

	// Previous line.
	m.topLN -= n
	start := m.BufStartNum()
	if m.topLN < start {
		m.topLN = start
		m.topLX = 0
		return
	}
	listX := m.leftMostX(m.topLN + m.firstLine())
	if len(listX) > 0 {
		m.topLX = listX[len(listX)-1]
		return
	}
	m.topLX = 0
}

// moveDown	moves down by the specified number of lines.
func (m *Document) moveDown(n int) {
	lN := m.topLN
	if !m.WrapMode {
		lN += n
		m.limitMoveDown(0, lN)
		return
	}

	// WrapMode
	listX := m.leftMostX(m.topLN + m.firstLine())
	for _, x := range listX {
		if x > m.topLX {
			m.limitMoveDown(x, lN)
			return
		}
	}

	// Next line.
	lN += n
	m.topLX = 0
	m.limitMoveDown(m.topLX, lN)
}

// moveNextSection moves to the next section.
func (m *Document) moveNextSection() error {
	// Move by page, if there is no section delimiter.
	if m.SectionDelimiter == "" {
		m.movePgDn()
		return nil
	}

	lN, err := m.nextSection(m.topLN + m.firstLine())
	if err != nil {
		m.movePgDn()
		return ErrNoDelimiter
	}
	m.moveLine((lN - m.firstLine()) + m.SectionStartPosition)
	return nil
}

// nextSection returns the line number of the previous section.
func (m *Document) nextSection(n int) (int, error) {
	lN := n + (1 - m.SectionStartPosition)
	searcher := NewSearcher(m.SectionDelimiter, m.SectionDelimiterReg, true, true)
	ctx := context.Background()
	defer ctx.Done()
	n, err := m.SearchLine(ctx, searcher, lN)
	if err != nil {
		return n, err
	}
	return n, nil
}

// movePrevSection moves to the previous section.
func (m *Document) movePrevSection() error {
	// Move by page, if there is no section delimiter.
	if m.SectionDelimiter == "" {
		m.movePgUp()
		return nil
	}

	lN, err := m.prevSection(m.topLN + m.firstLine())
	if err != nil {
		m.moveTop()
		return err
	}
	m.moveLine(lN)
	return nil
}

// prevSection returns the line number of the previous section.
func (m *Document) prevSection(n int) (int, error) {
	lN := n - (1 + m.SectionStartPosition)
	searcher := NewSearcher(m.SectionDelimiter, m.SectionDelimiterReg, true, true)
	ctx := context.Background()
	defer ctx.Done()
	n, err := m.BackSearchLine(ctx, searcher, lN)
	if err != nil {
		return 0, err
	}
	n = (n - m.firstLine()) + m.SectionStartPosition
	n = max(n, m.BufStartNum())
	return n, nil
}

// moveLastSection moves to the last section.
func (m *Document) moveLastSection() {
	// +1 to avoid if the bottom line is a session delimiter.
	lN := m.BufEndNum() - 2
	searcher := NewSearcher(m.SectionDelimiter, m.SectionDelimiterReg, true, true)
	ctx := context.Background()
	defer ctx.Done()
	n, err := m.BackSearchLine(ctx, searcher, lN)
	if err != nil {
		log.Printf("last section:%v", err)
		return
	}
	n = (n - m.firstLine()) + m.SectionStartPosition
	m.moveLine(n)
}

// leftMostX returns a list of left - most x positions when wrapping.
// Returns nil if there is no line number.
func (m *Document) leftMostX(lN int) []int {
	lc, err := m.contents(lN, m.TabWidth)
	if err != nil {
		return nil
	}
	return leftX(m.width, lc)
}

// leftX returns a list of left - most x positions when wrapping.
func leftX(width int, lc contents) []int {
	listX := make([]int, 0, (len(lc)/width)+1)
	listX = append(listX, 0)
	for n := width; n < len(lc); n += width {
		if lc[n-1].width == 2 {
			n--
		}
		listX = append(listX, n)
	}
	return listX
}
