package oviewer

import (
	"errors"
	"strings"
)

func (root *root) search(num int) (int, error) {
	for n := num; n < root.Model.endNum; n++ {
		if contains(root.Model.buffer[n], root.input, root.CaseSensitive) {
			return n, nil
		}
	}
	return 0, errors.New("not found")
}

func (root *root) backSearch(num int) (int, error) {
	for n := num; n >= 0; n-- {
		if contains(root.Model.buffer[n], root.input, root.CaseSensitive) {
			return n, nil
		}
	}
	return 0, errors.New("not found")
}

func contains(s, substr string, caseSensitive bool) bool {
	if !caseSensitive {
		strings.Contains(strings.ToLower(s), strings.ToLower(substr))
	}
	return strings.Contains(s, substr)
}

func rangePosition(s string, substr string, number int) (int, int) {
	var start, end int
	i := 0
	for n := 0; n < number-1; n++ {
		j := strings.Index(s[i:], substr)
		if j < 0 {
			return -1, -1
		}
		i += j + len(substr)
	}
	if number == 0 {
		de := strings.Index(s[i:], substr)
		end = i + de
	} else {
		ds := strings.Index(s[i:], substr)
		start = i + ds + 1
		de := strings.Index(s[start:], substr)
		if de < 0 {
			end = len(s)
		} else {
			end = start + de
		}
	}
	return start, end
}

func searchHighlight(s string, lc lineContents, substr string, caseSensitive bool) {
	if !caseSensitive {
		s = strings.ToLower(s)
		substr = strings.ToLower(substr)
	}

	for i := strings.Index(s, substr); i >= 0; {
		start := lc.cMap[i]
		end := lc.cMap[i+len(substr)]
		for ci := start; ci < end; ci++ {
			lc.contents[ci].style = lc.contents[ci].style.Reverse(true)
		}
		j := strings.Index(s[i+1:], substr)
		if j >= 0 {
			i += j + 1
		} else {
			break
		}
	}
}
