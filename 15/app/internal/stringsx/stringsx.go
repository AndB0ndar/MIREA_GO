package stringsx

import (
	"strings"
	"unicode/utf8"
)

func Clip(s string, maxLength int) string {
	if maxLength < 0 {
		maxLength = 0
	}

	if utf8.RuneCountInString(s) <= maxLength {
		return s
	}

	runes := []rune(s)
	if len(runes) > maxLength {
		return string(runes[:maxLength])
	}
	return s
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func ContainsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
