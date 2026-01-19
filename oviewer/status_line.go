package oviewer

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/rivo/uniseg"
)

// DefaultStatusLine is the number of lines in the status line.
const DefaultStatusLine = 1

// drawStatus draws a status line.
func (root *Root) drawStatus() {
	if root.scr.statusLineHeight == 0 {
		root.Screen.HideCursor()
		return
	}

	root.clearY(root.Doc.statusPos)
	root.drawLeftStatus()
	root.drawRightStatus()
}

// drawLeftStatus draws the status of the left side and positions the cursor.
func (root *Root) drawLeftStatus() {
	var pos int
	if root.input.Event.Mode() == Normal {
		pos = root.normalLeftStatus()
	} else {
		pos = root.inputLeftStatus()
	}
	root.Screen.ShowCursor(pos, root.Doc.statusPos)
}

// normalLeftStatus returns the status of the left side of the normal mode.
func (root *Root) normalLeftStatus() int {
	sColor := color.White
	numVisible := false
	if root.DocumentLen() > 1 && root.Doc.documentType != DocHelp && root.Doc.documentType != DocLog {
		numVisible = true
		if root.CurrentDoc != 0 {
			sColor = color.Color((root.CurrentDoc + 8) % 16)
		}
	}

	str := root.StrLeftStatus(numVisible)
	style := applyStyle(tcell.StyleDefault, root.Doc.Style.LeftStatus)
	if root.Doc.Normal.InvertColor {
		style = style.Foreground(color.IsValid + sColor).Reverse(true)
	}
	root.Screen.PutStrStyled(0, root.Doc.statusPos, str, style)

	cursorColor := color.GetColor(root.Doc.Style.LeftStatus.Foreground)
	root.Screen.SetCursorStyle(tcell.CursorStyle(root.Doc.Normal.CursorType), cursorColor)

	return uniseg.StringWidth(str)
}

// StrLeftStatus returns the left status string.
func (root *Root) StrLeftStatus(numVisible bool) string {
	var leftStatus strings.Builder

	if numVisible {
		leftStatus.WriteString("[")
		leftStatus.WriteString(strconv.Itoa(root.CurrentDoc))
		leftStatus.WriteString("]")
	}
	if root.Doc.pauseFollow {
		leftStatus.WriteString("||")
	}
	leftStatus.WriteString(root.statusMode())
	leftStatus.WriteString(root.displayTitle())
	leftStatus.WriteString(":")
	leftStatus.WriteString(root.message)
	return leftStatus.String()
}

// statusMode returns the status mode of the document.
func (root *Root) statusMode() string {
	if root.Doc.WatchMode {
		// Watch mode doubles as FollowSection mode.
		return "(Watch)"
	}
	if root.Doc.FollowSection {
		return "(Follow Section)"
	}
	if root.FollowAll {
		return "(Follow All)"
	}
	if root.Doc.FollowMode && root.Doc.FollowName {
		return "(Follow Name)"
	}
	if root.Doc.FollowMode {
		return "(Follow Mode)"
	}
	return ""
}

// inputLeftStatus draws the input status on the left side and returns the cursor position.
func (root *Root) inputLeftStatus() int {
	input := root.input
	prompt := root.inputPrompt()

	style := applyStyle(tcell.StyleDefault, root.Doc.Style.LeftStatus)
	root.Screen.PutStrStyled(0, root.Doc.statusPos, prompt+input.value, style)
	cursorColor := color.GetColor(root.Doc.Style.LeftStatus.Foreground)
	root.Screen.SetCursorStyle(tcell.CursorStyle(root.Doc.Input.CursorType), cursorColor)

	return uniseg.StringWidth(prompt) + input.cursorX
}

// inputPrompt returns a string describing the input field.
func (root *Root) inputPrompt() string {
	var prompt strings.Builder
	mode := root.input.Event.Mode()
	modePrompt := root.input.Event.Prompt()

	if mode == Search || mode == Backsearch || mode == Filter {
		prompt.WriteString(root.searchOpt)
	}
	prompt.WriteString(modePrompt)
	return prompt.String()
}

// drawRightStatus draws the status of the right side.
func (root *Root) drawRightStatus() {
	next := ""
	if !root.Doc.BufEOF() {
		next = "..."
	}
	numStr := fmt.Sprintf("(%d/%d%s)", root.Doc.firstLine()+root.Doc.topLN+1, root.Doc.BufEndNum(), next)
	if atomic.LoadInt32(&root.Doc.tmpFollow) == 1 {
		numStr = fmt.Sprintf("(?/%d%s)", root.Doc.storeEndNum(), next)
	}
	numWidth := uniseg.StringWidth(numStr)
	style := applyStyle(tcell.StyleDefault, root.Doc.Style.RightStatus)
	root.Screen.PutStrStyled(root.scr.vWidth-numWidth, root.Doc.statusPos, numStr, style)
}
