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
	previous
	goline
	header
	delimiter
	tabWidth
)

// Candidate represents a candidate list.
type Candidate struct {
	list []string
	p    int
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

// NewInput returns all the various inputs.
func NewInput() *Input {
	i := Input{}
	i.GoCList = &Candidate{
		list: []string{
			"1000",
			"100",
			"10",
			"0",
		},
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

// HandleInputEvent handles input events.
func (root *Root) HandleInputEvent(ev *tcell.EventKey) bool {
	// inputEvent returns input confirmed or not confirmed.
	ok := root.inputKeyEvent(ev)

	// Not confirmed or canceled.
	if !ok {
		return true
	}

	// confirmed.
	input := root.Input
	nev := input.EventInput.Confirm(input.value)
	go func() { root.Screen.PostEventWait(nev) }()

	input.mode = normal
	return true
}

// inputKeyEvent handles the keystrokes of the input.
func (root *Root) inputKeyEvent(ev *tcell.EventKey) bool {
	input := root.Input
	switch ev.Key() {
	case tcell.KeyEscape:
		input.mode = normal
		return true
	case tcell.KeyEnter:
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if input.cursorX == 0 {
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
		if input.cursorX > 0 {
			pos := stringWidth(input.value, input.cursorX)
			runes := []rune(input.value)
			if pos >= 0 {
				input.cursorX = runeWidth(string(runes[:pos]))
				if pos > 0 && runes[pos-1] == '\t' {
					input.cursorX--
				}
			}
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

func (input *Input) setMode(mode inputMode) {
	input.mode = mode
	input.value = ""
	input.cursorX = 0
	switch mode {
	case search:
		input.EventInput = &SearchInput{
			clist: input.SearchCList,
		}
	case previous:
		input.EventInput = &PreviousInput{
			clist: input.SearchCList,
		}
	case goline:
		input.EventInput = &GotoInput{
			clist: input.GoCList,
		}
	case header:
		input.EventInput = &HeaderInput{}
	case delimiter:
		input.EventInput = &DelimiterInput{
			clist: input.DelimiterCList,
		}
	case tabWidth:
		input.EventInput = &TABWidthInput{
			clist: input.TabWidthCList,
		}
	default:
		input.EventInput = &NormalInput{}
	}
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

type NormalInput struct {
	tcell.EventTime
}

func (n *NormalInput) Prompt() string {
	return ""
}

func (n *NormalInput) Confirm(i string) tcell.Event {
	return nil
}

func (n *NormalInput) Up(i string) string {
	return ""
}

func (n *NormalInput) Down(i string) string {
	return ""
}

type SearchInput struct {
	clist *Candidate
	tcell.EventTime
}

func (s *SearchInput) Prompt() string {
	return "/"
}

func (s *SearchInput) Confirm(i string) tcell.Event {
	s.clist.list = toLast(s.clist.list, i)
	s.clist.p = 0
	s.SetEventNow()
	return s
}

func (s *SearchInput) Up(i string) string {
	return s.clist.up()
}

func (s *SearchInput) Down(i string) string {
	return s.clist.down()
}

type PreviousInput struct {
	clist *Candidate
	tcell.EventTime
}

func (p *PreviousInput) Prompt() string {
	return "?"
}

func (p *PreviousInput) Confirm(s string) tcell.Event {
	p.clist.list = toLast(p.clist.list, s)
	p.clist.p = 0
	p.SetEventNow()
	return p
}

func (p *PreviousInput) Up(s string) string {
	return p.clist.up()
}

func (p *PreviousInput) Down(s string) string {
	return p.clist.down()
}

type GotoInput struct {
	clist *Candidate
	tcell.EventTime
}

func (g *GotoInput) Prompt() string {
	return "Goto line:"
}

func (g *GotoInput) Confirm(s string) tcell.Event {
	g.clist.list = toLast(g.clist.list, s)
	g.clist.p = 0
	g.SetEventNow()
	return g
}

func (g *GotoInput) Up(s string) string {
	return g.clist.up()
}

func (g *GotoInput) Down(s string) string {
	return g.clist.down()
}

type HeaderInput struct {
	tcell.EventTime
}

func (h *HeaderInput) Prompt() string {
	return "Header length:"
}

func (h *HeaderInput) Confirm(s string) tcell.Event {
	h.SetEventNow()
	return h
}

func (h *HeaderInput) Up(s string) string {
	n, err := strconv.Atoi(s)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n + 1)
}

func (h *HeaderInput) Down(s string) string {
	n, err := strconv.Atoi(s)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n - 1)
}

type DelimiterInput struct {
	clist *Candidate
	tcell.EventTime
}

func (d *DelimiterInput) Prompt() string {
	return "Delimiter:"
}

func (d *DelimiterInput) Confirm(s string) tcell.Event {
	d.clist.list = toLast(d.clist.list, s)
	d.clist.p = 0
	d.SetEventNow()
	return d
}

func (d *DelimiterInput) Up(s string) string {
	return d.clist.up()
}

func (d *DelimiterInput) Down(s string) string {
	return d.clist.down()
}

type TABWidthInput struct {
	clist *Candidate
	tcell.EventTime
}

func (t *TABWidthInput) Prompt() string {
	return "TAB width:"
}

func (t *TABWidthInput) Confirm(s string) tcell.Event {
	t.clist.list = toLast(t.clist.list, s)
	t.clist.p = 0
	t.SetEventNow()
	return t
}

func (t *TABWidthInput) Up(s string) string {
	return t.clist.up()
}

func (t *TABWidthInput) Down(s string) string {
	return t.clist.down()
}
