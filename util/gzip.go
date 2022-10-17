package util

import (
	"bytes"
	"compress/gzip"
	"log"
	"os"
)

// ReadGzippedTestFile reads a GZiped test file
func ReadGzippedTestFile(path string) *bytes.Buffer {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	defer gr.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(gr); err != nil {
		log.Fatal(err)
	}

	return buf
}
