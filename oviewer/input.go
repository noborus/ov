package oviewer

import (
	"context"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// InputMode represents the state of the input.
type InputMode int

const (
	// Normal is normal mode.
	Normal           InputMode = iota
	ViewMode                   // ViewMode is a view selection input mode.
	Search                     // Search is a search input mode.
	Backsearch                 // Backsearch is a backward search input mode.
	Goline                     // Goline is a move input mode.
	Header                     // Header is the number of headers input mode.
	Delimiter                  // Delimiter is a delimiter input mode.
	TabWidth                   // TabWidth is the tab number input mode.
	Watch                      // Watch is the watch interval input mode.
	SkipLines                  // SkipLines is the number of lines to skip.
	WriteBA                    // WriteBA is the number of ranges to write at quit.
	SectionDelimiter           // SectionDelimiter is a section delimiter input mode.
	SectionStart               // SectionStart is a section start position input mode.
	MultiColor                 // MultiColor is multi-word coloring.
	JumpTarget                 // JumpTarget is the position to display the search results.
	SaveBuffer                 // SaveBuffer is the save buffer.
	SectionNum                 // SectionNum is the section number.
)

// Input represents the status of various inputs.
// Retain each input list to save the input history.
type Input struct {
	Event Eventer

	// Candidate is prepared when the history is used as an input candidate.
	// Header and SkipLines use numbers up and down instead of candidate.
	DelimiterCandidate    *candidate
	ModeCandidate         *candidate
	SearchCandidate       *candidate
	GoCandidate           *candidate
	TabWidthCandidate     *candidate
	WatchCandidate        *candidate
	WriteBACandidate      *candidate
	SectionDelmCandidate  *candidate
	SectionStartCandidate *candidate
	MultiColorCandidate   *candidate
	JumpTargetCandidate   *candidate
	SaveBufferCandidate   *candidate

	value   string
	cursorX int
}

// NewInput returns all the various inputs.
func NewInput() *Input {
	i := Input{}

	i.ModeCandidate = viewModeCandidate()
	i.SearchCandidate = searchCandidate()
	i.GoCandidate = gotoCandidate()
	i.DelimiterCandidate = delimiterCandidate()
	i.TabWidthCandidate = tabWidthCandidate()
	i.WatchCandidate = watchCandidate()
	i.WriteBACandidate = baCandidate()
	i.SectionDelmCandidate = sectionDelimiterCandidate()
	i.SectionStartCandidate = sectionStartCandidate()
	i.MultiColorCandidate = multiColorCandidate()
	i.JumpTargetCandidate = jumpTargetCandidate()
	i.SaveBufferCandidate = saveBufferCandidate()

	i.Event = &eventNormal{}
	return &i
}

// InputEvent input key events.
func (root *Root) inputEvent(ctx context.Context, ev *tcell.EventKey) {
	// inputEvent returns input confirmed or not confirmed.
	// Not confirmed or canceled.
	evKey := root.inputKeyConfig.Capture(ev)
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

// keyEvent handles the keystrokes of the input.
func (input *Input) keyEvent(evKey *tcell.EventKey) bool {
	if evKey == nil {
		return false
	}
	switch evKey.Key() {
	case tcell.KeyEscape:
		input.value = ""
		input.Event = normal()
		return false
	case tcell.KeyEnter:
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if input.cursorX <= 0 {
			return false
		}
		pos := countToCursor(input.value, input.cursorX)
		runes := []rune(input.value)
		input.value = string(runes[:pos])
		input.cursorX = stringWidth(input.value)
		next := pos + 1
		for ; next < len(runes); next++ {
			if runewidth.RuneWidth(runes[next]) != 0 {
				break
			}
		}
		input.value += string(runes[next:])
	case tcell.KeyDelete:
		pos := countToCursor(input.value, input.cursorX)
		runes := []rune(input.value)
		dp := 1
		if input.cursorX == 0 {
			dp = 0
		}
		input.value = string(runes[:pos+dp])
		next := pos + 1
		for ; next < len(runes); next++ {
			if runewidth.RuneWidth(runes[next]) != 0 {
				break
			}
		}
		if len(runes) > next {
			input.value += string(runes[dp+next:])
		}
	case tcell.KeyLeft:
		if input.cursorX <= 0 {
			return false
		}
		pos := countToCursor(input.value, input.cursorX)
		runes := []rune(input.value)
		input.cursorX = stringWidth(string(runes[:pos]))
		if pos > 0 && runes[pos-1] == '\t' {
			input.cursorX--
		}
	case tcell.KeyRight:
		pos := countToCursor(input.value, input.cursorX+1)
		runes := []rune(input.value)
		if len(runes) > pos {
			input.cursorX = stringWidth(string(runes[:pos+1]))
		}
	case tcell.KeyTAB:
		pos := countToCursor(input.value, input.cursorX+1)
		runes := []rune(input.value)
		input.value = string(runes[:pos])
		input.value += "\t"
		input.cursorX += 2
		input.value += string(runes[pos:])
	case tcell.KeyRune:
		pos := countToCursor(input.value, input.cursorX+1)
		runes := []rune(input.value)
		input.value = string(runes[:pos])
		r := evKey.Rune()
		input.value += string(r)
		input.value += string(runes[pos:])
		input.cursorX += runewidth.RuneWidth(r)
	}
	return false
}

// inputCaseSensitive toggles case sensitivity.
func (root *Root) inputCaseSensitive() {
	root.Config.CaseSensitive = !root.Config.CaseSensitive
	if root.Config.CaseSensitive {
		root.Config.SmartCaseSensitive = false
	}
}

// inputSmartCaseSensitive toggles case sensitivity.
func (root *Root) inputSmartCaseSensitive() {
	root.Config.SmartCaseSensitive = !root.Config.SmartCaseSensitive
	if root.Config.SmartCaseSensitive {
		root.Config.CaseSensitive = false
	}
}

// inputIncSearch toggles incremental search.
func (root *Root) inputIncSearch() {
	root.Config.Incsearch = !root.Config.Incsearch
}

// inputRegexpSearch toggles regexp search.
func (root *Root) inputRegexpSearch() {
	root.Config.RegexpSearch = !root.Config.RegexpSearch
}

// inputPrevious searches the previous history.
func (root *Root) inputPrevious() {
	input := root.input
	input.value = input.Event.Up(input.value)
	runes := []rune(input.value)
	input.cursorX = stringWidth(string(runes))
}

// inputNext searches the next history.
func (root *Root) inputNext() {
	input := root.input
	input.value = input.Event.Down(input.value)
	runes := []rune(input.value)
	input.cursorX = stringWidth(string(runes))
}

// stringWidth returns the number of widths of the input.
// Tab is 2 characters.
func stringWidth(str string) int {
	width := 0
	for _, r := range str {
		width += runewidth.RuneWidth(r)
		if r == '\t' {
			width += 2
		}
	}
	return width
}

// countToCursor returns the number of characters in the input.
func countToCursor(str string, cursor int) int {
	width := 0
	i := 0
	for _, r := range str {
		width += runewidth.RuneWidth(r)
		if r == '\t' {
			width += 2
		}
		if width >= cursor {
			return i
		}
		i++
	}
	return i
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

func (c *candidate) toLast(str string) {
	if str == "" {
		return
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.list = toLast(c.list, str)
	c.p = 0
}

func (c *candidate) toAddTop(str string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.list = toAddTop(c.list, str)
}

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
