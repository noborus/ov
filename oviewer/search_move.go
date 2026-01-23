package oviewer

import (
	"context"
)

// searchGoTo moves to the specified line and position after searching.
// Go to the specified line +root.Doc.JumpTarget Go to.
// If the search term is off screen, move until the search term is visible.
func (m *Document) searchGoTo(lN int, start int, end int) {
	m.searchGoX(start, end)
	m.showGotoF = true
	m.topLN = lN - m.firstLine()
	m.moveYUp(m.jumpTargetHeight)
}

// Bottom line of jumpTarget when specifying a section.
func (m *Document) bottomJumpTarget() int {
	return m.statusPos - bottomMargin
}

// searchGoSection will go to the section with the matching term after searching.
// Move the JumpTarget so that it can be seen from the beginning of the section.
func (m *Document) searchGoSection(ctx context.Context, lN int, start int, end int) {
	m.searchGoX(start, end)

	sN, err := m.prevSection(ctx, lN)
	if err != nil {
		sN = 0
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

	if m.bottomJumpTarget() > y+m.headerHeight {
		m.jumpTargetHeight = y
		return
	}

	m.jumpTargetHeight = m.bottomJumpTarget() - (m.headerHeight + 1)
	m.moveYDown(y - m.jumpTargetHeight)
}

// searchGoX moves to the specified x position.
func (m *Document) searchGoX(start int, end int) {
	if m.bodyWidth == 0 {
		return
	}
	m.topLX = 0
	// If the search term is outside the height of the screen when in WrapMode.
	if nTh := start / m.bodyWidth; nTh > m.height {
		m.topLX = nTh * m.bodyWidth
	}

	// If the search term is outside the width of the screen when in NoWrapMode.
	if start < m.scrollX {
		m.scrollX = 0
	}
	if end > m.scrollX+m.bodyWidth-1 {
		m.scrollX = max(end-(m.bodyWidth-columnMargin), 0)
	}
}
