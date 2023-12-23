package oviewer

// searchGoTo moves to the specified line and position after searching.
// Go to the specified line +root.Doc.JumpTarget Go to.
// If the search term is off screen, move until the search term is visible.
func (m *Document) searchGoTo(lN int, x int) {
	m.searchGoX(x)

	m.topLN = lN - m.firstLine()
	m.moveYUp(m.jumpTargetNum)
}

// Bottom line of jumpTarget when specifying a section
func (m *Document) bottomJumpTarget() int {
	return m.statusPos - bottomMargin
}

// searchGoSection will go to the section with the matching term after searching.
// Move the JumpTarget so that it can be seen from the beginning of the section.
func (m *Document) searchGoSection(lN int, x int) {
	m.searchGoX(x)

	sN, err := m.prevSection(lN)
	if err != nil {
		sN = 0
	}
	if m.SectionHeader {
		sN = (sN - m.firstLine() + m.SectionHeaderNum) + m.SectionStartPosition
		sN = max(sN, m.BufStartNum())
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

	if m.bottomJumpTarget() > y+m.headerLen {
		m.jumpTargetNum = y
		return
	}

	m.jumpTargetNum = m.bottomJumpTarget() - (m.headerLen + 1)
	m.moveYDown(y - m.jumpTargetNum)
}

// searchGoX moves to the specified x position.
func (m *Document) searchGoX(x int) {
	m.topLX = 0
	// If the search term is outside the height of the screen when in WrapMode.
	if nTh := x / m.width; nTh > m.height {
		m.topLX = nTh * m.width
	}

	// If the search term is outside the width of the screen when in NoWrapMode.
	if x < m.x || x > m.x+m.width-1 {
		m.x = x
	}
}
