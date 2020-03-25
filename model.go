package zpager

import (
	"bufio"
	"io"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

type Model struct {
	TabWidth  int
	WrapMode  bool
	HeaderLen int
	PostWrite bool

	x      int
	y      int
	endY   int
	eof    bool
	text   []string
	header []string
	vSize  int
	vWidth int
	vHight int
	cache  *ristretto.Cache
}

type content struct {
	width int
	style tcell.Style
	mainc rune
	combc []rune
}

func NewModel() *Model {
	return &Model{
		text:   make([]string, 0),
		header: make([]string, 0),
		vSize:  1000,
	}
}

func (m *Model) ReadAll(r io.Reader) {
	scanner := bufio.NewScanner(r)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		for scanner.Scan() {
			m.text = append(m.text, scanner.Text())
			m.endY++
			if m.endY == m.vSize {
				ch <- struct{}{}
			}
		}
		m.eof = true
	}()

	select {
	case <-ch:
		return
	case <-time.After(500 * time.Millisecond):
		return
	}
}

func (m *Model) getContents(lineNum int) []content {
	var contents []content
	value, found := m.cache.Get(lineNum)
	if found {
		var ok bool
		if contents, ok = value.([]content); !ok {
			return nil
		}
	} else {
		contents = strToContent(m.text[lineNum], m.TabWidth)
		m.cache.Set(lineNum, contents, 1)
	}
	return contents
}

func strToContent(line string, tabWidth int) []content {
	var contents []content
	defaultContent := content{
		mainc: 0,
		combc: []rune{},
		width: 0,
		style: tcell.StyleDefault,
	}

	defaultStyle := tcell.StyleDefault
	style := defaultStyle
	n := 0
	for _, runeValue := range line {
		c := defaultContent
		switch runeValue {
		case '\n':
			continue
		case '\t':
			tabStop := tabWidth - (n % tabWidth)
			c.mainc = rune('\t')
			c.width = tabWidth
			c.style = style
			contents = append(contents, c)
			for i := 0; i < tabStop-1; i++ {
				contents = append(contents, defaultContent)
			}
			continue
		}
		switch runewidth.RuneWidth(runeValue) {
		case 0:
			if len(contents) > 0 {
				c2 := contents[len(contents)-1]
				c2.combc = append(c2.combc, runeValue)
				contents[len(contents)-1] = c2
			}
		case 1:
			c.mainc = runeValue
			c.width = 1
			c.style = style
			contents = append(contents, c)
			n++
		case 2:
			c.mainc = runeValue
			c.width = 2
			c.style = style
			contents = append(contents, c)
			contents = append(contents, defaultContent)
			n += 2
		}
	}
	return contents
}
