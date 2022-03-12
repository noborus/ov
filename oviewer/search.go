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

// Searcher interface provides a match method that determines
// if the search word matches the argument string.
type Searcher interface {
	Match(string) bool
}

// searchWord is a case-insensitive search.
type searchWord struct {
	word string
}

// searchWord Match is a case-insensitive search.
func (substr searchWord) Match(s string) bool {
	s = stripEscapeSequence(s)
	return strings.Contains(strings.ToLower(s), substr.word)
}

// sensitiveWord is a case-sensitive search.
type sensitiveWord struct {
	word string
}

// sensitiveWord Match is a case-sensitive search.
func (substr sensitiveWord) Match(s string) bool {
	s = stripEscapeSequence(s)
	return strings.Contains(s, substr.word)
}

// regexpWord is a regular expression search.
type regexpWord struct {
	word *regexp.Regexp
}

// regexpWord Match is a regular expression search.
func (substr regexpWord) Match(s string) bool {
	s = stripEscapeSequence(s)
	return substr.word.MatchString(s)
}

// stripRegexpES is a regular expression that excludes escape sequences.
var stripRegexpES = regexp.MustCompile("(\x1b\\[[\\d;*]*m)|.\b")

// stripEscapeSequence strips if it contains escape sequences.
func stripEscapeSequence(s string) string {
	// Remove EscapeSequence.
	if strings.ContainsAny(s, "\x1b\b") {
		s = stripRegexpES.ReplaceAllString(s, "")
	}
	return s
}

// NewSearcher returns the Searcher interface suitable for the search term.
func NewSearcher(word string, searchReg *regexp.Regexp, caseSensitive bool, regexpSearch bool) Searcher {
	if regexpSearch && word != regexp.QuoteMeta(word) {
		return regexpWord{
			word: searchReg,
		}
	}
	if caseSensitive {
		return sensitiveWord{
			word: word,
		}
	}
	return searchWord{
		word: strings.ToLower(word),
	}
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
	if err != nil {
		log.Printf("regexpCompile failed %s", r)
		return nil
	}
	return re
}

// rangePosition returns the range starting and ending from the s,substr string.
func rangePosition(s, substr string, number int) (int, int) {
	left := strings.Index(s, substr)
	// If left == 0
	// | a | b | c |
	// If left > 0
	// a | b | c
	if left > 0 {
		if number == 0 {
			return 0, left
		}
		number--
	}

	i := 0
	for n := 0; n < number; n++ {
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

// searchPosition returns the position where the search in the argument line matched.
// searchPosition uses cache.
func (root *Root) searchPosition(lN int, lineStr string) [][]int {
	m := root.Doc
	key := root.searchKey(lN)

	if value, found := m.cache.Get(key); found {
		poss, ok := value.([][]int)
		if !ok {
			return nil
		}
		return poss
	}

	var poss [][]int
	if root.Config.RegexpSearch {
		poss = searchPositionReg(lineStr, root.searchReg)
	} else {
		poss = searchPositionStr(root.Config.CaseSensitive, lineStr, root.searchWord)
	}

	m.cache.Set(key, poss, 3)
	return poss
}

// searcKey returns a search key for the cache.
func (root *Root) searchKey(lN int) string {
	s := "s"
	if root.Config.RegexpSearch {
		s += "r"
	}
	if root.Config.CaseSensitive {
		s += "i"
	}
	return fmt.Sprintf("search:%d:%s:%s", lN, s, root.searchWord)
}

// searchPositionReg returns an array of the beginning and end of the search string.
func searchPositionReg(s string, re *regexp.Regexp) [][]int {
	if re == nil || re.String() == "" {
		return nil
	}
	return re.FindAllIndex([]byte(s), -1)
}

// searchPosition returns an array of the beginning and end of the search string.
func searchPositionStr(caseSensitive bool, s string, substr string) [][]int {
	if substr == "" {
		return nil
	}

	var locs [][]int
	if !caseSensitive {
		s = strings.ToLower(s)
		substr = strings.ToLower(substr)
	}
	offSet := 0
	loc := strings.Index(s, substr)
	for loc != -1 {
		s = s[loc+len(substr):]
		locs = append(locs, []int{loc + offSet, loc + offSet + len(substr)})
		offSet += loc + len(substr)
		loc = strings.Index(s, substr)
	}
	return locs
}

// setSearcher is a wrapper for NewSearcher and returns a Searcher interface.
// Returns nil if there is no search term.
func (root *Root) setSearcher(word string, caseSensitive bool) Searcher {
	if word == "" {
		root.searchWord = ""
		root.searchReg = nil
		return nil
	}
	root.input.value = word
	root.searchWord = word
	root.searchReg = regexpCompile(root.searchWord, caseSensitive)

	return NewSearcher(root.searchWord, root.searchReg, caseSensitive, root.Config.RegexpSearch)
}

// searchMove searches forward/backward and moves to the nearest matching line.
func (root *Root) searchMove(ctx context.Context, forward bool, lN int, searcher Searcher) {
	if searcher == nil {
		return
	}
	root.setMessagef("search:%v (%v)Cancel", root.searchWord, strings.Join(root.cancelKeys, ","))
	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	root.cancelFunc = cancel

	eg.Go(func() error {
		return root.cancelWait()
	})

	eg.Go(func() error {
		var err error
		if forward {
			lN, err = root.Doc.SearchLine(ctx, searcher, lN)
		} else {
			lN, err = root.Doc.BackSearchLine(ctx, searcher, lN)
		}
		root.searchQuit()
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
	root.setMessagef("search:%v", root.searchWord)
}

// incSearch implements incremental forward/back search.
func (root *Root) incSearch(ctx context.Context, forward bool) {
	root.Doc.topLN = root.returnStartPosition()

	searcher := root.setSearcher(root.input.value, root.CaseSensitive)
	if searcher == nil {
		return
	}

	ctx = root.cancelRestart(ctx)

	go func() {
		lN := root.Doc.topLN + root.Doc.firstLine()
		var err error
		if forward {
			lN, err = root.Doc.SearchLine(ctx, searcher, lN)
		} else {
			lN, err = root.Doc.BackSearchLine(ctx, searcher, lN)
		}
		root.searchQuit()
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
func (root *Root) cancelRestart(ctx context.Context) context.Context {
	if root.cancelFunc != nil {
		root.cancelFunc()
	}
	ctx, cancel := context.WithCancel(ctx)
	root.cancelFunc = cancel
	return ctx
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
