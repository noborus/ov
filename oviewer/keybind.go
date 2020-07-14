package oviewer

import (
	"fmt"

	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cbind"
)

const (
	actionExit           = "exit"
	actionWriteExit      = "write exit"
	actionMoveDown       = "Down"
	actionSync           = "sync"
	actionMoveUp         = "UP"
	actionMoveTop        = "Top"
	actionMoveLeft       = "Left"
	actionMoveRight      = "Right"
	actionMoveHfLeft     = "halfLeft"
	actionMoveHfRight    = "halfRight"
	actionMoveBottom     = "Bottom"
	actionMovePgUp       = "movePgUp"
	actionMovePgDn       = "movePgDn"
	actionMoveHfUp       = "moveHfUp"
	actionMoveHfDn       = "moveHfDn"
	actionMoveMark       = "move Next Mark"
	actionMovePrevMark   = "move Prev Mark"
	actionWrap           = "ToggleWrapMode"
	actionColumnMode     = "ToggleColumnMode"
	actionAlternate      = "AlternateRows"
	actionMark           = "Mark"
	actionLineNumMode    = "LineNumMode"
	actionSearch         = "SearchMode"
	actionBackSearch     = "BackSearchMode"
	actionDelimiter      = "DelimiterMode"
	actionHeader         = "HeaderMode"
	actionTabWidth       = "TabWidthMode"
	actionGoLine         = "GoLineMode"
	actionNextSearch     = "NextSearch"
	actionNextBackSearch = "BackSearch"
)

func (root *Root) setHandler() map[string]func() {
	return map[string]func(){
		actionExit:           root.Quit,
		actionWriteExit:      root.WriteQuit,
		actionSync:           root.viewSync,
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
		actionSearch:         root.SetSearchMode,
		actionBackSearch:     root.SetBackSearchMode,
		actionDelimiter:      root.SetDelimiter,
		actionHeader:         root.SetHeader,
		actionTabWidth:       root.SetTabWidth,
		actionGoLine:         root.SetGoLine,
		actionNextSearch:     root.nextSearch,
		actionNextBackSearch: root.nextBackSearch,
	}
}

func (root *Root) setDefaultKeyBinds() map[string][]string {
	return map[string][]string{
		actionExit:           {"Escape", "q", "ctrl+c"},
		actionSync:           {"ctrl+l"},
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

func (root *Root) setKeyBind() error {
	c := root.KeyConfig

	actionHandlers := root.setHandler()
	keyBind := root.setDefaultKeyBinds()

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

func (root *Root) KeyCapture(ev *tcell.EventKey) bool {
	root.KeyConfig.Capture(ev)
	return true
}
