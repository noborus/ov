package oviewer

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/sync/errgroup"
)

// search is forward search.
func (root *Root) search(ctx context.Context, input string) {
	if input == "" {
		root.input.reg = nil
		return
	}
	root.input.reg = regexpComple(input, root.CaseSensitive)
	root.Doc.lineNum--
	root.nextSearch(ctx)
}

// backSearch is backward search.
func (root *Root) backSearch(ctx context.Context, input string) {
	if input == "" {
		root.input.reg = nil
		return
	}
	root.input.reg = regexpComple(input, root.CaseSensitive)
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

	re := root.input.reg
	if re == nil || re.String() == "" {
		return num, nil
	}
	for n := num; n < root.Doc.BufEndNum(); n++ {
		if contains(root.Doc.GetLine(n), re) {
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

	re := root.input.reg
	if re == nil || re.String() == "" {
		return num, nil
	}
	for n := num; n >= 0; n-- {
		if contains(root.Doc.GetLine(n), re) {
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
	return nil
}

// stripEscapeSequence is a regular expression that excludes escape sequences.
var stripEscapeSequence = regexp.MustCompile("(\x1b\\[[\\d;*]*m)|.\b")

// contains returns a bool containing the search string.
func contains(s string, re *regexp.Regexp) bool {
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
	pos = append(pos, re.FindAllIndex([]byte(s), -1)...)
	return pos
}
