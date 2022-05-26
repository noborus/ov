package oviewer

import (
	"fmt"
	"strings"

	"code.rocketnine.space/tslocum/cbind"
	"github.com/gdamore/tcell/v2"
)

const (
	actionExit           = "exit"
	actionWriteBA        = "set_write_exit"
	actionCancel         = "cancel"
	actionWriteExit      = "write_exit"
	actionSuspend        = "suspend"
	actionSync           = "sync"
	actionFollow         = "follow_mode"
	actionFollowAll      = "follow_all"
	actionFollowSection  = "follow_section"
	actionCloseFile      = "close_file"
	actionReload         = "reload"
	actionWatch          = "watch"
	actionWatchInterval  = "watch_interval"
	actionHelp           = "help"
	actionLogDoc         = "logdoc"
	actionMoveDown       = "down"
	actionMoveUp         = "up"
	actionMoveTop        = "top"
	actionMoveLeft       = "left"
	actionMoveRight      = "right"
	actionMoveHfLeft     = "half_left"
	actionMoveHfRight    = "half_right"
	actionMoveBottom     = "bottom"
	actionMovePgUp       = "page_up"
	actionMovePgDn       = "page_down"
	actionMoveHfUp       = "page_half_up"
	actionMoveHfDn       = "page_half_down"
	actionSection        = "section_delimiter"
	actionSectionStart   = "section_start"
	actionNextSection    = "next_section"
	actionLastSection    = "last_section"
	actionPrevSection    = "previous_section"
	actionMark           = "mark"
	actionRemoveMark     = "remove_mark"
	actionRemoveAllMark  = "remove_all_mark"
	actionMoveMark       = "next_mark"
	actionMovePrevMark   = "previous_mark"
	actionViewMode       = "set_view_mode"
	actionAlternate      = "alter_rows_mode"
	actionLineNumMode    = "line_number_mode"
	actionSearch         = "search"
	actionWrap           = "wrap_mode"
	actionColumnMode     = "column_mode"
	actionBackSearch     = "backsearch"
	actionDelimiter      = "delimiter"
	actionHeader         = "header"
	actionSkipLines      = "skip_lines"
	actionTabWidth       = "tabwidth"
	actionGoLine         = "goto"
	actionNextSearch     = "next_search"
	actionNextBackSearch = "next_backsearch"
	actionNextDoc        = "next_doc"
	actionPreviousDoc    = "previous_doc"
	actionCloseDoc       = "close_doc"
	actionToggleMouse    = "toggle_mouse"

	inputCaseSensitive = "input_casesensitive"
	inputIncSearch     = "input_incsearch"
	inputRegexpSearch  = "input_regexp_search"
)

func (root *Root) setHandler() map[string]func() {
	return map[string]func(){
		actionExit:           root.Quit,
		actionWriteBA:        root.setWriteBAMode,
		actionCancel:         root.Cancel,
		actionWriteExit:      root.WriteQuit,
		actionSuspend:        root.suspend,
		actionSync:           root.ViewSync,
		actionFollow:         root.toggleFollowMode,
		actionFollowAll:      root.toggleFollowAll,
		actionFollowSection:  root.toggleFollowSection,
		actionReload:         root.Reload,
		actionWatch:          root.toggleWatch,
		actionWatchInterval:  root.setWatchIntervalMode,
		actionCloseFile:      root.closeFile,
		actionHelp:           root.helpDisplay,
		actionLogDoc:         root.logDisplay,
		actionMoveDown:       root.moveDown,
		actionMoveUp:         root.moveUp,
		actionMoveTop:        root.moveTop,
		actionMoveBottom:     root.moveBottom,
		actionMovePgUp:       root.movePgUp,
		actionMovePgDn:       root.movePgDn,
		actionMoveHfUp:       root.moveHfUp,
		actionMoveHfDn:       root.moveHfDn,
		actionMoveLeft:       root.moveLeft,
		actionMoveRight:      root.moveRight,
		actionMoveHfLeft:     root.moveHfLeft,
		actionMoveHfRight:    root.moveHfRight,
		actionSection:        root.setSectionDelimiterMode,
		actionSectionStart:   root.setSectionStartMode,
		actionNextSection:    root.nextSction,
		actionPrevSection:    root.prevSection,
		actionLastSection:    root.lastSection,
		actionMoveMark:       root.markNext,
		actionMovePrevMark:   root.markPrev,
		actionViewMode:       root.setViewInputMode,
		actionWrap:           root.toggleWrapMode,
		actionColumnMode:     root.toggleColumnMode,
		actionAlternate:      root.toggleAlternateRows,
		actionLineNumMode:    root.toggleLineNumMode,
		actionMark:           root.addMark,
		actionRemoveMark:     root.removeMark,
		actionRemoveAllMark:  root.removeAllMark,
		actionSearch:         root.setSearchMode,
		actionBackSearch:     root.setBackSearchMode,
		actionDelimiter:      root.setDelimiterMode,
		actionHeader:         root.setHeaderMode,
		actionSkipLines:      root.setSkipLinesMode,
		actionTabWidth:       root.setTabWidthMode,
		actionGoLine:         root.setGoLineMode,
		actionNextSearch:     root.eventNextSearch,
		actionNextBackSearch: root.eventNextBackSearch,
		actionNextDoc:        root.nextDoc,
		actionPreviousDoc:    root.previousDoc,
		actionCloseDoc:       root.closeDocument,
		actionToggleMouse:    root.toggleMouse,
		inputCaseSensitive:   root.inputCaseSensitive,
		inputIncSearch:       root.inputIncSearch,
		inputRegexpSearch:    root.inputRegexpSearch,
	}
}

