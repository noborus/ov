package oviewer

import (
	"context"
	"log"
	"strconv"
	"strings"
	"sync"

	"codeberg.org/tslocum/cbind"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// InputMode represents the state of the input.
type InputMode int

const (
	// Normal is the default mode where no special input handling is active.
	Normal InputMode = iota
	// ViewMode is for view selection.
	ViewMode
	// Search is for searching.
	Search
	// Backsearch is for backward searching.
	Backsearch
	// Filter is for filtering.
	Filter
	// MarkByPattern is for marking by pattern.
	MarkByPattern
	// Goline is for moving to a specific line.
	Goline
	// Header is for setting the number of headers.
	Header
	// Delimiter is for setting the delimiter.
	Delimiter
	// TabWidth is for setting the tab width.
	TabWidth
	// Watch is for setting the watch interval.
	Watch
	// SkipLines is for setting the number of lines to skip.
	SkipLines
	// WriteBA is for setting the number of ranges to write at quit.
	WriteBA
	// SectionDelimiter is for setting the section delimiter.
	SectionDelimiter
	// SectionStart is for setting the section start position.
	SectionStart
	// MultiColor is for multi-word coloring.
	MultiColor
	// JumpTarget is for setting the position to display the search results.
	JumpTarget
	// SaveBuffer is for saving the buffer.
	SaveBuffer
	// SectionNum is for setting the section number.
	SectionNum
	// ConvertType is for setting the convert type.
	ConvertType
	// VerticalHeader is for setting the number of vertical headers.
	VerticalHeader
	// HeaderColumn is for setting the number of vertical header columns.
	HeaderColumn
	// MarkNum is for setting the mark number.
	MarkNum
)

// Input represents the status of various inputs.
// Retain each input list to save the input history.
type Input struct {
	Event Eventer

	// Candidate is prepared when the history is used as an input candidate.
	// Header and SkipLines use numbers up and down instead of candidate.
	Candidate map[InputMode]*candidate

	value   string
	cursorX int
}

// NewInput returns all the various inputs.
func NewInput() *Input {
	i := Input{}

	// Initialize the candidate list.
	i.Candidate = make(map[InputMode]*candidate)
	i.Candidate[ViewMode] = viewModeCandidate()
	i.Candidate[Search] = blankCandidate()
	i.Candidate[Goline] = blankCandidate()
	i.Candidate[Delimiter] = delimiterCandidate()
	i.Candidate[TabWidth] = tabWidthCandidate()
	i.Candidate[Watch] = watchCandidate()
	i.Candidate[WriteBA] = blankCandidate()
	i.Candidate[SectionDelimiter] = sectionDelimiterCandidate()
	i.Candidate[SectionStart] = sectionStartCandidate()
	i.Candidate[MultiColor] = multiColorCandidate()
	i.Candidate[JumpTarget] = jumpTargetCandidate()
	i.Candidate[SaveBuffer] = blankCandidate()
	i.Candidate[ConvertType] = converterCandidate()
	i.Candidate[MarkNum] = blankCandidate()

	i.Event = &eventNormal{}
	return &i
}

// InputEvent input key events.
func (root *Root) inputEvent(ctx context.Context, ev *tcell.EventKey) {
	// inputEvent returns input confirmed or not confirmed.
	// Not confirmed or canceled.
	evKey := root.inputCapture(ev)
	if ok := root.input.keyEvent(evKey); !ok {
		root.incrementalSearch(ctx)
		return
	}

	if root.cancelFunc != nil {
		root.cancelFunc()
		root.cancelFunc = nil
	}

	// Fires a confirmed event.
	input := root.input
	nev := input.Event.Confirm(input.value)
	root.postEvent(nev)
	input.Event = normal()
}

// inputCapture processes the given key event and returns the captured event.
func (root *Root) inputCapture(ev *tcell.EventKey) *tcell.EventKey {
	if root.shouldCaptureExclamationMark(ev) {
		return ev
	}
	return root.inputKeyConfig.Capture(ev)
}

// shouldCaptureExclamationMark checks if the current input mode is Filter and the key event is '!'.
// It returns true if the '!' key should be captured and not input, otherwise false.
func (root *Root) shouldCaptureExclamationMark(ev *tcell.EventKey) bool {
	if ev.Str() == "!" {
		if root.input.Event.Mode() != Filter {
			return true
		}
		if len(root.input.value) > 0 && root.input.value[len(root.input.value)-1] == '\\' {
			return true
		}
	}
	return false
}

// keyEvent handles the keystrokes of the input.
func (input *Input) keyEvent(evKey *tcell.EventKey) bool {
	if evKey == nil {
		return false
	}

	mod := evKey.Modifiers()
	key := evKey.Key()
	str := evKey.Str()

	if key == tcell.KeyRune {
		left, right := splitAtWidth(input.value, input.cursorX)
		input.value = left + str + right
		input.cursorX += uniseg.StringWidth(str)
		actualWidth := stringWidth(input.value)
		if input.cursorX > actualWidth {
			input.cursorX = actualWidth
		}
		return false
	}

	keyStr, err := cbind.Encode(mod, key, str)
	if err != nil {
		log.Printf("Error encoding key event: %v", err)
		return false
	}
	switch strings.ToLower(keyStr) {
	case "escape":
		input.value = ""
		input.Event = normal()
	case "enter":
		return true
	case "backspace", "backspace2", "ctrl+h":
		if input.cursorX > 0 {
			input.value, input.cursorX = deleteAtWidth(input.value, input.cursorX, true)
		}
	case "delete":
		input.value, _ = deleteAtWidth(input.value, input.cursorX, false)
	case "left":
		if input.cursorX > 0 {
			input.cursorX = cursorLeft(input.value, input.cursorX)
		}
	case "right":
		newCursor := cursorRight(input.value, input.cursorX)
		if newCursor != input.cursorX {
			input.cursorX = newCursor
		}
	case "tab":
		left, right := splitAtWidth(input.value, input.cursorX)
		input.value = left + "\t" + right
		input.cursorX += 2
	}
	return false
}

// stringWidth returns the number of widths of the input.
// Tab is 2 characters.
func stringWidth(str string) int {
	tabCount := strings.Count(str, "\t")
	return uniseg.StringWidth(str) + (tabCount * 2)
}

// graphemeWidth returns the display width of a grapheme.
// Tab characters are counted as 2, all other graphemes use uniseg.StringWidth.
func graphemeWidth(gr *uniseg.Graphemes) int {
	r := gr.Runes()
	if len(r) == 1 && r[0] == '\t' {
		return 2
	}
	return uniseg.StringWidth(gr.Str())
}

// splitAtWidth splits the string at the specified display width.
// It returns the left part (up to targetWidth) and the right part (after targetWidth).
func splitAtWidth(s string, targetWidth int) (string, string) {
	bytePos := 0
	width := 0
	gr := uniseg.NewGraphemes(s)

	for gr.Next() {
		if width >= targetWidth {
			break
		}
		gw := graphemeWidth(gr)
		if width+gw > targetWidth {
			break
		}
		width += gw
		bytePos += len(gr.Bytes())
	}
	return s[:bytePos], s[bytePos:]
}

// cursorLeft returns the cursor position moved left by one grapheme.
func cursorLeft(s string, cursorWidth int) int {
	width := 0
	prevWidth := 0
	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		gw := graphemeWidth(gr)
		if width >= cursorWidth {
			break
		}
		prevWidth = width
		width += gw
	}
	return prevWidth
}

