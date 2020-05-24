package oviewer

import (
	"errors"
	"regexp"
	"strings"
)

// Search is a forward search.
func (root *Root) Search() {
	root.inputRegexp = regexpComple(root.input, root.CaseSensitive)
	root.postSearch(root.search(root.Model.lineNum))
}

// NextSearch will re-run the forward search.
func (root *Root) NextSearch() {
	root.postSearch(root.search(root.Model.lineNum + root.Header + 1))
}

// BackSearch reverse search.
func (root *Root) BackSearch() {
	root.inputRegexp = regexpComple(root.input, root.CaseSensitive)
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
		if contains(root.Model.buffer[n], root.inputRegexp) {
			return n, nil
		}
	}
	return 0, errors.New("not found")
}

func (root *Root) backSearch(num int) (int, error) {
	for n := num; n >= 0; n-- {
		if contains(root.Model.buffer[n], root.inputRegexp) {
			return n, nil
		}
	}
	return 0, errors.New("not found")
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
	regexEscapeSequence = regexp.MustCompile("\x1b\\[[\\d;*]*m")
)

func contains(s string, re *regexp.Regexp) bool {
	if re == nil || re.String() == "" {
		return false
	}
	s = regexEscapeSequence.ReplaceAllString(s, "")
	return re.MatchString(s)
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

func searchPosition(s string, re *regexp.Regexp) []rangePos {
	if re == nil || re.String() == "" {
		return nil
	}
	var pos []rangePos
	s = regexEscapeSequence.ReplaceAllString(s, "")
	for _, f := range re.FindAllIndex([]byte(s), -1) {
		r := rangePos{f[0], f[1]}
		pos = append(pos, r)
	}
	return pos
}
