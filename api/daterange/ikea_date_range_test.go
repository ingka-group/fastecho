package daterange

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

func TestIKEAGetWhereClause(t *testing.T) {
	type whereClause struct {
		clause string
		args   []interface{}
	}

	type testStruct struct {
		name      string
		from      string
		to        string
		timeframe IKEATimeframe
		want      whereClause
	}

	tests := []testStruct{
		{
			name:      "ok - day",
			from:      "2024-12-01",
			to:        "2024-12-10",
			timeframe: IKEATimeframeDay,
			want: whereClause{
				clause: "date BETWEEN ? AND ?",
				args:   []interface{}{"2024-12-01", "2024-12-10"},
			},
		},
		{
			name:      "ok - weekly on same year",
			from:      "2024-12-01",
			to:        "2024-12-10",
			timeframe: IKEATimeframeWeek,
			want: whereClause{
				clause: "(ikea_year = ?) AND (ikea_week BETWEEN ? AND ?)",
				args:   []interface{}{2024, 49, 50},
			},
		},
		{
			name:      "ok - weekly on subsequent years",
			from:      "2024-12-01",
			to:        "2025-12-10",
			timeframe: IKEATimeframeWeek,
			want: whereClause{
				clause: "(ikea_year > ? AND ikea_year < ?) OR (ikea_year = ? AND ikea_week >= ?) OR (ikea_year = ? AND ikea_week <= ?)",
				args:   []interface{}{2024, 2025, 2024, 49, 2025, 50},
			},
		},
		{
			name:      "ok - weekly on non subsequent from and to years",
			from:      "2024-12-01",
			to:        "2027-12-10",
			timeframe: IKEATimeframeWeek,
			want: whereClause{
				clause: "(ikea_year > ? AND ikea_year < ?) OR (ikea_year = ? AND ikea_week >= ?) OR (ikea_year = ? AND ikea_week <= ?)",
				args:   []interface{}{2024, 2027, 2024, 49, 2027, 49},
			},
		},
		{
			name:      "ok - monthly aggregation on same year",
			from:      "2024-07-01",
			to:        "2024-12-10",
			timeframe: IKEATimeframeMonth,
			want: whereClause{
				clause: "iso_year = ? AND iso_month BETWEEN ? AND ?",
				args:   []interface{}{2024, 7, 12},
			},
		},
		{
			name:      "ok - monthly aggregation on different from and to years",
			from:      "2024-07-01",
			to:        "2025-12-10",
			timeframe: IKEATimeframeMonth,
			want: whereClause{
				clause: "(iso_year > ? AND iso_year < ?) OR (iso_year = ? AND iso_month >= ?) OR (iso_year = ? AND iso_month <= ?) ",
				args:   []interface{}{2024, 2025, 2024, 7, 2025, 12},
			},
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

			gotClause, gotArgs := tt.timeframe.GetWhereClause(from, to)

			if gotClause != tt.want.clause {
				t.Errorf("got: %s, want: %s", gotClause, tt.want.clause)
			}

			assert.Equal(t, tt.want.args, gotArgs)

			for i := range gotArgs {
				if gotArgs[i] != tt.want.args[i] {
					t.Errorf("got: %v, want: %v", gotArgs, tt.want.args)
				}
			}
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
