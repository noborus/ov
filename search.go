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

// Search for searchWord from contents and reverse.
func searchContents(cp *[]content, searchWord string, CaseSensitive bool) {
	contents := *cp
	if searchWord == "" {
		for n := range contents {
			contents[n].style = contents[n].style.Reverse(false)
		}
		return
	}
	s, cIndex := contentsToStr(contents)
	if !CaseSensitive {
		s = strings.ToLower(s)
		searchWord = strings.ToLower(searchWord)
	}
	for i := strings.Index(s, searchWord); i >= 0; {
		start := cIndex[i]
		end := cIndex[i+len(searchWord)]
		for ci := start; ci < end; ci++ {
			contents[ci].style = contents[ci].style.Reverse(true)
		}
		j := strings.Index(s[i+1:], searchWord)
		if j >= 0 {
			i += j + 1
		} else {
			break
		}
	}
}

// contentsToStr converts a content array into a single-line string.
// Returns a character string.
// Also return the position of characters and the position of contents in map.
func contentsToStr(contents []content) (string, map[int]int) {
	buf := make([]rune, 0)
	cIndex := make(map[int]int)
	byteLen := 0
	for n := range contents {
		if contents[n].width > 0 {
			cIndex[byteLen] = n
			buf = append(buf, contents[n].mainc)
			b := string(contents[n].mainc)
			byteLen += len(b)
		}
	}
	s := string(buf)
	cIndex[len(s)] = len(contents)
	return s, cIndex
}
