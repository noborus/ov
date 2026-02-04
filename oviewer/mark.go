package oviewer

import (
	"context"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v3"
)

// addMark adds the current line number to the mark.
func (root *Root) addMark(context.Context) {
	lN := root.firstBodyLine()
	root.Doc.marked = root.Doc.marked.remove(lN)
	lineC := root.lineContent(lN)
	mark := Mark{lineNum: lN, contents: lineC.lc}
	root.Doc.marked = append(root.Doc.marked, mark)
	root.Doc.markedPoint = len(root.Doc.marked) - 1
	root.setMessagef("Marked to line %d", lN-root.Doc.firstLine()+1)
}

// firstBodyLine returns the first line number of the body.
func (root *Root) firstBodyLine() int {
	ln := root.scr.lineNumber(root.Doc.headerHeight + root.Doc.sectionHeaderHeight)
	return ln.number
}

// addMarks adds multiple line numbers to the mark list at once.
func (root *Root) addMarks(_ context.Context, marks MarkedList) {
	root.Doc.markedPoint = -1
	if len(marks) == 0 {
		root.setMessagef("Added %d marks", 0)
		return
	}

	exists := make(map[int]bool, len(root.Doc.marked))
	for _, m := range root.Doc.marked {
		exists[m.lineNum] = true
	}

	added := 0
	for _, m := range marks {
		if _, ok := exists[m.lineNum]; ok {
			continue
		}
		exists[m.lineNum] = true
		root.Doc.marked = append(root.Doc.marked, m)
		added++
	}
	root.setMessagef("Added %d marks", added)
}

// removeMark removes the current line number from the mark list.
func (root *Root) removeMark(context.Context) {
	lN := root.firstBodyLine()
	marked := root.Doc.marked.remove(lN)
	if len(root.Doc.marked) == len(marked) {
		root.setMessagef("Not marked line %d", lN-root.Doc.firstLine()+1)
		return
	}
	root.Doc.marked = marked
	root.Doc.markedPoint -= 1
	root.setMessagef("Remove the mark at line %d", lN-root.Doc.firstLine()+1)
}

// removeAllMark removes all marks.
func (root *Root) removeAllMark(context.Context) {
	root.Doc.marked = nil
	root.Doc.markedPoint = -1
	root.setMessage("Remove all marks")
}

// goMarkNumber moves to the specified mark number or relative position (+n/-n).
func (root *Root) goMarkNumber(input string) {
	if root.previousSidebarMode != SidebarModeMark {
		root.toggleSidebar(context.Background(), root.previousSidebarMode)
	}
	if len(root.Doc.marked) == 0 {
		return
	}
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return
	}
	idx := calcMarkIndex(input, root.Doc.markedPoint)
	idx = min(max(0, idx), len(root.Doc.marked)-1)
	root.Doc.markedPoint = idx
	root.goLineNumber(root.Doc.marked[root.Doc.markedPoint].lineNum)
}

// calcMarkIndex calculates the mark index from the input string.
// If the input is a relative position (+n/-n), it calculates the index accordingly.
func calcMarkIndex(input string, current int) int {
	if len(input) == 0 {
		return current
	}
	if input[0] == '+' || input[0] == '-' {
		n, err := strconv.Atoi(input)
		if err != nil {
			return current
		}
		return current + n
	}
	n, err := strconv.Atoi(input)
	if err != nil {
		return current
	}
	return n
}

// nextMark moves to the next mark.
func (root *Root) nextMark(context.Context) {
	if len(root.Doc.marked) == 0 {
		return
	}

	if len(root.Doc.marked) > root.Doc.markedPoint+1 {
		root.Doc.markedPoint++
	} else {
		root.Doc.markedPoint = 0
	}
	root.goLineNumber(root.Doc.marked[root.Doc.markedPoint].lineNum)
}

// prevMark moves to the previous mark.
func (root *Root) prevMark(context.Context) {
	if len(root.Doc.marked) == 0 {
		return
	}

	if root.Doc.markedPoint > 0 {
		root.Doc.markedPoint--
	} else {
		root.Doc.markedPoint = len(root.Doc.marked) - 1
	}
	root.goLineNumber(root.Doc.marked[root.Doc.markedPoint].lineNum)
}

func (list MarkedList) remove(lineNumber int) MarkedList {
	newList := make(MarkedList, 0, len(list))
	for _, m := range list {
		if m.lineNum != lineNumber {
			newList = append(newList, m)
		}
	}
	return newList
}

func (list MarkedList) contains(lineNumber int) bool {
	for _, mark := range list {
		if mark.lineNum == lineNumber {
			return true
		}
	}
	return false
}

// eventAddMarks represents an event to add multiple marks.
type eventAddMarks struct {
	tcell.EventTime
	marks MarkedList
}

// markByPattern marks lines matching the pattern.
func (root *Root) markByPattern(ctx context.Context, pattern string) {
	searcher := root.setSearcher(pattern, root.Config.CaseSensitive)
	if searcher == nil {
		return
	}
	go root.markByPatternImpl(ctx, pattern, searcher)
}

// markByPatternImpl marks lines matching the pattern.
func (root *Root) markByPatternImpl(ctx context.Context, pattern string, searcher Searcher) {
	var marks MarkedList
	root.setMessagef("mark by pattern %v", pattern)

	for ln := root.Doc.BufStartNum(); ln < root.Doc.BufEndNum(); ln++ {
		select {
		case <-ctx.Done():
			return
		default:
		}
		lineC := root.lineContent(ln)
		if searcher.MatchString(lineC.str) {
			mark := Mark{lineNum: ln, contents: lineC.lc}
			marks = append(marks, mark)
		}
	}
	root.sendAddMarks(marks)
}

// sendAddMarks fires the eventAddMarks event.
func (root *Root) sendAddMarks(marks MarkedList) {
	ev := &eventAddMarks{}
	ev.marks = marks
	ev.SetEventNow()
	root.postEvent(ev)
}
