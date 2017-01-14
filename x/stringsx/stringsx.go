package stringsx

import (
	"strings"
	"unicode"
)

// DefaultIfBlank uses the provided default value if s is blank.
func DefaultIfBlank(s, defaultValue string) string {
	if len(strings.TrimSpace(s)) == 0 {
		return defaultValue
	}
	return s
}

// Contains returns true iff s matches one of the strings in v
func Contains(s string, v ...string) bool {
	for _, x := range v {
		if s == x {
			return true
		}
	}
	return false
}

// ToPrivate lowercases the first letter.
func ToPrivate(s string) string {
	if s == "" {
		return ""
	}

	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func ToPublic(s string) string {
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
