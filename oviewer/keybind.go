package oviewer

import (
	"context"
	"fmt"
	"io"
	"strings"

	"code.rocketnine.space/tslocum/cbind"
	"github.com/gdamore/tcell/v2"
)

// The name of the action to assign the key to.
// The string is displayed in help.
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
	actionPlain          = "plain_mode"
	actionRainbow        = "rainbow_mode"
	actionCloseFile      = "close_file"
	actionReload         = "reload"
	actionWatch          = "watch"
	actionWatchInterval  = "watch_interval"
	actionHelp           = "help"
	actionLogDoc         = "logdoc"
	actionMoveDown       = "down"
	actionMoveUp         = "up"
	actionMoveTop        = "top"
	actionMoveWidthLeft  = "width_left"
	actionMoveWidthRight = "width_right"
	actionMoveLeft       = "left"
	actionMoveRight      = "right"
	actionMoveHfLeft     = "half_left"
	actionMoveHfRight    = "half_right"
	actionMoveBeginLeft  = "begin_left"
	actionMoveEndRight   = "end_right"
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
	actionFilter         = "filter"
	actionWrap           = "wrap_mode"
	actionColumnMode     = "column_mode"
	actionColumnWidth    = "column_width"
	actionBackSearch     = "backsearch"
	actionDelimiter      = "delimiter"
	actionHeader         = "header"
	actionSkipLines      = "skip_lines"
	actionTabWidth       = "tabwidth"
	actionGoLine         = "goto"
	actionSectionNum     = "section_header_num"
	actionNextSearch     = "next_search"
	actionNextBackSearch = "next_backsearch"
	actionNextDoc        = "next_doc"
	actionPreviousDoc    = "previous_doc"
	actionCloseDoc       = "close_doc"
	actionCloseAllFilter = "close_all_filter"
	actionToggleMouse    = "toggle_mouse"
	actionMultiColor     = "multi_color"
	actionJumpTarget     = "jump_target"
	actionSaveBuffer     = "save_buffer"
	actionHideOther      = "hide_other"

	inputCaseSensitive      = "input_casesensitive"
	inputSmartCaseSensitive = "input_smart_casesensitive"
	inputIncSearch          = "input_incsearch"
	inputRegexpSearch       = "input_regexp_search"
	inputNonMatch           = "input_non_match"
	inputPrevious           = "input_previous"
	inputNext               = "input_next"
	inputCopy               = "input_copy"
	inputPaste              = "input_paste"
)

// handlers returns a map of the action's handlers.
func (root *Root) handlers() map[string]func(context.Context) {
	return map[string]func(context.Context){
		actionExit:           root.Quit,
		actionWriteBA:        root.setWriteBAMode,
		actionCancel:         root.Cancel,
		actionWriteExit:      root.WriteQuit,
		actionSuspend:        root.suspend,
		actionSync:           root.ViewSync,
		actionFollow:         root.toggleFollowMode,
		actionFollowAll:      root.toggleFollowAll,
		actionFollowSection:  root.toggleFollowSection,
		actionPlain:          root.togglePlain,
		actionRainbow:        root.toggleRainbow,
		actionReload:         root.Reload,
		actionWatch:          root.toggleWatch,
		actionWatchInterval:  root.setWatchIntervalMode,
		actionCloseFile:      root.closeFile,
		actionHelp:           root.helpDisplay,
		actionLogDoc:         root.logDisplay,
		actionMoveDown:       root.moveDownOne,
		actionMoveUp:         root.moveUpOne,
		actionMoveTop:        root.moveTop,
		actionMoveBottom:     root.moveBottom,
		actionMovePgUp:       root.movePgUp,
		actionMovePgDn:       root.movePgDn,
		actionMoveHfUp:       root.moveHfUp,
		actionMoveHfDn:       root.moveHfDn,
		actionMoveWidthLeft:  root.moveWidthLeft,
		actionMoveWidthRight: root.moveWidthRight,
		actionMoveLeft:       root.moveLeftOne,
		actionMoveRight:      root.moveRightOne,
		actionMoveHfLeft:     root.moveHfLeft,
		actionMoveHfRight:    root.moveHfRight,
		actionMoveBeginLeft:  root.moveBeginLeft,
		actionMoveEndRight:   root.moveEndRight,
		actionSection:        root.setSectionDelimiterMode,
		actionSectionStart:   root.setSectionStartMode,
		actionNextSection:    root.nextSection,
		actionPrevSection:    root.prevSection,
		actionLastSection:    root.lastSection,
		actionMoveMark:       root.markNext,
		actionMovePrevMark:   root.markPrev,
		actionViewMode:       root.setViewInputMode,
		actionWrap:           root.toggleWrapMode,
		actionColumnMode:     root.toggleColumnMode,
		actionColumnWidth:    root.toggleColumnWidth,
		actionAlternate:      root.toggleAlternateRows,
		actionLineNumMode:    root.toggleLineNumMode,
		actionMark:           root.addMark,
		actionRemoveMark:     root.removeMark,
		actionRemoveAllMark:  root.removeAllMark,
		actionSearch:         root.setForwardSearchMode,
		actionBackSearch:     root.setBackSearchMode,
		actionFilter:         root.setSearchFilterMode,
		actionDelimiter:      root.setDelimiterMode,
		actionHeader:         root.setHeaderMode,
		actionSkipLines:      root.setSkipLinesMode,
		actionTabWidth:       root.setTabWidthMode,
		actionGoLine:         root.setGoLineMode,
		actionSectionNum:     root.setSectionNumMode,
		actionNextSearch:     root.sendNextSearch,
		actionNextBackSearch: root.sendNextBackSearch,
		actionNextDoc:        root.nextDoc,
		actionPreviousDoc:    root.previousDoc,
		actionCloseDoc:       root.closeDocument,
		actionCloseAllFilter: root.closeAllFilter,
		actionToggleMouse:    root.toggleMouse,
		actionMultiColor:     root.setMultiColorMode,
		actionJumpTarget:     root.setJumpTargetMode,
		actionSaveBuffer:     root.setSaveBuffer,
		actionHideOther:      root.toggleHideOtherSection,

		inputCaseSensitive:      root.inputCaseSensitive,
		inputSmartCaseSensitive: root.inputSmartCaseSensitive,
		inputIncSearch:          root.inputIncSearch,
		inputRegexpSearch:       root.inputRegexpSearch,
		inputNonMatch:           root.inputNonMatch,
		inputPrevious:           root.inputPrevious,
		inputNext:               root.inputNext,
		inputCopy:               root.CopySelect,
		inputPaste:              root.Paste,
	}
}