// cursorRight returns the cursor position moved right by one grapheme.
// If the cursor is already at the end, it returns the current position.
func cursorRight(s string, cursorWidth int) int {
	width := 0
	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		gw := graphemeWidth(gr)
		if width >= cursorWidth {
			return width + gw
		}
		width += gw
	}
	return cursorWidth // Already at the end
}

// deleteAtWidth deletes a grapheme at the cursor position.
// If backspace is true, it deletes the grapheme before the cursor (Backspace behavior).
// If backspace is false, it deletes the grapheme at the cursor (Delete behavior).
// It returns the new string and the new cursor position.
func deleteAtWidth(s string, cursorWidth int, backspace bool) (string, int) {
	var result strings.Builder
	width := 0
	newCursorWidth := cursorWidth
	deleted := false

	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		grapheme := gr.Str()
		gw := graphemeWidth(gr)

		if backspace {
			// Backspace: delete the grapheme before cursor
			if !deleted && width+gw >= cursorWidth && width < cursorWidth {
				// This is the grapheme to delete
				newCursorWidth = width
				deleted = true
				continue
			}
		} else {
			// Delete: delete the grapheme at cursor
			if !deleted && width >= cursorWidth {
				// This is the grapheme to delete
				deleted = true
				continue
			}
		}

		result.WriteString(grapheme)
		width += gw
	}

	return result.String(), newCursorWidth
}

