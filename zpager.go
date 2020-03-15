package zpager

import (
	"fmt"
	"io"
	"os"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
	"github.com/mattn/go-runewidth"
)

type root struct {
	parent *views.Application
	main   *views.CellView
	model  *model
	status *views.SimpleStyledTextBar
	title  *views.TextBar
	views.Panel
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
	block    []string
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

func (root *root) updateKeys() {
	root.parent.Update()
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
		m.lineRune = setLineRune(m.block[y])
		m.line = y
	}
	if x >= len(m.lineRune) {
		return 0, tcell.StyleDefault, nil, 1
	}

	return m.lineRune[x], tcell.StyleDefault, nil, 1
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

func Run(reader io.Reader) {
	root := newRoot()
	app := &views.Application{}
	root.parent = app
	app.SetStyle(tcell.StyleDefault)
	app.SetRootWidget(root)

	root.model.block = make([]string, 0)
	r := uncompressedReader(reader)
	go ReadAll(&root.model.block, r)
	defer r.Close()

	if e := app.Run(); e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		os.Exit(1)
	}
}
