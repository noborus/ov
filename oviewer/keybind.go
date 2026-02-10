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
	actionAlignFormat    = "align_format"
	actionRawFormat      = "raw_format"
	actionFixedColumn    = "fixed_column"
	actionShrinkColumn   = "shrink_column"
	actionRightAlign     = "right_align"
	actionRuler          = "toggle_ruler"
	actionWriteOriginal  = "write_original"

	// Sidebar actions.
	actionSidebarHelp    = "sidebar_help"
	actionSidebarMarks   = "sidebar_marks"
	actionSidebarDocList = "sidebar_doc_list"
	actionSidebarUp      = "sidebar_up"
	actionSidebarDown    = "sidebar_down"
	actionSidebarLeft    = "sidebar_left"
	actionSidebarRight   = "sidebar_right"

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
	actionMarkNumber     = "mark_number"
	actionMarkByPattern  = "mark_by_pattern"

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

		// Sidebar actions.
		actionSidebarHelp:    root.toggleSidebarHelp,
		actionSidebarMarks:   root.toggleSidebarMarks,
		actionSidebarDocList: root.toggleSidebarDocList,
		actionSidebarUp:      root.sidebarUp,
		actionSidebarDown:    root.sidebarDown,
		actionSidebarLeft:    root.sidebarLeft,
		actionSidebarRight:   root.sidebarRight,

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
		actionMarkNumber:     root.inputMarkNumber,
		actionMarkByPattern:  root.inputMarkByPattern,

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
		actionSidebarHelp:    {"alt+h"},
		actionSidebarMarks:   {"alt+m"},
		actionSidebarDocList: {"alt+l"},
		actionSidebarUp:      {"shift+Up"},
		actionSidebarDown:    {"shift+Down"},
		actionSidebarLeft:    {"shift+Left"},
		actionSidebarRight:   {"shift+Right"},

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
	{Group: GroupChange, Action: actionWrap, Description: "wrap/nowrap toggle"},
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
		keys := k[bind.Action]
		b.WriteString(fmt.Sprintf(" %-30s * %s\n", "["+strings.Join(keys, "], [")+"]", bind.Description))
	}
	return b.String()
}

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

func writeHeaderTh(w io.Writer) {
	fmt.Fprintf(w, " %-30s %s\n", "Key", "Action")
}

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
	for _, mapping := range keyActions {
		if len(mapping.action) > 1 {
			duplicates = append(duplicates, mapping)
		}
	}

	return duplicates, errors
}
