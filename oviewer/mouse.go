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

var (
	WheelScrollNum      = 2                      // WheelScrollNum is the number of lines to scroll with the mouse wheel.
	DoubleClickInterval = 500 * time.Millisecond // Double click detection time
	DoubleClickDistance = 2                      // Maximum allowed movement distance (in characters) for double click detection
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
	// Avoid redrawing with other mouse events.
	// Skipping redraw here prevents unnecessary screen updates for mouse events
	// that do not affect the view, improving performance and reducing flicker.
	root.skipDraw = true
}

// wheelUp moves the mouse wheel up.
func (root *Root) wheelUp(context.Context) {
	root.setMessage("")
	root.moveUp(WheelScrollNum)
}

// wheelDown moves the mouse wheel down.
func (root *Root) wheelDown(context.Context) {
	root.setMessage("")
	root.moveDown(WheelScrollNum)
}

func (root *Root) wheelRight(ctx context.Context) {
	root.setMessage("")
	if root.Doc.ColumnMode {
		root.moveRightOne(ctx)
	} else {
		root.moveRight(root.Doc.HScrollWidthNum)
	}
}

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

// handlePrimaryButtonClick handles primary button click events (common logic)
func (root *Root) handlePrimaryButtonClick(ctx context.Context, ev *tcell.EventMouse) bool {
	x, y := ev.Position()
	now := time.Now()

	// Check for double click
	if root.isDoubleClick(x, y, now) {
		root.handleDoubleClick(ctx, ev)
		return true
	}

	// Update click state for single click
	root.updateClickState(x, y, now)

	// Start new selection
	root.startSelection(x, y, ev.Modifiers())
	return true
}

// startSelection initializes a new selection
func (root *Root) startSelection(x, y int, modifiers tcell.ModMask) {
	root.scr.x1, root.scr.y1 = x, y
	root.scr.x2, root.scr.y2 = x, y

	root.scr.mouseSelect = SelectActive
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
		root.selectRange(ctx, ev)
		return true
	case tcell.ButtonNone:
		// Check if mouse has moved enough to be considered a selection
		if root.hasMinimumSelection() {
			root.scr.mouseSelect = SelectCopied
			root.scr.mousePressed = false
			root.CopySelect(ctx)
		} else {
			// Reset selection if no meaningful movement occurred
			root.resetSelect()
		}
		return true
	default:
		return true
	}
}

// Add new function to properly reset click state
// resetClickState resets the click state
func (root *Root) resetClickState() {
	root.clickState.clickCount = 0
	root.clickState.lastClickTime = time.Time{}
	root.clickState.lastClickX = 0
	root.clickState.lastClickY = 0
}

// isDoubleClick checks if the current click is a double click
func (root *Root) isDoubleClick(x, y int, now time.Time) bool {
	if root.clickState.clickCount == 0 {
		return false
	}
	timeDiff := now.Sub(root.clickState.lastClickTime)
	if timeDiff > DoubleClickInterval {
		root.resetClickState() // Reset if too much time has passed
		return false
	}

	distanceX := abs(x - root.clickState.lastClickX)
	distanceY := abs(y - root.clickState.lastClickY)

	return distanceX <= DoubleClickDistance && distanceY <= DoubleClickDistance
}

// updateClickState updates the click state
func (root *Root) updateClickState(x, y int, now time.Time) {
	root.clickState.lastClickTime = now
	root.clickState.lastClickX = x
	root.clickState.lastClickY = y
	root.clickState.clickCount++
}

// handleDoubleClick handles double click events for word selection
func (root *Root) handleDoubleClick(ctx context.Context, ev *tcell.EventMouse) {
	x, y := ev.Position()

	// Find word boundaries
	startX, endX := root.findWordBoundaries(x, y)

	root.scr.x1, root.scr.y1 = startX, y
	root.scr.x2, root.scr.y2 = endX, y
	root.scr.mouseSelect = SelectCopied
	root.scr.mousePressed = false
	root.CopySelect(ctx)
	// Reset click state after handling double click
	root.resetClickState()
}

// findWordBoundaries finds word boundaries for the given position
func (root *Root) findWordBoundaries(x, y int) (int, int) {
	ln := root.scr.lineNumber(y)
	line, ok := root.scr.lines[ln.number]
	if !ok || !line.valid {
		return x, x
	}

	// Convert screen coordinates to content coordinates
	contentX := root.Doc.x + x + root.scr.branchWidth(line.lc, ln.wrap)

	// Get character type at current position
	charType := root.getCharTypeAt(line, contentX)

	// Search left boundary
	startX := x
	for startX > 0 {
		testContentX := root.Doc.x + (startX - 1) + root.scr.branchWidth(line.lc, ln.wrap)
		if root.getCharTypeAt(line, testContentX) != charType {
			break
		}
		startX--
	}

	// Search right boundary
	endX := x
	for endX < root.scr.vWidth-1 {
		testContentX := root.Doc.x + (endX + 1) + root.scr.branchWidth(line.lc, ln.wrap)
		if root.getCharTypeAt(line, testContentX) != charType {
			break
		}
		endX++
	}

	return startX, endX
}

