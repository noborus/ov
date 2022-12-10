package oviewer

import (
	"fmt"
	"io"
	"strings"

	"github.com/jwalton/gchalk"
)

// NewHelp generates a document for help.
func NewHelp(k KeyBind) (*Document, error) {
	m, err := NewDocument()
	if err != nil {
		return nil, err
	}

	m.append("\t\t\t" + gchalk.WithUnderline().Bold("ov help"))

	str := strings.Split(KeyBindString(k), "\n")
	m.append(str...)
	m.FileName = "Help"
	m.eof = 1
	m.preventReload = true
	m.seekable = false
	return m, err
}

// KeyBindString returns keybind as a string for help.
func KeyBindString(k KeyBind) string {
	var b strings.Builder
	fmt.Fprint(&b, gchalk.Bold("\n\tKey binding\n"))
	fmt.Fprint(&b, "\n")
	k.writeKeyBind(&b, actionExit, "quit")
	k.writeKeyBind(&b, actionCancel, "cancel")
	k.writeKeyBind(&b, actionWriteExit, "output screen and quit")
	k.writeKeyBind(&b, actionWriteBA, "set output screen and quit")
	k.writeKeyBind(&b, actionSuspend, "suspend")
	k.writeKeyBind(&b, actionHelp, "display help screen")
	k.writeKeyBind(&b, actionLogDoc, "display log screen")
	k.writeKeyBind(&b, actionSync, "screen sync")
	k.writeKeyBind(&b, actionFollow, "follow mode toggle")
	k.writeKeyBind(&b, actionFollowAll, "follow all mode toggle")
	k.writeKeyBind(&b, actionToggleMouse, "enable/disable mouse")

	fmt.Fprint(&b, gchalk.Bold("\n\tMoving\n"))
	fmt.Fprint(&b, "\n")
	k.writeKeyBind(&b, actionMoveDown, "forward by one line")
	k.writeKeyBind(&b, actionMoveUp, "backward by one line")
	k.writeKeyBind(&b, actionMoveTop, "go to top of document")
	k.writeKeyBind(&b, actionMoveBottom, "go to end of document")
	k.writeKeyBind(&b, actionMovePgDn, "forward by page")
	k.writeKeyBind(&b, actionMovePgUp, "backward by page")
	k.writeKeyBind(&b, actionMoveHfDn, "forward a half page")
	k.writeKeyBind(&b, actionMoveHfUp, "backward a half page")
	k.writeKeyBind(&b, actionMoveLeft, "scroll to left")
	k.writeKeyBind(&b, actionMoveRight, "scroll to right")
	k.writeKeyBind(&b, actionMoveHfLeft, "scroll left half screen")
	k.writeKeyBind(&b, actionMoveHfRight, "scroll right half screen")
	k.writeKeyBind(&b, actionGoLine, "go to line(input number)")

	fmt.Fprint(&b, gchalk.Bold("\n\tMove document\n"))
	fmt.Fprint(&b, "\n")
	k.writeKeyBind(&b, actionNextDoc, "next document")
	k.writeKeyBind(&b, actionPreviousDoc, "previous document")
	k.writeKeyBind(&b, actionCloseDoc, "close current document")

	fmt.Fprint(&b, gchalk.Bold("\n\tMark position\n"))
	fmt.Fprint(&b, "\n")
	k.writeKeyBind(&b, actionMark, "mark current position")
	k.writeKeyBind(&b, actionRemoveMark, "remove mark current position")
	k.writeKeyBind(&b, actionRemoveAllMark, "remove all mark")
	k.writeKeyBind(&b, actionMoveMark, "move to next marked position")
	k.writeKeyBind(&b, actionMovePrevMark, "move to previous marked position")

	fmt.Fprint(&b, gchalk.Bold("\n\tSearch\n"))
	fmt.Fprint(&b, "\n")
	k.writeKeyBind(&b, actionSearch, "forward search mode")
	k.writeKeyBind(&b, actionBackSearch, "backward search mode")
	k.writeKeyBind(&b, actionNextSearch, "repeat forward search")
	k.writeKeyBind(&b, actionNextBackSearch, "repeat backward search")

	fmt.Fprint(&b, gchalk.Bold("\n\tChange display\n"))
	fmt.Fprint(&b, "\n")
	k.writeKeyBind(&b, actionWrap, "wrap/nowrap toggle")
	k.writeKeyBind(&b, actionColumnMode, "column mode toggle")
	k.writeKeyBind(&b, actionAlternate, "alternate rows of style toggle")
	k.writeKeyBind(&b, actionLineNumMode, "line number toggle")
	k.writeKeyBind(&b, actionPlain, "original decoration toggle")

	fmt.Fprint(&b, gchalk.Bold("\n\tChange Display with Input\n"))
	fmt.Fprint(&b, "\n")
	k.writeKeyBind(&b, actionViewMode, "view mode selection")
	k.writeKeyBind(&b, actionDelimiter, "column delimiter string")
	k.writeKeyBind(&b, actionHeader, "number of header lines")
	k.writeKeyBind(&b, actionSkipLines, "number of skip lines")
	k.writeKeyBind(&b, actionTabWidth, "TAB width")
	k.writeKeyBind(&b, actionMultiColor, "multi color highlight")
	k.writeKeyBind(&b, actionJumpTarget, "jump target")

	fmt.Fprint(&b, gchalk.Bold("\n\tSection\n"))
	fmt.Fprint(&b, "\n")
	k.writeKeyBind(&b, actionSection, "section delimiter regular expression")
	k.writeKeyBind(&b, actionSectionStart, "section start position")
	k.writeKeyBind(&b, actionNextSection, "next section")
	k.writeKeyBind(&b, actionPrevSection, "previous section")
	k.writeKeyBind(&b, actionLastSection, "last section")
	k.writeKeyBind(&b, actionFollowSection, "follow section mode toggle")

	fmt.Fprint(&b, gchalk.Bold("\n\tClose and reload\n"))
	fmt.Fprint(&b, "\n")
	k.writeKeyBind(&b, actionCloseFile, "close file")
	k.writeKeyBind(&b, actionReload, "reload file")
	k.writeKeyBind(&b, actionWatch, "watch mode")
	k.writeKeyBind(&b, actionWatchInterval, "set watch interval")

	fmt.Fprint(&b, gchalk.Bold("\n\tKey binding when typing\n"))
	fmt.Fprint(&b, "\n")
	k.writeKeyBind(&b, inputCaseSensitive, "case-sensitive toggle")
	k.writeKeyBind(&b, inputRegexpSearch, "regular expression search toggle")
	k.writeKeyBind(&b, inputIncSearch, "incremental search toggle")
	return b.String()
}

func (k KeyBind) writeKeyBind(w io.Writer, action string, detail string) {
	fmt.Fprintf(w, " %-28s * %s\n", "["+strings.Join(k[action], "], [")+"]", detail)
}
