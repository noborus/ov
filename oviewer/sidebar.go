package oviewer

import (
	"context"
	"fmt"
	"strings"
)

// SidebarItem represents an item to display in the sidebar.
type SidebarItem struct {
	Label     string   // Always-visible label (e.g., index, description)
	Contents  contents // Contents to display
	IsCurrent bool
}

type SidebarMode int

// sidebarMode represents the mode of the sidebar.
const (
	// SidebarModeNone is no sidebar.
	SidebarModeNone SidebarMode = iota
	// SidebarModeHelp is the help sidebar.
	SidebarModeHelp
	// SidebarModeMarks is the mark list sidebar.
	SidebarModeMarks
	// SidebarModeDocList is the document list sidebar.
	SidebarModeDocList
	// SidebarModeSections is the section list sidebar.
	SidebarModeSections
)

// String returns the string representation of the SidebarMode.
func (s SidebarMode) String() string {
	switch s {
	case SidebarModeHelp:
		return "Help"
	case SidebarModeMarks:
		return "Marks"
	case SidebarModeDocList:
		return "Documents"
	case SidebarModeSections:
		return "Sections"
	default:
		return "none"
	}
}

const (
	minSidebarWidth     = 12
	maxSidebarWidth     = 100
	defaultSidebarWidth = "20%"
)

// sidebarScroll holds scroll positions for sidebar.
type sidebarScroll struct {
	x        int
	y        int
	currentY int // CurrentY is the current Y position.
}

// prepareSidebarItems creates the sidebarItems slice for display.
func (root *Root) prepareSidebarItems() {
	if root.sidebarWidth <= 0 {
		return
	}
	var items []SidebarItem
	switch root.sidebarMode {
	case SidebarModeHelp:
		items = root.sidebarItemsForHelp()
	case SidebarModeMarks:
		items = root.sidebarItemsForMarks()
	case SidebarModeDocList:
		items = root.sidebarItemsForDocList()
	case SidebarModeSections:
		items = root.sidebarItemsForSections()
	}
	root.SidebarItems = items
}

// sidebarItemsForHelp creates SidebarItems for the help sidebar.
func (root *Root) sidebarItemsForHelp() []SidebarItem {
	var items []SidebarItem
	keyBinds := GetKeyBinds(root.Config)
	descriptions := keyBinds.GetKeyBindDescriptions(GroupAll)
	totalLines := len(descriptions) * 2
	root.adjustSidebarScroll(SidebarModeHelp, totalLines, 0)
	scroll := root.sidebarScrolls[SidebarModeHelp]
	start := scroll.y
	end := min(start+root.scr.vHeight, totalLines)
	for line := start; line < end; line++ {
		i := line / 2
		desc := descriptions[i]
		if line%2 == 0 {
			content := StrToContents("["+desc[1]+"]", 0)
			items = append(items, SidebarItem{Label: "", Contents: content, IsCurrent: false})
		} else {
			contentDesc := StrToContents("  "+desc[0], 0)
			items = append(items, SidebarItem{Label: "", Contents: contentDesc, IsCurrent: false})
		}
	}
	root.SidebarHelpItems = items
	return items
}

// sidebarItemsForMarks creates SidebarItems for the mark sidebar.
func (root *Root) sidebarItemsForMarks() []SidebarItem {
	var items []SidebarItem
	length := root.sidebarWidth - 4
	marks := root.Doc.marked
	current := root.Doc.markedPoint
	root.adjustSidebarScroll(SidebarModeMarks, len(marks), current)
	scroll := root.sidebarScrolls[SidebarModeMarks]
	start := scroll.y
	end := min(start+root.scr.vHeight, len(marks))
	for i := start; i < end; i++ {
		mark := marks[i]
		isCurrent := (i == current)
		lc := mark.contents.TrimLeft()
		numContents := StrToContents(fmt.Sprintf("%d ", mark.lineNum), 0)
		lc = append(numContents, lc...)
		if len(lc) < length {
			spaces := StrToContents(strings.Repeat(" ", length-len(lc)), 0)
			lc = append(lc, spaces...)
		}
		label := fmt.Sprintf("%2d ", i)
		items = append(items, SidebarItem{
			Label:     label,
			Contents:  lc,
			IsCurrent: isCurrent,
		})
	}
	return items
}

