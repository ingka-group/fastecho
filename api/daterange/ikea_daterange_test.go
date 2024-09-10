package daterange

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

func TestIKEADateRange_ImplementsInterface(t *testing.T) {
	dateRange := IKEADateRange{}
	var i interface{} = dateRange

	_, ok := i.(DateRange)
	if !ok {
		t.Errorf("IKEADateRange does not implement DateRange interface")
	}
}

func TestIKEADateRange_GetWhereClauseSQL(t *testing.T) {
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
			want:      "(ikea_year = 2024 AND ikea_week BETWEEN 49 AND 50)",
		},
		{
			name:      "ok: weekly on subsequent years",
			from:      "2024-12-01",
			to:        "2025-12-10",
			timeframe: TimeframeWeek,
			want:      "((ikea_year > 2024 AND ikea_year < 2025) OR (ikea_year = 2024 AND ikea_week >= 49) OR (ikea_year = 2025 AND ikea_week <= 50))",
		},
		{
			name:      "ok: weekly on non subsequent from and to years",
			from:      "2024-12-01",
			to:        "2027-12-10",
			timeframe: TimeframeWeek,
			want:      "((ikea_year > 2024 AND ikea_year < 2027) OR (ikea_year = 2024 AND ikea_week >= 49) OR (ikea_year = 2027 AND ikea_week <= 49))",
		},
		{
			name:      "ok: monthly aggregation on same year",
			from:      "2024-07-01",
			to:        "2024-12-10",
			timeframe: TimeframeMonth,
			want:      "(iso_year = 2024 AND iso_month BETWEEN 7 AND 12)",
		},
		{
			name:      "ok: monthly aggregation on different from and to years",
			from:      "2024-07-01",
			to:        "2025-12-10",
			timeframe: TimeframeMonth,
			want:      "((iso_year > 2024 AND iso_year < 2025) OR (iso_year = 2024 AND iso_month >= 7) OR (iso_year = 2025 AND iso_month <= 12))",
		},
		{
			name:      "ok: year aggregation",
			from:      "2024-07-01",
			to:        "2025-12-10",
			timeframe: TimeframeYear,
			want:      "(financial_year BETWEEN 2024 AND 2026)",
		},
		{
			name:      "ok: year aggregation in same iso year but different financial year",
			from:      "2024-07-01",
			to:        "2024-10-10",
			timeframe: TimeframeYear,
			want:      "(financial_year BETWEEN 2024 AND 2025)",
		},
		{
			name:      "ok: year aggregation in same iso year and same financial year",
			from:      "2024-07-01",
			to:        "2024-08-30",
			timeframe: TimeframeYear,
			want:      "(financial_year BETWEEN 2024 AND 2024)",
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

			d := IKEADateRange{
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

func TestIKEADateRange_GetTimeColumns(t *testing.T) {
	tests := []struct {
		name      string
		timeframe Timeframe
		alias     bool
		want      []string
	}{
		{
			name:      "ok: day",
			timeframe: TimeframeDay,
			want:      []string{"date"},
		},
		{
			name:      "ok: week",
			timeframe: TimeframeWeek,
			want:      []string{"ikea_year", "ikea_week"},
		},
		{
			name:      "ok: month",
			timeframe: TimeframeMonth,
			want:      []string{"iso_year", "iso_month"},
		},
		{
			name:      "ok: year",
			timeframe: TimeframeYear,
			want:      []string{"financial_year"},
		},
		{
			name:      "ok: week (alias)",
			timeframe: TimeframeWeek,
			alias:     true,
			want:      []string{"ikea_year AS year", "ikea_week AS week"},
		},
		{
			name:      "ok: month (alias)",
			timeframe: TimeframeMonth,
			alias:     true,
			want:      []string{"iso_year AS year", "iso_month AS month"},
		},
		{
			name:      "ok: year (alias)",
			timeframe: TimeframeYear,
			alias:     true,
			want:      []string{"financial_year AS year"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := IKEADateRange{
				Timeframe: tt.timeframe,
			}
			assert.Equalf(t, tt.want, d.GetTimeColumns(tt.alias), "GetTimeColumns(%v)", tt.alias)
		})
	}
}
