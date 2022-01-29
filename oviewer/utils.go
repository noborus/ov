package oviewer

// max returns the larger value of the argument.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller value of the argument.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// removeStr removes the value of the specified string from slice.
func removeStr(list []string, s string) []string {
	if len(s) == 0 {
		return list
	}

	for n, l := range list {
		if l == s {
			list = append(list[:n], list[n+1:]...)
		}
	}
	return list
}

// removeStr removes the specified int value from slice.
func removeInt(list []int, c int) []int {
	for n, l := range list {
		if l == c {
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
