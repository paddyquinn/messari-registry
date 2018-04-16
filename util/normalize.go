package util

import "strings"

// Normalize trims space and lowercases the passed string.
func Normalize(str string) *string {
	normalizedStr := strings.ToLower(strings.TrimSpace(str))
	return &normalizedStr
}
