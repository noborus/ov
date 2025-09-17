package oviewer

import (
	"context"
	"log"
	"sync/atomic"
)

// lastLineMargin is the margin of the last line.
const lastLineMargin = 1

// moveLine moves to the specified line.
func (m *Document) moveLine(lN int) int {
	if m.BufEndNum() == 0 {
		return 0
	}
	m.showGotoF = true

	if endLN := m.BufEndNum() - 1; lN > endLN {
		lN = endLN
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
		m.requestBottom()
	}

	lN := m.BufEndNum()
	height := m.height - lastLineMargin
	if m.BufEOF() {
		if chunkNum, _ := chunkLineNum(lN - height); chunkNum > 0 {
			m.requestLoadSync(chunkNum)
		}
	}
	m.topLX, m.topLN = m.bottomLineNum(lN, height)
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
func (m *Document) movePgUp() bool {
	return m.moveLimitYUp(m.height)
}

// movePgDn moves down one screen.
func (m *Document) movePgDn() bool {
	return m.moveYDown(m.height)
}

// moveHfUp moves up half a screen.
func (m *Document) moveHfUp() bool {
	return m.moveLimitYUp(m.height / 2)
}

// moveHfDn moves down half a screen.
func (m *Document) moveHfDn() bool {
	return m.moveYDown(m.height / 2)
}

// limitMoveDown limits the movement of the cursor when moving down.
// Returns true if reached the bottom line.
func (m *Document) limitMoveDown(lX int, lN int) bool {
	if lN+m.height < m.BufEndNum()-m.SkipLines {
		m.topLX = lX
		m.topLN = lN
		return false
	}

	tX, tN := m.bottomLineNum(m.BufEndNum(), m.height-lastLineMargin)
	if lN < tN || (lN == tN && lX < tX) {
		m.topLX = lX
		m.topLN = lN
		return false
	}
	// move to bottom
	if m.topLN < tN || (m.topLN == tN && m.topLX < tX) {
		m.topLX = tX
		m.topLN = tN
	}
	return true
}

// numOfWrap returns the number of wrap from lX and lN.
func (m *Document) numOfWrap(lX int, lN int) int {
	listX := m.leftMostX(lN)
	if len(listX) == 0 {
		return 0
	}
	for n, v := range listX {
		if v >= lX {
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

// bottomLineNum returns the display start line
// when the last line number as an argument.
func (m *Document) bottomLineNum(lN int, height int) (int, int) {
	if lN < m.headerHeight {
		return 0, 0
	}

	bodyHeight := (height - (m.headerHeight))

	if !m.WrapMode {
		return 0, lN - bodyHeight - m.firstLine()
	}
	// WrapMode
	topLX, topLN := m.numUp(0, lN, bodyHeight)
	return topLX, topLN - m.firstLine()
}

// numUp finds lX, lN when the number of lines is moved up from lX, lN.
func (m *Document) numUp(lX int, lN int, upY int) (int, int) {
	listX := m.leftMostX(lN)
	n := numOfReverseSlice(listX, lX)
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

// moveYDown moves down by the specified number of y.
func (m *Document) moveYDown(moveY int) bool {
	if !m.WrapMode {
		return m.limitMoveDown(0, m.topLN+moveY)
	}

	// WrapMode
	if m.topLN < 0 {
		return m.limitMoveDown(m.topLX, m.topLN+1)
	}
	lN := m.topLN + m.firstLine()
	lX := m.topLX
	listX := m.leftMostX(lN)
	n := numOfReverseSlice(listX, lX)
	for y := 0; y <= moveY; y++ {
		if n >= len(listX) {
			lN++
			if lN > m.BufEndNum() {
				break
			}
			listX = m.leftMostX(lN)
			n = 0
		}
		lX = 0
		if len(listX) > 0 && n < len(listX) {
			lX = listX[n]
		}
		n++
	}
	return m.limitMoveDown(lX, lN-m.firstLine())
}

// limitMoveUp moves up by the specified number of y.
// The movement is limited to the top of the document.
func (m *Document) moveLimitYUp(moveY int) bool {
	m.moveYUp(moveY)
	if m.topLN < m.BufStartNum() {
		m.moveTop()
		return true
	}
	return false
}

// moveYUp moves up by the specified number of y.
func (m *Document) moveYUp(moveY int) {
	if !m.WrapMode {
		m.topLN -= moveY
		return
	}

	// WrapMode
	lX, lN := m.numUp(m.topLX, m.topLN+m.firstLine(), moveY)
	m.topLX = lX
	m.topLN = lN - m.firstLine()
}

// leftMostX returns a list of left - most x positions when wrapping.
// Returns nil if there is no line number.
func (m *Document) leftMostX(lN int) []int {
	lc, err := m.contents(lN)
	if err != nil {
		return nil
	}
	return leftX(m.width, lc)
}

// leftX returns a list of left - most x positions when wrapping.
func leftX(width int, lc contents) []int {
	if width <= 0 {
		return []int{0}
	}
	if len(lc) == 0 {
		return []int{0}
	}
	end := len(lc)
	listX := make([]int, 0, (end/width)+1)
	listX = append(listX, 0)
	for n := width; n < end; n += width {
		if lc.IsFullWidth(n - 1) {
			n--
		}
		listX = append(listX, n)
	}
	return listX
}

// nextSection returns the line number of the next section.
func (m *Document) nextSection(ctx context.Context, lN int) (int, error) {
	searcher := NewSearcher(m.SectionDelimiter, m.SectionDelimiterReg, true, true)
	return m.SearchLine(ctx, searcher, lN+1)
}

// prevSection returns the line number of the previous section.
func (m *Document) prevSection(ctx context.Context, lN int) (int, error) {
	searcher := NewSearcher(m.SectionDelimiter, m.SectionDelimiterReg, true, true)
	return m.BackSearchLine(ctx, searcher, lN-1)
}

// moveNextSection moves to the next section.
func (m *Document) moveNextSection(ctx context.Context) error {
	if m.SectionDelimiter == "" {
		return ErrNoDelimiter
	}

	lN, err := m.nextSection(ctx, m.topLN+m.firstLine()-m.SectionStartPosition)
	if err != nil {
		return ErrNoMoreSection
	}
	m.moveLine((lN - m.firstLine() + m.SectionStartPosition))
	return nil
}

// movePrevSection moves to the previous section.
func (m *Document) movePrevSection(ctx context.Context) error {
	return m.movePrevSectionLN(ctx, m.topLN)
}

// movePrevSectionLN moves to the previous section by line number.
func (m *Document) movePrevSectionLN(ctx context.Context, start int) error {
	if m.SectionDelimiter == "" {
		return ErrNoDelimiter
	}

	lN, err := m.prevSection(ctx, start+m.firstLine()-m.SectionStartPosition)
	if err != nil {
		return ErrNoMoreSection
	}
	lN = (lN - m.firstLine() + m.SectionStartPosition)
	lN = max(lN, m.BufStartNum())
	m.moveLine(lN)
	return nil
}

// moveLastSection moves to the last section.
func (m *Document) moveLastSection(ctx context.Context) {
	if err := m.movePrevSectionLN(ctx, m.BufEndNum()+1); err != nil {
		log.Printf("last section: %v\n", err)
	}
}
