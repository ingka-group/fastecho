package stringutils

import (
	"strconv"
	"strings"
)

// IsEmpty checks whether the given string is empty
func IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// ToInt converts a string to an int
func ToInt(str string) (int, error) {
	num, err := strconv.Atoi(str)
	if err != nil {
		return -1, err
	}

	return num, nil
}

// ExistsInSlice returns whether a given string exists in the given slice
func ExistsInSlice(str string, sl []string) bool {
	for _, item := range sl {
		if item == str {
			return true
		}
	}
	return false
}
