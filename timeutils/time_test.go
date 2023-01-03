package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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

func TestToISODateStr(t *testing.T) {
	tests := []struct {
		name   string
		given  time.Time
		expect string
	}{
		{
			name:   "ok",
			given:  time.Date(2006, time.January, 2, 0, 0, 0, 0, time.UTC),
			expect: "2006-01-02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expect, ToISODateStr(tt.given))
		})
	}
}
