package oviewer

import (
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// saveBuffer saves the buffer to the specified file.
func (root *Root) saveBuffer(input string) {
	fileName := strings.TrimSpace(input)

	_, err := os.Stat(fileName)
	if err == nil {
		root.setMessagef("overwrite? (O)overwrite (N)cancel:")
		if !root.saveConfirm() {
			root.setMessagef("save cancel")
			return
		}
	}

	file, err := os.Create(fileName)
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
func (root *Root) saveConfirm() bool {
	for {
		ev := root.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyRune {
				switch ev.Rune() {
				case 'o', 'O':
					return true
				case 'n', 'N', 'q', 'Q':
					return false
				}
			} else if ev.Key() == tcell.KeyEscape {
				return false
			}
		}
	}
}
