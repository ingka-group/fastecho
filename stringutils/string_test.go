package stringutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToInt(t *testing.T) {
	var tests = []struct {
		name      string
		givenStr  string
		expectInt int
		expectErr bool
	}{
		{
			name:      "ok",
			givenStr:  "100",
			expectInt: 100,
		},
		{
			name:      "ok_negative_number",
			givenStr:  "-100",
			expectInt: -100,
		},
		{
			name:      "error_string_not_numeric",
			givenStr:  "100STR",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num, err := ToInt(tt.givenStr)
			if err != nil {
				assert.True(t, tt.expectErr)
			} else {
				assert.False(t, tt.expectErr)
				assert.Equal(t, tt.expectInt, num)
			}
		})
	}
}

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

func TestExistsInSlice(t *testing.T) {
	type args struct {
		givenStr string
		givenSl  []string
	}

	tests := []struct {
		name   string
		args   args
		expect bool
	}{
		{
			name: "ok",
			args: args{
				givenStr: "100-ABC",
				givenSl:  []string{"100-ABC", "200-DEF"},
			},
			expect: true,
		},
		{
			name: "does_not_exist",
			args: args{
				givenStr: "100-ABC",
				givenSl:  []string{"200-DEF"},
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, ExistsInSlice(tt.args.givenStr, tt.args.givenSl))
		})
	}
}
