package oviewer

import (
	"regexp"
	"strings"
)

// search is forward search.
func (root *Root) search(input string) {
	root.input.reg = regexpComple(input, root.CaseSensitive)
	root.goSearchLine(root.searchLine(root.Doc.lineNum))
}

// backSearch is backward search.
func (root *Root) backSearch(input string) {
	root.input.reg = regexpComple(input, root.CaseSensitive)
	root.goSearchLine(root.backSearchLine(root.Doc.lineNum))
}

// nextSearch is forward search again.
func (root *Root) nextSearch() {
	root.goSearchLine(root.searchLine(root.Doc.lineNum + root.Doc.Header + 1))
}

// nextBackSearch is backward　search again..
func (root *Root) nextBackSearch() {
	root.goSearchLine(root.backSearchLine(root.Doc.lineNum + root.Doc.Header - 1))
}

// goSearchLine moves to the line found.
func (root *Root) goSearchLine(lineNum int, err error) {
	if err != nil {
		root.message = err.Error()
		return
	}
	root.moveLine(lineNum - root.Doc.Header)
}

// searchLine is searches below from the specified line.
func (root *Root) searchLine(num int) (int, error) {
	for n := num; n < root.Doc.BufEndNum(); n++ {
		if contains(root.Doc.GetLine(n), root.input.reg) {
			return n, nil
		}
	}
	return 0, ErrNotFound
}

// backsearch is searches upward from the specified line.
func (root *Root) backSearchLine(num int) (int, error) {
	if num > root.Doc.BufEndNum() {
		num = root.Doc.BufEndNum() - 1
	}
	for n := num; n >= 0; n-- {
		if contains(root.Doc.GetLine(n), root.input.reg) {
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

// stripEscapeSequence is a regular expression that excludes escape sequences.
var stripEscapeSequence = regexp.MustCompile("(\x1b\\[[\\d;*]*m)|.\b")

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
