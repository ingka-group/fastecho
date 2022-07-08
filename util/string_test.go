package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		givenStr string
		expect   bool
	}{
		{
			name:     "ok",
			givenStr: "non-empty-str",
			expect:   false,
		},
		{
			name:     "empty_str",
			givenStr: "",
			expect:   true,
		},
		{
			name:     "empty_str_whitespace",
			givenStr: "  ",
			expect:   true,
		},
		{
			name:     "empty_str_\\char",
			givenStr: "\r\t\n",
			expect:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, IsEmpty(tt.givenStr))
		})
	}
}
