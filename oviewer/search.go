package oviewer

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"golang.org/x/sync/errgroup"
)

// SearchType represents the type of search.
type SearchType int

const (
	searchSensitive SearchType = iota
	searchInsensitive
	searchRegexp
)

// search is forward search.
func (root *Root) search(ctx context.Context, input string) {
	if input == "" {
		root.input.reg = nil
		return
	}
	root.input.value = input
	root.Doc.lineNum--
	root.nextSearch(ctx)
}

// backSearch is backward search.
func (root *Root) backSearch(ctx context.Context, input string) {
	if input == "" {
		root.input.reg = nil
		return
	}
	root.input.value = input
	root.nextBackSearch(ctx)
}

// nextSearch is forward search again.
func (root *Root) nextSearch(ctx context.Context) {
	root.setMessage(fmt.Sprintf("search:%v (%v)Cancel", root.input.value, strings.Join(root.cancelKeys, ",")))

	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg.Go(func() error {
		return root.cancelWait(cancel)
	})

	eg.Go(func() error {
		lineNum, err := root.searchLine(ctx, root.Doc.lineNum+root.Doc.Header+1)
		if err != nil {
			return err
		}
		root.moveLine(lineNum - root.Doc.Header)
		return nil
	})

	if err := eg.Wait(); err != nil {
		root.setMessage(err.Error())
		return
	}
	root.setMessage(fmt.Sprintf("search:%v", root.input.value))
}

// nextBackSearch is backward　search again.
func (root *Root) nextBackSearch(ctx context.Context) {
	root.setMessage(fmt.Sprintf("search:%v (%v)Cancel", root.input.value, strings.Join(root.cancelKeys, ",")))

	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg.Go(func() error {
		return root.cancelWait(cancel)
	})

	eg.Go(func() error {
		lineNum, err := root.backSearchLine(ctx, root.Doc.lineNum+root.Doc.Header-1)
		if err != nil {
			return err
		}
		root.moveLine(lineNum - root.Doc.Header)
		return nil
	})

	if err := eg.Wait(); err != nil {
		root.setMessage(err.Error())
		return
	}
	root.setMessage(fmt.Sprintf("search:%v", root.input.value))
}

// searchLine is searches below from the specified line.
func (root *Root) searchLine(ctx context.Context, num int) (int, error) {
	defer root.searchQuit()
	num = max(num, 0)

	if root.input.value == "" {
		return num, nil
	}

	root.input.reg = regexpComple(root.input.value, root.CaseSensitive)
	if root.input.reg == nil {
		return num, nil
	}

	searchType := getSearchType(root.input.value, root.CaseSensitive)

	for n := num; n < root.Doc.BufEndNum(); n++ {
		if root.contains(root.Doc.GetLine(n), searchType) {
			return n, nil
		}
		select {
		case <-ctx.Done():
			return 0, ErrCancel
		default:
		}
	}
	return 0, ErrNotFound
}

// backsearch is searches upward from the specified line.
func (root *Root) backSearchLine(ctx context.Context, num int) (int, error) {
	defer root.searchQuit()
	num = min(num, root.Doc.BufEndNum()-1)

	root.input.reg = regexpComple(root.input.value, root.CaseSensitive)
	if root.input.reg == nil {
		return num, nil
	}

	searchType := getSearchType(root.input.value, root.CaseSensitive)

	for n := num; n >= 0; n-- {
		if root.contains(root.Doc.GetLine(n), searchType) {
			return n, nil
		}
		select {
		case <-ctx.Done():
			return 0, ErrCancel
		default:
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
		return re
	}
	log.Printf("regexpCompile failed %s", r)
	return nil
}

// stripEscapeSequence is a regular expression that excludes escape sequences.
var stripEscapeSequence = regexp.MustCompile("(\x1b\\[[\\d;*]*m)|.\b")

// contains returns a bool containing the search string.
func (root *Root) contains(s string, t SearchType) bool {
	if strings.ContainsAny(s, "\x1b\b") {
		s = stripEscapeSequence.ReplaceAllString(s, "")
	}
	switch t {
	case searchSensitive:
		return strings.Contains(s, root.input.value)
	case searchInsensitive:
		return strings.Contains(strings.ToLower(s), strings.ToLower(root.input.value))
	default:
		return root.input.reg.MatchString(s)
	}
}

func getSearchType(t string, caseSensitive bool) SearchType {
	searchType := searchRegexp
	if t == regexp.QuoteMeta(t) {
		if caseSensitive {
			searchType = searchSensitive
		} else {
			searchType = searchInsensitive
		}
	}
	return searchType
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
	pos = append(pos, re.FindAllIndex([]byte(s), -1)...)
	return pos
}
