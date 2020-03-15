package zpager

import (
	"fmt"

	"github.com/gdamore/tcell"
)

func (root *root) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape:
			root.parent.Quit()
			return true
		case tcell.KeyEnter:
			root.model.y++
			root.updateKeys()
			return true
		case tcell.KeyLeft:
			root.model.x--
			root.updateKeys()
			return true
		case tcell.KeyRight:
			root.model.x++
			root.updateKeys()
			return true
		case tcell.KeyDown:
			root.model.y++
			root.updateKeys()
			return true
		case tcell.KeyUp:
			root.model.y--
			root.updateKeys()
			return true
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'Q', 'q':
				root.parent.Quit()
				return true
			case 'S', 's':
				root.model.y = root.model.y + 10000
				root.updateKeys()
				return true
			case 'V', 'v':
				root.model.y = root.model.y + 10
				root.updateKeys()
				return true
			case 'N', 'n':
				root.model.y = root.model.y + 1
				root.updateKeys()
				return true
			case 'P', 'p':
				root.model.y = root.model.y - 1
				root.updateKeys()
				return true
			case 'H', 'h':
				root.SetTitle(root.title)
				root.updateKeys()
				return true
			case 'E', 'e':
				root.status.SetCenter(fmt.Sprintf("x: %d y:%d", root.model.endx, root.model.endy))
				root.SetStatus(root.status)
				return true
			case 'D', 'd':
				root.RemoveWidget(root.title)
				root.RemoveWidget(root.status)
				return true
			}
		}
	}
	return true
}
