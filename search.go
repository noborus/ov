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
		return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
	}
	return strings.Contains(s, substr)
}

type rangePos struct {
	start int
	end   int
}

func rangePosition(s string, substr string, number int) rangePos {
	r := rangePos{0, 0}
	i := 0
	for n := 0; n < number-1; n++ {
		j := strings.Index(s[i:], substr)
		if j < 0 {
			return rangePos{-1, -1}
		}
		i += j + len(substr)
	}
	if number == 0 {
		de := strings.Index(s[i:], substr)
		r.end = i + de
	} else {
		ds := strings.Index(s[i:], substr)
		r.start = i + ds + 1
		de := -1
		if r.start < len(s) {
			de = strings.Index(s[r.start:], substr)
		}
		if de < 0 {
			r.end = len(s)
		} else {
			r.end = r.start + de
		}
	}
	return r
}

func searchPosition(s string, substr string, caseSensitive bool) []rangePos {
	var pos []rangePos
	if !caseSensitive {
		s = strings.ToLower(s)
		substr = strings.ToLower(substr)
	}

	for i := strings.Index(s, substr); i >= 0; {
		r := rangePos{i, i + len(substr)}
		pos = append(pos, r)
		j := strings.Index(s[i+1:], substr)
		if j >= 0 {
			i += j + 1
		} else {
			break
		}
	}
	return pos
}
