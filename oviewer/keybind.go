package oviewer

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cbind"
)

const (
	actionExit           = "Exit"
	actionWriteExit      = "Write exit"
	actionSync           = "Sync"
	actionHelp           = "Help"
	actionMoveDown       = "Down"
	actionMoveUp         = "Up"
	actionMoveTop        = "Top"
	actionMoveLeft       = "Left"
	actionMoveRight      = "Right"
	actionMoveHfLeft     = "Half left"
	actionMoveHfRight    = "Half right"
	actionMoveBottom     = "Bottom"
	actionMovePgUp       = "Page up"
	actionMovePgDn       = "Page down"
	actionMoveHfUp       = "Page half up"
	actionMoveHfDn       = "Page half down"
	actionMark           = "Mark"
	actionMoveMark       = "Move to next mark"
	actionMovePrevMark   = "Move to previous mark"
	actionAlternate      = "Toggle alternating rows mode"
	actionLineNumMode    = "Toggle line number mode"
	actionSearch         = "Search"
	actionWrap           = "Toggle wrap/nowrap mode"
	actionColumnMode     = "Toggle column mode"
	actionBackSearch     = "Backsearch"
	actionDelimiter      = "Input delimiter"
	actionHeader         = "Input header len"
	actionTabWidth       = "Input tabwidth"
	actionGoLine         = "Input line number to move"
	actionNextSearch     = "Next search"
	actionNextBackSearch = "Next backsearch"
)

func (root *Root) setHandler() map[string]func() {
	return map[string]func(){
		actionExit:           root.Quit,
		actionWriteExit:      root.WriteQuit,
		actionSync:           root.viewSync,
		actionHelp:           root.Help,
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
		actionWrap:           root.toggleWrapMode,
		actionColumnMode:     root.toggleColumnMode,
		actionAlternate:      root.toggleAlternateRows,
		actionLineNumMode:    root.toggleLineNumMode,
		actionMark:           root.markLineNum,
		actionSearch:         root.setSearchMode,
		actionBackSearch:     root.setBackSearchMode,
		actionDelimiter:      root.setDelimiterMode,
		actionHeader:         root.setHeaderMode,
		actionTabWidth:       root.setTabWidthMode,
		actionGoLine:         root.setGoLineMode,
		actionNextSearch:     root.nextSearch,
		actionNextBackSearch: root.nextBackSearch,
	}
}

type KeyBind map[string][]string

func SetDefaultKeyBinds() map[string][]string {
	return map[string][]string{
		actionExit:           {"Escape", "q", "ctrl+c"},
		actionWriteExit:      {"Q"},
		actionSync:           {"ctrl+l"},
		actionHelp:           {"h"},
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
		actionWrap:           {"w", "W"},
		actionColumnMode:     {"c"},
		actionAlternate:      {"C"},
		actionLineNumMode:    {"G"},
		actionMark:           {"m"},
		actionSearch:         {"/"},
		actionBackSearch:     {"?"},
		actionDelimiter:      {"d"},
		actionHeader:         {"H"},
		actionTabWidth:       {"t"},
		actionGoLine:         {"g"},
		actionNextSearch:     {"n"},
		actionNextBackSearch: {"N"},
	}
}

func (root *Root) setKeyBind(keyBind map[string][]string) error {
	c := root.keyConfig

	actionHandlers := root.setHandler()

	for a, keys := range keyBind {
		handler := actionHandlers[a]
		if handler == nil {
			return fmt.Errorf("failed to set keybind for %s: unknown action", a)
		}
		for _, k := range keys {
			mod, key, ch, err := cbind.Decode(k)
			if err != nil {
				return fmt.Errorf("failed to set keybind %s for %s: %s", k, a, err)
			}

			if key == tcell.KeyRune {
				c.SetRune(mod, ch, wrapEventHandler(handler))
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

func KeyBindString(k KeyBind) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "\n\tKey binding\n\n")
	k.writeKeyBind(&b, actionExit)
	k.writeKeyBind(&b, actionWriteExit)
	k.writeKeyBind(&b, actionHelp)
	k.writeKeyBind(&b, actionSync)

	fmt.Fprintf(&b, "\n\tMoving\n\n")
	k.writeKeyBind(&b, actionMoveDown)
	k.writeKeyBind(&b, actionMoveUp)
	k.writeKeyBind(&b, actionMoveTop)
	k.writeKeyBind(&b, actionMoveBottom)
	k.writeKeyBind(&b, actionMovePgUp)
	k.writeKeyBind(&b, actionMovePgDn)
	k.writeKeyBind(&b, actionMoveHfUp)
	k.writeKeyBind(&b, actionMoveHfDn)
	k.writeKeyBind(&b, actionMoveLeft)
	k.writeKeyBind(&b, actionMoveRight)
	k.writeKeyBind(&b, actionMoveHfLeft)
	k.writeKeyBind(&b, actionMoveHfRight)
	k.writeKeyBind(&b, actionGoLine)

	fmt.Fprintf(&b, "\n\tMark position\n\n")
	k.writeKeyBind(&b, actionMark)
	k.writeKeyBind(&b, actionMoveMark)
	k.writeKeyBind(&b, actionMovePrevMark)

	fmt.Fprintf(&b, "\n\tSearch\n\n")
	k.writeKeyBind(&b, actionSearch)
	k.writeKeyBind(&b, actionBackSearch)
	k.writeKeyBind(&b, actionNextSearch)
	k.writeKeyBind(&b, actionNextBackSearch)

	fmt.Fprintf(&b, "\n\tChange display\n\n")
	k.writeKeyBind(&b, actionWrap)
	k.writeKeyBind(&b, actionColumnMode)
	k.writeKeyBind(&b, actionAlternate)
	k.writeKeyBind(&b, actionLineNumMode)

	fmt.Fprintf(&b, "\n\tChange Display with Input\n\n")
	k.writeKeyBind(&b, actionDelimiter)
	k.writeKeyBind(&b, actionHeader)
	k.writeKeyBind(&b, actionTabWidth)

	return b.String()
}

func (k KeyBind) writeKeyBind(w io.Writer, action string) {
	fmt.Fprintf(w, "  %-26s: %s\n", "["+strings.Join(k[action], "], [")+"]", action)
}
