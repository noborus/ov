package zpager

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
	"github.com/klauspost/compress/zstd"
	"github.com/mattn/go-runewidth"
	"github.com/pierrec/lz4"
	"github.com/ulikunitz/xz"
)

type root struct {
	parent *views.Application
	main   *views.CellView
	model  *model
	status *views.SimpleStyledTextBar
	title  *views.TextBar
	views.Panel
}

func (root *root) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape:
			root.parent.Quit()
			return true
		case tcell.KeyEnter:
			root.model.y++
			root.updateKeys()
			return true
		case tcell.KeyLeft:
			root.model.x--
			root.updateKeys()
			return true
		case tcell.KeyRight:
			root.model.x++
			root.updateKeys()
			return true
		case tcell.KeyDown:
			root.model.y++
			root.updateKeys()
			return true
		case tcell.KeyUp:
			root.model.y--
			root.updateKeys()
			return true
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'Q', 'q':
				root.parent.Quit()
				return true
			case 'S', 's':
				root.model.y = root.model.y + 10000
				root.updateKeys()
				return true
			case 'V', 'v':
				root.model.y = root.model.y + 10
				root.updateKeys()
				return true
			case 'N', 'n':
				root.model.y = root.model.y + 1
				root.updateKeys()
				return true
			case 'P', 'p':
				root.model.y = root.model.y - 1
				root.updateKeys()
				return true
			case 'H', 'h':
				root.SetTitle(root.title)
				root.updateKeys()
				return true
			case 'E', 'e':
				root.status.SetCenter(fmt.Sprintf("y:%d", root.model.y))
				root.SetStatus(root.status)
				return true
			case 'D', 'd':
				root.RemoveWidget(root.title)
				root.RemoveWidget(root.status)
				return true
			}
		}
	}
	return true
}

type model struct {
	x        int
	y        int
	endx     int
	endy     int
	hide     bool
	enab     bool
	line     int
	lineRune []rune
	loc      string
	block    [][]byte
}

func (m *model) GetBounds() (int, int) {
	return m.endx, m.endy
}

func (m *model) MoveCursor(offx, offy int) {
	m.x += offx
	m.y += offy
	m.limitCursor()
}

func (m *model) limitCursor() {
	if m.x < 0 {
		m.x = 0
	}
	if m.x > m.endx-1 {
		m.x = m.endx - 1
	}
	if m.y < 0 {
		m.y = 0
	}
	if m.y > m.endy-1 {
		m.y = m.endy - 1
	}
	m.loc = fmt.Sprintf("Cursor is %d,%d", m.x, m.y)
}

func (m *model) GetCursor() (int, int, bool, bool) {
	return m.x, m.y, m.enab, !m.hide
}

func (m *model) SetCursor(x int, y int) {
	m.x = x
	m.y = y

	m.limitCursor()
}

func (m *root) updateKeys() {
	mm := m.model
	_, by := mm.GetBounds()
	if mm.y >= len(mm.block) {
		mm.y = len(mm.block) - by
	}
	if mm.x >= 200 {
		mm.x = mm.endx
	}
	m.parent.Update()
}

func setLineRune(str string) []rune {
	var lineRune []rune
	for _, runeValue := range str {
		switch runewidth.RuneWidth(runeValue) {
		case 0:
			lineRune = append(lineRune, rune(' '))
		case 1:
			lineRune = append(lineRune, runeValue)
		case 2:
			lineRune = append(lineRune, runeValue)
			lineRune = append(lineRune, rune(' '))
		}
	}
	return lineRune
}

func (m *model) GetCell(vx, vy int) (rune, tcell.Style, []rune, int) {
	y := vy
	x := vx
	if m.y > 0 {
		y = m.y + vy
	}
	if m.x > 0 {
		x = m.x + vx
	}
	if x < 0 || y < 0 || y >= len(m.block) {
		return 0, tcell.StyleDefault, nil, 1
	}

	if y != m.line {
		m.lineRune = setLineRune(string(m.block[y]))
		m.line = y
	}
	if x >= len(m.lineRune) {
		return 0, tcell.StyleDefault, nil, 1
	}

	return m.lineRune[x], tcell.StyleDefault, nil, 1
}

func (m *model) SetCell(r io.Reader) {
	m.block = make([][]byte, 0)
	m.line = -1
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b0 := scanner.Bytes()
		b1 := make([]byte, len(b0))
		copy(b1, b0)
		m.block = append(m.block, b1)
	}
}
func newRoot() *root {
	root := &root{}

	root.main = views.NewCellView()
	m := &model{}
	root.main.SetModel(m)
	root.model = m
	root.main.SetStyle(tcell.StyleDefault)
	root.Panel.SetContent(root.main)

	status := views.NewSimpleStyledTextBar()
	status.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorBlue).
		Foreground(tcell.ColorYellow))
	status.SetCenter("Status bar")
	root.status = status

	title := views.NewTextBar()
	title.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorTeal).
		Foreground(tcell.ColorWhite))
	title.SetCenter("Title", tcell.StyleDefault)
	root.title = title
	return root
}

func uncompressedReader(reader io.Reader) io.ReadCloser {
	var err error
	buf := [7]byte{}
	n, err := io.ReadAtLeast(reader, buf[:], len(buf))
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return ioutil.NopCloser(bytes.NewReader(buf[:n]))
		}
		return ioutil.NopCloser(bytes.NewReader(nil))
	}

	rd := io.MultiReader(bytes.NewReader(buf[:n]), reader)
	var r io.ReadCloser
	switch {
	case bytes.Equal(buf[:3], []byte{0x1f, 0x8b, 0x8}):
		r, err = gzip.NewReader(rd)
	case bytes.Equal(buf[:3], []byte{0x42, 0x5A, 0x68}):
		r = ioutil.NopCloser(bzip2.NewReader(rd))
	case bytes.Equal(buf[:4], []byte{0x28, 0xb5, 0x2f, 0xfd}):
		var zr *zstd.Decoder
		zr, err = zstd.NewReader(rd)
		r = ioutil.NopCloser(zr)
	case bytes.Equal(buf[:4], []byte{0x04, 0x22, 0x4d, 0x18}):
		r = ioutil.NopCloser(lz4.NewReader(rd))
	case bytes.Equal(buf[:7], []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x0, 0x0}):
		var zr *xz.Reader
		zr, err = xz.NewReader(rd)
		r = ioutil.NopCloser(zr)
	}
	if err != nil || r == nil {
		r = ioutil.NopCloser(rd)
	}
	return r
}

func Run() {
	root := newRoot()
	app := &views.Application{}
	root.parent = app
	app.SetStyle(tcell.StyleDefault)
	app.SetRootWidget(root)

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	r := uncompressedReader(file)
	go root.model.SetCell(r)
	defer r.Close()

	if e := app.Run(); e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		os.Exit(1)
	}
}
