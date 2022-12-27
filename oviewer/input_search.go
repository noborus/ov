package oviewer

import "github.com/gdamore/tcell/v2"

// setSearchMode sets the inputMode to Search.
func (root *Root) setSearchMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = Search
	input.EventInput = newSearchInput(input.SearchCandidate)
	root.OriginPos = root.Doc.topLN
}

// searchCandidate returns the candidate to set to default.
func searchCandidate() *candidate {
	return &candidate{
		list: []string{},
	}
}

// searchInput represents the search input mode.
type searchInput struct {
	value string
	clist *candidate
	tcell.EventTime
}

// newSearchInput returns SearchInput.
func newSearchInput(clist *candidate) *searchInput {
	return &searchInput{
		value:     "",
		clist:     clist,
		EventTime: tcell.EventTime{},
	}
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

// setBackSearchMode sets the inputMode to Backsearch.
func (root *Root) setBackSearchMode() {
	input := root.input
	input.value = ""
	input.cursorX = 0
	input.mode = Backsearch
	input.EventInput = newBackSearchInput(input.SearchCandidate)
	root.OriginPos = root.Doc.topLN
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
