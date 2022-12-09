package oviewer

import (
	"golang.org/x/exp/constraints"
)

// max returns the larger value of the argument.
func max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller value of the argument.
func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// remove removes the value of the specified string from slice.
func remove[T comparable](list []T, s T) []T {
	for n, l := range list {
		if l == s {
			list = append(list[:n], list[n+1:]...)
		}
	}
	return list
}

// containsInt returns true if the specified int is included.
func containsInt(list []int, e int) bool {
	for _, n := range list {
		if e == n {
			return true
		}
	}
	return false
}

// toLast toLast moves the specified string to the end.
func toLast(list []string, s string) []string {
	if len(s) == 0 {
		return list
	}

	list = remove(list, s)
	list = append(list, s)
	return list
}
