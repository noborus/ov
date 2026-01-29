package oviewer

import (
	"context"
	"fmt"
	"io"
	"maps"
	"strings"

	"codeberg.org/tslocum/cbind"
	"github.com/gdamore/tcell/v3"
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
	actionSidebarHelp    = "sidebar_help"
	actionMarkList       = "show_mark_list"
	actionDocList        = "show_doc_list"
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
		actionSidebarHelp:    root.toggleSidebarHelp,
		actionMarkList:       root.toggleShowMarkList,
		actionDocList:        root.toggleShowDocList,
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

// KeyBind represents a mapping from action names to their associated key sequences.
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
		actionSidebarHelp:    {"alt+h"},
		actionMarkList:       {","},
		actionDocList:        {"f"},

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

type KeyBindDescription struct {
	Group       string
	Action      string
	Description string
}

var keyBindDescriptions = []KeyBindDescription{
	// General
	{Group: "", Action: actionExit, Description: "quit"},
	{Group: "", Action: actionCancel, Description: "cancel"},
	{Group: "", Action: actionWriteExit, Description: "output screen and quit"},
	{Group: "", Action: actionWriteBA, Description: "set output screen and quit"},
	{Group: "", Action: actionWriteOriginal, Description: "set output original screen and quit"},
	{Group: "", Action: actionSuspend, Description: "suspend"},
	{Group: "", Action: actionEdit, Description: "edit current document"},
	{Group: "", Action: actionHelp, Description: "display help screen"},
	{Group: "", Action: actionSidebarHelp, Description: "toggle sidebar help"},
	{Group: "", Action: actionLogDoc, Description: "display log screen"},
	{Group: "", Action: actionSync, Description: "screen sync"},
	{Group: "", Action: actionFollow, Description: "follow mode toggle"},
	{Group: "", Action: actionFollowAll, Description: "follow all mode toggle"},
	{Group: "", Action: actionToggleMouse, Description: "enable/disable mouse"},
	{Group: "", Action: actionSaveBuffer, Description: "save buffer to file"},

	// Moving
	{Group: "Moving", Action: actionMoveDown, Description: "forward by one line"},
	{Group: "Moving", Action: actionMoveUp, Description: "backward by one line"},
	{Group: "Moving", Action: actionMoveTop, Description: "go to top of document"},
	{Group: "Moving", Action: actionMoveBottom, Description: "go to end of document"},
	{Group: "Moving", Action: actionMovePgDn, Description: "forward by page"},
	{Group: "Moving", Action: actionMovePgUp, Description: "backward by page"},
	{Group: "Moving", Action: actionMoveHfDn, Description: "forward a half page"},
	{Group: "Moving", Action: actionMoveHfUp, Description: "backward a half page"},
	{Group: "Moving", Action: actionMoveLeft, Description: "scroll to left"},
	{Group: "Moving", Action: actionMoveRight, Description: "scroll to right"},
	{Group: "Moving", Action: actionMoveHfLeft, Description: "scroll left half screen"},
	{Group: "Moving", Action: actionMoveHfRight, Description: "scroll right half screen"},
	{Group: "Moving", Action: actionMoveWidthLeft, Description: "scroll left specified width"},
	{Group: "Moving", Action: actionMoveWidthRight, Description: "scroll right specified width"},
	{Group: "Moving", Action: actionMoveBeginLeft, Description: "go to beginning of line"},
	{Group: "Moving", Action: actionMoveEndRight, Description: "go to end of line"},
	{Group: "Moving", Action: actionGoLine, Description: "go to line(input number or `.n` or `n%` allowed)"},

	// Move document
	{Group: "Move document", Action: actionNextDoc, Description: "next document"},
	{Group: "Move document", Action: actionPreviousDoc, Description: "previous document"},
	{Group: "Move document", Action: actionCloseDoc, Description: "close current document"},
	{Group: "Move document", Action: actionCloseAllFilter, Description: "close all filtered documents"},
	{Group: "Move document", Action: actionDocList, Description: "show document list sidebar"},

	// Mark position
	{Group: "Mark position", Action: actionMark, Description: "mark current position"},
	{Group: "Mark position", Action: actionRemoveMark, Description: "remove mark current position"},
	{Group: "Mark position", Action: actionRemoveAllMark, Description: "remove all mark"},
	{Group: "Mark position", Action: actionMoveMark, Description: "move to next marked position"},
	{Group: "Mark position", Action: actionMovePrevMark, Description: "move to previous marked position"},
	{Group: "Mark position", Action: actionMarkList, Description: "show mark list sidebar"},

	// Search
	{Group: "Search", Action: actionSearch, Description: "forward search mode"},
	{Group: "Search", Action: actionBackSearch, Description: "backward search mode"},
	{Group: "Search", Action: actionNextSearch, Description: "repeat forward search"},
	{Group: "Search", Action: actionNextBackSearch, Description: "repeat backward search"},
	{Group: "Search", Action: actionFilter, Description: "filter search mode"},

	// Change display
	{Group: "Change display", Action: actionWrap, Description: "wrap/nowrap toggle"},
	{Group: "Change display", Action: actionColumnMode, Description: "column mode toggle"},
	{Group: "Change display", Action: actionColumnWidth, Description: "column width toggle"},
	{Group: "Change display", Action: actionRainbow, Description: "column rainbow toggle"},
	{Group: "Change display", Action: actionAlternate, Description: "alternate rows of style toggle"},
	{Group: "Change display", Action: actionLineNumMode, Description: "line number toggle"},
	{Group: "Change display", Action: actionPlain, Description: "original decoration toggle(plain)"},
	{Group: "Change display", Action: actionAlignFormat, Description: "align columns"},
	{Group: "Change display", Action: actionRawFormat, Description: "raw output"},
	{Group: "Change display", Action: actionRuler, Description: "ruler toggle"},
	{Group: "Change display", Action: actionStatusLine, Description: "status line toggle"},

	// Change Display with Input
	{Group: "Change Display with Input", Action: actionViewMode, Description: "view mode selection"},
	{Group: "Change Display with Input", Action: actionDelimiter, Description: "column delimiter string"},
	{Group: "Change Display with Input", Action: actionHeader, Description: "number of header lines"},
	{Group: "Change Display with Input", Action: actionSkipLines, Description: "number of skip lines"},
	{Group: "Change Display with Input", Action: actionTabWidth, Description: "TAB width"},
	{Group: "Change Display with Input", Action: actionMultiColor, Description: "multi color highlight"},
	{Group: "Change Display with Input", Action: actionJumpTarget, Description: "jump target(`.n` or `n%` or `section` allowed)"},
	{Group: "Change Display with Input", Action: actionConvertType, Description: "convert type selection"},
	{Group: "Change Display with Input", Action: actionVerticalHeader, Description: "number of vertical header"},
	{Group: "Change Display with Input", Action: actionHeaderColumn, Description: "number of header column"},

	// Column operation
	{Group: "Column operation", Action: actionFixedColumn, Description: "header column fixed toggle"},
	{Group: "Column operation", Action: actionShrinkColumn, Description: "shrink column toggle(align mode only)"},
	{Group: "Column operation", Action: actionRightAlign, Description: "right align column toggle(align mode only)"},

	// Section operation
	{Group: "Section operation", Action: actionSection, Description: "section delimiter regular expression"},
	{Group: "Section operation", Action: actionSectionStart, Description: "section start position"},
	{Group: "Section operation", Action: actionNextSection, Description: "next section"},
	{Group: "Section operation", Action: actionPrevSection, Description: "previous section"},
	{Group: "Section operation", Action: actionLastSection, Description: "last section"},
	{Group: "Section operation", Action: actionFollowSection, Description: "follow section mode toggle"},
	{Group: "Section operation", Action: actionSectionNum, Description: "number of section header lines"},
	{Group: "Section operation", Action: actionHideOther, Description: `hide "other" section toggle`},

	// Close and reload
	{Group: "Close and reload", Action: actionCloseFile, Description: "close file"},
	{Group: "Close and reload", Action: actionReload, Description: "reload file"},
	{Group: "Close and reload", Action: actionWatch, Description: "watch mode"},
	{Group: "Close and reload", Action: actionWatchInterval, Description: "set watch interval"},

	// Key binding when typing
	{Group: "Key binding when typing", Action: inputCaseSensitive, Description: "case-sensitive toggle"},
	{Group: "Key binding when typing", Action: inputSmartCaseSensitive, Description: "smart case-sensitive toggle"},
	{Group: "Key binding when typing", Action: inputRegexpSearch, Description: "regular expression search toggle"},
	{Group: "Key binding when typing", Action: inputIncSearch, Description: "incremental search toggle"},
	{Group: "Key binding when typing", Action: inputNonMatch, Description: "non-match toggle"},
	{Group: "Key binding when typing", Action: inputPrevious, Description: "previous candidate"},
	{Group: "Key binding when typing", Action: inputNext, Description: "next candidate"},
	{Group: "Key binding when typing", Action: inputCopy, Description: "copy to clipboard"},
	{Group: "Key binding when typing", Action: inputPaste, Description: "paste from clipboard"},
}

// String returns keybind as a string for help.
func (k KeyBind) String() string {
	var b strings.Builder
	writeHeaderTh(&b)
	group := "nogroup"
	for _, bind := range keyBindDescriptions {
		if bind.Group != group {
			group = bind.Group
			writeHeader(&b, group)
		}
		keys := k[bind.Action]
		b.WriteString(fmt.Sprintf(" %-30s * %s\n", "["+strings.Join(keys, "], [")+"]", bind.Description))
	}
	return b.String()
}

func (k KeyBind) GetKeyBindDescriptions(group string) [][]string {
	var descriptions [][]string

	if group != "all" {
		for _, bind := range keyBindDescriptions {
			if bind.Group != group {
				continue
			}
			descriptions = append(descriptions, []string{bind.Description, strings.Join(k[bind.Action], ", ")})
		}
		return descriptions
	}

	for _, bind := range keyBindDescriptions {
		descriptions = append(descriptions, []string{bind.Description, strings.Join(k[bind.Action], ", ")})
	}
	return descriptions
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
		if err := c.Set(k, wrapEventHandler(ctx, handler)); err != nil {
			return fmt.Errorf("%w [%s] for %s: %w", ErrFailedKeyBind, k, name, err)
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
	if root.keyConfig.Capture(ev) == nil {
		return true
	}
	if root.Config.Debug {
		root.setMessageLogf("key \"%s\" not assigned", root.formatKeyName(ev))
	}
	return false
}

// formatKeyName formats the key name for better readability
func (root *Root) formatKeyName(ev *tcell.EventKey) string {
	// For special keys, use the standard name
	return ev.Name()
}

// keyActionMapping represents a key that is assigned to multiple actions.
type keyActionMapping struct {
	key    string
	action []string
}

// normalizeKey normalizes a key string by decoding and re-encoding it.
// This ensures consistent key representation for duplicate detection.
func normalizeKey(key string) (string, error) {
	dm, dk, dc, err := cbind.Decode(key)
	if err != nil {
		return "", err
	}

	encoded, err := cbind.Encode(dm, dk, dc)
	if err != nil {
		return "", err
	}

	// If the original was ctrl+J and encode result is Enter, keep ctrl+j
	lowerKey := strings.ToLower(key)
	if encoded == "Enter" && lowerKey == "ctrl+j" {
		return "ctrl+j", nil
	}

	return encoded, nil
}

// normalizeKeyWithPrefix normalizes a key and adds input_ prefix if the action is an input action.
func normalizeKeyWithPrefix(key, action string) (string, error) {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(action, "input_") {
		return "input_" + normalizedKey, nil
	}
	return normalizedKey, nil
}

// findDuplicateKeyBind scans the provided KeyBind map and returns a slice of duplicate,
// where each duplicate contains a key and the list of actions that share that key.
// Input: keyBind - a map of action names to their associated key sequences.
// Output: a slice of keyActionMapping structs for keys assigned to multiple actions and any errors encountered.
func findDuplicateKeyBind(keyBind KeyBind) ([]keyActionMapping, []error) {
	keyActions := make(map[string]keyActionMapping)
	var errors []error

	for action, keys := range keyBind {
		for _, key := range keys {
			normalizedKey, err := normalizeKeyWithPrefix(key, action)
			if err != nil {
				errors = append(errors, fmt.Errorf("%w: key %s for action %s", ErrInvalidKey, key, action))
				continue // Skip this key and continue with the next one.
			}

			if existing, exists := keyActions[normalizedKey]; exists {
				keyActions[normalizedKey] = keyActionMapping{
					key:    normalizedKey,
					action: append(existing.action, action),
				}
			} else {
				keyActions[normalizedKey] = keyActionMapping{
					key:    normalizedKey,
					action: []string{action},
				}
			}
		}
	}

	// Return only keys with multiple actions (duplicates)
	duplicates := make([]keyActionMapping, 0, len(keyActions))
	for _, mapping := range keyActions {
		if len(mapping.action) > 1 {
			duplicates = append(duplicates, mapping)
		}
	}

	return duplicates, errors
}
