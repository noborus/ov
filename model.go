package oviewer

import (
	"bufio"
	"io"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

// The Model structure contains the values
// for the logical screen.
type Model struct {
	x          int
	lineNum    int
	yy         int
	endNum     int
	eof        bool
	header     []string
	buffer     []string
	beforeSize int
	vWidth     int
	vHight     int
	cache      *ristretto.Cache
}

//NewModel returns Model.
func NewModel() *Model {
	return &Model{
		buffer:     make([]string, 0),
		header:     make([]string, 0),
		beforeSize: 1000,
	}
}

// ReadAll reads all from the reader to the buffer.
// It returns if beforeSize is accumulated in buffer
// before the end of read.
func (m *Model) ReadAll(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		for scanner.Scan() {
			m.buffer = append(m.buffer, scanner.Text())
			m.endNum++
			if m.endNum == m.beforeSize {
				ch <- struct{}{}
			}
		}
		if err := scanner.Err(); err != nil {
			m.eof = false
			return
		}
		m.eof = true
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(500 * time.Millisecond):
		return nil
	}
}

// content represents one character on the terminal.
type content struct {
	width int
	style tcell.Style
	mainc rune
	combc []rune
}

// getContents returns one row of contents from buffer.
func (m *Model) getContents(lineNum int, tabWidth int) []content {
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
		contents = strToContents(m.buffer[lineNum], tabWidth)
		m.cache.Set(lineNum, contents, 1)
	}
	return contents
}

// strToContents converts a single-line string into a content array.
func strToContents(line string, tabWidth int) []content {
	contents := []content{}
	defaultStyle := tcell.StyleDefault
	defaultContent := content{
		mainc: 0,
		combc: []rune{},
		width: 0,
		style: defaultStyle,
	}

	style := defaultStyle
	x := 0
	for _, runeValue := range line {
		c := defaultContent
		switch runeValue {
		case '\n':
			continue
		case '\t':
			tabStop := tabWidth - (x % tabWidth)
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
			x++
		case 2:
			c.mainc = runeValue
			c.width = 2
			c.style = style
			contents = append(contents, c)
			contents = append(contents, defaultContent)
			x += 2
		}
	}
	return contents
}

// contentsToStr converts a content array into a single-line string.
func contentsToStr(contents []content) (string, map[int]int) {
	buf := make([]rune, 0)
	cIndex := make(map[int]int)
	byteLen := 0
	for n := range contents {
		if contents[n].width > 0 {
			cIndex[byteLen] = n
			buf = append(buf, contents[n].mainc)
			b := string(contents[n].mainc)
			byteLen += len(b)
		}
	}
	s := string(buf)
	cIndex[len(s)] = len(contents)
	return s, cIndex
}
