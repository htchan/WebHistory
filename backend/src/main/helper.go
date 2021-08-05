package main

import (
	"strings"
)

func isSubSet(s1 string, s2 string) bool {
	if len(s1) < len(s2) { return false }
	for _, char := range strings.Split(s2) {
		if !strings.Contains(s1, char) {
			return false
		}
	}
	return true
}
