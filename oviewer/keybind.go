package oviewer

import (
	"context"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"codeberg.org/tslocum/cbind"
	"github.com/gdamore/tcell/v3"
)

// The name of the action to assign the key to.
// The string is displayed in help.
const (
	// General
	actionExit          = "exit"
	actionCancel        = "cancel"
	actionWriteExit     = "write_exit"
	actionWriteBA       = "set_write_exit"
	actionWriteOriginal = "write_original"
	actionSuspend       = "suspend"
	actionEdit          = "edit"
	actionHelp          = "help"
	actionLogDoc        = "logdoc"
	actionSync          = "sync"
	actionFollow        = "follow_mode"
	actionFollowAll     = "follow_all"
	actionToggleMouse   = "toggle_mouse"
	actionSaveBuffer    = "save_buffer"

	// Moving
	actionMoveDown       = "down"
	actionMoveUp         = "up"
	actionMoveTop        = "top"
	actionMoveBottom     = "bottom"
	actionMovePgDn       = "page_down"
	actionMovePgUp       = "page_up"
	actionMoveHfDn       = "page_half_down"
	actionMoveHfUp       = "page_half_up"
	actionMoveLeft       = "left"
	actionMoveRight      = "right"
	actionMoveHfLeft     = "half_left"
	actionMoveHfRight    = "half_right"
	actionMoveWidthLeft  = "width_left"
	actionMoveWidthRight = "width_right"
	actionMoveBeginLeft  = "begin_left"
	actionMoveEndRight   = "end_right"
	actionGoLine         = "goto"
	actionMarkNumber     = "mark_number"

	// Sidebar
	actionSidebarHelp     = "sidebar_help"
	actionSidebarMarks    = "sidebar_marks"
	actionSidebarDocList  = "sidebar_doc_list"
	actionSidebarSections = "sidebar_sections"
	actionSidebarUp       = "sidebar_up"
	actionSidebarDown     = "sidebar_down"
	actionSidebarLeft     = "sidebar_left"
	actionSidebarRight    = "sidebar_right"

	// Move document
	actionNextDoc        = "next_doc"
	actionPreviousDoc    = "previous_doc"
	actionCloseDoc       = "close_doc"
	actionCloseAllFilter = "close_all_filter"

	// Mark position
	actionMark          = "mark"
	actionRemoveMark    = "remove_mark"
	actionRemoveAllMark = "remove_all_mark"
	actionMoveMark      = "next_mark"
	actionMovePrevMark  = "previous_mark"
	actionMarkByPattern = "mark_by_pattern"

	// Search
	actionSearch         = "search"
	actionBackSearch     = "backsearch"
	actionNextSearch     = "next_search"
	actionNextBackSearch = "next_backsearch"
	actionFilter         = "filter"

	// Change display
	actionWrap        = "wrap_mode"
	actionWordWrap    = "word_wrap_mode"
	actionColumnMode  = "column_mode"
	actionColumnWidth = "column_width"
	actionRainbow     = "rainbow_mode"
	actionAlternate   = "alter_rows_mode"
	actionLineNumMode = "line_number_mode"
	actionPlain       = "plain_mode"
	actionAlignFormat = "align_format"
	actionRawFormat   = "raw_format"
	actionRuler       = "toggle_ruler"
	actionStatusLine  = "status_line"

	// Change Display with Input
	actionViewMode       = "set_view_mode"
	actionDelimiter      = "delimiter"
	actionHeader         = "header"
	actionSkipLines      = "skip_lines"
	actionTabWidth       = "tabwidth"
	actionMultiColor     = "multi_color"
	actionJumpTarget     = "jump_target"
	actionConvertType    = "convert_type"
	actionVerticalHeader = "vertical_header"
	actionHeaderColumn   = "header_column"

	// Column operation
	actionFixedColumn  = "fixed_column"
	actionShrinkColumn = "shrink_column"
	actionRightAlign   = "right_align"

	// Section operation
	actionSection       = "section_delimiter"
	actionSectionStart  = "section_start"
	actionNextSection   = "next_section"
	actionPrevSection   = "previous_section"
	actionLastSection   = "last_section"
	actionFollowSection = "follow_section"
	actionSectionNum    = "section_header_num"
	actionHideOther     = "hide_other"

	// Close and reload
	actionCloseFile     = "close_file"
	actionReload        = "reload"
	actionWatch         = "watch"
	actionWatchInterval = "watch_interval"

	// Key binding when typing
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
		// General
		actionExit:          root.Quit,
		actionCancel:        root.Cancel,
		actionWriteExit:     root.WriteQuit,
		actionWriteBA:       root.inputWriteBA,
		actionWriteOriginal: root.toggleWriteOriginal,
		actionSuspend:       root.suspend,
		actionEdit:          root.edit,
		actionHelp:          root.helpDisplay,
		actionLogDoc:        root.logDisplay,
		actionSync:          root.ViewSync,
		actionFollow:        root.toggleFollowMode,
		actionFollowAll:     root.toggleFollowAll,
		actionToggleMouse:   root.toggleMouse,
		actionSaveBuffer:    root.inputSaveBuffer,

		// Moving
		actionMoveDown:       root.moveDownOne,
		actionMoveUp:         root.moveUpOne,
		actionMoveTop:        root.moveTop,
		actionMoveBottom:     root.moveBottom,
		actionMovePgDn:       root.movePgDn,
		actionMovePgUp:       root.movePgUp,
		actionMoveHfDn:       root.moveHfDn,
		actionMoveHfUp:       root.moveHfUp,
		actionMoveLeft:       root.moveLeftOne,
		actionMoveRight:      root.moveRightOne,
		actionMoveHfLeft:     root.moveHfLeft,
		actionMoveHfRight:    root.moveHfRight,
		actionMoveWidthLeft:  root.moveWidthLeft,
		actionMoveWidthRight: root.moveWidthRight,
		actionMoveBeginLeft:  root.moveBeginLeft,
		actionMoveEndRight:   root.moveEndRight,
		actionGoLine:         root.inputGoLine,
		actionMarkNumber:     root.inputMarkNumber,

		// Sidebar
		actionSidebarHelp:     root.toggleSidebarHelp,
		actionSidebarMarks:    root.toggleSidebarMarks,
		actionSidebarDocList:  root.toggleSidebarDocList,
		actionSidebarSections: root.toggleSidebarSections,
		actionSidebarUp:       root.sidebarUp,
		actionSidebarDown:     root.sidebarDown,
		actionSidebarLeft:     root.sidebarLeft,
		actionSidebarRight:    root.sidebarRight,

		// Move document
		actionNextDoc:        root.nextDoc,
		actionPreviousDoc:    root.previousDoc,
		actionCloseDoc:       root.closeDocument,
		actionCloseAllFilter: root.closeAllFilter,

		// Mark position
		actionMark:          root.addMark,
		actionRemoveMark:    root.removeMark,
		actionRemoveAllMark: root.removeAllMark,
		actionMoveMark:      root.nextMark,
		actionMovePrevMark:  root.prevMark,
		actionMarkByPattern: root.inputMarkByPattern,

		// Search
		actionSearch:         root.inputForwardSearch,
		actionBackSearch:     root.inputBackSearch,
		actionNextSearch:     root.sendNextSearch,
		actionNextBackSearch: root.sendNextBackSearch,
		actionFilter:         root.inputSearchFilter,

		// Change display
		actionWrap:        root.toggleWrapMode,
		actionWordWrap:    root.toggleWordWrap,
		actionColumnMode:  root.toggleColumnMode,
		actionColumnWidth: root.toggleColumnWidth,
		actionRainbow:     root.toggleRainbow,
		actionAlternate:   root.toggleAlternateRows,
		actionLineNumMode: root.toggleLineNumMode,
		actionPlain:       root.togglePlain,
		actionAlignFormat: root.alignFormat,
		actionRawFormat:   root.rawFormat,
		actionRuler:       root.toggleRuler,
		actionStatusLine:  root.toggleStatusLine,

		// Change Display with Input
		actionViewMode:       root.inputViewMode,
		actionDelimiter:      root.inputDelimiter,
		actionHeader:         root.inputHeader,
		actionSkipLines:      root.inputSkipLines,
		actionTabWidth:       root.inputTabWidth,
		actionMultiColor:     root.inputMultiColor,
		actionJumpTarget:     root.inputJumpTarget,
		actionConvertType:    root.inputConvert,
		actionVerticalHeader: root.inputVerticalHeader,
		actionHeaderColumn:   root.inputHeaderColumn,

		// Column operation
		actionFixedColumn:  root.toggleFixedColumn,
		actionShrinkColumn: root.toggleShrinkColumn,
		actionRightAlign:   root.toggleRightAlign,

		// Section operation
		actionSection:       root.inputSectionDelimiter,
		actionSectionStart:  root.inputSectionStart,
		actionNextSection:   root.nextSection,
		actionPrevSection:   root.prevSection,
		actionLastSection:   root.lastSection,
		actionFollowSection: root.toggleFollowSection,
		actionSectionNum:    root.inputSectionNum,
		actionHideOther:     root.toggleHideOtherSection,

		// Close and reload
		actionCloseFile:     root.closeFile,
		actionReload:        root.Reload,
		actionWatch:         root.toggleWatch,
		actionWatchInterval: root.inputWatchInterval,

		// Key binding when typing
		inputCaseSensitive:      root.toggleCaseSensitive,
		inputSmartCaseSensitive: root.toggleSmartCaseSensitive,
		inputRegexpSearch:       root.toggleRegexpSearch,
		inputIncSearch:          root.toggleIncSearch,
		inputNonMatch:           root.toggleNonMatch,
		inputPrevious:           root.candidatePrevious,
		inputNext:               root.candidateNext,
		inputCopy:               root.CopySelect,
		inputPaste:              root.Paste,
	}
}