// reset resets the input.
func (input *Input) reset() {
	input.value = ""
	input.cursorX = 0
}

func (input *Input) previous() {
	input.value = input.Event.Up(input.value)
	input.cursorX = stringWidth(input.value)
}

func (input *Input) next() {
	input.value = input.Event.Down(input.value)
	input.cursorX = stringWidth(input.value)
}

// toggleCaseSensitive toggles case sensitivity.
func (root *Root) toggleCaseSensitive(context.Context) {
	root.Config.CaseSensitive = !root.Config.CaseSensitive
	if root.Config.CaseSensitive {
		root.Config.SmartCaseSensitive = false
	}
	root.setPromptOpt()
}

// toggleSmartCaseSensitive toggles case sensitivity.
func (root *Root) toggleSmartCaseSensitive(context.Context) {
	root.Config.SmartCaseSensitive = !root.Config.SmartCaseSensitive
	if root.Config.SmartCaseSensitive {
		root.Config.CaseSensitive = false
	}
	root.setPromptOpt()
}

// toggleIncSearch toggles incremental search.
func (root *Root) toggleIncSearch(context.Context) {
	root.Config.Incsearch = !root.Config.Incsearch
	root.setPromptOpt()
}

// toggleRegexpSearch toggles regexp search.
func (root *Root) toggleRegexpSearch(context.Context) {
	root.Config.RegexpSearch = !root.Config.RegexpSearch
	root.setPromptOpt()
}

func (root *Root) toggleNonMatch(context.Context) {
	root.Doc.nonMatch = !root.Doc.nonMatch
	root.setPromptOpt()
}

// candidatePrevious searches the previous history.
func (root *Root) candidatePrevious(context.Context) {
	root.input.previous()
}

// candidateNext searches the next history.
func (root *Root) candidateNext(context.Context) {
	root.input.next()
}

// Eventer is a generic interface for inputs.
type Eventer interface {
	// Mode returns the input mode.
	Mode() InputMode
	// Prompt returns the prompt string in the input field.
	Prompt() string
	// Confirm returns the event when the input is confirmed.
	Confirm(i string) tcell.Event
	// Up returns strings when the up key is pressed during input.
	Up(i string) string
	// Down returns strings when the down key is pressed during input.
	Down(i string) string
}

// candidate represents a input candidate list.
type candidate struct {
	mux  sync.Mutex
	list []string
	p    int
}

// toLast returns the candidate list with the specified string at the end.
func (c *candidate) toLast(str string) {
	if str == "" {
		return
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.list = toLast(c.list, str)
	c.p = 0
}

// toAddTop returns the candidate list with the specified string at the top.
func (c *candidate) toAddTop(str string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.list = toAddTop(c.list, str)
}

// toAddLast returns the candidate list with the specified string at the end.
func (c *candidate) toAddLast(str string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.list = toAddLast(c.list, str)
}

// up returns the previous candidate.
func (c *candidate) up() string {
	c.mux.Lock()
	defer c.mux.Unlock()
	if len(c.list) == 0 {
		return ""
	}

	if c.p > 0 {
		c.p--
		return c.list[c.p]
	}

	c.p = len(c.list) - 1
	return c.list[c.p]
}

// down returns the next candidate.
func (c *candidate) down() string {
	c.mux.Lock()
	defer c.mux.Unlock()
	if len(c.list) == 0 {
		return ""
	}

	if len(c.list) > c.p+1 {
		c.p++
		return c.list[c.p]
	}

	c.p = 0
	return c.list[c.p]
}

// blankCandidate returns the candidate to set to default.
func blankCandidate() *candidate {
	return &candidate{
		list: []string{},
	}
}

// upNum returns the incremented number as a string; if the input is not a valid integer, it returns "0".
// "0" is returned when the input cannot be parsed as an integer, indicating an invalid or empty input.
func upNum(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n + 1)
}

// downNum returns the number of the next candidate.
// If the input string is not a valid integer or the value is less than or equal to 0,
// it returns "0" to prevent negative or invalid numbers.
func downNum(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil || n <= 0 {
		return "0"
	}
	return strconv.Itoa(n - 1)
}
