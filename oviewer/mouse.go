package oviewer

import (
	"context"
	"log"
	"strings"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// ClickType represents the type of mouse click.
type ClickType int

const (
	ClickSingle ClickType = iota
	ClickDouble
	ClickTriple
	ClickExpired // Time expired, should reset state
)

const (
	charTypeWhitespace   = 0
	charTypeAlphanumeric = 1
	charTypeOther        = 2
)

var (
	ClickInterval = 500 * time.Millisecond // Double/triple click detection time
	ClickDistance = 2                      // Maximum allowed movement distance (in screen coordinates/cells) for double/triple click detection
)

// ClickState holds the state for click detection.
type ClickState struct {
	lastClickTime time.Time
	lastClickX    int
	lastClickY    int
	clickCount    int
}

// mouseEvent handles mouse events.
func (root *Root) mouseEvent(ctx context.Context, ev *tcell.EventMouse) {
	button := ev.Buttons()
	mod := ev.Modifiers()

	var horizontal bool
	if mod&tcell.ModShift != 0 {
		horizontal = true
	}

	if button&tcell.WheelUp != 0 {
		if horizontal {
			root.wheelLeft(ctx)
			return
		}
		root.wheelUp(ctx)
		return
	}

	if button&tcell.WheelDown != 0 {
		if horizontal {
			root.wheelRight(ctx)
			return
		}
		root.wheelDown(ctx)
		return
	}

	if button&tcell.WheelLeft != 0 {
		root.wheelLeft(ctx)
		return
	}
	if button&tcell.WheelRight != 0 {
		root.wheelRight(ctx)
		return
	}

	if root.mouseSelect(ctx, ev) {
		return
	}
	// Skip redraw for non-selection mouse events.
	root.skipDraw = true
}

// wheelUp moves the mouse wheel up.
func (root *Root) wheelUp(context.Context) {
	root.setMessage("")
	root.moveUp(root.Doc.VScrollLines)
}

// wheelDown moves the mouse wheel down.
func (root *Root) wheelDown(context.Context) {
	root.setMessage("")
	root.moveDown(root.Doc.VScrollLines)
}

// wheelRight moves the mouse wheel right.
func (root *Root) wheelRight(ctx context.Context) {
	root.setMessage("")
	if root.Doc.ColumnMode {
		root.moveRightOne(ctx)
	} else {
		root.moveRight(root.Doc.HScrollWidthNum)
	}
}

// wheelLeft moves the mouse wheel left.
func (root *Root) wheelLeft(ctx context.Context) {
	root.setMessage("")
	if root.Doc.ColumnMode {
		root.moveLeftOne(ctx)
	} else {
		root.moveLeft(root.Doc.HScrollWidthNum)
	}
}

// mouseSelect handles mouse selection.
func (root *Root) mouseSelect(ctx context.Context, ev *tcell.EventMouse) bool {
	button := ev.Buttons()

	// Handle middle button paste regardless of state
	if button == tcell.ButtonMiddle {
		root.Paste(ctx)
		root.setMessage("Paste")
		return true
	}

	switch root.scr.mouseSelect {
	case SelectNone:
		return root.handleSelectNone(ctx, ev, button)
	case SelectActive:
		return root.handleSelectActive(ctx, ev, button)
	case SelectCopied:
		return root.handleSelectCopied(ctx, ev, button)
	default:
		return false
	}
}

// handleSelectNone handles mouse events when no selection is active.
func (root *Root) handleSelectNone(ctx context.Context, ev *tcell.EventMouse, button tcell.ButtonMask) bool {
	if button != tcell.ButtonPrimary {
		return false
	}

	return root.handlePrimaryButtonClick(ctx, ev)
}

// handleSelectCopied handles mouse events when selection is copied.
func (root *Root) handleSelectCopied(ctx context.Context, ev *tcell.EventMouse, button tcell.ButtonMask) bool {
	if button != tcell.ButtonPrimary {
		return false
	}

	// Reset previous selection before handling new click
	root.resetSelect()
	return root.handlePrimaryButtonClick(ctx, ev)
}

// handlePrimaryButtonClick handles single/double/triple click events.
func (root *Root) handlePrimaryButtonClick(ctx context.Context, ev *tcell.EventMouse) bool {
	x, y := ev.Position()
	mod := ev.Modifiers()
	now := time.Now()

	// Alt+click extension has priority over click type detection
	if mod&tcell.ModAlt != 0 && root.scr.hasAnchorPoint {
		// Update click state for Alt+click (to maintain consistency)
		root.updateClickState(x, y, now)
		return root.extendSelectionWithAltClick(ev)
	}

	clickType := root.checkClickType(x, y, now)

	switch clickType {
	case ClickTriple:
		root.handleTripleClick(ctx, ev)
		root.resetClickState()
		return true
	case ClickDouble:
		root.handleDoubleClick(ctx, ev)
		root.updateClickState(x, y, now)
		return true
	case ClickExpired:
		root.resetClickState()
		fallthrough
	case ClickSingle:
		root.updateClickState(x, y, now)
		root.startSelection(x, y, ev.Modifiers())
		return true
	}

	return false
}

// extendSelectionWithAltClick extends the current selection to the Alt+click position.
// This implements the anchor-and-extend selection workflow where the first click
// sets an anchor point and Alt+click extends the selection to the new position.
func (root *Root) extendSelectionWithAltClick(ev *tcell.EventMouse) bool {
	x, y := ev.Position()
	root.scr.x2, root.scr.y2 = root.adjustPositionForWideChar(x, y)
	root.scr.mouseSelect = SelectActive
	return true
}

// startSelection initializes a new selection.
func (root *Root) startSelection(x, y int, modifiers tcell.ModMask) {
	x, y = root.adjustPositionForWideChar(x, y)
	root.scr.x1, root.scr.y1 = x, y
	root.scr.x2, root.scr.y2 = x, y

	root.scr.mouseSelect = SelectActive
	root.scr.hasAnchorPoint = true
	root.scr.mousePressed = true
	root.scr.mouseRectangle = false
	// If Ctrl is pressed, enable rectangle selection
	if modifiers&tcell.ModCtrl != 0 {
		root.scr.mouseRectangle = true
	}
}

// handleSelectActive handles mouse events during active selection.
func (root *Root) handleSelectActive(ctx context.Context, ev *tcell.EventMouse, button tcell.ButtonMask) bool {
	switch button {
	case tcell.ButtonPrimary:
		x, y := ev.Position()
		root.scr.x2, root.scr.y2 = root.adjustPositionForWideChar(x, y)
		return true
	case tcell.ButtonNone:
		// Check if mouse has moved enough to be considered a selection
		if root.scr.x1 != root.scr.x2 || root.scr.y1 != root.scr.y2 {
			root.scr.mouseSelect = SelectCopied
			root.scr.mousePressed = false
			root.CopySelect(ctx)
		} else {
			// No selection made, reset
			root.scr.mouseSelect = SelectNone
			root.scr.mousePressed = false
		}
		return true
	default:
		return true
	}
}

// adjustPositionForWideChar adjusts the position if a wide character is clicked.
func (root *Root) adjustPositionForWideChar(x, y int) (int, int) {
	p, _, _, _ := root.Screen.GetContent(x, y)
	if p == ' ' && x > 0 {
		_, _, _, w := root.Screen.GetContent(x-1, y) // Adjust for clicking the second half of a wide character
		if w == 2 {
			x--
		}
	}
	return x, y
}

// resetClickState resets the click state.
func (root *Root) resetClickState() {
	root.clickState.clickCount = 0
	root.clickState.lastClickTime = time.Time{}
	root.clickState.lastClickX = 0
	root.clickState.lastClickY = 0
}

// checkClickType determines the type of click based on timing and position.
func (root *Root) checkClickType(x, y int, now time.Time) ClickType {
	if root.clickState.clickCount == 0 {
		return ClickSingle
	}

	timeDiff := now.Sub(root.clickState.lastClickTime)
	if timeDiff > ClickInterval {
		return ClickExpired
	}

	distanceX := abs(x - root.clickState.lastClickX)
	distanceY := abs(y - root.clickState.lastClickY)

	if distanceX <= ClickDistance && distanceY <= ClickDistance {
		switch root.clickState.clickCount {
		case 1:
			return ClickDouble
		case 2:
			return ClickTriple
		}
	}
	return ClickSingle
}

// handleDoubleClick handles double click events for word/column selection (triple click uses handleTripleClick).
func (root *Root) handleDoubleClick(ctx context.Context, ev *tcell.EventMouse) {
	root.doubleClickSetRange(ev)

	// Check if a valid selection range was created
	if root.scr.x1 == root.scr.x2 && root.scr.y1 == root.scr.y2 {
		// No selection range found, clear anchor and reset
		root.resetSelect()
		return
	}
	root.scr.mouseSelect = SelectCopied
	root.scr.mousePressed = false
	root.CopySelect(ctx)
}

// doubleClickSetRange sets the selection range for double click events.
func (root *Root) doubleClickSetRange(ev *tcell.EventMouse) {
	x, y := ev.Position()
	ln := root.scr.lineNumber(y)
	lineC, ok := root.scr.lines[ln.number]
	if !ok || !lineC.valid {
		root.scr.x1, root.scr.y1 = x, y
		root.scr.x2, root.scr.y2 = x, y
		return
	}

	if len(lineC.columnRanges) > 0 {
		startX, startY, endX, endY := root.findColumnBoundaries(x, y, lineC, ln)
		if startX != endX || startY != endY {
			root.scr.x1, root.scr.y1 = startX, startY
			root.scr.x2, root.scr.y2 = endX, endY
			return
		}
		// If no column found, fall through to word selection.
	}

	startX, startY, endX, endY := root.findWordBoundariesInLine(x, y, lineC, ln)
	root.scr.x1, root.scr.y1 = startX, startY
	root.scr.x2, root.scr.y2 = endX, endY
}

// updateClickState updates the click state.
func (root *Root) updateClickState(x, y int, now time.Time) {
	root.clickState.lastClickTime = now
	root.clickState.lastClickX = x
	root.clickState.lastClickY = y
	root.clickState.clickCount++
}

// findColumnBoundaries selects the entire column at the given position.
func (root *Root) findColumnBoundaries(x, y int, lineC LineC, ln LineNumber) (int, int, int, int) {
	contentX := root.Doc.x + (x - root.scr.startX) + root.scr.branchWidth(lineC.lc, ln.wrap)
	startCol, endCol, ok := findColumnRange(contentX, lineC.columnRanges)
	if !ok {
		return x, y, x, y
	}
	startX := x - (contentX - startCol)
	startY := y
	for startX < root.scr.startX {
		startY--
		startX = root.scr.vWidth - (root.scr.startX - startX)
	}
	endX := x + (endCol - contentX) - 1
	endY := y
	if !root.Doc.WrapMode {
		return startX, startY, endX, endY
	}
	// Handle wrapping for endX
	wrappedX := root.scr.vWidth - 1
	for endX > root.scr.vWidth-1 {
		prevIdx := wrappedX
		if prevIdx >= 0 && prevIdx < len(lineC.lc) && lineC.lc[prevIdx].width == 2 {
			endX++
		}
		endY++
		endX = root.scr.startX + (endX - root.scr.vWidth)
		wrappedX += root.scr.vWidth
	}
	return startX, startY, endX, endY
}

// findWordBoundariesInLine selects the word at the given position.
func (root *Root) findWordBoundariesInLine(x, y int, lineC LineC, ln LineNumber) (int, int, int, int) {
	x = x - root.scr.startX
	contentX := x + root.scr.branchWidth(lineC.lc, ln.wrap)
	if contentX < 0 || contentX >= len(lineC.lc) {
		// Out of bounds: return clicked position only
		return x + root.scr.startX, y, x + root.scr.startX, y
	}
	charType := getCharTypeAt(lineC, contentX)
	startX := x
	startY := y
	for ln.wrap >= 0 {
		testContentX := (startX - 1) + root.scr.branchWidth(lineC.lc, ln.wrap)
		if getCharTypeAt(lineC, testContentX) != charType {
			break
		}
		startX--
		if startX < root.scr.startX {
			ln.wrap--
			startX = root.scr.vWidth
			startY--
		}
	}
	endX := startX
	endY := startY
	for endX < len(lineC.lc)-1 {
		testContentX := (endX + 1) + root.scr.branchWidth(lineC.lc, ln.wrap)
		if getCharTypeAt(lineC, testContentX) != charType {
			break
		}
		endX++
		if root.Doc.WrapMode && endX >= root.scr.vWidth-1 {
			endX = root.scr.startX
			endY++
			ln.wrap++
		}
	}
	return startX + root.scr.startX, startY, endX + root.scr.startX, endY
}

// findColumnRange returns the start and end of the column range containing contentX, or ok=false if not found.
func findColumnRange(contentX int, columnRanges []columnRange) (start, end int, ok bool) {
	for _, colRange := range columnRanges {
		if contentX >= colRange.start && contentX <= colRange.end {
			return colRange.start, colRange.end, true
		}
	}
	return 0, 0, false
}

// getCharTypeAt returns character type at the given content position.
func getCharTypeAt(line LineC, contentX int) int {
	if contentX < 0 || contentX >= len(line.lc) {
		return charTypeWhitespace // Out of bounds treated as whitespace
	}
	content := line.lc[contentX]
	char := content.mainc
	width := content.width
	if width == 0 && char == 0 {
		if contentX == 0 {
			return charTypeOther
		}
		// If width is 0, check the previous character.
		char = line.lc[contentX-1].mainc
	}
	if char == ' ' || char == '\t' {
		return charTypeWhitespace
	}
	if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_' {
		return charTypeAlphanumeric
	}
	return charTypeOther
}