// KeyBind is the mapping of action and key.
type KeyBind map[string][]string

// defaultKeyBinds are the default keybindings.
func defaultKeyBinds() KeyBind {
	return map[string][]string{
		actionExit:           {"Escape", "q"},
		actionWriteBA:        {"ctrl+q"},
		actionCancel:         {"ctrl+c"},
		actionWriteExit:      {"Q"},
		actionSync:           {"ctrl+l"},
		actionFollow:         {"ctrl+f"},
		actionFollowAll:      {"ctrl+a"},
		actionFollowSection:  {"F2"},
		actionPlain:          {"ctrl+e"},
		actionRainbow:        {"ctrl+r"},
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
		actionMoveWidthLeft:  {"ctrl+shift+left"},
		actionMoveWidthRight: {"ctrl+shift+right"},
		actionMoveLeft:       {"left"},
		actionMoveRight:      {"right"},
		actionMoveHfLeft:     {"ctrl+left"},
		actionMoveHfRight:    {"ctrl+right"},
		actionMoveBeginLeft:  {"shift+Home"},
		actionMoveEndRight:   {"shift+End"},
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
		actionColumnWidth:    {"alt+o"},
		actionAlternate:      {"C"},
		actionLineNumMode:    {"G"},
		actionMark:           {"m"},
		actionRemoveAllMark:  {"ctrl+delete"},
		actionRemoveMark:     {"M"},
		actionSearch:         {"/"},
		actionBackSearch:     {"?"},
		actionFilter:         {"&"},
		actionDelimiter:      {"d"},
		actionHeader:         {"H"},
		actionSkipLines:      {"ctrl+s"},
		actionTabWidth:       {"t"},
		actionGoLine:         {"g"},
		actionSectionNum:     {"F7"},
		actionNextSearch:     {"n"},
		actionNextBackSearch: {"N"},
		actionNextDoc:        {"]"},
		actionPreviousDoc:    {"["},
		actionCloseDoc:       {"ctrl+k"},
		actionCloseAllFilter: {"K"},
		actionToggleMouse:    {"ctrl+alt+r"},
		actionSuspend:        {"ctrl+z"},
		actionMultiColor:     {"."},
		actionJumpTarget:     {"j"},
		actionSaveBuffer:     {"S"},
		actionHideOther:      {"alt+-"},

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

// String returns keybind as a string for help.
func (k KeyBind) String() string {
	var b strings.Builder
	writeHeaderTh(&b)
	k.writeKeyBind(&b, actionExit, "quit")
	k.writeKeyBind(&b, actionCancel, "cancel")
	k.writeKeyBind(&b, actionWriteExit, "output screen and quit")
	k.writeKeyBind(&b, actionWriteBA, "set output screen and quit")
	k.writeKeyBind(&b, actionSuspend, "suspend")
	k.writeKeyBind(&b, actionHelp, "display help screen")
	k.writeKeyBind(&b, actionLogDoc, "display log screen")
	k.writeKeyBind(&b, actionSync, "screen sync")
	k.writeKeyBind(&b, actionFollow, "follow mode toggle")
	k.writeKeyBind(&b, actionFollowAll, "follow all mode toggle")
	k.writeKeyBind(&b, actionToggleMouse, "enable/disable mouse")
	k.writeKeyBind(&b, actionSaveBuffer, "save buffer to file")

	writeHeader(&b, "Moving")
	k.writeKeyBind(&b, actionMoveDown, "forward by one line")
	k.writeKeyBind(&b, actionMoveUp, "backward by one line")
	k.writeKeyBind(&b, actionMoveTop, "go to top of document")
	k.writeKeyBind(&b, actionMoveBottom, "go to end of document")
	k.writeKeyBind(&b, actionMovePgDn, "forward by page")
	k.writeKeyBind(&b, actionMovePgUp, "backward by page")
	k.writeKeyBind(&b, actionMoveHfDn, "forward a half page")
	k.writeKeyBind(&b, actionMoveHfUp, "backward a half page")
	k.writeKeyBind(&b, actionMoveLeft, "scroll to left")
	k.writeKeyBind(&b, actionMoveRight, "scroll to right")
	k.writeKeyBind(&b, actionMoveHfLeft, "scroll left half screen")
	k.writeKeyBind(&b, actionMoveHfRight, "scroll right half screen")
	k.writeKeyBind(&b, actionMoveWidthLeft, "scroll left specified width")
	k.writeKeyBind(&b, actionMoveWidthRight, "scroll right specified width")
	k.writeKeyBind(&b, actionMoveBeginLeft, "go to beginning of line")
	k.writeKeyBind(&b, actionMoveEndRight, "go to end of line")
	k.writeKeyBind(&b, actionGoLine, "go to line(input number or `.n` or `n%` allowed)")

	writeHeader(&b, "Move document")
	k.writeKeyBind(&b, actionNextDoc, "next document")
	k.writeKeyBind(&b, actionPreviousDoc, "previous document")
	k.writeKeyBind(&b, actionCloseDoc, "close current document")
	k.writeKeyBind(&b, actionCloseAllFilter, "close all filtered documents")

	writeHeader(&b, "Mark position")
	k.writeKeyBind(&b, actionMark, "mark current position")
	k.writeKeyBind(&b, actionRemoveMark, "remove mark current position")
	k.writeKeyBind(&b, actionRemoveAllMark, "remove all mark")
	k.writeKeyBind(&b, actionMoveMark, "move to next marked position")
	k.writeKeyBind(&b, actionMovePrevMark, "move to previous marked position")

	writeHeader(&b, "Search")
	k.writeKeyBind(&b, actionSearch, "forward search mode")
	k.writeKeyBind(&b, actionBackSearch, "backward search mode")
	k.writeKeyBind(&b, actionNextSearch, "repeat forward search")
	k.writeKeyBind(&b, actionNextBackSearch, "repeat backward search")
	k.writeKeyBind(&b, actionFilter, "filter search mode")

	writeHeader(&b, "Change display")
	k.writeKeyBind(&b, actionWrap, "wrap/nowrap toggle")
	k.writeKeyBind(&b, actionColumnMode, "column mode toggle")
	k.writeKeyBind(&b, actionColumnWidth, "column width toggle")
	k.writeKeyBind(&b, actionRainbow, "column rainbow toggle")
	k.writeKeyBind(&b, actionAlternate, "alternate rows of style toggle")
	k.writeKeyBind(&b, actionLineNumMode, "line number toggle")
	k.writeKeyBind(&b, actionPlain, "original decoration toggle(plain)")
	k.writeKeyBind(&b, actionHideOther, "toggle hide other section")

	writeHeader(&b, "Change Display with Input")
	k.writeKeyBind(&b, actionViewMode, "view mode selection")
	k.writeKeyBind(&b, actionDelimiter, "column delimiter string")
	k.writeKeyBind(&b, actionHeader, "number of header lines")
	k.writeKeyBind(&b, actionSkipLines, "number of skip lines")
	k.writeKeyBind(&b, actionTabWidth, "TAB width")
	k.writeKeyBind(&b, actionMultiColor, "multi color highlight")
	k.writeKeyBind(&b, actionJumpTarget, "jump target(`.n` or `n%` or `section` allowed)")

	writeHeader(&b, "Section")
	k.writeKeyBind(&b, actionSection, "section delimiter regular expression")
	k.writeKeyBind(&b, actionSectionStart, "section start position")
	k.writeKeyBind(&b, actionNextSection, "next section")
	k.writeKeyBind(&b, actionPrevSection, "previous section")
	k.writeKeyBind(&b, actionLastSection, "last section")
	k.writeKeyBind(&b, actionFollowSection, "follow section mode toggle")
	k.writeKeyBind(&b, actionSectionNum, "section header number")

	writeHeader(&b, "Close and reload")
	k.writeKeyBind(&b, actionCloseFile, "close file")
	k.writeKeyBind(&b, actionReload, "reload file")
	k.writeKeyBind(&b, actionWatch, "watch mode")
	k.writeKeyBind(&b, actionWatchInterval, "set watch interval")

	writeHeader(&b, "Key binding when typing")
	k.writeKeyBind(&b, inputCaseSensitive, "case-sensitive toggle")
	k.writeKeyBind(&b, inputSmartCaseSensitive, "smart case-sensitive toggle")
	k.writeKeyBind(&b, inputRegexpSearch, "regular expression search toggle")
	k.writeKeyBind(&b, inputIncSearch, "incremental search toggle")
	k.writeKeyBind(&b, inputNonMatch, "non-match toggle")
	k.writeKeyBind(&b, inputPrevious, "previous candidate")
	k.writeKeyBind(&b, inputNext, "next candidate")
	k.writeKeyBind(&b, inputCopy, "copy to clipboard.")
	k.writeKeyBind(&b, inputPaste, "paste from clipboard")
	return b.String()
}

func writeHeaderTh(w io.Writer) {
	fmt.Fprintf(w, " %-30s %s\n", "Key", "Action")
}
func writeHeader(w io.Writer, header string) {
	fmt.Fprintf(w, "\n\t%s\n", header)
}

func (k KeyBind) writeKeyBind(w io.Writer, action string, detail string) {
	fmt.Fprintf(w, " %-30s * %s\n", "["+strings.Join(k[action], "], [")+"]", detail)
}

// GetKeyBinds returns the current key mapping.
func GetKeyBinds(config Config) KeyBind {
	keyBind := make(map[string][]string)

	if strings.ToLower(config.DefaultKeyBind) != "disable" {
		keyBind = defaultKeyBinds()
	}

	// Overwrite with config file.
	for k, v := range config.Keybind {
		keyBind[k] = v
	}
	return keyBind
}

// setHandlers sets keys to action handlers.
func (root *Root) setHandlers(ctx context.Context, keyBind KeyBind) error {
	c := root.keyConfig
	in := root.inputKeyConfig

	actionHandlers := root.handlers()

	for name, keys := range keyBind {
		handler := actionHandlers[name]
		if handler == nil {
			return fmt.Errorf("%w for [%s] unknown action", ErrFailedKeyBind, name)
		}

		if strings.HasPrefix(name, "input_") {
			if err := setHandler(ctx, in, name, keys, handler); err != nil {
				return err
			}
			continue
		}
		if err := setHandler(ctx, c, name, keys, handler); err != nil {
			return err
		}
	}
	return nil
}

// setHandler sets multiple keys in one action handler.
func setHandler(ctx context.Context, c *cbind.Configuration, name string, keys []string, handler func(context.Context)) error {
	for _, k := range keys {
		mod, key, ch, err := cbind.Decode(k)
		if err != nil {
			return fmt.Errorf("%w [%s] for %s: %w", ErrFailedKeyBind, k, name, err)
		}
		if key == tcell.KeyRune {
			c.SetRune(mod, ch, wrapEventHandler(ctx, handler))
			// Added "shift+N" instead of 'N' to get it on windows.
			if 0x21 <= ch && ch <= 0x60 {
				c.SetRune(mod|tcell.ModShift, ch, wrapEventHandler(ctx, handler))
			}
		} else {
			// ctrl+h, Backspace and ctrl+Backspace can only be assigned one handler.
			if key == tcell.KeyBackspace || key == tcell.KeyBackspace2 {
				c.SetKey(tcell.ModNone, tcell.KeyBackspace, wrapEventHandler(ctx, handler))
				c.SetKey(tcell.ModNone, tcell.KeyBackspace2, wrapEventHandler(ctx, handler))
				c.SetKey(tcell.ModCtrl, tcell.KeyBackspace, wrapEventHandler(ctx, handler))
				c.SetKey(tcell.ModCtrl, tcell.KeyBackspace2, wrapEventHandler(ctx, handler))
				continue
			}
			c.SetKey(mod, key, wrapEventHandler(ctx, handler))
		}
	}
	return nil
}

// wrapEventHandler is a wrapper for matching func types.
func wrapEventHandler(ctx context.Context, f func(context.Context)) func(_ *tcell.EventKey) *tcell.EventKey {
	return func(_ *tcell.EventKey) *tcell.EventKey {
		f(ctx)
		return nil
	}
}

// keyCapture does the actual key action.
func (root *Root) keyCapture(ev *tcell.EventKey) bool {
	root.keyConfig.Capture(ev)
	return true
}