// KeyBind represents a mapping from action names to their associated key sequences.
type KeyBind map[string][]string

type Group int

const (
	GroupAll     Group = -1
	GroupGeneral Group = iota
	GroupMoving
	GroupSidebar
	GroupDocList
	GroupMark
	GroupSearch
	GroupChange
	GroupChangeInput
	GroupColumn
	GroupSection
	GroupClose
	GroupTyping
)

func (g Group) String() string {
	switch g {
	case GroupAll:
		return "All"
	case GroupGeneral:
		return "General"
	case GroupMoving:
		return "Moving"
	case GroupSidebar:
		return "Sidebar"
	case GroupDocList:
		return "Move document"
	case GroupMark:
		return "Mark position"
	case GroupSearch:
		return "Search"
	case GroupChange:
		return "Change display"
	case GroupChangeInput:
		return "Change Display with Input"
	case GroupColumn:
		return "Column operation"
	case GroupSection:
		return "Section operation"
	case GroupClose:
		return "Close and reload"
	case GroupTyping:
		return "Key binding when typing"
	default:
		return ""
	}
}

type KeyBindDescription struct {
	Group       Group
	Action      string
	Description string
}

var keyBindDescriptions = []KeyBindDescription{
	// General
	{Group: GroupGeneral, Action: actionExit, Description: "quit"},
	{Group: GroupGeneral, Action: actionCancel, Description: "cancel"},
	{Group: GroupGeneral, Action: actionWriteExit, Description: "output screen and quit"},
	{Group: GroupGeneral, Action: actionWriteBA, Description: "set output screen and quit"},
	{Group: GroupGeneral, Action: actionWriteOriginal, Description: "set output original screen and quit"},
	{Group: GroupGeneral, Action: actionSuspend, Description: "suspend"},
	{Group: GroupGeneral, Action: actionEdit, Description: "edit current document"},
	{Group: GroupGeneral, Action: actionHelp, Description: "display help screen"},
	{Group: GroupGeneral, Action: actionLogDoc, Description: "display log screen"},
	{Group: GroupGeneral, Action: actionSync, Description: "screen sync"},
	{Group: GroupGeneral, Action: actionFollow, Description: "follow mode toggle"},
	{Group: GroupGeneral, Action: actionFollowAll, Description: "follow all mode toggle"},
	{Group: GroupGeneral, Action: actionToggleMouse, Description: "enable/disable mouse"},
	{Group: GroupGeneral, Action: actionSaveBuffer, Description: "save buffer to file"},

	// Moving
	{Group: GroupMoving, Action: actionMoveDown, Description: "forward by one line"},
	{Group: GroupMoving, Action: actionMoveUp, Description: "backward by one line"},
	{Group: GroupMoving, Action: actionMoveTop, Description: "go to top of document"},
	{Group: GroupMoving, Action: actionMoveBottom, Description: "go to end of document"},
	{Group: GroupMoving, Action: actionMovePgDn, Description: "forward by page"},
	{Group: GroupMoving, Action: actionMovePgUp, Description: "backward by page"},
	{Group: GroupMoving, Action: actionMoveHfDn, Description: "forward a half page"},
	{Group: GroupMoving, Action: actionMoveHfUp, Description: "backward a half page"},
	{Group: GroupMoving, Action: actionMoveLeft, Description: "scroll left"},
	{Group: GroupMoving, Action: actionMoveRight, Description: "scroll right"},
	{Group: GroupMoving, Action: actionMoveHfLeft, Description: "scroll left half screen"},
	{Group: GroupMoving, Action: actionMoveHfRight, Description: "scroll right half screen"},
	{Group: GroupMoving, Action: actionMoveWidthLeft, Description: "scroll left specified width"},
	{Group: GroupMoving, Action: actionMoveWidthRight, Description: "scroll right specified width"},
	{Group: GroupMoving, Action: actionMoveBeginLeft, Description: "go to beginning of line"},
	{Group: GroupMoving, Action: actionMoveEndRight, Description: "go to end of line"},
	{Group: GroupMoving, Action: actionGoLine, Description: "go to line(input number or `.n` or `n%` allowed)"},
	{Group: GroupMoving, Action: actionMarkNumber, Description: "go to mark number(input number allowed)"},

	// Sidebar
	{Group: GroupSidebar, Action: actionSidebarHelp, Description: "toggle help in sidebar"},
	{Group: GroupSidebar, Action: actionSidebarMarks, Description: "toggle mark list in sidebar"},
	{Group: GroupSidebar, Action: actionSidebarDocList, Description: "toggle document list in sidebar"},
	{Group: GroupSidebar, Action: actionSidebarSections, Description: "toggle section list in sidebar"},
	{Group: GroupSidebar, Action: actionSidebarUp, Description: "scroll up in sidebar"},
	{Group: GroupSidebar, Action: actionSidebarDown, Description: "scroll down in sidebar"},
	{Group: GroupSidebar, Action: actionSidebarLeft, Description: "scroll left in sidebar"},
	{Group: GroupSidebar, Action: actionSidebarRight, Description: "scroll right in sidebar"},

	// Move document
	{Group: GroupDocList, Action: actionNextDoc, Description: "next document"},
	{Group: GroupDocList, Action: actionPreviousDoc, Description: "previous document"},
	{Group: GroupDocList, Action: actionCloseDoc, Description: "close current document"},
	{Group: GroupDocList, Action: actionCloseAllFilter, Description: "close all filtered documents"},

	// Mark position
	{Group: GroupMark, Action: actionMark, Description: "mark current position"},
	{Group: GroupMark, Action: actionRemoveMark, Description: "remove mark current position"},
	{Group: GroupMark, Action: actionRemoveAllMark, Description: "remove all mark"},
	{Group: GroupMark, Action: actionMoveMark, Description: "move to next marked position"},
	{Group: GroupMark, Action: actionMovePrevMark, Description: "move to previous marked position"},
	{Group: GroupMark, Action: actionMarkByPattern, Description: "mark by pattern mode"},

	// Search
	{Group: GroupSearch, Action: actionSearch, Description: "forward search mode"},
	{Group: GroupSearch, Action: actionBackSearch, Description: "backward search mode"},
	{Group: GroupSearch, Action: actionNextSearch, Description: "repeat forward search"},
	{Group: GroupSearch, Action: actionNextBackSearch, Description: "repeat backward search"},
	{Group: GroupSearch, Action: actionFilter, Description: "filter search mode"},

	// Change display
	{Group: GroupChange, Action: actionWrap, Description: "wrap toggle (character based)"},
	{Group: GroupChange, Action: actionWordWrap, Description: "word wrap toggle"},
	{Group: GroupChange, Action: actionColumnMode, Description: "column mode toggle"},
	{Group: GroupChange, Action: actionColumnWidth, Description: "column width toggle"},
	{Group: GroupChange, Action: actionRainbow, Description: "column rainbow toggle"},
	{Group: GroupChange, Action: actionAlternate, Description: "alternate rows of style toggle"},
	{Group: GroupChange, Action: actionLineNumMode, Description: "line number toggle"},
	{Group: GroupChange, Action: actionPlain, Description: "original decoration toggle(plain)"},
	{Group: GroupChange, Action: actionAlignFormat, Description: "align columns"},
	{Group: GroupChange, Action: actionRawFormat, Description: "raw output"},
	{Group: GroupChange, Action: actionRuler, Description: "ruler toggle"},
	{Group: GroupChange, Action: actionStatusLine, Description: "status line toggle"},

	// Change Display with Input
	{Group: GroupChangeInput, Action: actionViewMode, Description: "view mode selection"},
	{Group: GroupChangeInput, Action: actionDelimiter, Description: "column delimiter string"},
	{Group: GroupChangeInput, Action: actionHeader, Description: "number of header lines"},
	{Group: GroupChangeInput, Action: actionSkipLines, Description: "number of skip lines"},
	{Group: GroupChangeInput, Action: actionTabWidth, Description: "TAB width"},
	{Group: GroupChangeInput, Action: actionMultiColor, Description: "multi color highlight"},
	{Group: GroupChangeInput, Action: actionJumpTarget, Description: "jump target(`.n` or `n%` or `section` allowed)"},
	{Group: GroupChangeInput, Action: actionConvertType, Description: "convert type selection"},
	{Group: GroupChangeInput, Action: actionVerticalHeader, Description: "number of vertical header"},
	{Group: GroupChangeInput, Action: actionHeaderColumn, Description: "number of header column"},

	// Column operation
	{Group: GroupColumn, Action: actionFixedColumn, Description: "header column fixed toggle"},
	{Group: GroupColumn, Action: actionShrinkColumn, Description: "shrink column toggle(align mode only)"},
	{Group: GroupColumn, Action: actionRightAlign, Description: "right align column toggle(align mode only)"},

	// Section operation
	{Group: GroupSection, Action: actionSection, Description: "section delimiter regular expression"},
	{Group: GroupSection, Action: actionSectionStart, Description: "section start position"},
	{Group: GroupSection, Action: actionNextSection, Description: "next section"},
	{Group: GroupSection, Action: actionPrevSection, Description: "previous section"},
	{Group: GroupSection, Action: actionLastSection, Description: "last section"},
	{Group: GroupSection, Action: actionFollowSection, Description: "follow section mode toggle"},
	{Group: GroupSection, Action: actionSectionNum, Description: "number of section header lines"},
	{Group: GroupSection, Action: actionHideOther, Description: `hide "other" section toggle`},

	// Close and reload
	{Group: GroupClose, Action: actionCloseFile, Description: "close file"},
	{Group: GroupClose, Action: actionReload, Description: "reload file"},
	{Group: GroupClose, Action: actionWatch, Description: "watch mode"},
	{Group: GroupClose, Action: actionWatchInterval, Description: "set watch interval"},

	// Key binding when typing
	{Group: GroupTyping, Action: inputCaseSensitive, Description: "case-sensitive toggle"},
	{Group: GroupTyping, Action: inputSmartCaseSensitive, Description: "smart case-sensitive toggle"},
	{Group: GroupTyping, Action: inputRegexpSearch, Description: "regular expression search toggle"},
	{Group: GroupTyping, Action: inputIncSearch, Description: "incremental search toggle"},
	{Group: GroupTyping, Action: inputNonMatch, Description: "non-match toggle"},
	{Group: GroupTyping, Action: inputPrevious, Description: "previous candidate"},
	{Group: GroupTyping, Action: inputNext, Description: "next candidate"},
	{Group: GroupTyping, Action: inputCopy, Description: "copy to clipboard"},
	{Group: GroupTyping, Action: inputPaste, Description: "paste from clipboard"},
}

