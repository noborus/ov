package oviewer

import (
	"regexp"
	"strings"
)

// NextSearch will re-run the forward search.
func (root *Root) NextSearch() {
	root.postSearch(root.search(root.Model.lineNum + root.Header + 1))
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
		if contains(root.Model.buffer[n], root.Input.reg) {
			return n, nil
		}
	}
	return 0, ErrNotFound
}

func (root *Root) backSearch(num int) (int, error) {
	for n := num; n >= 0; n-- {
		if contains(root.Model.buffer[n], root.Input.reg) {
			return n, nil
		}
	}
	return 0, ErrNotFound
}

func regexpComple(r string, caseSensitive bool) *regexp.Regexp {
	if !caseSensitive {
		r = "(?i)" + r
	}
	re, err := regexp.Compile(r)
	if err == nil {
		return re
	}
	r = regexp.QuoteMeta(r)
	re, err = regexp.Compile(r)
	if err == nil {
		return nil
	}
	return re
}

var (
	stripEscapeSequence = regexp.MustCompile("\x1b\\[[\\d;*]*m")
)

func contains(s string, re *regexp.Regexp) bool {
	if re == nil || re.String() == "" {
		return false
	}
	s = stripEscapeSequence.ReplaceAllString(s, "")
	return re.MatchString(s)
}

func rangePosition(s, substr string, number int) (int, int) {
	i := 0

	if number == 0 {
		de := strings.Index(s[i:], substr)
		return 0, i + de
	}

	for n := 0; n < number-1; n++ {
		j := strings.Index(s[i:], substr)
		if j < 0 {
			return -1, -1
		}
		i += j + len(substr)
	}

	ds := strings.Index(s[i:], substr)
	if ds < 0 {
		return -1, -1
	}
	start := i + ds + len(substr)
	de := -1
	if start < len(s) {
		de = strings.Index(s[start:], substr)
	}

	end := start + de
	if de < 0 {
		end = len(s)
	}
	return start, end
}

func searchPosition(s string, re *regexp.Regexp) [][]int {
	if re == nil || re.String() == "" {
		return nil
	}
	var pos [][]int
	s = stripEscapeSequence.ReplaceAllString(s, "")
	pos = append(pos, re.FindAllIndex([]byte(s), -1)...)
	return pos
}
