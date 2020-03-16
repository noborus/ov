package zpager

import (
	"bufio"
	"io"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
)

type Model struct {
	TabWidth int
	WrapMode bool

	x      int
	y      int
	endY   int
	eof    bool
	text   []string
	vSize  int
	vWidth int
	vHight int
	vView  [][]content
}

type content struct {
	mainc     rune
	combc     []rune
	width     int
	highlight bool
}

func NewModel() *Model {
	return &Model{
		text:  make([]string, 0),
		vSize: 1000,
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

func (m *Model) noWrapContent(subStr string) {
	lY := m.y
	lX := m.x
	if lX <= -10 {
		lX = -10
		m.x = -10
	}
	contents := make([][]content, 0, m.vHight)
	maxX := 0
	for y := 0; y < m.vHight; y++ {
		if lY+y >= len(m.text) {
			break
		}
		content := strToContent(m.text[lY+y], subStr, m.TabWidth)
		if len(content) > maxX {
			maxX = len(content)
		}
		contents = append(contents, content)
	}
	if maxX-1 < lX {
		m.x = maxX - 1
		lX = m.x
	}
	for y, content := range contents {
		for x := 0; x < m.vWidth; x++ {
			if lX+x < 0 {
				continue
			}
			if lX+x >= len(content) {
				break
			}
			m.vView[y][x] = content[lX+x]
		}
	}
}

func (m *Model) wrapContent(subStr string) {
	lY := m.y
	if lY < 0 {
		lY = 0
	}
	lX := 0
	y := 0
	x := 0
	for {
		contents := strToContent(m.text[lY], subStr, m.TabWidth)
		lX = m.x
		for {
			if lX < 0 {
				x++
				lX++
				continue
			}
			if len(contents) == 0 {
				break
			}
			m.vView[y][x] = contents[lX]
			x++
			// Wrap
			if x >= m.vWidth {
				x = 0
				y++
				if y >= m.vHight {
					return
				}
			}
			lX++
			// EOL
			if lX >= len(contents) {
				x = 0
				break
			}
		}
		y++
		// Reach the bottom
		if y >= m.vHight {
			return
		}
		lY++
		// EOF
		if lY >= len(m.text) {
			return
		}
	}
}

func strToContent(line string, subStr string, tabWidth int) []content {
	var contents []content
	str := strings.ReplaceAll(line, subStr, "\n"+subStr+"\n")
	hlFlag := false
	defaultContent := content{
		mainc:     0,
		combc:     []rune{},
		width:     0,
		highlight: false,
	}

	n := 0
	for _, runeValue := range str {
		c := defaultContent
		if runeValue == '\n' {
			if !hlFlag {
				hlFlag = true
			} else {
				hlFlag = false
			}
			continue
		} else if runeValue == '\t' {
			tabStop := tabWidth - (n % tabWidth)
			c.mainc = rune(' ')
			c.width = 1
			c.highlight = hlFlag
			for i := 0; i < tabStop; i++ {
				contents = append(contents, c)
			}
			continue
		}
		switch runewidth.RuneWidth(runeValue) {
		case 0:
			c.mainc = rune(' ')
			c.width = 1
			c.highlight = hlFlag
			contents = append(contents, c)
			n++
		case 1:
			c.mainc = runeValue
			c.width = 1
			c.highlight = hlFlag
			contents = append(contents, c)
			n++
		case 2:
			c.mainc = runeValue
			c.width = 2
			c.highlight = hlFlag
			contents = append(contents, c)
			contents = append(contents, defaultContent)
			n++
			n++
		}
	}
	return contents
}