// String returns keybind as a string for help.
func (k KeyBind) String() string {
	var b strings.Builder
	writeHeaderTh(&b)
	group := Group(-1)
	for _, bind := range keyBindDescriptions {
		if bind.Group != group {
			group = bind.Group
			writeHeader(&b, group.String())
		}
		keys := normalizeKeysForDisplay(k[bind.Action])
		b.WriteString(fmt.Sprintf(" %-30s * %s\n", "["+strings.Join(keys, "], [")+"]", bind.Description))
	}
	return b.String()
}

// normalizeKeysForDisplay normalizes a slice of key strings for display purposes.
func normalizeKeysForDisplay(keys []string) []string {
	normalized := make([]string, 0, len(keys))
	for _, key := range keys {
		normalizedKey, err := normalizeKey(key)
		if err != nil {
			normalized = append(normalized, key)
			continue
		}
		normalized = append(normalized, normalizedKey)
	}
	return normalized
}

// GetKeyBindDescriptions returns a slice of key binding descriptions for the specified group.
func (k KeyBind) GetKeyBindDescriptions(group Group) [][]string {
	var descriptions [][]string

	if group != -1 {
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

// writeHeaderTh writes the header for the key binding display to the provided writer.
func writeHeaderTh(w io.Writer) {
	fmt.Fprintf(w, " %-30s %s\n", "Key", "Action")
}

// writeHeader writes a section header for the key binding display to the provided writer.
func writeHeader(w io.Writer, header string) {
	fmt.Fprintf(w, "\n\t%s\n", header)
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

	switch strings.ToLower(config.DefaultKeyBind) {
	case "disable":
		// no default keybindings
	case "less":
		keyBind = LessKeyBinds()
	default:
		keyBind = DefaultKeyBinds()
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
		// sidebar operations are enabled even while typing.
		if strings.HasPrefix(name, "sidebar_") {
			if err := setHandler(ctx, in, name, keys, handler); err != nil {
				return err
			}
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

// formatKeyName formats the key name for better readability.
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

// normalizeKeyWithPrefix normalizes a key and adds input_ or sidebar_ prefix if the action is an input action.
func normalizeKeyWithPrefix(key, action string) (string, error) {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(action, "input_") {
		return "input_" + normalizedKey, nil
	}
	if strings.HasPrefix(action, "sidebar_") {
		return "sidebar_" + normalizedKey, nil
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
	for _, k := range slices.Sorted(maps.Keys(keyActions)) {
		mapping := keyActions[k]
		if len(mapping.action) > 1 {
			slices.Sort(mapping.action)
			duplicates = append(duplicates, mapping)
		}
	}

	return duplicates, errors
}
