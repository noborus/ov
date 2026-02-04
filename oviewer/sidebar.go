package oviewer

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// SidebarItem represents an item to display in the sidebar.
type SidebarItem struct {
	Contents  contents
	IsCurrent bool
}

type SidebarMode int

// sidebarMode represents the mode of the sidebar.
const (
	// SidebarModeNone is no sidebar.
	SidebarModeNone SidebarMode = iota
	// SidebarModeDocList is the document list sidebar.
	SidebarModeDocList
	// SidebarModeMark is the mark sidebar.
	SidebarModeMark
	// SidebarModeHelp is the help sidebar.
	SidebarModeHelp
)

// String returns the string representation of the SidebarMode.
func (s SidebarMode) String() string {
	switch s {
	case SidebarModeDocList:
		return "Files"
	case SidebarModeMark:
		return "Marks"
	case SidebarModeHelp:
		return "Help"
	default:
		return "none"
	}
}

const (
	minSidebarWidth     = 12
	maxSidebarWidth     = 100
	defaultSidebarWidth = 20
)

var defaultSidebarWidthString = strconv.Itoa(defaultSidebarWidth)

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
	case SidebarModeMark:
		items = root.sidebarItemsForMark()
	case SidebarModeDocList:
		items = root.sidebarItemsForDocList()
	case SidebarModeHelp:
		items = root.sidebarItemsForHelp()
	}
	root.SidebarItems = items
}

// sidebarItemsForMark creates SidebarItems for the mark sidebar.
func (root *Root) sidebarItemsForMark() []SidebarItem {
	var items []SidebarItem
	length := root.sidebarWidth - 4
	marks := root.Doc.marked
	current := root.Doc.markedPoint
	root.adjustSidebarScroll(SidebarModeMark, len(marks), current)
	scroll := root.sidebarScrolls[SidebarModeMark]
	start := scroll.y
	end := min(start+root.scr.vHeight, len(marks))
	for i := start; i < end; i++ {
		mark := marks[i]
		isCurrent := (i == current)
		lc := mark.contents.TrimLeft()
		if len(lc) < length {
			spaces := StrToContents(strings.Repeat(" ", length-len(lc)), 0)
			lc = append(lc, spaces...)
		}
		numContents := StrToContents(fmt.Sprintf("%2d %d ", i, mark.lineNum), 0)
		lc = append(numContents, lc...)
		items = append(items, SidebarItem{Contents: lc, IsCurrent: isCurrent})
	}
	return items
}

// sidebarItemsForDocList creates SidebarItems for the document list sidebar.
func (root *Root) sidebarItemsForDocList() []SidebarItem {
	var items []SidebarItem
	length := root.sidebarWidth - 5
	current := root.CurrentDoc
	root.adjustSidebarScroll(SidebarModeDocList, len(root.DocList), current)
	for i, doc := range root.DocList {
		text := fmt.Sprintf("%2d %s", i, doc.FileName)
		displayName := StrToContents(text, 0)
		if len(displayName) < length {
			spaces := StrToContents(strings.Repeat(" ", length-len(displayName)), 0)
			displayName = append(displayName, spaces...)
		}
		isCurrent := (i == current)
		items = append(items, SidebarItem{Contents: displayName, IsCurrent: isCurrent})
	}
	return items
}

// sidebarItemsForHelp creates SidebarItems for the help sidebar.
func (root *Root) sidebarItemsForHelp() []SidebarItem {
	if root.SidebarHelpItems != nil {
		return root.SidebarHelpItems
	}
	root.sidebarScrolls[SidebarModeHelp] = sidebarScroll{x: 0, y: 0, currentY: 0}
	var items []SidebarItem
	length := 100
	keyBinds := GetKeyBinds(root.Config)
	descriptions := keyBinds.GetKeyBindDescriptions(GroupAll)
	for _, desc := range descriptions {
		line := "[" + desc[1] + "]"
		content := StrToContents(line, 0)
		if len(content) < length {
			spaces := StrToContents(strings.Repeat(" ", length-len(content)), 0)
			content = append(content, spaces...)
		}
		items = append(items, SidebarItem{Contents: content, IsCurrent: false})

		contentDesc := StrToContents("  "+desc[0], 0)
		if len(contentDesc) < length {
			spaces := StrToContents(strings.Repeat(" ", length-len(contentDesc)), 0)
			contentDesc = append(contentDesc, spaces...)
		}
		items = append(items, SidebarItem{Contents: contentDesc, IsCurrent: false})
	}
	root.SidebarHelpItems = items
	return items
}

func (root *Root) sidebarUp(_ context.Context) {
	scroll := root.sidebarScrolls[root.sidebarMode]
	if scroll.y > 0 {
		scroll.y--
		root.sidebarScrolls[root.sidebarMode] = scroll
	}
}

func (root *Root) sidebarDown(_ context.Context) {
	scroll := root.sidebarScrolls[root.sidebarMode]
	scroll.y++
	root.sidebarScrolls[root.sidebarMode] = scroll
}

func (root *Root) sidebarLeft(_ context.Context) {
	scroll := root.sidebarScrolls[root.sidebarMode]
	scroll.x--
	scroll.x = max(scroll.x, 0)
	root.sidebarScrolls[root.sidebarMode] = scroll
}

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
