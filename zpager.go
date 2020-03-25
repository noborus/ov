package zpager

import (
	"fmt"
	"io"
	"os"

	"github.com/dgraph-io/ristretto"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
	"golang.org/x/crypto/ssh/terminal"
)

type root struct {
	model     *Model
	statusPos int
	fileName  string
	mode      Mode
	input     string
	tcell.Screen
}

type Mode int

const (
	normal Mode = iota
	search
	previous
	goline
	header
)

func (root *root) PrepareView() {
	m := root.model
	screen := root.Screen
	m.vWidth, m.vHight = screen.Size()
	m.vHight--
	root.statusPos = m.vHight
	m.vView = make([][]content, m.vHight)
	for y := 0; y < m.vHight; y++ {
		m.vView[y] = make([]content, m.vWidth)
	}
}

func (root *root) Draw() {
	m := root.model
	screen := root.Screen
	if len(m.text) == 0 || len(m.vView) == 0 {
		root.statusDraw()
		root.Show()
		return
	}

	space := content{
		mainc: ' ',
		combc: nil,
		width: 1,
	}
	for y := 0; y < m.vHight; y++ {
		for x := 0; x < m.vWidth; x++ {
			m.vView[y][x] = space
		}
	}

	if m.y+m.vHight > len(m.text) {
		m.y = len(m.text) - m.vHight
	}
	if m.y < 0 {
		m.y = 0
	}

	searchWord := ""
	if root.mode == normal {
		searchWord = root.input
	}
	m.header = m.text[:m.HeaderLen]
	if m.WrapMode {
		m.wrapContent(searchWord)
	} else {
		m.noWrapContent(searchWord)
	}

	for y := 0; y < m.vHight; y++ {
		for x := 0; x < m.vWidth; x++ {
			content := m.vView[y][x]
			screen.SetContent(x, y, content.mainc, content.combc, content.style)
		}
	}
	root.statusDraw()
	root.Show()
}

func (root *root) setContentString(vx int, vy int, contents []content, style tcell.Style) {
	screen := root.Screen
	for x, content := range contents {
		screen.SetContent(vx+x, vy, content.mainc, content.combc, style)
	}
}

func (root *root) statusDraw() {
	screen := root.Screen
	style := tcell.StyleDefault
	next := "..."
	if root.model.eof {
		next = ""
	}
	for x := 0; x < root.model.vWidth; x++ {
		screen.SetContent(x, root.statusPos, 0, nil, style)
	}

	leftStatus := ""
	leftStyle := style
	switch root.mode {
	case search:
		leftStatus = "/" + root.input
	case previous:
		leftStatus = "?" + root.input
	case goline:
		leftStatus = "Goto line:" + root.input
	case header:
		leftStatus = "Header length:" + root.input
	default:
		leftStatus = fmt.Sprintf("%s:", root.fileName)
		leftStyle = style.Reverse(true)
	}
	leftContents := strToContent(leftStatus, root.model.TabWidth)
	root.setContentString(0, root.statusPos, leftContents, leftStyle)

	screen.ShowCursor(runewidth.StringWidth(leftStatus), root.statusPos)

	rightStatus := fmt.Sprintf("(%d/%d%s)", root.model.y, root.model.endY, next)
	rightContents := strToContent(rightStatus, root.model.TabWidth)
	root.setContentString(root.model.vWidth-len(rightStatus), root.statusPos, rightContents, style)
}

type eventAppQuit struct {
	tcell.EventTime
}

func (root *root) Quit() {
	ev := &eventAppQuit{}
	ev.SetEventNow()
	go func() { root.Screen.PostEventWait(ev) }()
}

func (root *root) Resize() {
	root.Sync()
}

func (root *root) Sync() {
	root.PrepareView()
	root.Draw()
}

func (root *root) Run() {
	screen := root.Screen

loop:
	for {
		root.Draw()
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *eventAppQuit:
			break loop
		case *tcell.EventKey:
			root.HandleEvent(ev)
		case *tcell.EventResize:
			root.Resize()
		}
	}
}

func Run(m *Model, args []string) error {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return err
	}
	m.cache = cache
	var reader io.Reader
	fileName := ""
	switch len(args) {
	case 0:
		if terminal.IsTerminal(0) {
			return fmt.Errorf("missing filename")
		}
		fileName = "(STDIN)"
		reader = uncompressedReader(os.Stdin)
	case 1:
		fileName = args[0]
		r, err := os.Open(fileName)
		if err != nil {
			return err
		}
		reader = uncompressedReader(r)
	default:
		readers := make([]io.Reader, 0)
		for _, fileName := range args {
			r, err := os.Open(fileName)
			if err != nil {
				return err
			}
			readers = append(readers, uncompressedReader(r))
			reader = io.MultiReader(readers...)
		}
	}
	m.ReadAll(reader)

	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	err = screen.Init()
	if err != nil {
		return err
	}

	root := &root{}
	root.Screen = screen
	root.fileName = fileName
	root.model = m

	defer func() {
		root.Screen.Fini()
		if root.model.PostWrite {
			for i := 0; i < len(m.text); i++ {
				if i > m.vHight-1 {
					break
				}
				fmt.Println(m.text[m.y+i])
			}
		}
	}()

	screen.Clear()
	root.Sync()
	root.Run()

	return nil
}
