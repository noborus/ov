package oviewer

import (
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// saveSelection represents the state of the save selection.
type saveSelection string

const (
	// saveCancel is a save cancel.
	saveCancel saveSelection = "cancel"
	// saveOverWrite is a save overwrite.
	saveOverWrite saveSelection = "overwrite"
	// saveAppend is a save append.
	saveAppend saveSelection = "append"
	// saveIgnore is a save ignore.
	saveIgnore saveSelection = "ignore"
)

// saveBuffer saves the buffer to the specified file.
func (root *Root) saveBuffer(input string) {
	fileName := strings.TrimSpace(input)

	flag, err := root.saveFlag(fileName)
	if err != nil {
		root.setMessage("save cancel")
		return
	}
	perm := os.FileMode(0o644)
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

func (root *Root) saveFlag(fileName string) (int, error) {
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
			return 0, ErrCancel
		}
	}
	return flag, nil
}

// saveConfirm waits for the user to confirm the save.
func (root *Root) saveConfirm() saveSelection {
	for {
		ev := root.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			s := saveConfirmKey(ev)
			if s != saveIgnore {
				return s
			}
		}
	}
}

func saveConfirmKey(ev *tcell.EventKey) saveSelection {
	switch ev.Key() {
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'o', 'O':
			return saveOverWrite
		case 'a', 'A':
			return saveAppend
		case 'n', 'N', 'q', 'Q':
			return saveCancel
		}
	case tcell.KeyEscape:
		return saveCancel
	}
	return saveIgnore
}
