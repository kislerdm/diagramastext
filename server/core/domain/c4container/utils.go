package c4container

import "strings"

func stringCleaner(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimLeft(s, `"`)
	s = strings.TrimRight(s, `"`)
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
