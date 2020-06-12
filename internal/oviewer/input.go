package oviewer

import (
	"regexp"
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

type Input struct {
	mode    Mode
	value   string
	reg     *regexp.Regexp
	cursorX int

	SearchCList    *Candidate
	GoCList        *Candidate
	DelimiterCList *Candidate
	TabWidthCList  *Candidate
}

// Mode represents the state of the input.
type Mode int

const (
	normal Mode = iota
	search
	previous
	goline
	header
	delimiter
	tabWidth
)

type Candidate struct {
	list []string
	p    int
}

type EventInput interface {
	Prompt() string
	Confirm(i string) tcell.Event
	Up(i string) string
	Down(i string) string
}

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
	return &i
}

// HandleInputEvent handles input events.
func (root *Root) HandleInputEvent(ev *tcell.EventKey) bool {
	// Input not confirmed or canceled.
	ok := root.inputEvent(ev)
	if !ok {
		// Input is not confirmed.
		return true
	}

	// Input is confirmed.
	nev := root.EventInput.Confirm(root.Input.value)
	go func() { root.Screen.PostEventWait(nev) }()

	root.Input.mode = normal
	return true
}

func (root *Root) inputEvent(ev *tcell.EventKey) bool {
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
		input.value = root.EventInput.Up(input.value)
		runes := []rune(input.value)
		input.cursorX = runeWidth(string(runes))
	case tcell.KeyDown:
		input.value = root.EventInput.Down(input.value)
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

func (root *Root) setMode(mode Mode) {
	input := root.Input
	input.mode = mode
	input.value = ""
	input.cursorX = 0
	switch mode {
	case search:
		root.EventInput = &SearchInput{
			clist: input.SearchCList,
		}
	case previous:
		root.EventInput = &PreviousInput{
			clist: input.SearchCList,
		}
	case goline:
		root.EventInput = &GotoInput{
			clist: input.GoCList,
		}
	case header:
		root.EventInput = &HeaderInput{}
	case delimiter:
		root.EventInput = &DelimiterInput{
			clist: input.DelimiterCList,
		}
	case tabWidth:
		root.EventInput = &TABWidthInput{
			clist: input.TabWidthCList,
		}
	default:
		root.EventInput = &NormalInput{}
	}
	root.ShowCursor(input.cursorX, root.statusPos)
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
	s.clist.list = append(s.clist.list, i)
	s.SetEventNow()
	return s
}

func (s *SearchInput) Up(i string) string {
	if len(s.clist.list) == 0 {
		return ""
	}
	if len(s.clist.list) > s.clist.p+1 {
		s.clist.p++
		return s.clist.list[s.clist.p]
	}
	s.clist.p = 0
	return s.clist.list[s.clist.p]
}

func (s *SearchInput) Down(i string) string {
	if len(s.clist.list) == 0 {
		return ""
	}
	if s.clist.p > 0 {
		s.clist.p--
		return s.clist.list[s.clist.p]
	}
	s.clist.p = len(s.clist.list) - 1
	return s.clist.list[s.clist.p]
}

type PreviousInput struct {
	clist *Candidate
	tcell.EventTime
}

func (p *PreviousInput) Prompt() string {
	return "?"
}

func (p *PreviousInput) Confirm(i string) tcell.Event {
	p.clist.list = append(p.clist.list, i)
	p.SetEventNow()
	return p
}

func (p *PreviousInput) Up(i string) string {
	if len(p.clist.list) == 0 {
		return ""
	}
	if len(p.clist.list) > p.clist.p+1 {
		p.clist.p++
		return p.clist.list[p.clist.p]
	}
	p.clist.p = 0
	return p.clist.list[p.clist.p]
}

func (p *PreviousInput) Down(i string) string {
	if len(p.clist.list) == 0 {
		return ""
	}
	if p.clist.p > 0 {
		p.clist.p--
		return p.clist.list[p.clist.p]
	}
	p.clist.p = len(p.clist.list) - 1
	return p.clist.list[p.clist.p]
}

type GotoInput struct {
	clist *Candidate
	tcell.EventTime
}

func (g *GotoInput) Prompt() string {
	return "Goto line:"
}

func (g *GotoInput) Confirm(i string) tcell.Event {
	g.clist.list = append(g.clist.list, i)
	g.SetEventNow()
	return g
}

func (g *GotoInput) Up(i string) string {
	if len(g.clist.list) > g.clist.p+1 {
		g.clist.p++
		return g.clist.list[g.clist.p]
	}
	g.clist.p = 0
	return g.clist.list[g.clist.p]
}

func (g *GotoInput) Down(i string) string {
	if g.clist.p > 0 {
		g.clist.p--
		return g.clist.list[g.clist.p]
	}
	g.clist.p = len(g.clist.list) - 1
	return g.clist.list[g.clist.p]
}

type HeaderInput struct {
	tcell.EventTime
}

func (h *HeaderInput) Prompt() string {
	return "Header length:"
}

func (h *HeaderInput) Confirm(i string) tcell.Event {
	h.SetEventNow()
	return h
}

func (h *HeaderInput) Up(i string) string {
	n, err := strconv.Atoi(i)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n + 1)
}

func (h *HeaderInput) Down(i string) string {
	n, err := strconv.Atoi(i)
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

func (d *DelimiterInput) Confirm(i string) tcell.Event {
	d.clist.list = append(d.clist.list, i)
	d.SetEventNow()
	return d
}

func (d *DelimiterInput) Up(i string) string {
	if d.clist.p > 0 {
		d.clist.p--
		return d.clist.list[d.clist.p]
	}
	d.clist.p = len(d.clist.list) - 1
	return d.clist.list[d.clist.p]
}

func (d *DelimiterInput) Down(i string) string {
	if len(d.clist.list) > d.clist.p+1 {
		d.clist.p++
		return d.clist.list[d.clist.p]
	}
	d.clist.p = 0
	return d.clist.list[d.clist.p]
}

type TABWidthInput struct {
	clist *Candidate
	tcell.EventTime
}

func (t *TABWidthInput) Prompt() string {
	return "TAB width:"
}

func (t *TABWidthInput) Confirm(i string) tcell.Event {
	t.clist.list = append(t.clist.list, i)
	t.SetEventNow()
	return t
}

func (t *TABWidthInput) Up(i string) string {
	if t.clist.p > 0 {
		t.clist.p--
		return t.clist.list[t.clist.p]
	}
	t.clist.p = len(t.clist.list) - 1
	return t.clist.list[t.clist.p]
}

func (t *TABWidthInput) Down(i string) string {
	if len(t.clist.list) > t.clist.p+1 {
		t.clist.p++
		return t.clist.list[t.clist.p]
	}
	t.clist.p = 0
	return t.clist.list[t.clist.p]
}
