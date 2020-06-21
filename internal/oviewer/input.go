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

	mode    inputMode
	value   string
	reg     *regexp.Regexp
	cursorX int

	SearchCList    *Candidate
	GoCList        *Candidate
	DelimiterCList *Candidate
	TabWidthCList  *Candidate
}

// inputMode represents the state of the input.
type inputMode int

const (
	normal inputMode = iota
	search
	backsearch
	goline
	header
	delimiter
	tabWidth
)

// InputKeyEvent input key events.
func (root *Root) InputKeyEvent(ev *tcell.EventKey) bool {
	// inputEvent returns input confirmed or not confirmed.
	ok := root.inputKeyEvent(ev)

	// Not confirmed or canceled.
	if !ok {
		return true
	}

	input := root.Input
	// confirmed.
	nev := input.EventInput.Confirm(input.value)
	go func() {
		root.Screen.PostEventWait(nev)
	}()

	input.mode = normal
	return true
}

// inputKeyEvent handles the keystrokes of the input.
func (root *Root) inputKeyEvent(ev *tcell.EventKey) bool {
	input := root.Input

	switch ev.Key() {
	case tcell.KeyEscape:
		input.mode = normal
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
		input.value += string(runes[pos+1:])
	case tcell.KeyDelete:
		pos := stringWidth(input.value, input.cursorX)
		runes := []rune(input.value)
		dp := 1
		if input.cursorX == 0 {
			dp = 0
		}
		input.value = string(runes[:pos+dp])
		if len(runes) > pos+1 {
			input.value += string(runes[pos+dp+1:])
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

// Candidate represents a candidate list.
type Candidate struct {
	list []string
	p    int
}

// NewInput returns all the various inputs.
func NewInput() *Input {
	i := Input{}
	i.GoCList = &Candidate{
		list: []string{},
	}
	i.DelimiterCList = &Candidate{
		list: []string{
			"│",
			"\t",
			"|",
			",",
		},
	}
	i.TabWidthCList = &Candidate{
		list: []string{
			"3",
			"2",
			"4",
			"8",
		},
	}
	i.SearchCList = &Candidate{
		list: []string{},
	}
	i.EventInput = &NormalInput{}
	return &i
}

// SetMode changes the input mode.
func (input *Input) SetMode(mode inputMode) {
	input.mode = mode
	input.value = ""
	input.cursorX = 0

	switch mode {
	case search:
		input.EventInput = NewSearchInput(input.SearchCList)
	case backsearch:
		input.EventInput = NewBackSearchInput(input.SearchCList)
	case goline:
		input.EventInput = NewGotoInput(input.GoCList)
	case header:
		input.EventInput = NewHeaderInput()
	case delimiter:
		input.EventInput = NewDelimiterInput(input.DelimiterCList)
	case tabWidth:
		input.EventInput = NewTABWidthInput(input.TabWidthCList)
	default:
		input.EventInput = NewNormalInput()
	}
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

// NormalInput represents the normal input mode.
// This is a dummy as it normally does not accept input.
type NormalInput struct {
	tcell.EventTime
}

// NewNormalInput returns NormalInput.
func NewNormalInput() *NormalInput {
	return &NormalInput{}
}

// Prompt returns the prompt string in the input field.
func (n *NormalInput) Prompt() string {
	return ""
}

// Confirm returns the event when the input is confirmed.
func (n *NormalInput) Confirm(str string) tcell.Event {
	return nil
}

// Up returns strings when the up key is pressed during input.
func (n *NormalInput) Up(str string) string {
	return ""
}

// Down returns strings when the down key is pressed during input.
func (n *NormalInput) Down(str string) string {
	return ""
}

// SearchInput represents the search input mode.
type SearchInput struct {
	clist *Candidate
	tcell.EventTime
}

// NewSearchInput returns SearchInput.
func NewSearchInput(clist *Candidate) *SearchInput {
	return &SearchInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (s *SearchInput) Prompt() string {
	return "/"
}

// Confirm returns the event when the input is confirmed.
func (s *SearchInput) Confirm(str string) tcell.Event {
	s.clist.list = toLast(s.clist.list, str)
	s.clist.p = 0
	s.SetEventNow()
	return s
}

// Up returns strings when the up key is pressed during input.
func (s *SearchInput) Up(str string) string {
	return s.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (s *SearchInput) Down(str string) string {
	return s.clist.down()
}

// BackSearchInput represents the back search input mode.
type BackSearchInput struct {
	clist *Candidate
	tcell.EventTime
}

// NewBackSearchInput returns BackSearchInput.
func NewBackSearchInput(clist *Candidate) *BackSearchInput {
	return &BackSearchInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (b *BackSearchInput) Prompt() string {
	return "?"
}

// Confirm returns the event when the input is confirmed.
func (b *BackSearchInput) Confirm(str string) tcell.Event {
	b.clist.list = toLast(b.clist.list, str)
	b.clist.p = 0
	b.SetEventNow()
	return b
}

// Up returns strings when the up key is pressed during input.
func (b *BackSearchInput) Up(str string) string {
	return b.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (b *BackSearchInput) Down(str string) string {
	return b.clist.down()
}

// GotoInput represents the goto input mode.
type GotoInput struct {
	clist *Candidate
	tcell.EventTime
}

// NewGotoInput returns GotoInput.
func NewGotoInput(clist *Candidate) *GotoInput {
	return &GotoInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (g *GotoInput) Prompt() string {
	return "Goto line:"
}

// Confirm returns the event when the input is confirmed.
func (g *GotoInput) Confirm(str string) tcell.Event {
	g.clist.list = toLast(g.clist.list, str)
	g.clist.p = 0
	g.SetEventNow()
	return g
}

// Up returns strings when the up key is pressed during input.
func (g *GotoInput) Up(str string) string {
	return g.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (g *GotoInput) Down(str string) string {
	return g.clist.down()
}

// HeaderInput represents the goto input mode.
type HeaderInput struct {
	tcell.EventTime
}

// NewHeaderInput returns HeaderInput.
func NewHeaderInput() *HeaderInput {
	return &HeaderInput{}
}

// Prompt returns the prompt string in the input field.
func (h *HeaderInput) Prompt() string {
	return "Header length:"
}

// Confirm returns the event when the input is confirmed.
func (h *HeaderInput) Confirm(str string) tcell.Event {
	h.SetEventNow()
	return h
}

// Up returns strings when the up key is pressed during input.
func (h *HeaderInput) Up(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n + 1)
}

// Down returns strings when the down key is pressed during input.
func (h *HeaderInput) Down(str string) string {
	n, err := strconv.Atoi(str)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n - 1)
}

// DelimiterInput represents the delimiter input mode.
type DelimiterInput struct {
	clist *Candidate
	tcell.EventTime
}

// NewDelimiterInput returns DelimiterInput.
func NewDelimiterInput(clist *Candidate) *DelimiterInput {
	return &DelimiterInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (d *DelimiterInput) Prompt() string {
	return "Delimiter:"
}

// Confirm returns the event when the input is confirmed.
func (d *DelimiterInput) Confirm(str string) tcell.Event {
	d.clist.list = toLast(d.clist.list, str)
	d.clist.p = 0
	d.SetEventNow()
	return d
}

// Up returns strings when the up key is pressed during input.
func (d *DelimiterInput) Up(str string) string {
	return d.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (d *DelimiterInput) Down(str string) string {
	return d.clist.down()
}

// TABWidthInput represents the TABWidth input mode.
type TABWidthInput struct {
	clist *Candidate
	tcell.EventTime
}

// NewTABWidthInput returns TABWidthInput.
func NewTABWidthInput(clist *Candidate) *TABWidthInput {
	return &TABWidthInput{clist: clist}
}

// Prompt returns the prompt string in the input field.
func (t *TABWidthInput) Prompt() string {
	return "TAB width:"
}

// Confirm returns the event when the input is confirmed.
func (t *TABWidthInput) Confirm(str string) tcell.Event {
	t.clist.list = toLast(t.clist.list, str)
	t.clist.p = 0
	t.SetEventNow()
	return t
}

// Up returns strings when the up key is pressed during input.
func (t *TABWidthInput) Up(str string) string {
	return t.clist.up()
}

// Down returns strings when the down key is pressed during input.
func (t *TABWidthInput) Down(str string) string {
	return t.clist.down()
}

func (c *Candidate) up() string {
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

func (c *Candidate) down() string {
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
