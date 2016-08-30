package generators

import "strings"

func defaultIfBlank(s, defaultValue string) string {
	if len(strings.TrimSpace(s)) == 0 {
		return defaultValue
	}
	return s
}
