package test

import (
	"fmt"
	"log"
	"os"
)

type Fixtures struct{}

func (f Fixtures) ReadResponse(s string) string {
	exPath, err := testpath()
	if err != nil {
		log.Fatalf("could not file test path: %v", err)
	}

	path := fmt.Sprintf("%s/fixtures/responses/%s.json", exPath, s)
	buf, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("could not read file '%s': %v", path, err)
	}

	return string(buf)
}