// KeyBind is the mapping of action and key.
type KeyBind map[string][]string

// GetKeyBinds returns the current key mapping.
func GetKeyBinds(bind map[string][]string) map[string][]string {
	keyBind := map[string][]string{
		actionExit:           {"Escape", "q"},
		actionWriteBA:        {"ctrl+q"},
		actionCancel:         {"ctrl+c"},
		actionWriteExit:      {"Q"},
		actionSync:           {"ctrl+l"},
		actionFollow:         {"ctrl+f"},
		actionFollowAll:      {"ctrl+a"},
		actionFollowSection:  {"F2"},
		actionCloseFile:      {"ctrl+F9", "ctrl+alt+s"},
		actionReload:         {"F5", "ctrl+alt+l"},
		actionWatch:          {"F4", "ctrl+alt+w"},
		actionWatchInterval:  {"ctrl+w"},
		actionHelp:           {"h", "ctrl+F1", "ctrl+alt+c"},
		actionLogDoc:         {"ctrl+F2", "ctrl+alt+e"},
		actionMoveDown:       {"Enter", "Down", "ctrl+N"},
		actionMoveUp:         {"Up", "ctrl+p"},
		actionMoveTop:        {"Home"},
		actionMoveBottom:     {"End"},
		actionMovePgUp:       {"PageUp", "ctrl+b"},
		actionMovePgDn:       {"PageDown", "ctrl+v"},
		actionMoveHfUp:       {"ctrl+u"},
		actionMoveHfDn:       {"ctrl+d"},
		actionMoveLeft:       {"left"},
		actionMoveRight:      {"right"},
		actionMoveHfLeft:     {"ctrl+left"},
		actionMoveHfRight:    {"ctrl+right"},
		actionSection:        {"alt+d"},
		actionSectionStart:   {"ctrl+F3", "alt+s"},
		actionNextSection:    {"space"},
		actionPrevSection:    {"^"},
		actionLastSection:    {"9"},
		actionMoveMark:       {">"},
		actionMovePrevMark:   {"<"},
		actionViewMode:       {"p", "P"},
		actionWrap:           {"w", "W"},
		actionColumnMode:     {"c"},
		actionAlternate:      {"C"},
		actionLineNumMode:    {"G"},
		actionMark:           {"m"},
		actionRemoveAllMark:  {"ctrl+delete"},
		actionRemoveMark:     {"M"},
		actionSearch:         {"/"},
		actionBackSearch:     {"?"},
		actionDelimiter:      {"d"},
		actionHeader:         {"H"},
		actionSkipLines:      {"ctrl+s"},
		actionTabWidth:       {"t"},
		actionGoLine:         {"g"},
		actionNextSearch:     {"n"},
		actionNextBackSearch: {"N"},
		actionNextDoc:        {"]"},
		actionPreviousDoc:    {"["},
		actionCloseDoc:       {"ctrl+k"},
		actionToggleMouse:    {"ctrl+alt+r"},
		actionSuspend:        {"ctrl+z"},

		inputCaseSensitive: {"alt+c"},
		inputIncSearch:     {"alt+i"},
		inputRegexpSearch:  {"alt+r"},
	}

	for k, v := range bind {
		keyBind[k] = v
	}

	return keyBind
}

func (root *Root) setKeyBind(keyBind map[string][]string) error {
	c := root.keyConfig
	in := root.inputKeyConfig

	actionHandlers := root.setHandler()

	for name, keys := range keyBind {
		handler := actionHandlers[name]
		if handler == nil {
			return fmt.Errorf("%w for [%s] unknown action", ErrFailedKeyBind, name)
		}

		if strings.HasPrefix(name, "input_") {
			if err := setHandler(in, name, keys, handler); err != nil {
				return err
			}
			continue
		}
		if err := setHandler(c, name, keys, handler); err != nil {
			return err
		}
	}
	return nil
}

func setHandler(c *cbind.Configuration, name string, keys []string, handler func()) error {
	for _, k := range keys {
		mod, key, ch, err := cbind.Decode(k)
		if err != nil {
			return fmt.Errorf("%w [%s] for %s: %s", ErrFailedKeyBind, k, name, err)
		}
		if key == tcell.KeyRune {
			c.SetRune(mod, ch, wrapEventHandler(handler))
			// Added "shift+N" instead of 'N' to get it on windows.
			if 'A' <= ch && ch <= 'Z' {
				c.SetRune(mod|tcell.ModShift, ch, wrapEventHandler(handler))
			}
		} else {
			c.SetKey(mod, key, wrapEventHandler(handler))
		}
	}
	return nil
}

func wrapEventHandler(f func()) func(_ *tcell.EventKey) *tcell.EventKey {
	return func(_ *tcell.EventKey) *tcell.EventKey {
		f()
		return nil
	}
}

func (root *Root) keyCapture(ev *tcell.EventKey) bool {
	root.keyConfig.Capture(ev)
	return true
}
