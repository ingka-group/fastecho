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
			name:      "ok: negative number",
			givenStr:  "-100",
			expectInt: -100,
		},
		{
			name:      "error: string is not numeric",
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
			name:     "ok: empty string",
			givenStr: "",
			expect:   true,
		},
		{
			name:     "ok: empty string whitespace",
			givenStr: "  ",
			expect:   true,
		},
		{
			name:     "ok: empty string with \\",
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
			name: "ok: does not exist",
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
