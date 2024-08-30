package test

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"
)

// testpath returns a full path for the directory of a test file that called this function,
// so it can be used to build a path to binary files like fixtures next to the test files,
// which gives us an option to store fixtures in the same package with test.
func testpath() (string, error) {
	for i := 0; i < 32; i++ {
		_, caller, _, ok := runtime.Caller(i)
		if ok && strings.HasSuffix(caller, "_test.go") {
			return filepath.Dir(caller), nil
		}
	}

	return "", errors.New("cannot determine filesystem path for current test file")
}
