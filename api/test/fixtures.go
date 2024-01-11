package test

import (
	"fmt"
	"log"
	"os"
)

// Fixtures is a helper for reading fixtures.
type Fixtures struct{}

// ReadResponse reads the response from a file.
func (f Fixtures) ReadResponse(s string) string {
	return f.ReadFixture(s+".json", "responses")
}

// ReadRequestBody reads the request body from a file.
func (f Fixtures) ReadRequestBody(s string) string {
	return f.ReadFixture(s+".json", "requests")
}

// ReadFixture reads a fixture from a file.
func (f Fixtures) ReadFixture(filename, dir string) string {
	executionPath, err := testpath()
	if err != nil {
		log.Fatalf("could not file test path: %v", err)
	}

	path := fmt.Sprintf("%s/fixtures/%s/%s", executionPath, dir, filename)
	buf, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("could not read file '%s': %v", path, err)
	}

	return string(buf)
}
