package oviewer

import (
	"errors"
	"strings"
)

// Search is a forward search.
func (root *Root) Search() {
	root.postSearch(root.search(root.Model.lineNum))
}

// NextSearch will re-run the forward search.
func (root *Root) NextSearch() {
	root.postSearch(root.search(root.Model.lineNum + root.Header + 1))
}

// BackSearch reverse search.
func (root *Root) BackSearch() {
	root.postSearch(root.backSearch(root.Model.lineNum))
}

// NextBackSearch will re-run the reverse search.
func (root *Root) NextBackSearch() {
	root.postSearch(root.backSearch(root.Model.lineNum + root.Header - 1))
}

func (root *Root) postSearch(lineNum int, err error) {
	if err != nil {
		root.message = err.Error()
		return
	}
	root.moveNum(lineNum - root.Header)
}

func (root *Root) search(num int) (int, error) {
	for n := num; n < root.Model.BufEndNum(); n++ {
		if contains(root.Model.buffer[n], root.input, root.CaseSensitive) {
			return n, nil
		}
	}
	return 0, errors.New("not found")
}

func (root *Root) backSearch(num int) (int, error) {
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

func rangePosition(s, substr string, number int) rangePos {
	r := rangePos{0, 0}
	i := 0

	if number == 0 {
		de := strings.Index(s[i:], substr)
		r.end = i + de
		return r
	}

	for n := 0; n < number-1; n++ {
		j := strings.Index(s[i:], substr)
		if j < 0 {
			return rangePos{-1, -1}
		}
		i += j + len(substr)
	}

	ds := strings.Index(s[i:], substr)
	if ds < 0 {
		return rangePos{-1, -1}
	}
	r.start = i + ds + len(substr)
	de := -1
	if r.start < len(s) {
		de = strings.Index(s[r.start:], substr)
	}

	r.end = r.start + de
	if de < 0 {
		r.end = len(s)
	}
	return r
}

func searchPosition(s, substr string, caseSensitive bool) []rangePos {
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
