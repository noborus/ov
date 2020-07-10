package oviewer

import (
	"regexp"
	"strings"
)

// Search is a Up search.
func (root *Root) Search(input string) {
	root.Input.reg = regexpComple(input, root.CaseSensitive)
	root.goSearchLine(root.search(root.Model.lineNum))
}

// BackSearch reverse search.
func (root *Root) BackSearch(input string) {
	root.Input.reg = regexpComple(input, root.CaseSensitive)
	root.goSearchLine(root.backSearch(root.Model.lineNum))
}

// NextSearch will re-run the forward search.
func (root *Root) NextSearch() {
	root.goSearchLine(root.search(root.Model.lineNum + root.Header + 1))
}

// NextBackSearch will re-run the reverse search.
func (root *Root) NextBackSearch() {
	root.goSearchLine(root.backSearch(root.Model.lineNum + root.Header - 1))
}

// goSearchLine moves to the line found.
func (root *Root) goSearchLine(lineNum int, err error) {
	if err != nil {
		root.message = err.Error()
		return
	}
	root.moveNum(lineNum - root.Header)
}

// search is searches below from the specified line.
func (root *Root) search(num int) (int, error) {
	for n := num; n < root.Model.BufEndNum(); n++ {
		if contains(root.Model.buffer[n], root.Input.reg) {
			return n, nil
		}
	}
	return 0, ErrNotFound
}

// backsearch is searches upward from the specified line.
func (root *Root) backSearch(num int) (int, error) {
	for n := num; n >= 0; n-- {
		if contains(root.Model.buffer[n], root.Input.reg) {
			return n, nil
		}
	}
	return 0, ErrNotFound
}

// regexpComple is regexp.Compile the search string.
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
	// stripEscapeSequence is a regular expression that excludes escape sequences.
	stripEscapeSequence = regexp.MustCompile("\x1b\\[[\\d;*]*m")
)

// contains returns a bool containing the search string.
func contains(s string, re *regexp.Regexp) bool {
	if re == nil || re.String() == "" {
		return false
	}
	s = stripEscapeSequence.ReplaceAllString(s, "")
	return re.MatchString(s)
}

// rangePosition returns the range starting and ending from the s,substr string.
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

// searchPosition returns an array of the beginning and end of the search string.
func searchPosition(s string, re *regexp.Regexp) [][]int {
	if re == nil || re.String() == "" {
		return nil
	}

	var pos [][]int
	s = stripEscapeSequence.ReplaceAllString(s, "")
	pos = append(pos, re.FindAllIndex([]byte(s), -1)...)
	return pos
}
