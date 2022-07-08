package util

import "strings"

func IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}
