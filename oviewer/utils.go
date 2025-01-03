package oviewer

import (
	"regexp"
	"strings"
)

// remove removes the value of the specified string from slice.
func remove[T comparable](list []T, s T) []T {
	for n, l := range list {
		if l == s {
			list = append(list[:n], list[n+1:]...)
		}
	}
	return list
}

// contains returns true if the specified value is included.
func contains[T comparable](list []T, e T) bool {
	for _, n := range list {
		if e == n {
			return true
		}
	}
	return false
}

// toAddTop adds the string if it is not in list.
func toAddTop(list []string, s string) []string {
	if len(s) == 0 {
		return list
	}
	if !contains(list, s) {
		list = append([]string{s}, list...)
	}
	return list
}

// toAddLast adds a string to the end if it is not in list.
func toAddLast(list []string, s string) []string {
	if len(s) == 0 {
		return list
	}
	if !contains(list, s) {
		list = append(list, s)
	}
	return list
}

// toLast moves the specified string to the end.
func toLast(list []string, s string) []string {
	if len(s) == 0 {
		return list
	}

	list = remove(list, s)
	list = append(list, s)
	return list
}

// allIndex is a wrapper that returns either a regular expression index or a string index.
func allIndex(s string, substr string, reg *regexp.Regexp) [][]int {
	if reg != nil {
		return reg.FindAllStringIndex(s, -1)
	}
	return allStringIndex(s, substr)
}

// allStringIndex returns all matching string positions.
func allStringIndex(s string, substr string) [][]int {
	if len(substr) == 0 {
		return nil
	}
	var result [][]int
	width := len(substr)
	for pos, offSet := strings.Index(s, substr), 0; pos != -1; {
		s = s[pos+width:]
		result = append(result, []int{pos + offSet, pos + offSet + width})
		offSet += pos + width

		if len(s) > 0 && s[0] == '"' {
			qpos := strings.Index(s[1:], `"`)
			s = s[qpos+2:]
			offSet += qpos + 2
		}

		pos = strings.Index(s, substr)
	}
	return result
}
