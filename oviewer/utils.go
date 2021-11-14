package oviewer

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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

func removeInt(list []int, c int) []int {
	for n, l := range list {
		if l == c {
			list = append(list[:n], list[n+1:]...)
		}
	}
	return list
}

func containsInt(list []int, e int) bool {
	for _, n := range list {
		if e == n {
			return true
		}
	}
	return false
}
