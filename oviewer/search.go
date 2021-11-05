package oviewer

import (
	"context"
	"errors"
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

// forwardSearch is forward search.
func (root *Root) forwardSearch(ctx context.Context, lN int, input string) {
	if input == "" {
		root.searchWord = ""
		root.searchReg = nil
		return
	}
	root.input.value = input
	root.searchWord = root.input.value
	root.searchReg = regexpCompile(root.searchWord, root.CaseSensitive)
	root.setMessage(fmt.Sprintf("search:%v (%v)Cancel", root.searchWord, strings.Join(root.cancelKeys, ",")))

	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	root.cancelFunc = cancel

	eg.Go(func() error {
		return root.cancelWait()
	})

	searchWord := root.searchWord
	searchReg := root.searchReg
	eg.Go(func() error {
		lN, err := root.searchLine(ctx, searchWord, searchReg, lN)
		if err != nil {
			return err
		}
		root.moveLine(lN - root.Doc.firstLine())
		return nil
	})

	if err := eg.Wait(); err != nil {
		root.setMessage(err.Error())
		return
	}
	root.setMessage(fmt.Sprintf("search:%v", searchWord))
}

// backSearch is backward search.
func (root *Root) backSearch(ctx context.Context, lN int, input string) {
	if input == "" {
		root.searchWord = ""
		root.searchReg = nil
		return
	}
	root.input.value = input
	root.searchWord = root.input.value
	root.searchReg = regexpCompile(root.searchWord, root.CaseSensitive)
	root.setMessage(fmt.Sprintf("search:%v (%v)Cancel", root.searchWord, strings.Join(root.cancelKeys, ",")))

	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	root.cancelFunc = cancel

	eg.Go(func() error {
		return root.cancelWait()
	})

	searchWord := root.searchWord
	searchReg := root.searchReg
	eg.Go(func() error {
		lN, err := root.backSearchLine(ctx, searchWord, searchReg, lN)
		if err != nil {
			return err
		}
		root.moveLine(lN - root.Doc.firstLine())
		return nil
	})

	if err := eg.Wait(); err != nil {
		root.setMessage(err.Error())
		return
	}
	root.setMessage(fmt.Sprintf("search:%v", root.searchWord))
}

//incSearch implements incremental search.
func (root *Root) incSearch(ctx context.Context) {
	root.cancelRestart(ctx)

	root.searchWord = root.input.value
	root.searchReg = regexpCompile(root.searchWord, root.CaseSensitive)
	root.Doc.topLN = root.returnStartPosition()

	go func() {
		lN, err := root.searchLine(ctx, root.searchWord, root.searchReg, root.Doc.topLN+root.Doc.firstLine())
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Println(err)
			}
			return
		}
		root.MoveLine(lN - root.Doc.firstLine() + 1)
	}()
}

//incSearch implements incremental search.
func (root *Root) incBackSearch(ctx context.Context) {
	root.cancelRestart(ctx)

	root.searchWord = root.input.value
	root.searchReg = regexpCompile(root.searchWord, root.CaseSensitive)
	root.Doc.topLN = root.returnStartPosition()

	go func() {
		lN, err := root.backSearchLine(ctx, root.searchWord, root.searchReg, root.Doc.topLN+root.Doc.firstLine())
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Println(err)
			}
			return
		}
		root.MoveLine(lN - root.Doc.firstLine() + 1)
	}()
}

// cancelRestart calls the cancel function and sets the cancel function again.
func (root *Root) cancelRestart(ctx context.Context) {
	if root.cancelFunc != nil {
		root.cancelFunc()
	}
	ctx, cancel := context.WithCancel(ctx)
	root.cancelFunc = cancel
}

// returnStartPosition checks the input value and returns the start position.
func (root *Root) returnStartPosition() int {
	start := root.Doc.topLN
	if !strings.Contains(root.searchWord, root.OriginStr) {
		start = root.OriginPos
	}
	root.OriginStr = root.searchWord
	return start
}

// searchLine is searches below from the specified line.
func (root *Root) searchLine(ctx context.Context, searchWord string, searchReg *regexp.Regexp, num int) (int, error) {
	defer root.searchQuit()
	num = max(num, 0)

	searchType := getSearchType(searchWord, root.CaseSensitive, root.Config.RegexpSearch)

	for n := num; n < root.Doc.BufEndNum(); n++ {
		if root.contains(root.Doc.GetLine(n), searchWord, searchReg, searchType) {
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
func (root *Root) backSearchLine(ctx context.Context, searchWord string, searchReg *regexp.Regexp, num int) (int, error) {
	defer root.searchQuit()
	num = min(num, root.Doc.BufEndNum()-1)

	searchType := getSearchType(searchWord, root.CaseSensitive, root.Config.RegexpSearch)

	for n := num; n >= 0; n-- {
		if root.contains(root.Doc.GetLine(n), searchWord, searchReg, searchType) {
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

// regexpCompile is regexp.Compile the search string.
func regexpCompile(r string, caseSensitive bool) *regexp.Regexp {
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
func (root *Root) contains(s string, substr string, subReg *regexp.Regexp, t SearchType) bool {
	// Remove EscapeSequence.
	if strings.ContainsAny(s, "\x1b\b") {
		s = stripEscapeSequence.ReplaceAllString(s, "")
	}

	switch t {
	case searchSensitive:
		return strings.Contains(s, substr)
	case searchInsensitive:
		return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
	default: // Regular expressions.
		return subReg.MatchString(s)
	}
}

func getSearchType(t string, caseSensitive bool, regexpSearch bool) SearchType {
	if regexpSearch && t != regexp.QuoteMeta(t) {
		return searchRegexp
	}
	if caseSensitive {
		return searchSensitive
	}
	return searchInsensitive
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

// searchPositionReg returns an array of the beginning and end of the search string.
func searchPositionReg(s string, re *regexp.Regexp) [][]int {
	if re == nil || re.String() == "" {
		return nil
	}

	return re.FindAllIndex([]byte(s), -1)
}

// searchPosition returns an array of the beginning and end of the search string.
func searchPosition(caseSensitive bool, searchText string, substr string) [][]int {
	if substr == "" {
		return nil
	}

	var locs [][]int
	if !caseSensitive {
		searchText = strings.ToLower(searchText)
		substr = strings.ToLower(substr)
	}
	offSet := 0
	loc := strings.Index(searchText, substr)
	for loc != -1 {
		searchText = searchText[loc+len(substr):]
		locs = append(locs, []int{loc + offSet, loc + offSet + len(substr)})
		offSet += loc + len(substr)
		loc = strings.Index(searchText, substr)
	}
	return locs
}
