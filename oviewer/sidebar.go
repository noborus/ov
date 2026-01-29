package oviewer

import (
	"fmt"
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

// prepareSidebarItems creates the sidebarItems slice for display.
func (root *Root) prepareSidebarItems() {
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
	var items []SidebarItem
	length := root.sidebarWidth - 2
	keyBinds := GetKeyBinds(root.Config)
	descriptions := keyBinds.GetKeyBindDescriptions("")
	for _, desc := range descriptions {
		line := "[" + desc[1] + "]"
		content := StrToContents(line, 0)
		if len(content) > length {
			content = content[:length]
		}
		items = append(items, SidebarItem{Contents: content, IsCurrent: false})
		contentDesc := StrToContents("  "+desc[0], 0)
		if len(contentDesc) > length {
			contentDesc = contentDesc[:length]
		}
		items = append(items, SidebarItem{Contents: contentDesc, IsCurrent: false})
	}
	root.SidebarHelpItems = items
	return items
}
