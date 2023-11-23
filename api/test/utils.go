package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

// minify minifies the given JSON input.
func minify(input string) string {
	src := []byte(input)
	var buff = new(bytes.Buffer)

	err := json.Compact(buff, src)
	if err != nil {
		log.Fatalf("failure encountered compacting json: %v", err)
	}

	ret, err := io.ReadAll(buff)
	if err != nil {
		log.Fatalf("read buffer error encountered: %v", err)
	}

	return string(ret)
}

// testpath returns a full path for the directory of a test file that called this function,
// so it can be used to build a path to binary files like fixtures next to the test files,
// which gives us an option to store fixtures in the same package with tests.
func testpath() (string, error) {
	for i := 0; i < 32; i++ {
		_, caller, _, ok := runtime.Caller(i)
		if ok && strings.HasSuffix(caller, "_test.go") {
			return filepath.Dir(caller), nil
		}
	}

	return "", errors.New("wake up mr freeman")
}
