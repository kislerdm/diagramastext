package c4container

import "strings"

func stringCleaner(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
