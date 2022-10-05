package util

import "strings"

// IsEmpty checks whether the given string is empty
func IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}