// handleTripleClick handles triple click events (selects the whole line).
func (root *Root) handleTripleClick(ctx context.Context, ev *tcell.EventMouse) {
	x, y := ev.Position()
	ln := root.scr.lineNumber(y)
	ln.wrap = 0
	// Start with y at the beginning of the line.
	y = root.scr.lineY(ln)
	lineC, ok := root.scr.lines[ln.number]
	if !ok || !lineC.valid {
		root.scr.x1, root.scr.y1 = x, y
		root.scr.x2, root.scr.y2 = x, y
	} else {
		root.scr.x1, root.scr.y1 = root.scr.startX, y
		lastY := y
		for endY := y; endY < len(root.scr.numbers) && root.scr.lineNumber(endY).number == ln.number; endY++ {
			lastY = endY
		}
		root.scr.x2, root.scr.y2 = root.scr.startX+root.scr.vWidth-1, lastY
	}
	root.scr.mouseSelect = SelectCopied
	root.scr.mousePressed = false
	root.CopySelect(ctx)
}

// resetSelect resets the selection.
func (root *Root) resetSelect() {
	root.scr.mouseSelect = SelectNone
	root.scr.x1, root.scr.y1 = 0, 0
	root.scr.x2, root.scr.y2 = 0, 0
	root.scr.mousePressed = false
	root.scr.hasAnchorPoint = false
	root.setMessage("")
}

