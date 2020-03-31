package oviewer

import (
	"errors"
	"strings"
)

func (root *root) search(num int) (int, error) {
	for n := num; n < root.Model.endNum; n++ {
		if strings.Contains(root.Model.buffer[n], root.input) {
			return n, nil
		}
	}
	return 0, errors.New("not found")
}

func (root *root) backSearch(num int) (int, error) {
	for n := num; n >= 0; n-- {
		if strings.Contains(root.Model.buffer[n], root.input) {
			return n, nil
		}
	}
	return 0, errors.New("not found")
}
