// Package oviewer provides a pager for terminals.
package oviewer

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dgraph-io/ristretto"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
	"golang.org/x/crypto/ssh/terminal"
)

type root struct {
	model         *Model
	wrapHeaderLen int
	bottomPos     int
	statusPos     int
	fileName      string
	mode          Mode
	input         string

	minStartPos int

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
	root.setWrapHeaderLen()
	root.statusPos = m.vHight - 1
}

func (root *root) Draw() {
	m := root.model
	screen := root.Screen
	if len(m.buffer) == 0 || m.vHight == 0 {
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
			screen.SetContent(x, y, space.mainc, space.combc, space.style)
		}
	}

	bottom := root.bottomLineNum(m.endY) - m.HeaderLen
	if m.lineNum > bottom+1 {
		m.lineNum = bottom + 1
	}
	if m.lineNum < 0 {
		m.lineNum = 0
	}

	searchWord := ""
	if root.mode == normal {
		searchWord = root.input
	}

	lY := 0
	lX := 0
	hy := 0
	for lY < m.HeaderLen {
		contents := m.getContents(lY)
		if len(contents) == 0 {
			lY++
			continue
		}
		for n := range contents {
			contents[n].style = contents[n].style.Bold(true)
		}
		if m.WrapMode {
			lX, lY = root.wrapContents(hy, lX, lY, contents)
		} else {
			lX, lY = root.noWrapContents(hy, m.x, lY, contents)
		}
		hy++
	}

	lX = m.yy * m.vWidth
	for y := root.HeaderLen(); y < m.vHight; y++ {
		contents := m.getContents(m.lineNum + lY)
		if len(contents) == 0 {
			lY++
			continue
		}
		doReverse(&contents, searchWord)
		if m.WrapMode {
			lX, lY = root.wrapContents(y, lX, lY, contents)
		} else {
			lX, lY = root.noWrapContents(y, m.x, lY, contents)
		}
	}
	root.bottomPos = m.lineNum + lY - 1
	root.statusDraw()
	//debug := fmt.Sprintf("m.y:%v header:%v bottom:%v\n", m.lineNum, m.HeaderLen, bottom)
	//root.setContentString(30, root.statusPos, strToContents(debug, 0), tcell.StyleDefault)
	root.Show()
}

func (root *root) wrapContents(y int, lX int, lY int, contents []content) (rX int, rY int) {
	wX := 0
	for {
		if lX+wX >= len(contents) {
			// EOL
			lX = 0
			lY++
			break
		}
		content := contents[lX+wX]
		if wX+content.width > root.model.vWidth {
			// next line
			lX += wX
			break
		}
		root.Screen.SetContent(wX, y, content.mainc, content.combc, content.style)
		wX++
	}
	return lX, lY
}

func (root *root) noWrapContents(y int, lX int, lY int, contents []content) (rX int, rY int) {
	if lX < root.minStartPos {
		lX = root.minStartPos
	}
	for x := 0; x < root.model.vWidth; x++ {
		if lX+x < 0 {
			continue
		}
		if lX+x >= len(contents) {
			break
		}
		content := contents[lX+x]
		root.Screen.SetContent(x, y, content.mainc, content.combc, content.style)
	}
	lY++
	return lX, lY
}

func (root *root) setWrapHeaderLen() {
	m := root.model
	root.wrapHeaderLen = 0
	for y := 0; y < m.HeaderLen; y++ {
		root.wrapHeaderLen += 1 + (len(m.getContents(y)) / m.vWidth)
	}
}

func (root *root) HeaderLen() int {
	if root.model.WrapMode {
		return root.wrapHeaderLen
	}
	return root.model.HeaderLen
}

func (root *root) bottomLineNum(num int) int {
	m := root.model
	if !m.WrapMode {
		if num <= m.vHight {
			return 0
		}
		return num - (m.vHight - root.model.HeaderLen) + 1
	}

	for y := m.vHight - root.wrapHeaderLen; y > 0; {
		y -= 1 + (len(m.getContents(num)) / m.vWidth)
		num--
	}
	num++
	return num
}

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

func doReverse(cp *[]content, searchWord string) {
	contents := *cp
	for n := range contents {
		contents[n].style = contents[n].style.Reverse(false)
	}
	if searchWord == "" {
		return
	}
	s, cIndex := contentsToStr(contents)
	for i := strings.Index(s, searchWord); i >= 0; {
		start := cIndex[i]
		end := cIndex[i+len(searchWord)]
		for ci := start; ci < end; ci++ {
			contents[ci].style = contents[ci].style.Reverse(true)
		}
		j := strings.Index(s[i+1:], searchWord)
		if j >= 0 {
			i += j + 1
		} else {
			break
		}
	}
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
	leftContents := strToContents(leftStatus, root.model.TabWidth)
	root.setContentString(0, root.statusPos, leftContents, leftStyle)

	screen.ShowCursor(runewidth.StringWidth(leftStatus), root.statusPos)

	rightStatus := fmt.Sprintf("(%d/%d%s)", root.model.lineNum, root.model.endY, next)
	rightContents := strToContents(rightStatus, root.model.TabWidth)
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

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		<-c
		root.Quit()
	}()

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
	root.minStartPos = -10
	defer func() {
		root.Screen.Fini()
		if root.model.PostWrite {
			for i := 0; i < m.vHight; i++ {
				if m.lineNum+i >= len(m.buffer) {
					break
				}
				fmt.Println(m.buffer[m.lineNum+i])
			}
		}
	}()

	screen.Clear()
	root.Sync()
	root.Run()

	return nil
}
