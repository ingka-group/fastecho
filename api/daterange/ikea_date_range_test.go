package daterange

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

func TestIKEAGetWhereClause(t *testing.T) {
	type testStruct struct {
		name      string
		from      string
		to        string
		timeframe IKEATimeframe
		want      string
	}

	tests := []testStruct{
		{
			name:      "ok - day",
			from:      "2024-12-01",
			to:        "2024-12-10",
			timeframe: IKEATimeframeDay,
			want:      "date BETWEEN '2024-12-01' AND '2024-12-10'",
		},
		{
			name:      "ok - weekly on same year",
			from:      "2024-12-01",
			to:        "2024-12-10",
			timeframe: IKEATimeframeWeek,
			want:      "ikea_year = 2024 AND ikea_week BETWEEN 49 AND 50",
		},
		{
			name:      "ok - weekly on subsequent years",
			from:      "2024-12-01",
			to:        "2025-12-10",
			timeframe: IKEATimeframeWeek,
			want:      "(ikea_year > 2024 AND ikea_year < 2025) OR (ikea_year = 2024 AND ikea_week >= 49) OR (ikea_year = 2025 AND ikea_week <= 50)",
		},
		{
			name:      "ok - weekly on non subsequent from and to years",
			from:      "2024-12-01",
			to:        "2027-12-10",
			timeframe: IKEATimeframeWeek,
			want:      "(ikea_year > 2024 AND ikea_year < 2027) OR (ikea_year = 2024 AND ikea_week >= 49) OR (ikea_year = 2027 AND ikea_week <= 49)",
		},
		{
			name:      "ok - monthly aggregation on same year",
			from:      "2024-07-01",
			to:        "2024-12-10",
			timeframe: IKEATimeframeMonth,
			want:      "iso_year = 2024 AND iso_month BETWEEN 7 AND 12",
		},
		{
			name:      "ok - monthly aggregation on different from and to years",
			from:      "2024-07-01",
			to:        "2025-12-10",
			timeframe: IKEATimeframeMonth,
			want:      "(iso_year > 2024 AND iso_year < 2025) OR (iso_year = 2024 AND iso_month >= 7) OR (iso_year = 2025 AND iso_month <= 12)",
		},
		{
			name:      "ok - year aggregation",
			from:      "2024-07-01",
			to:        "2025-12-10",
			timeframe: IKEATimeframeYear,
			want:      "financial_year BETWEEN 2024 AND 2025",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, err := getIsoDate(tt.from)
			if err != nil {
				t.Errorf("error parsing from date in test: %v", err)
			}

			to, err := getIsoDate(tt.to)
			if err != nil {
				t.Errorf("error parsing to date in test: %v", err)
			}

			stmt := tt.timeframe.GetWhereClause(from, to)

			assert.Equal(t, tt.want, stmt)

		})
	}
}

func getIsoDate(str string) (date.ISODate, error) {
	dateVal, err := time.Parse("2006-01-02", str)
	res := date.FromTime(dateVal)
	return res, err
}

func TestIKEADateRangeImplementsDateRangeInterface(t *testing.T) {
	isoRange := IKEADateRange{}
	var i interface{} = isoRange

	_, ok := i.(DateRange)
	if !ok {
		t.Errorf("ISODateRange does not implement DateRange interface")
	}
}