// eventCopySelect represents a mouse select event.
type eventCopySelect struct {
	tcell.EventTime
}

// CopySelect fires the eventCopySelect event.
func (root *Root) CopySelect(context.Context) {
	root.sendCopySelect()
}

func (root *Root) sendCopySelect() {
	ev := &eventCopySelect{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// copyToClipboard writes the selection to the clipboard.
func (root *Root) copyToClipboard(_ context.Context) {
	str, err := root.rangeToString(root.scr.x1, root.scr.y1, root.scr.x2, root.scr.y2)
	if err != nil {
		root.debugMessage("copyToClipboard: " + err.Error())
		return
	}
	if len(str) == 0 {
		return
	}

	root.copyClipboard(str)
	root.setMessage("Copy")
}

// copyClipboard copies the string to the clipboard.
// The method of copying to the clipboard is determined by the configuration.
// The default method is to use the clipboard package.
// "OSC52” can be specified (available only if the terminal supports it).
func (root *Root) copyClipboard(str string) {
	switch root.Config.ClipboardMethod {
	case "OSC52":
		root.Screen.SetClipboard([]byte(str))
	default:
		if err := clipboard.WriteAll(str); err != nil {
			log.Printf("copyToClipboard: %v\n", err)
		}
	}
}

type eventPaste struct {
	tcell.EventTime
}

// Paste fires the eventPaste event.
func (root *Root) Paste(context.Context) {
	root.sendPaste()
}

func (root *Root) sendPaste() {
	ev := &eventPaste{}
	ev.SetEventNow()
	root.postEvent(ev)
}

// pasteFromClipboard writes a string from the clipboard.
func (root *Root) pasteFromClipboard(context.Context) {
	input := root.input
	if input.Event.Mode() == Normal {
		return
	}

	str, err := clipboard.ReadAll()
	if err != nil {
		log.Printf("pasteFromClipboard: %v\n", err)
		return
	}

	pos := countToCursor(input.value, input.cursorX+1)
	runes := []rune(input.value)
	input.value = string(runes[:pos])
	input.value += str
	input.value += string(runes[pos:])
	input.cursorX += runewidth.StringWidth(str)
}

func (root *Root) rangeToString(startX, startY, endX, endY int) (string, error) {
	startX, startY, endX, endY = normalizeRange(startX, startY, endX, endY)
	if root.scr.mouseRectangle {
		return root.scr.rectangleToString(root.Doc, startX, startY, endX, endY)
	}
	return root.scr.lineRangeToString(root.Doc, startX, startY, endX, endY)
}

func normalizeRange(startX, startY, endX, endY int) (int, int, int, int) {
	if endY < startY {
		startY, endY = endY, startY
		startX, endX = endX, startX
	}
	if startY == endY {
		if endX < startX {
			startX, endX = endX, startX
		}
	}
	return startX, startY, endX, endY
}

// lineRangeToString returns the selection.
// The arguments must be startX <= endX and startY <= endY.
// ...◼◻◻◻◻◻
// ◻◻◻◻◻◻◻◻
// ◻◻◻◼......
func (scr SCR) lineRangeToString(m *Document, startX, startY, endX, endY int) (string, error) {
	l1 := scr.lineNumber(startY)
	lineC1, ok := scr.lines[l1.number]
	if !ok || !lineC1.valid {
		return "", ErrOutOfRange
	}
	wx1 := scr.branchWidth(lineC1.lc, l1.wrap)

	var l2 LineNumber
	for y := endY; ; y-- {
		l := scr.lineNumber(y)
		if l.number < m.BufEndNum() {
			l2 = l
			endY = y
			break
		}
	}

	lineC2, ok := scr.lines[l2.number]
	if !ok || !lineC2.valid {
		return "", ErrOutOfRange
	}
	wx2 := scr.branchWidth(lineC2.lc, l2.wrap)

	if l1.number == l2.number {
		x1 := m.x + startX + wx1
		x2 := m.x + endX + wx2
		return scr.selectLine(lineC1, x1, x2+1), nil
	}

	var buff strings.Builder

	first := scr.selectLine(lineC1, m.x+startX+wx1, -1)
	buff.WriteString(first)
	buff.WriteByte('\n')

	for y := startY + 1; y < endY; y++ {
		ln := scr.lineNumber(y)
		if ln.number == l1.number || ln.number == l2.number || ln.wrap > 0 {
			continue
		}
		lineC, ok := scr.lines[ln.number]
		if !ok || !lineC.valid {
			break
		}
		str := scr.selectLine(lineC, 0, -1)
		buff.WriteString(str)
		buff.WriteByte('\n')
	}

	last := scr.selectLine(lineC2, 0, m.x+endX+wx2+1)
	buff.WriteString(last)
	return buff.String(), nil
}

// rectangleToByte returns a rectangular range.
// The arguments must be startX <= endX and startY <= endY.
// ...◼◻◻...
// ...◻◻◻...
// ...◻◻◼...
func (scr SCR) rectangleToString(m *Document, startX, startY, endX, endY int) (string, error) {
	var buff strings.Builder
	for y := startY; y <= endY; y++ {
		ln := scr.lineNumber(y)
		lineC, ok := scr.lines[ln.number]
		if !ok || !lineC.valid {
			break
		}
		wx := scr.branchWidth(lineC.lc, ln.wrap)
		str := scr.selectLine(lineC, m.x+startX+wx, m.x+endX+wx+1)
		buff.WriteString(str)
		buff.WriteByte('\n')
	}
	return buff.String(), nil
}

// branchWidth returns the leftmost position of the number of wrapped line.
// If the wide character is at the right end, it may wrap one character forward.
func (scr SCR) branchWidth(lc contents, branch int) int {
	i := 0
	w := scr.startX
	x := 0
	for n := range lc {
		c := lc[n]
		if w+c.width > scr.vWidth {
			i++
			w = scr.startX
		}
		if i >= branch {
			break
		}
		w += c.width
		x += c.width
	}
	return x
}

// selectLine returns a string in the specified range on one line.
// startX: start position (relative to the visible area), endX: end position (relative to the visible area, exclusive).
// The arguments must be startX <= endX.
func (scr SCR) selectLine(lineC LineC, startX int, endX int) string {
	maxX := len(lineC.lc)
	// -1 is a special max value.
	if endX == -1 {
		endX = maxX
	}
	startX -= scr.startX
	endX -= scr.startX
	startX = max(0, startX)
	endX = max(0, endX)
	startX = min(startX, maxX)
	endX = min(endX, maxX)
	start := lineC.pos.n(startX)
	end := lineC.pos.n(endX)

	if start > len(lineC.str) || end > len(lineC.str) || start > end {
		log.Printf("selectLine:len(%d):start(%d):end(%d)\n", len(lineC.str), start, end)
		return ""
	}
	return lineC.str[start:end]
}
