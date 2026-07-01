package oviewer

import "maps"

//go:generate go run ../tools/gen-keybind

// DefaultKeyBinds are the default keybindings.
func DefaultKeyBinds() KeyBind {
	return map[string][]string{
		actionExit:           {"Escape", "q"},
		actionCancel:         {"ctrl+c"},
		actionWriteExit:      {"Q"},
		actionSuspend:        {"ctrl+z"},
		actionEdit:           {"alt+v"},
		actionSync:           {"ctrl+l"},
		actionFollow:         {"ctrl+f"},
		actionFollowAll:      {"ctrl+a"},
		actionFollowSection:  {"F2"},
		actionPlain:          {"ctrl+e"},
		actionRainbow:        {"ctrl+r"},
		actionCloseFile:      {"ctrl+F9", "ctrl+alt+s"},
		actionReload:         {"F5", "ctrl+alt+l"},
		actionWatch:          {"F4", "ctrl+alt+w"},
		actionHelp:           {"h", "ctrl+F1", "ctrl+alt+c"},
		actionLogDoc:         {"ctrl+F2", "ctrl+alt+e"},
		actionMark:           {"m"},
		actionRemoveMark:     {"M"},
		actionRemoveAllMark:  {"ctrl+delete"},
		actionAlternate:      {"C"},
		actionLineNumMode:    {"G"},
		actionWrap:           {"w", "W"},
		actionWordWrap:       {"alt+w"},
		actionColumnMode:     {"c"},
		actionColumnWidth:    {"alt+o"},
		actionNextSearch:     {"n"},
		actionNextBackSearch: {"N"},
		actionNextDoc:        {"]"},
		actionPreviousDoc:    {"["},
		actionCloseDoc:       {"ctrl+k"},
		actionCloseAllFilter: {"K"},
		actionToggleMouse:    {"ctrl+F8", "ctrl+alt+r"},
		actionHideOther:      {"alt+-"},
		actionAlignFormat:    {"alt+f"},
		actionRawFormat:      {"alt+r"},
		actionFixedColumn:    {"F"},
		actionShrinkColumn:   {"s"},
		actionRightAlign:     {"alt+a"},
		actionRuler:          {"alt+shift+F9"},
		actionWriteOriginal:  {"alt+shift+F8"},
		actionStatusLine:     {"ctrl+F10"},

		// Move actions.
		actionMoveDown:       {"Enter", "Down", "ctrl+n"},
		actionMoveUp:         {"Up", "ctrl+p"},
		actionMoveTop:        {"Home"},
		actionMoveWidthLeft:  {"alt+left"},
		actionMoveWidthRight: {"alt+right"},
		actionMoveLeft:       {"left"},
		actionMoveRight:      {"right"},
		actionMoveHfLeft:     {"ctrl+left"},
		actionMoveHfRight:    {"ctrl+right"},
		actionMoveBeginLeft:  {"shift+Home"},
		actionMoveEndRight:   {"shift+End"},
		actionMoveBottom:     {"End"},
		actionMovePgUp:       {"PageUp", "ctrl+b"},
		actionMovePgDn:       {"PageDown", "ctrl+v"},
		actionMoveHfUp:       {"ctrl+u"},
		actionMoveHfDn:       {"ctrl+d"},
		actionNextSection:    {"space"},
		actionLastSection:    {"9"},
		actionPrevSection:    {"^"},
		actionMoveMark:       {">"},
		actionMovePrevMark:   {"<"},

		// sidebar actions.
		actionSidebarHelp:     {"alt+h"},
		actionSidebarMarks:    {"alt+m"},
		actionSidebarDocList:  {"alt+l"},
		actionSidebarSections: {"alt+u"},
		actionSidebarStyles:   {"alt+i"},
		actionSidebarUp:       {"shift+Up"},
		actionSidebarDown:     {"shift+Down"},
		actionSidebarLeft:     {"shift+Left"},
		actionSidebarRight:    {"shift+Right"},

		// Actions that enter input mode.
		actionConvertType:    {"alt+t"},
		actionDelimiter:      {"d"},
		actionGoLine:         {"g"},
		actionHeaderColumn:   {"Y"},
		actionHeader:         {"H"},
		actionJumpTarget:     {"j"},
		actionMultiColor:     {"."},
		actionSaveBuffer:     {"S"},
		actionSearch:         {"/"},
		actionBackSearch:     {"?"},
		actionFilter:         {"&"},
		actionSection:        {"alt+d"},
		actionSectionNum:     {"F7"},
		actionSectionStart:   {"ctrl+F3", "alt+s"},
		actionSkipLines:      {"ctrl+s"},
		actionTabWidth:       {"t"},
		actionVerticalHeader: {"y"},
		actionViewMode:       {"p", "P"},
		actionWatchInterval:  {"ctrl+w"},
		actionWriteBA:        {"ctrl+q"},
		actionMarkNumber:     {","},
		actionMarkByPattern:  {"*"},
		actionStyleToggle:    {"o"},

		// input actions.
		inputCaseSensitive:      {"alt+c"},
		inputSmartCaseSensitive: {"alt+s"},
		inputIncSearch:          {"alt+i"},
		inputRegexpSearch:       {"alt+r"},
		inputNonMatch:           {"!"},
		inputPrevious:           {"Up"},
		inputNext:               {"Down"},
		inputCopy:               {"ctrl+c"},
		inputPaste:              {"ctrl+v"},
	}
}

