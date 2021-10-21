package oviewer

import (
	"fmt"
	"io"
	"strings"

	"code.rocketnine.space/tslocum/cbind"
	"github.com/gdamore/tcell/v2"
)

const (
	actionExit           = "exit"
	actionCancel         = "cancel"
	actionWriteExit      = "write_exit"
	actionSync           = "sync"
	actionFollow         = "follow_mode"
	actionFollowAll      = "follow_all"
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
	actionMark           = "mark"
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
)

func (root *Root) setHandler() map[string]func() {
	return map[string]func(){
		actionExit:           root.Quit,
		actionCancel:         root.Cancel,
		actionWriteExit:      root.WriteQuit,
		actionSync:           root.ViewSync,
		actionFollow:         root.toggleFollowMode,
		actionFollowAll:      root.toggleFollowAll,
		actionHelp:           root.Help,
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
		actionMoveMark:       root.markNext,
		actionMovePrevMark:   root.markPrev,
		actionViewMode:       root.setViewInputMode,
		actionWrap:           root.toggleWrapMode,
		actionColumnMode:     root.toggleColumnMode,
		actionAlternate:      root.toggleAlternateRows,
		actionLineNumMode:    root.toggleLineNumMode,
		actionMark:           root.markLineNum,
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
	}
}

// KeyBind is the mapping of action and key.
type KeyBind map[string][]string

// GetKeyBinds returns the current key mapping.
func GetKeyBinds(bind map[string][]string) map[string][]string {
	keyBind := map[string][]string{
		actionExit:           {"Escape", "q"},
		actionCancel:         {"ctrl+c"},
		actionWriteExit:      {"Q"},
		actionSync:           {"ctrl+l"},
		actionFollow:         {"ctrl+f"},
		actionFollowAll:      {"ctrl+a"},
		actionHelp:           {"h"},
		actionLogDoc:         {"ctrl+alt+e"},
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
		actionMoveMark:       {">"},
		actionMovePrevMark:   {"<"},
		actionViewMode:       {"p", "P"},
		actionWrap:           {"w", "W"},
		actionColumnMode:     {"c"},
		actionAlternate:      {"C"},
		actionLineNumMode:    {"G"},
		actionMark:           {"m"},
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
	}

	for k, v := range bind {
		keyBind[k] = v
	}

	return keyBind
}

func (root *Root) setKeyBind(keyBind map[string][]string) error {
	c := root.keyConfig

	actionHandlers := root.setHandler()

	for a, keys := range keyBind {
		handler := actionHandlers[a]
		if handler == nil {
			return fmt.Errorf("%w for [%s] unknown action", ErrFailedKeyBind, a)
		}
		for _, k := range keys {
			mod, key, ch, err := cbind.Decode(k)
			if err != nil {
				return fmt.Errorf("%w [%s] for %s: %s", ErrFailedKeyBind, k, a, err)
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

// KeyBindString returns keybind as a string for help.
func KeyBindString(k KeyBind) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n\tKey binding\n\n")
	k.writeKeyBind(&b, actionExit, "quit")
	k.writeKeyBind(&b, actionCancel, "cancel")
	k.writeKeyBind(&b, actionWriteExit, "output screen and quit")
	k.writeKeyBind(&b, actionHelp, "display help screen")
	k.writeKeyBind(&b, actionLogDoc, "display log screen")
	k.writeKeyBind(&b, actionSync, "screen sync")
	k.writeKeyBind(&b, actionFollow, "follow mode toggle")
	k.writeKeyBind(&b, actionFollowAll, "follow all mode toggle")
	k.writeKeyBind(&b, actionToggleMouse, "enable/disable mouse")
	k.writeKeyBind(&b, actionCloseDoc, "close current document")

	fmt.Fprintf(&b, "\n\tMoving\n\n")
	k.writeKeyBind(&b, actionMoveDown, "forward by one line")
	k.writeKeyBind(&b, actionMoveUp, "backward by one line")
	k.writeKeyBind(&b, actionMoveTop, "go to begin of line")
	k.writeKeyBind(&b, actionMoveBottom, "go to end of line")
	k.writeKeyBind(&b, actionMovePgDn, "forward by page")
	k.writeKeyBind(&b, actionMovePgUp, "backward by page")
	k.writeKeyBind(&b, actionMoveHfDn, "forward a half page")
	k.writeKeyBind(&b, actionMoveHfUp, "backward a half page")
	k.writeKeyBind(&b, actionMoveLeft, "scroll to left")
	k.writeKeyBind(&b, actionMoveRight, "scroll to right")
	k.writeKeyBind(&b, actionMoveHfLeft, "scroll left half screen")
	k.writeKeyBind(&b, actionMoveHfRight, "scroll right half screen")
	k.writeKeyBind(&b, actionGoLine, "number of go to line")
	k.writeKeyBind(&b, actionNextDoc, "next document")
	k.writeKeyBind(&b, actionPreviousDoc, "previous document")

	fmt.Fprintf(&b, "\n\tMark position\n\n")
	k.writeKeyBind(&b, actionMark, "mark current position")
	k.writeKeyBind(&b, actionMoveMark, "move to next marked position")
	k.writeKeyBind(&b, actionMovePrevMark, "move to previous marked position")

	fmt.Fprintf(&b, "\n\tSearch\n\n")
	k.writeKeyBind(&b, actionSearch, "forward search mode")
	k.writeKeyBind(&b, actionBackSearch, "backward search mode")
	k.writeKeyBind(&b, actionNextSearch, "repeat forward search")
	k.writeKeyBind(&b, actionNextBackSearch, "repeat backward search")

	fmt.Fprintf(&b, "\n\tChange display\n\n")
	k.writeKeyBind(&b, actionWrap, "wrap/nowrap toggle")
	k.writeKeyBind(&b, actionColumnMode, "column mode toggle")
	k.writeKeyBind(&b, actionAlternate, "color to alternate rows toggle")
	k.writeKeyBind(&b, actionLineNumMode, "line number toggle")

	fmt.Fprintf(&b, "\n\tChange Display with Input\n\n")
	k.writeKeyBind(&b, actionViewMode, "view mode selection")
	k.writeKeyBind(&b, actionDelimiter, "delimiter string")
	k.writeKeyBind(&b, actionHeader, "number of header lines")
	k.writeKeyBind(&b, actionSkipLines, "number of skip lines")
	k.writeKeyBind(&b, actionTabWidth, "TAB width")

	return b.String()
}

func (k KeyBind) writeKeyBind(w io.Writer, action string, detail string) {
	fmt.Fprintf(w, "  %-26s * %s\n", "["+strings.Join(k[action], "], [")+"]", detail)
}
