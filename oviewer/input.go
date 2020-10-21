package oviewer

import (
	"regexp"
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

// Input represents the status of various inputs.
// Retain each input list to save the input history.
type Input struct {
	EventInput EventInput

	mode    InputMode
	value   string
	reg     *regexp.Regexp
	cursorX int

	SearchCandidate    *candidate
	GoCandidate        *candidate
	DelimiterCandidate *candidate
	TabWidthCandidate  *candidate
}

// InputMode represents the state of the input.
type InputMode int

const (
	// Normal is normal mode.
	Normal InputMode = iota
	// Help is Help screen mode.
	Help
	// LogDoc is Error screen mode.
	LogDoc
	// Search is a search input mode.
	Search
	// Backsearch is a backward search input mode.
	Backsearch
	// Goline is a move input mode.
	Goline
	// Header is the number of headers input mode.
	Header
	// Delimiter is a delimiter input mode.
	Delimiter
	// TabWidth is the tab number input mode.
	TabWidth
)

// InputEvent input key events.
func (root *Root) inputEvent(ev *tcell.EventKey) {
	// inputEvent returns input confirmed or not confirmed.
	ok := root.inputKeyEvent(ev)

	// Not confirmed or canceled.
	if !ok {
		return
	}

	input := root.input
	// confirmed.
	nev := input.EventInput.Confirm(input.value)
	go func() {
		root.Screen.PostEventWait(nev)
	}()

	input.mode = Normal
	input.EventInput = newNormalInput()
}

// inputKeyEvent handles the keystrokes of the input.
func (root *Root) inputKeyEvent(ev *tcell.EventKey) bool {
	input := root.input

	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlG, tcell.KeyCtrlC:
		input.mode = Normal
		return false
	case tcell.KeyEnter:
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if input.cursorX <= 0 {
			return false
		}
		pos := stringWidth(input.value, input.cursorX)
		runes := []rune(input.value)
		input.value = string(runes[:pos])
		input.cursorX = runeWidth(input.value)
		next := pos + 1
		for ; next < len(runes); next++ {
			if runewidth.RuneWidth(runes[next]) != 0 {
				break
			}
		}
		input.value += string(runes[next:])
	case tcell.KeyDelete:
		pos := stringWidth(input.value, input.cursorX)
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
		pos := stringWidth(input.value, input.cursorX)
		runes := []rune(input.value)
		input.cursorX = runeWidth(string(runes[:pos]))
		if pos > 0 && runes[pos-1] == '\t' {
			input.cursorX--
		}
	case tcell.KeyRight:
		pos := stringWidth(input.value, input.cursorX+1)
		runes := []rune(input.value)
		input.cursorX = runeWidth(string(runes[:pos+1]))
	case tcell.KeyUp:
		input.value = input.EventInput.Up(input.value)
		runes := []rune(input.value)
		input.cursorX = runeWidth(string(runes))
	case tcell.KeyDown:
		input.value = input.EventInput.Down(input.value)
		runes := []rune(input.value)
		input.cursorX = runeWidth(string(runes))
	case tcell.KeyTAB:
		pos := stringWidth(input.value, input.cursorX+1)
		runes := []rune(input.value)
		input.value = string(runes[:pos])
		input.value += "\t"
		input.cursorX += 2
		input.value += string(runes[pos:])
	case tcell.KeyCtrlA:
		root.CaseSensitive = !root.CaseSensitive
	case tcell.KeyRune:
		pos := stringWidth(input.value, input.cursorX+1)
		runes := []rune(input.value)
		input.value = string(runes[:pos])
		r := ev.Rune()
		input.value += string(r)
		input.value += string(runes[pos:])
		input.cursorX += runewidth.RuneWidth(r)
	}
	return false
}

// stringWidth returns the number of characters in the input.
func stringWidth(str string, cursor int) int {
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

// runeWidth returns the number of widths of the input.
func runeWidth(str string) int {
	width := 0
	for _, r := range str {
		width += runewidth.RuneWidth(r)
		if r == '\t' {
			width += 2
		}
	}
	return width
}

// candidate represents a input candidate list.
type candidate struct {
	list []string
	p    int
}

// NewInput returns all the various inputs.
func NewInput() *Input {
	i := Input{}
	i.GoCandidate = &candidate{
		list: []string{},
	}
	i.DelimiterCandidate = &candidate{
		list: []string{
			"│",
			"\t",
			"|",
			",",
		},
	}
	i.TabWidthCandidate = &candidate{
		list: []string{
			"3",
			"2",
			"4",
			"8",
		},
	}
	i.SearchCandidate = &candidate{
		list: []string{},
	}
	i.EventInput = &normalInput{}
	return &i
}

func (root *Root) setSearchMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = Search
	input.EventInput = newSearchInput(input.SearchCandidate)
}

func (root *Root) setBackSearchMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = Search
	input.EventInput = newBackSearchInput(input.SearchCandidate)
}

func (root *Root) setDelimiterMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = Delimiter
	input.EventInput = newDelimiterInput(input.DelimiterCandidate)
}

func (root *Root) setHeaderMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = Header
	input.EventInput = newHeaderInput()
}

func (root *Root) setTabWidthMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = TabWidth
	input.EventInput = newTabWidthInput(input.TabWidthCandidate)
}

func (root *Root) setGoLineMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = Goline
	input.EventInput = newGotoInput(input.GoCandidate)
}

