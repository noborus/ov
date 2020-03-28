package oviewer

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

	x       int
	lineNum int
	yy      int
	endY    int
	eof     bool
	header  []string
	buffer  []string
	vSize   int
	vWidth  int
	vHight  int
	cache   *ristretto.Cache
}

type content struct {
	width int
	style tcell.Style
	mainc rune
	combc []rune
}

func NewModel() *Model {
	return &Model{
		buffer: make([]string, 0),
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
			m.buffer = append(m.buffer, scanner.Text())
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
	if lineNum < 0 || lineNum >= len(m.buffer) {
		return nil
	}
	value, found := m.cache.Get(lineNum)
	if found {
		var ok bool
		if contents, ok = value.([]content); !ok {
			return nil
		}
	} else {
		contents = strToContents(m.buffer[lineNum], m.TabWidth)
		m.cache.Set(lineNum, contents, 1)
	}
	return contents
}

func strToContents(line string, tabWidth int) []content {
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
			c.mainc = rune(' ')
			c.width = 1
			c.style = style
			for i := 0; i < tabStop; i++ {
				contents = append(contents, c)
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
