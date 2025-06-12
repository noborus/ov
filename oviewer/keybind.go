package oviewer

import (
	"context"
	"fmt"
	"io"
	"maps"
	"strings"

	"codeberg.org/tslocum/cbind"
	"github.com/gdamore/tcell/v2"
)

// The name of the action to assign the key to.
// The string is displayed in help.
const (
	actionExit           = "exit"
	actionCancel         = "cancel"
	actionWriteExit      = "write_exit"
	actionSuspend        = "suspend"
	actionEdit           = "edit"
	actionSync           = "sync"
	actionFollow         = "follow_mode"
	actionFollowAll      = "follow_all"
	actionFollowSection  = "follow_section"
	actionPlain          = "plain_mode"
	actionRainbow        = "rainbow_mode"
	actionCloseFile      = "close_file"
	actionReload         = "reload"
	actionWatch          = "watch"
	actionHelp           = "help"
	actionLogDoc         = "logdoc"
	actionMark           = "mark"
	actionRemoveMark     = "remove_mark"
	actionRemoveAllMark  = "remove_all_mark"
	actionAlternate      = "alter_rows_mode"
	actionLineNumMode    = "line_number_mode"
	actionWrap           = "wrap_mode"
	actionColumnMode     = "column_mode"
	actionColumnWidth    = "column_width"
	actionNextSearch     = "next_search"
	actionNextBackSearch = "next_backsearch"
	actionNextDoc        = "next_doc"
	actionPreviousDoc    = "previous_doc"
	actionCloseDoc       = "close_doc"
	actionCloseAllFilter = "close_all_filter"
	actionToggleMouse    = "toggle_mouse"
	actionHideOther      = "hide_other"
	actionStatusLine     = "status_line"
	actionAlignFormat    = "align_format"
	actionRawFormat      = "raw_format"
	actionFixedColumn    = "fixed_column"
	actionShrinkColumn   = "shrink_column"
	actionRightAlign     = "right_align"
	actionRuler          = "toggle_ruler"
	actionWriteOriginal  = "write_original"

	// Move actions.
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
	actionNextSection    = "next_section"
	actionLastSection    = "last_section"
	actionPrevSection    = "previous_section"
	actionMoveMark       = "next_mark"
	actionMovePrevMark   = "previous_mark"

	// Actions that enter input mode.
	actionConvertType    = "convert_type"
	actionDelimiter      = "delimiter"
	actionGoLine         = "goto"
	actionHeaderColumn   = "header_column"
	actionHeader         = "header"
	actionJumpTarget     = "jump_target"
	actionMultiColor     = "multi_color"
	actionSaveBuffer     = "save_buffer"
	actionSearch         = "search"
	actionBackSearch     = "backsearch"
	actionFilter         = "filter"
	actionSection        = "section_delimiter"
	actionSectionNum     = "section_header_num"
	actionSectionStart   = "section_start"
	actionSkipLines      = "skip_lines"
	actionTabWidth       = "tabwidth"
	actionVerticalHeader = "vertical_header"
	actionViewMode       = "set_view_mode"
	actionWatchInterval  = "watch_interval"
	actionWriteBA        = "set_write_exit"

	// input actions.
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
		actionCancel:         root.Cancel,
		actionWriteExit:      root.WriteQuit,
		actionSuspend:        root.suspend,
		actionEdit:           root.edit,
		actionSync:           root.ViewSync,
		actionFollow:         root.toggleFollowMode,
		actionFollowAll:      root.toggleFollowAll,
		actionFollowSection:  root.toggleFollowSection,
		actionPlain:          root.togglePlain,
		actionRainbow:        root.toggleRainbow,
		actionCloseFile:      root.closeFile,
		actionReload:         root.Reload,
		actionWatch:          root.toggleWatch,
		actionHelp:           root.helpDisplay,
		actionLogDoc:         root.logDisplay,
		actionMark:           root.addMark,
		actionRemoveMark:     root.removeMark,
		actionRemoveAllMark:  root.removeAllMark,
		actionAlternate:      root.toggleAlternateRows,
		actionLineNumMode:    root.toggleLineNumMode,
		actionWrap:           root.toggleWrapMode,
		actionColumnMode:     root.toggleColumnMode,
		actionColumnWidth:    root.toggleColumnWidth,
		actionNextSearch:     root.sendNextSearch,
		actionNextBackSearch: root.sendNextBackSearch,
		actionNextDoc:        root.nextDoc,
		actionPreviousDoc:    root.previousDoc,
		actionCloseDoc:       root.closeDocument,
		actionCloseAllFilter: root.closeAllFilter,
		actionToggleMouse:    root.toggleMouse,
		actionHideOther:      root.toggleHideOtherSection,
		actionStatusLine:     root.toggleStatusLine,
		actionAlignFormat:    root.alignFormat,
		actionRawFormat:      root.rawFormat,
		actionFixedColumn:    root.toggleFixedColumn,
		actionShrinkColumn:   root.toggleShrinkColumn,
		actionRightAlign:     root.toggleRightAlign,
		actionRuler:          root.toggleRuler,
		actionWriteOriginal:  root.toggleWriteOriginal,

		// Move actions.
		actionMoveDown:       root.moveDownOne,
		actionMoveUp:         root.moveUpOne,
		actionMoveTop:        root.moveTop,
		actionMoveWidthLeft:  root.moveWidthLeft,
		actionMoveWidthRight: root.moveWidthRight,
		actionMoveLeft:       root.moveLeftOne,
		actionMoveRight:      root.moveRightOne,
		actionMoveHfLeft:     root.moveHfLeft,
		actionMoveHfRight:    root.moveHfRight,
		actionMoveBeginLeft:  root.moveBeginLeft,
		actionMoveEndRight:   root.moveEndRight,
		actionMoveBottom:     root.moveBottom,
		actionMovePgUp:       root.movePgUp,
		actionMovePgDn:       root.movePgDn,
		actionMoveHfUp:       root.moveHfUp,
		actionMoveHfDn:       root.moveHfDn,
		actionNextSection:    root.nextSection,
		actionLastSection:    root.lastSection,
		actionPrevSection:    root.prevSection,
		actionMoveMark:       root.nextMark,
		actionMovePrevMark:   root.prevMark,

		// Actions that enter input mode.
		actionConvertType:    root.inputConvert,
		actionDelimiter:      root.inputDelimiter,
		actionGoLine:         root.inputGoLine,
		actionHeaderColumn:   root.inputHeaderColumn,
		actionHeader:         root.inputHeader,
		actionJumpTarget:     root.inputJumpTarget,
		actionMultiColor:     root.inputMultiColor,
		actionSaveBuffer:     root.inputSaveBuffer,
		actionSearch:         root.inputForwardSearch,
		actionBackSearch:     root.inputBackSearch,
		actionFilter:         root.inputSearchFilter,
		actionSection:        root.inputSectionDelimiter,
		actionSectionNum:     root.inputSectionNum,
		actionSectionStart:   root.inputSectionStart,
		actionSkipLines:      root.inputSkipLines,
		actionTabWidth:       root.inputTabWidth,
		actionVerticalHeader: root.inputVerticalHeader,
		actionViewMode:       root.inputViewMode,
		actionWatchInterval:  root.inputWatchInterval,
		actionWriteBA:        root.inputWriteBA,

		// input actions.
		inputCaseSensitive:      root.toggleCaseSensitive,
		inputSmartCaseSensitive: root.toggleSmartCaseSensitive,
		inputIncSearch:          root.toggleIncSearch,
		inputRegexpSearch:       root.toggleRegexpSearch,
		inputNonMatch:           root.toggleNonMatch,
		inputPrevious:           root.candidatePrevious,
		inputNext:               root.candidateNext,
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
		actionColumnMode:     {"c"},
		actionColumnWidth:    {"alt+o"},
		actionNextSearch:     {"n"},
		actionNextBackSearch: {"N"},
		actionNextDoc:        {"]"},
		actionPreviousDoc:    {"["},
		actionCloseDoc:       {"ctrl+k"},
		actionCloseAllFilter: {"K"},
		actionToggleMouse:    {"ctrl+alt+r"},
		actionHideOther:      {"alt+-"},
		actionAlignFormat:    {"alt+F"},
		actionRawFormat:      {"alt+R"},
		actionFixedColumn:    {"F"},
		actionShrinkColumn:   {"s"},
		actionRightAlign:     {"alt+a"},
		actionRuler:          {"alt+shift+F9"},
		actionWriteOriginal:  {"alt+shift+F8"},
		actionStatusLine:     {"ctrl+F10"},

		// Move actions.
		actionMoveDown:       {"Enter", "Down", "ctrl+N"},
		actionMoveUp:         {"Up", "ctrl+p"},
		actionMoveTop:        {"Home"},
		actionMoveWidthLeft:  {"ctrl+shift+left"},
		actionMoveWidthRight: {"ctrl+shift+right"},
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

// String returns keybind as a string for help.
func (k KeyBind) String() string {
	var b strings.Builder
	writeHeaderTh(&b)
	k.writeKeyBind(&b, actionExit, "quit")
	k.writeKeyBind(&b, actionCancel, "cancel")
	k.writeKeyBind(&b, actionWriteExit, "output screen and quit")
	k.writeKeyBind(&b, actionWriteBA, "set output screen and quit")
	k.writeKeyBind(&b, actionWriteOriginal, "set output original screen and quit")
	k.writeKeyBind(&b, actionSuspend, "suspend")
	k.writeKeyBind(&b, actionEdit, "edit current document")
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
	k.writeKeyBind(&b, actionAlignFormat, "align columns")
	k.writeKeyBind(&b, actionRawFormat, "raw output")
	k.writeKeyBind(&b, actionRuler, "ruler toggle")
	k.writeKeyBind(&b, actionStatusLine, "status line toggle")

	writeHeader(&b, "Change Display with Input")
	k.writeKeyBind(&b, actionViewMode, "view mode selection")
	k.writeKeyBind(&b, actionDelimiter, "column delimiter string")
	k.writeKeyBind(&b, actionHeader, "number of header lines")
	k.writeKeyBind(&b, actionSkipLines, "number of skip lines")
	k.writeKeyBind(&b, actionTabWidth, "TAB width")
	k.writeKeyBind(&b, actionMultiColor, "multi color highlight")
	k.writeKeyBind(&b, actionJumpTarget, "jump target(`.n` or `n%` or `section` allowed)")
	k.writeKeyBind(&b, actionConvertType, "convert type selection")
	k.writeKeyBind(&b, actionVerticalHeader, "number of vertical header")
	k.writeKeyBind(&b, actionHeaderColumn, "number of header column")

	writeHeader(&b, "Column operation")
	k.writeKeyBind(&b, actionFixedColumn, "header column fixed toggle")
	k.writeKeyBind(&b, actionShrinkColumn, "shrink column toggle(align mode only)")
	k.writeKeyBind(&b, actionRightAlign, "right align column toggle(align mode only)")

	writeHeader(&b, "Section operation")
	k.writeKeyBind(&b, actionSection, "section delimiter regular expression")
	k.writeKeyBind(&b, actionSectionStart, "section start position")
	k.writeKeyBind(&b, actionNextSection, "next section")
	k.writeKeyBind(&b, actionPrevSection, "previous section")
	k.writeKeyBind(&b, actionLastSection, "last section")
	k.writeKeyBind(&b, actionFollowSection, "follow section mode toggle")
	k.writeKeyBind(&b, actionSectionNum, "number of section header lines")
	k.writeKeyBind(&b, actionHideOther, `hide "other" section toggle`)

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
	k.writeKeyBind(&b, inputCopy, "copy to clipboard")
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

// GetKeyBinds returns the current key mapping based on the provided configuration.
// If the default key bindings are not disabled in the configuration, it initializes
// the key bindings with the default values and then overwrites them with any custom
// key bindings specified in the configuration file.
//
// Parameters:
//   - config: The configuration object that contains the key binding settings.
//
// Returns:
//   - KeyBind: A map where the keys are action names and the values are slices of
//     strings representing the keys assigned to those actions.
func GetKeyBinds(config Config) KeyBind {
	keyBind := make(map[string][]string)

	if strings.ToLower(config.DefaultKeyBind) != "disable" {
		keyBind = defaultKeyBinds()
	}

	// Overwrite with config file.
	maps.Copy(keyBind, config.Keybind)
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

type duplicate struct {
	key    string
	action []string
}

func findDuplicateKeyBind(keyBind KeyBind) []duplicate {
	keyActions := make(map[string]duplicate)
	for k, v := range keyBind {
		for _, key := range v {
			if strings.HasPrefix(k, "input_") {
				key = "input_" + key
			}
			if _, ok := keyActions[key]; ok {
				keyActions[key] = duplicate{key: key, action: append(keyActions[key].action, k)}
			} else {
				keyActions[key] = duplicate{key: key, action: []string{k}}
			}
		}
	}
	q := make([]duplicate, 0, len(keyActions))
	for _, v := range keyActions {
		if len(v.action) > 1 {
			q = append(q, v)
		}
	}
	return q
}