// EventInput is a generic interface for inputs.
type EventInput interface {
	// Prompt returns the prompt string in the input field.
	Prompt() string
	// Confirm returns the event when the input is confirmed.
	Confirm(i string) tcell.Event
	// Up returns strings when the up key is pressed during input.
	Up(i string) string
	// Down returns strings when the down key is pressed during input.
	Down(i string) string
}

// normalInput represents the normal input mode.
// This is a dummy as it normally does not accept input.
type normalInput struct {
	tcell.EventTime
}

// newNormalInput returns normalInput.
func newNormalInput() *normalInput {
	return &normalInput{}
}

// Prompt returns the prompt string in the input field.
func (n *normalInput) Prompt() string {
	return ""
}

// Confirm returns the event when the input is confirmed.
func (n *normalInput) Confirm(str string) tcell.Event {
	return nil
}

// Up returns strings when the up key is pressed during input.
func (n *normalInput) Up(str string) string {
	return ""
}

// Down returns strings when the down key is pressed during input.
func (n *normalInput) Down(str string) string {
	return ""
}

// searchInput represents the search input mode.
type searchInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newSearchInput returns SearchInput.
func newSearchInput(clist *candidate) *searchInput {
	return &searchInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (s *searchInput) Prompt() string {
	return "/"
}

// Confirm returns the event when the input is confirmed.
func (s *searchInput) Confirm(str string) tcell.Event {
	s.value = str
	s.clist.list = toLast(s.clist.list, str)
	s.clist.p = 0
	s.SetEventNow()
	return s
}

// Up returns strings when the up key is pressed during input.
func (s *searchInput) Up(str string) string {
	return s.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (s *searchInput) Down(str string) string {
	return s.clist.down()
}

// backSearchInput represents the back search input mode.
type backSearchInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newBackSearchInput returns BackSearchInput.
func newBackSearchInput(clist *candidate) *backSearchInput {
	return &backSearchInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (b *backSearchInput) Prompt() string {
	return "?"
}

// Confirm returns the event when the input is confirmed.
func (b *backSearchInput) Confirm(str string) tcell.Event {
	b.value = str
	b.clist.list = toLast(b.clist.list, str)
	b.clist.p = 0
	b.SetEventNow()
	return b
}

// Up returns strings when the up key is pressed during input.
func (b *backSearchInput) Up(str string) string {
	return b.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (b *backSearchInput) Down(str string) string {
	return b.clist.down()
}

// gotoInput represents the goto input mode.
type gotoInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newGotoInput returns GotoInput.
func newGotoInput(clist *candidate) *gotoInput {
	return &gotoInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (g *gotoInput) Prompt() string {
	return "Goto line:"
}

// Confirm returns the event when the input is confirmed.
func (g *gotoInput) Confirm(str string) tcell.Event {
	g.value = str
	g.clist.list = toLast(g.clist.list, str)
	g.clist.p = 0
	g.SetEventNow()
	return g
}

// Up returns strings when the up key is pressed during input.
func (g *gotoInput) Up(str string) string {
	return g.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (g *gotoInput) Down(str string) string {
	return g.clist.down()
}

// headerInput represents the goto input mode.
type headerInput struct {
	value string
	tcell.EventTime
}

// newHeaderInput returns HeaderInput.
func newHeaderInput() *headerInput {
	return &headerInput{}
}

// Prompt returns the prompt string in the input field.
func (h *headerInput) Prompt() string {
	return "Header length:"
}

// Confirm returns the event when the input is confirmed.
func (h *headerInput) Confirm(str string) tcell.Event {
	h.value = str
	h.SetEventNow()
	return h
}

// Up returns strings when the up key is pressed during input.
func (h *headerInput) Up(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n + 1)
}

// Down returns strings when the down key is pressed during input.
func (h *headerInput) Down(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n - 1)
}

// delimiterInput represents the delimiter input mode.
type delimiterInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newDelimiterInput returns DelimiterInput.
func newDelimiterInput(clist *candidate) *delimiterInput {
	return &delimiterInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (d *delimiterInput) Prompt() string {
	return "Delimiter:"
}

// Confirm returns the event when the input is confirmed.
func (d *delimiterInput) Confirm(str string) tcell.Event {
	d.value = str
	d.clist.list = toLast(d.clist.list, str)
	d.clist.p = 0
	d.SetEventNow()
	return d
}

// Up returns strings when the up key is pressed during input.
func (d *delimiterInput) Up(str string) string {
	return d.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (d *delimiterInput) Down(str string) string {
	return d.clist.down()
}

// tabWidthInput represents the TABWidth input mode.
type tabWidthInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newTabWidthInput returns TABWidthInput.
func newTabWidthInput(clist *candidate) *tabWidthInput {
	return &tabWidthInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (t *tabWidthInput) Prompt() string {
	return "TAB width:"
}

// Confirm returns the event when the input is confirmed.
func (t *tabWidthInput) Confirm(str string) tcell.Event {
	t.value = str
	t.clist.list = toLast(t.clist.list, str)
	t.clist.p = 0
	t.SetEventNow()
	return t
}

// Up returns strings when the up key is pressed during input.
func (t *tabWidthInput) Up(str string) string {
	return t.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (t *tabWidthInput) Down(str string) string {
	return t.clist.down()
}

func (c *candidate) up() string {
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

func (c *candidate) down() string {
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

func toLast(list []string, s string) []string {
	if len(s) == 0 {
		return list
	}

	for n, l := range list {
		if l == s {
			list = append(list[:n], list[n+1:]...)
		}
	}

	list = append(list, s)
	return list
}
