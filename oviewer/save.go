package oviewer

import (
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type saveSelection string

const (
	saveCancel    saveSelection = "cancel"
	saveOverWrite saveSelection = "overwrite"
	saveAppend    saveSelection = "append"
)

// saveBuffer saves the buffer to the specified file.
func (root *Root) saveBuffer(input string) {
	fileName := strings.TrimSpace(input)

	perm := os.FileMode(0644)
	flag := os.O_WRONLY | os.O_CREATE
	_, err := os.Stat(fileName)
	if err == nil {
		root.setMessagef("overwrite? (O)overwrite, (A)Append, (N)cancel:")
		switch root.saveConfirm() {
		case saveOverWrite:
			flag = os.O_WRONLY | os.O_TRUNC
		case saveAppend:
			flag |= os.O_APPEND
		case saveCancel:
			root.setMessage("save cancel")
			return
		}
	}

	file, err := os.OpenFile(fileName, flag, perm)
	if err != nil {
		root.setMessageLogf("cannot save: %s:%s", fileName, err)
		return
	}
	defer file.Close()

	if err := root.Doc.Export(file, root.Doc.BufStartNum(), root.Doc.BufEndNum()); err != nil {
		root.setMessageLogf("cannot save: %s:%s", fileName, err)
		return
	}

	root.setMessageLogf("saved %s", fileName)
}

// saveConfirm waits for the user to confirm the save.
func (root *Root) saveConfirm() saveSelection {
	for {
		ev := root.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyRune {
				switch ev.Rune() {
				case 'o', 'O':
					return saveOverWrite
				case 'a', 'A':
					return saveAppend
				case 'n', 'N', 'q', 'Q':
					return saveCancel
				}
			} else if ev.Key() == tcell.KeyEscape {
				return saveCancel
			}
		}
	}
}
