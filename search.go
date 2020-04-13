package oviewer

import (
	"errors"
	"strings"
)

func (root *root) search(num int) (int, error) {
	for n := num; n < root.Model.endNum; n++ {
		if root.contains(root.Model.buffer[n], root.input) {
			return n, nil
		}
	}
	return 0, errors.New("not found")
}

func (root *root) backSearch(num int) (int, error) {
	for n := num; n >= 0; n-- {
		if root.contains(root.Model.buffer[n], root.input) {
			return n, nil
		}
	}
	return 0, errors.New("not found")
}

func (root *root) contains(s, substr string) bool {
	if !root.CaseSensitive {
		strings.Contains(strings.ToLower(s), strings.ToLower(substr))
	}
	return strings.Contains(s, substr)
}

func searchHighlight(line string, lc lineContents, searchWord string, CaseSensitive bool) {
	s := line
	if !CaseSensitive {
		s = strings.ToLower(s)
		searchWord = strings.ToLower(searchWord)
	}

	for i := strings.Index(s, searchWord); i >= 0; {
		start := lc.cMap[i]
		end := lc.cMap[i+len(searchWord)]
		for ci := start; ci < end; ci++ {
			lc.contents[ci].style = lc.contents[ci].style.Reverse(true)
		}
		j := strings.Index(s[i+1:], searchWord)
		if j >= 0 {
			i += j + 1
		} else {
			break
		}
	}
}