// LessKeyBinds are keybindings preset for less-like operations.
func LessKeyBinds() KeyBind {
	keyBind := DefaultKeyBinds()

	maps.Copy(keyBind, KeyBind{
		actionEdit:           {"v"},
		actionSync:           {"r", "ctrl+l"},
		actionReload:         {"R", "ctrl+r"},
		actionWatch:          {"T", "ctrl+alt+w"},
		actionFollow:         {"F"},
		actionHelp:           {"h", "ctrl+alt+c"},
		actionLogDoc:         {"ctrl+alt+e"},
		actionLineNumMode:    {"alt+n"},
		actionAlignFormat:    {"ctrl+alt+f"},
		actionRawFormat:      {"ctrl+alt+g"},
		actionFixedColumn:    {"alt+f"},
		actionShrinkColumn:   {"alt+x"},
		actionPlain:          {"ctrl+F7"},
		actionRainbow:        {"ctrl+F4"},
		actionCloseDoc:       {"alt+k"},
		actionCloseAllFilter: {"ctrl+alt+k"},

		// Move actions.
		actionMoveDown:     {"e", "ctrl+e", "j", "J", "ctrl+j", "Enter", "Down"},
		actionMoveUp:       {"y", "Y", "ctrl+y", "k", "K", "ctrl+k", "Up"},
		actionMoveTop:      {"Home", "g", "<"},
		actionMoveBottom:   {"End", ">", "G"},
		actionMovePgUp:     {"PageUp", "b", "alt+v"},
		actionMovePgDn:     {"PageDown", "ctrl+v", "alt+space", "f", "z"},
		actionMoveHfUp:     {"u", "ctrl+u"},
		actionMoveHfDn:     {"d", "ctrl+d"},
		actionMoveMark:     {"alt+>"},
		actionMovePrevMark: {"alt+<"},

		// Actions that enter input mode.
		actionDelimiter:      {"F8"},
		actionGoLine:         {":"},
		actionHeaderColumn:   {"ctrl+alt+d"},
		actionJumpTarget:     {"alt+j"},
		actionSaveBuffer:     {"s"},
		actionConvertType:    {"ctrl+alt+t"},
		actionVerticalHeader: {"ctrl+alt+b"},
	})

	return keyBind
}

// EmacsKeyBinds are keybindings preset for Emacs-like operations.
func EmacsKeyBinds() KeyBind {
	keyBind := DefaultKeyBinds()

	maps.Copy(keyBind, KeyBind{
		actionCancel:    {"ctrl+c", "ctrl+g"},
		actionEdit:      {"v"},
		actionFollow:    {"alt+ctrl+f"},
		actionFollowAll: {"alt+ctrl+a"},

		// Move actions.
		actionMoveLeft:    {"left", "ctrl+b"},
		actionMoveRight:   {"right", "ctrl+f"},
		actionMovePgUp:    {"PageUp", "alt+v"},
		actionMovePgDn:    {"PageDown", "ctrl+v", "space"},
		actionMoveTop:     {"Home", "alt+shift+,", "alt+<"},
		actionMoveBottom:  {"End", "alt+shift+.", "alt+>"},
		actionNextSection: {"ctrl+F6"},

		// Actions that enter input mode.
		actionSearch:    {"/", "ctrl+s"},
		actionSkipLines: {"ctrl+F5"},
	})

	return keyBind
}
