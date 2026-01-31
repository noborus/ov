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
	scrolls := root.sidebarScrolls[SidebarModeMark]
	scrollX := scrolls.X
	scrollY := scrolls.Y
	if scrolls.CurrentY != root.Doc.markedPoint {
		if root.Doc.markedPoint-scrollY < 0 {
			scrollY = root.Doc.markedPoint
		} else if root.Doc.markedPoint-scrollY >= root.scr.vHeight-4 {
			scrollY = root.Doc.markedPoint - root.scr.vHeight + 4 + 1
		}
		root.sidebarScrolls[SidebarModeMark] = SidebarScroll{X: scrollX, Y: scrollY, CurrentY: root.Doc.markedPoint}
	}
	marks := root.Doc.marked
	for i, mark := range marks {
		isCurrent := (i == root.Doc.markedPoint)
		lc := mark.contents.TrimLeft()
		if len(lc) < length {
			spaces := StrToContents(strings.Repeat(" ", length-len(lc)), 0)
			lc = append(lc, spaces...)
		}
		numContents := StrToContents(fmt.Sprintf("%2d ", mark.lineNum), 0)
		lc = append(numContents, lc...)
		items = append(items, SidebarItem{Contents: lc, IsCurrent: isCurrent})
	}
	return items
}

// sidebarItemsForDocList creates SidebarItems for the document list sidebar.
func (root *Root) sidebarItemsForDocList() []SidebarItem {
	var items []SidebarItem
	length := root.sidebarWidth - 5
	scrolls := root.sidebarScrolls[SidebarModeDocList]
	scrollX := scrolls.X
	scrollY := scrolls.Y
	if scrolls.CurrentY != root.CurrentDoc {
		if root.CurrentDoc-scrollY < 0 {
			scrollY = root.CurrentDoc
		} else if root.CurrentDoc-scrollY >= root.scr.vHeight-4 {
			scrollY = root.CurrentDoc - root.scr.vHeight - 4 + 1
		}
		root.sidebarScrolls[SidebarModeDocList] = SidebarScroll{X: scrollX, Y: scrollY, CurrentY: root.CurrentDoc}
	}
	for i, doc := range root.DocList {
		displayName := StrToContents(doc.FileName, 0)
		if len(displayName) < length {
			spaces := StrToContents(strings.Repeat(" ", length-len(displayName)), 0)
			displayName = append(displayName, spaces...)
		} else if len(displayName) > length {
			displayName = displayName[len(displayName)-(length):]
		}
		text := fmt.Sprintf("%2d %-20s", i, displayName)
		isCurrent := (i == root.CurrentDoc)
		content := StrToContents(text, 0)
		items = append(items, SidebarItem{Contents: content, IsCurrent: isCurrent})
	}
	return items
}

// sidebarItemsForHelp creates SidebarItems for the help sidebar.
func (root *Root) sidebarItemsForHelp() []SidebarItem {
	if root.SidebarHelpItems != nil {
		return root.SidebarHelpItems
	}
	root.sidebarScrolls[SidebarModeHelp] = SidebarScroll{X: 0, Y: 0, CurrentY: 0}
	var items []SidebarItem
	length := 100
	keyBinds := GetKeyBinds(root.Config)
	descriptions := keyBinds.GetKeyBindDescriptions(GroupAll)
	for _, desc := range descriptions {
		line := "[" + desc[1] + "]"
		content := StrToContents(line, 0)
		if len(content) > length {
			content = content[:length]
		} else {
			spaces := StrToContents(strings.Repeat(" ", length-len(content)), 0)
			content = append(content, spaces...)
		}
		items = append(items, SidebarItem{Contents: content, IsCurrent: false})

		contentDesc := StrToContents("  "+desc[0], 0)
		if len(contentDesc) > length {
			contentDesc = contentDesc[:length]
		} else {
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
	if scroll.Y > 0 {
		scroll.Y--
		root.sidebarScrolls[root.sidebarMode] = scroll
	}
}

func (root *Root) sidebarDown(_ context.Context) {
	scroll := root.sidebarScrolls[root.sidebarMode]
	scroll.Y++
	root.sidebarScrolls[root.sidebarMode] = scroll
}

func (root *Root) sidebarLeft(_ context.Context) {
	scroll := root.sidebarScrolls[root.sidebarMode]
	scroll.X--
	if scroll.X < 0 {
		scroll.X = 0
	}
	root.sidebarScrolls[root.sidebarMode] = scroll
}

func (root *Root) sidebarRight(_ context.Context) {
	scroll := root.sidebarScrolls[root.sidebarMode]
	scroll.X++
	root.sidebarScrolls[root.sidebarMode] = scroll
}