// getCharTypeAt returns character type at the given content position
func (root *Root) getCharTypeAt(line LineC, contentX int) int {
	if contentX < 0 || contentX >= len(line.lc) {
		return 0 // Out of range is treated as whitespace
	}

	char := line.lc[contentX].mainc
	if char == ' ' || char == '\t' {
		return 1 // Whitespace
	}
	if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_' {
		return 2 // Alphanumeric and underscore
	}
	return 3 // Other characters
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// hasMinimumSelection checks if the selection is large enough to be meaningful
func (root *Root) hasMinimumSelection() bool {
	// Require movement of at least 1 position in either direction
	return root.scr.x1 != root.scr.x2 || root.scr.y1 != root.scr.y2
}

// selectRange saves the position by selecting the range with the mouse.
// The selection range is represented by (x1, y1), (x2, y2).
func (root *Root) selectRange(_ context.Context, ev *tcell.EventMouse) {
	root.scr.x2, root.scr.y2 = ev.Position()
}

// resetSelect resets the selection.
func (root *Root) resetSelect() {
	root.scr.mouseSelect = SelectNone
	root.scr.x1, root.scr.y1 = 0, 0
	root.scr.x2, root.scr.y2 = 0, 0
	root.scr.mousePressed = false
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

func (root *Root) rangeToString(x1, y1, x2, y2 int) (string, error) {
	x1, y1, x2, y2 = normalizeRange(x1, y1, x2, y2)
	if root.scr.mouseRectangle {
		return root.scr.rectangleToString(root.Doc, x1, y1, x2, y2)
	}
	return root.scr.lineRangeToString(root.Doc, x1, y1, x2, y2)
}

func normalizeRange(x1, y1, x2, y2 int) (int, int, int, int) {
	if y2 < y1 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}
	if y1 == y2 {
		if x2 < x1 {
			x1, x2 = x2, x1
		}
	}
	return x1, y1, x2, y2
}

// lineRangeToString returns the selection.
// The arguments must be x1 <= x2 and y1 <= y2.
// ...◼◻◻◻◻◻
// ◻◻◻◻◻◻◻◻
// ◻◻◻◼......
func (scr SCR) lineRangeToString(m *Document, x1, y1, x2, y2 int) (string, error) {
	l1 := scr.lineNumber(y1)
	line1, ok := scr.lines[l1.number]
	if !ok || !line1.valid {
		return "", ErrOutOfRange
	}
	wx1 := scr.branchWidth(line1.lc, l1.wrap)

	var l2 LineNumber
	for y := y2; ; y-- {
		l := scr.lineNumber(y)
		if l.number < m.BufEndNum() {
			l2 = l
			y2 = y
			break
		}
	}

	line2, ok := scr.lines[l2.number]
	if !ok || !line2.valid {
		return "", ErrOutOfRange
	}
	wx2 := scr.branchWidth(line2.lc, l2.wrap)

	if l1.number == l2.number {
		x1 := m.x + x1 + wx1
		x2 := m.x + x2 + wx2
		return scr.selectLine(line1, x1, x2+1), nil
	}

	var buff strings.Builder

	first := scr.selectLine(line1, m.x+x1+wx1, -1)
	buff.WriteString(first)
	buff.WriteByte('\n')

	for y := y1 + 1; y < y2; y++ {
		ln := scr.lineNumber(y)
		if ln.number == l1.number || ln.number == l2.number || ln.wrap > 0 {
			continue
		}
		line, ok := scr.lines[ln.number]
		if !ok || !line.valid {
			break
		}
		str := scr.selectLine(line, 0, -1)
		buff.WriteString(str)
		buff.WriteByte('\n')
	}

	last := scr.selectLine(line2, 0, m.x+x2+wx2+1)
	buff.WriteString(last)
	return buff.String(), nil
}

// rectangleToByte returns a rectangular range.
// The arguments must be x1 <= x2 and y1 <= y2.
// ...◼◻◻...
// ...◻◻◻...
// ...◻◻◼...
func (scr SCR) rectangleToString(m *Document, x1, y1, x2, y2 int) (string, error) {
	var buff strings.Builder
	for y := y1; y <= y2; y++ {
		ln := scr.lineNumber(y)
		lineC, ok := scr.lines[ln.number]
		if !ok || !lineC.valid {
			break
		}
		wx := scr.branchWidth(lineC.lc, ln.wrap)
		str := scr.selectLine(lineC, m.x+x1+wx, m.x+x2+wx+1)
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
// The arguments must be x1 <= x2.
func (scr SCR) selectLine(lineC LineC, x1 int, x2 int) string {
	maxX := len(lineC.lc)
	// -1 is a special max value.
	if x2 == -1 {
		x2 = maxX
	}
	x1 -= scr.startX
	x2 -= scr.startX
	x1 = max(0, x1)
	x2 = max(0, x2)
	x1 = min(x1, maxX)
	x2 = min(x2, maxX)
	start := lineC.pos.n(x1)
	end := lineC.pos.n(x2)

	if start > len(lineC.str) || end > len(lineC.str) || start > end {
		log.Printf("selectLine:len(%d):start(%d):end(%d)\n", len(lineC.str), start, end)
		return ""
	}
	return lineC.str[start:end]
}
