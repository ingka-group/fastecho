package daterange

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

func TestISODateRange_ImplementsInterface(t *testing.T) {
	dateRange := IKEADateRange{}
	var i interface{} = dateRange

	_, ok := i.(DateRange)
	if !ok {
		t.Errorf("ISODateRange does not implement DateRange interface")
	}
}

func TestISODateRange_GetWhereClauseSQL(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		to        string
		timeframe Timeframe
		want      string
	}{
		{
			name:      "ok: day",
			from:      "2024-12-01",
			to:        "2024-12-10",
			timeframe: TimeframeDay,
			want:      "(date BETWEEN '2024-12-01' AND '2024-12-10')",
		},
		{
			name:      "ok: weekly on same year",
			from:      "2024-12-01",
			to:        "2024-12-10",
			timeframe: TimeframeWeek,
			want:      "(year = 2024 AND week BETWEEN 48 AND 50)",
		},
		{
			name:      "ok: weekly on subsequent years",
			from:      "2024-12-01",
			to:        "2025-12-10",
			timeframe: TimeframeWeek,
			want:      "((year > 2024 AND year < 2025) OR (year = 2024 AND week >= 48) OR (year = 2025 AND week <= 50))",
		},
		{
			name:      "ok: weekly on non subsequent from and to years",
			from:      "2024-12-01",
			to:        "2027-12-10",
			timeframe: TimeframeWeek,
			want:      "((year > 2024 AND year < 2027) OR (year = 2024 AND week >= 48) OR (year = 2027 AND week <= 49))",
		},
		{
			name:      "ok: monthly aggregation on same year",
			from:      "2024-07-01",
			to:        "2024-12-10",
			timeframe: TimeframeMonth,
			want:      "(year = 2024 AND month BETWEEN 7 AND 12)",
		},
		{
			name:      "ok: monthly aggregation on different from and to years",
			from:      "2024-07-01",
			to:        "2025-12-10",
			timeframe: TimeframeMonth,
			want:      "((year > 2024 AND year < 2025) OR (year = 2024 AND month >= 7) OR (year = 2025 AND month <= 12))",
		},
		{
			name:      "ok: year aggregation",
			from:      "2024-07-01",
			to:        "2025-12-10",
			timeframe: TimeframeYear,
			want:      "(year BETWEEN 2024 AND 2025)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, err := date.FromStr(tt.from)
			if err != nil {
				t.Fatal(err)
			}

			to, err := date.FromStr(tt.to)
			if err != nil {
				t.Fatal(err)
			}

			d := ISODateRange{
				BasicDateRange: BasicDateRange{
					From: *from,
					To:   *to,
				},
				Timeframe: tt.timeframe,
			}
			assert.Equal(t, tt.want, d.GetWhereClauseSQL())
		})
	}
}
