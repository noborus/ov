package oviewer

import (
	"context"
	"log"
	"sync/atomic"
)

// moveLine moves to the specified line.
func (m *Document) moveLine(lN int) int {
	lN = min(lN, m.BufEndNum())
	m.topLN = lN
	m.topLX = 0
	return lN
}

// moveTop moves to the top.
func (m *Document) moveTop() {
	m.moveLine(m.BufStartNum())
}

func (m *Document) moveBottom() {
	if m.seekable && atomic.LoadInt32(&m.store.eof) == 0 && atomic.LoadInt32(&m.tmpFollow) == 0 {
		// If the file is seekable, move to the end of the file.
		m.requestEnd()
	}
	m.topLX, m.topLN = m.bottomLineNum(m.height-2, m.BufEndNum())
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

func (m *Document) movePgUp() {
	m.moveNumUp(m.height)
	if m.topLN < m.BufStartNum() {
		m.moveTop()
	}
}

func (m *Document) movePgDn() {
	y := m.bottomLN - m.firstLine()
	x := m.bottomLX
	m.limitMoveDown(x, y)
}

func (m *Document) moveHfUp() {
	m.moveNumUp(m.height / 2)
	if m.topLN < m.BufStartNum() {
		m.moveTop()
	}
}

func (m *Document) moveHfDn() {
	m.moveNumDown(m.height / 2)
}

// limitMoveDown limits the movement of the cursor when moving down.
func (m *Document) limitMoveDown(x int, y int) {
	if y+m.height < m.BufEndNum()-m.SkipLines {
		m.topLN = y
		m.topLX = x
		return
	}

	tx, tn := m.bottomLineNum(m.height-2, m.BufEndNum())
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

// Moves up by the specified number of y.
func (m *Document) moveNumUp(moveY int) {
	if !m.WrapMode {
		m.topLN -= moveY
		return
	}

	// WrapMode
	num := m.topLN + m.firstLine()
	m.topLX, num = m.findNumUp(m.topLX, num, moveY)
	m.topLN = num - m.firstLine()
}

// bottomLineNum returns the display start line
// when the last line number as an argument.
func (m *Document) bottomLineNum(height int, lN int) (int, int) {
	if lN < m.headerLen {
		return 0, 0
	}
	if !m.WrapMode {
		return 0, lN - (height + m.firstLine())
	}
	// WrapMode
	lX, lN := m.findNumUp(height, 0, lN)
	return lX, lN - m.firstLine()
}

// findNumUp finds lX, lN when the number of lines is moved up from lX, lN.
func (m *Document) findNumUp(upY int, lX int, lN int) (int, int) {
	listX := m.leftMostX(lN)
	n := numOfSlice(listX, lX)

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

// Moves down by the specified number of y.
func (m *Document) moveNumDown(moveY int) {
	num := m.topLN + m.firstLine()
	if !m.WrapMode {
		m.limitMoveDown(0, num+moveY)
		return
	}

	// WrapMode
	x := m.topLX
	listX := m.leftMostX(num)
	n := numOfReverseSlice(listX, x)

	for y := 0; y < moveY; y++ {
		if n >= len(listX) {
			num++
			if num > m.BufEndNum() {
				break
			}
			listX = m.leftMostX(num)
			n = 0
		}
		x = 0
		if len(listX) > 0 && n < len(listX) {
			x = listX[n]
		}
		n++
	}

	m.limitMoveDown(x, num-m.Header)
}

func (m *Document) moveUp(n int) {
	if m.topLN <= m.BufStartNum() && m.topLX == 0 {
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

func (m *Document) moveDown(n int) {
	num := m.topLN

	if !m.WrapMode {
		num += n
		m.limitMoveDown(0, num)
		return
	}

	// WrapMode
	listX := m.leftMostX(m.topLN + m.firstLine())
	for _, x := range listX {
		if x > m.topLX {
			m.limitMoveDown(x, num)
			return
		}
	}

	// Next line.
	num += n
	m.topLX = 0
	m.limitMoveDown(m.topLX, num)
}

// nextSection returns the line number of the previous section.
func (m *Document) nextSection(n int) (int, error) {
	num := n + (1 - m.SectionStartPosition)
	searcher := NewSearcher(m.SectionDelimiter, m.SectionDelimiterReg, true, true)
	ctx := context.Background()
	defer ctx.Done()
	n, err := m.SearchLine(ctx, searcher, num)
	if err != nil {
		return n, err
	}
	return n, nil
}

// prevSection returns the line number of the previous section.
func (m *Document) prevSection(n int) (int, error) {
	num := n - (1 + m.SectionStartPosition)
	searcher := NewSearcher(m.SectionDelimiter, m.SectionDelimiterReg, true, true)
	ctx := context.Background()
	defer ctx.Done()
	n, err := m.BackSearchLine(ctx, searcher, num)
	if err != nil {
		return 0, err
	}
	n = (n - m.firstLine()) + m.SectionStartPosition
	n = max(n, m.BufStartNum())
	return n, nil
}

func (m *Document) lastSection() {
	// +1 to avoid if the bottom line is a session delimiter.
	num := m.BufEndNum() - 2
	searcher := NewSearcher(m.SectionDelimiter, m.SectionDelimiterReg, true, true)
	ctx := context.Background()
	defer ctx.Done()
	n, err := m.BackSearchLine(ctx, searcher, num)
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
	return leftMostX(m.width, lc)
}

func leftMostX(width int, lc contents) []int {
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
