package oviewer

import (
	"bytes"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"slices"
	"strings"
)

// remove removes all occurrences of the specified value from slice.
func remove[T comparable](list []T, s T) []T {
	return slices.DeleteFunc(list, func(v T) bool {
		return v == s
	})
}

// abs returns the absolute value of an integer.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// toAddTop adds the string if it is not in list.
func toAddTop(list []string, s string) []string {
	if len(s) == 0 {
		return list
	}
	if !slices.Contains(list, s) {
		return slices.Insert(list, 0, s)
	}
	return list
}

// toAddLast adds a string to the end if it is not in list.
func toAddLast(list []string, s string) []string {
	if len(s) == 0 {
		return list
	}
	if !slices.Contains(list, s) {
		list = append(list, s)
	}
	return list
}

// toLast moves the specified string to the end.
func toLast(list []string, s string) []string {
	if len(s) == 0 {
		return list
	}

	list = slices.DeleteFunc(list, func(v string) bool {
		return v == s
	})
	return append(list, s)
}

// allIndex is a wrapper that returns either a regular expression index or a string index.
func allIndex(s string, substr string, reg *regexp.Regexp) [][]int {
	if reg != nil {
		return reg.FindAllStringIndex(s, -1)
	}
	return allStringIndex(s, substr)
}

// allStringIndex returns all matching string positions.
// Delimiters inside a double-quoted field are not treated as delimiters.
func allStringIndex(s string, substr string) [][]int {
	if len(substr) == 0 {
		return nil
	}
	var result [][]int
	width := len(substr)
	offSet := 0
	// The first field may also be quoted, so skip it before the first search.
	if len(s) > 0 && s[0] == '"' {
		s, offSet = skipQuoted(s, offSet)
	}
	for pos := strings.Index(s, substr); pos != -1; pos = strings.Index(s, substr) {
		s = s[pos+width:]
		result = append(result, []int{pos + offSet, pos + offSet + width})
		offSet += pos + width

		if len(s) > 0 && s[0] == '"' {
			s, offSet = skipQuoted(s, offSet)
		}
	}
	return result
}

// skipQuoted skips the double-quoted field at the beginning of s.
// It returns the remainder of s and the updated offset.
// If the quote is not closed, only the opening quote is skipped.
func skipQuoted(s string, offSet int) (string, int) {
	qpos := strings.Index(s[1:], `"`)
	return s[qpos+2:], offSet + qpos + 2
}

// writeLine writes a line to w.
// It adds a newline to the end of the line.
// It logs write errors.
func writeLine(w io.Writer, line []byte) {
	if _, err := w.Write(line); err != nil {
		log.Printf("%s:%s", line, err)
		return
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		log.Printf("%s:%s", line, err)
	}
}

// stripEscapeSequenceString removes escape sequences and backspaces from a string.
// It is used to identify and remove these sequences from strings or byte slices.
var stripRegexpES = regexp.MustCompile("(\x1b\\[[\\d;*]*m)|.\\x08")

func stripEscapeSequenceString(src string) string {
	if !strings.ContainsAny(src, "\x1b\\x08") {
		return src
	}
	// Remove EscapeSequence.
	return stripRegexpES.ReplaceAllString(src, "")
}

// stripEscapeSequence strips if it contains escape sequences.
func stripEscapeSequenceBytes(src []byte) []byte {
	if !bytes.ContainsAny(src, "\x1b\x08") {
		return src
	}
	// Remove EscapeSequence.
	return stripRegexpES.ReplaceAll(src, []byte(""))
}

// getShell returns the shell name based on the operating system.
func getShell() string {
	if runtime.GOOS == "windows" {
		return "CMD.EXE"
	}
	return "/bin/sh"
}

// getTTY returns the TTY file based on the operating system.
func getTTY() (*os.File, error) {
	if runtime.GOOS == "windows" {
		return os.Open("CONIN$")
	}
	return os.Open("/dev/tty")
}

// firstRune returns the first rune of the string.
func firstRune(s string) rune {
	if len(s) == 0 {
		return rune(0)
	}
	return []rune(s)[0]
}