// sidebarItemsForDocList creates SidebarItems for the document list sidebar.
func (root *Root) sidebarItemsForDocList() []SidebarItem {
	var items []SidebarItem
	length := root.sidebarWidth - 5
	current := root.CurrentDoc
	root.adjustSidebarScroll(SidebarModeDocList, len(root.DocList), current)
	scroll := root.sidebarScrolls[SidebarModeDocList]
	start := scroll.y
	end := min(start+root.scr.vHeight, len(root.DocList))
	for i := start; i < end; i++ {
		doc := root.DocList[i]
		displayName := StrToContents(doc.FileName, 0)
		if len(displayName) < length {
			spaces := StrToContents(strings.Repeat(" ", length-len(displayName)), 0)
			displayName = append(displayName, spaces...)
		}
		isCurrent := (i == current)
		label := fmt.Sprintf("%2d ", i)
		items = append(items, SidebarItem{
			Label:     label,
			Contents:  displayName,
			IsCurrent: isCurrent,
		})
	}
	return items
}

// sidebarItemsForSections returns SidebarItems for sectionList.
func (root *Root) sidebarItemsForSections() []SidebarItem {
	var items []SidebarItem
	length := root.sidebarWidth - 4
	sections := root.Doc.sectionList
	lN := root.Doc.topLN + root.Doc.firstLine()
	current := -1
	for i, section := range sections {
		if section.lineNum >= lN {
			current = i
			break
		}
	}
	root.adjustSidebarScroll(SidebarModeSections, len(sections), current)
	scroll := root.sidebarScrolls[SidebarModeSections]
	start := scroll.y
	end := min(start+root.scr.vHeight, len(sections))
	for i := start; i < end; i++ {
		section := sections[i]
		lc := section.contents.TrimLeft()
		numContents := StrToContents(fmt.Sprintf("%d ", section.lineNum), 0)
		lc = append(numContents, lc...)
		if len(lc) < length {
			spaces := StrToContents(strings.Repeat(" ", length-len(lc)), 0)
			lc = append(lc, spaces...)
		}
		label := fmt.Sprintf("%2d ", i)
		items = append(items, SidebarItem{
			Label:     label,
			Contents:  lc,
			IsCurrent: (i == current),
		})
	}
	return items
}

// sidebarUp scrolls the sidebar up.
func (root *Root) sidebarUp(_ context.Context) {
	scroll := root.sidebarScrolls[root.sidebarMode]
	if scroll.y > 0 {
		scroll.y--
		root.sidebarScrolls[root.sidebarMode] = scroll
	}
}

// sidebarDown scrolls the sidebar down.
func (root *Root) sidebarDown(_ context.Context) {
	scroll := root.sidebarScrolls[root.sidebarMode]
	scroll.y++
	root.sidebarScrolls[root.sidebarMode] = scroll
}

// sidebarLeft scrolls the sidebar left.
func (root *Root) sidebarLeft(_ context.Context) {
	scroll := root.sidebarScrolls[root.sidebarMode]
	scroll.x--
	scroll.x = max(scroll.x, 0)
	root.sidebarScrolls[root.sidebarMode] = scroll
}

// sidebarRight scrolls the sidebar right.
func (root *Root) sidebarRight(_ context.Context) {
	scroll := root.sidebarScrolls[root.sidebarMode]
	scroll.x++
	root.sidebarScrolls[root.sidebarMode] = scroll
}

// adjustSidebarScroll adjusts scrollY so that currentIndex is visible, only if currentIndex has changed.
func (root *Root) adjustSidebarScroll(mode SidebarMode, itemsLen, currentIndex int) {
	if root.sidebarScrolls == nil {
		return
	}
	scroll := root.sidebarScrolls[mode]
	height := root.scr.vHeight - 4
	scroll.y = max(scroll.y, 0)
	scroll.y = min(scroll.y, max(itemsLen-height, 0))
	if scroll.currentY == currentIndex {
		root.sidebarScrolls[mode] = scroll
		return
	}

	if currentIndex < scroll.y {
		scroll.y = currentIndex
	} else if currentIndex >= scroll.y+height {
		scroll.y = currentIndex - height + 1
	}
	scroll.y = max(scroll.y, 0)
	maxY := max(itemsLen-height, 0)
	scroll.y = min(scroll.y, maxY)
	scroll.currentY = currentIndex
	root.sidebarScrolls[mode] = scroll
}
