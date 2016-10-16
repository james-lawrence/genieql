package stringsx

import "strings"

// DefaultIfBlank uses the provided default value if s is blank.
func DefaultIfBlank(s, defaultValue string) string {
	if len(strings.TrimSpace(s)) == 0 {
		return defaultValue
	}
	return s
}
