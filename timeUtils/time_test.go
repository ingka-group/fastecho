package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExtractDate(t *testing.T) {
	tests := []struct {
		name     string
		givenStr string
		expect   string
	}{
		{
			name:     "ok",
			givenStr: "2006-01-02T15:04:05.999999999Z07:00",
			expect:   "2006-01-02",
		},
		{
			name:     "no time part",
			givenStr: "2006-01-02",
			expect:   "2006-01-02",
		},
		{
			name:     "empty_str",
			givenStr: "",
			expect:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expect, ExtractDate(tt.givenStr))
		})
	}
}

func TestToDate(t *testing.T) {
	tests := []struct {
		name        string
		givenStr    string
		expect      time.Time
		expectError bool
	}{
		{
			name:     "ok",
			givenStr: "2006-01-02T15:04:05.999999999Z01:00",
			expect:   time.Date(2006, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "no time part",
			givenStr: "2006-01-02",
			expect:   time.Date(2006, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "not a date",
			givenStr:    "blablaT123",
			expectError: true,
		},
		{
			name:        "empty_str",
			givenStr:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			time, err := ToDate(tt.givenStr)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expect, time)
			}
		})
	}
}

func TestDateToDDWWMMYYYY(t *testing.T) {
	type DDMMYYYY struct {
		day   int
		week  int
		month int
		year  int
	}

	tests := []struct {
		name      string
		givenTime time.Time
		expect    DDMMYYYY
	}{
		{
			name:      "ok",
			givenTime: time.Date(2006, time.January, 2, 0, 0, 0, 0, time.UTC),
			expect: DDMMYYYY{
				day:   2,
				week:  1,
				month: 1,
				year:  2006,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, w, m, y := DateToDDWWMMYYYY(tt.givenTime)

			require.Equal(t, tt.expect.day, d)
			require.Equal(t, tt.expect.week, w)
			require.Equal(t, tt.expect.month, m)
			require.Equal(t, tt.expect.year, y)
		})
	}
}
