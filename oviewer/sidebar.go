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
	marks := root.Doc.marked
	for i, mark := range marks {
		isCurrent := (i == root.Doc.markedPoint)
		items = append(items, SidebarItem{Contents: mark.content, IsCurrent: isCurrent})
	}
	return items
}

// sidebarItemsForDocList creates SidebarItems for the document list sidebar.
func (root *Root) sidebarItemsForDocList() []SidebarItem {
	var items []SidebarItem
	for i, doc := range root.DocList {
		displayName := doc.FileName
		text := fmt.Sprintf("%2d %s", i, displayName)
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
	keyBinds := GetKeyBinds(root.Config)
	for description, keys := range keyBinds {
		line := "[" + strings.Join(keys, " ") + "]"
		content := StrToContents(line, 0)
		items = append(items, SidebarItem{Contents: content, IsCurrent: false})
		contentDesc := StrToContents("  "+description, 0)
		items = append(items, SidebarItem{Contents: contentDesc, IsCurrent: false})
	}
	root.SidebarHelpItems = items
	return items
}
