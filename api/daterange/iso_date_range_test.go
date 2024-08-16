package daterange

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

func TestGetStringQuery(t *testing.T) {
	type args struct {
		from      date.ISODate
		to        date.ISODate
		timeframe ISOTimeframe
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ok - day",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: ISOTimeframeDay,
			},
			want: "date BETWEEN '2021-01-01' AND '2021-02-01'",
		},
		{
			name: "ok - week - same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 4, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: ISOTimeframeWeek,
			},
			want: "year = 2021 AND week BETWEEN 1 AND 5",
		},
		{
			name: "ok - week - week 53 same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: ISOTimeframeWeek,
			},
			want: "(year > 2020 AND year < 2021) OR (year = 2020 AND week >= 53) OR (year = 2021 AND week <= 5)",
		},
		{
			name: "ok - week - succeeding years",
			args: args{
				from:      date.ISODate{Time: time.Date(2020, 12, 4, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 1, 4, 0, 0, 0, 0, time.UTC)},
				timeframe: ISOTimeframeWeek,
			},
			want: "(year > 2020 AND year < 2021) OR (year = 2020 AND week >= 49) OR (year = 2021 AND week <= 1)",
		},
		{
			name: "ok - week - gap year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC)},
				timeframe: ISOTimeframeWeek,
			},
			want: "(year > 2020 AND year < 2023) OR (year = 2020 AND week >= 53) OR (year = 2023 AND week <= 1)",
		},
		{
			name: "ok - month - same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: ISOTimeframeMonth,
			},
			want: "year = 2021 AND month BETWEEN 1 AND 2",
		},
		{
			name: "ok - month - succeeding years",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 11, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: ISOTimeframeMonth,
			},
			want: "(year > 2021 AND year < 2022) OR (year = 2021 AND month >= 11) OR (year = 2022 AND month <= 1)",
		},
		{
			name: "ok - month - gap years",
			args: args{
				from:      date.ISODate{Time: time.Date(2020, 11, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: ISOTimeframeMonth,
			},
			want: "(year > 2020 AND year < 2022) OR (year = 2020 AND month >= 11) OR (year = 2022 AND month <= 1)",
		},
		{
			name: "ok - year - same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: ISOTimeframeYear,
			},
			want: "year BETWEEN 2022 AND 2022",
		},
		{
			name: "ok - year - year range",
			args: args{
				from:      date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: ISOTimeframeYear,
			},
			want: "year BETWEEN 2022 AND 2024",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := tt.args.timeframe.GetWhereClause(tt.args.from, tt.args.to)
			assert.Equal(t, tt.want, stmt)
		})
	}
}

func TestISODateRangeImplementsDateRangeInterface(t *testing.T) {
	isoRange := ISODateRange{}
	var i interface{} = isoRange

	_, ok := i.(DateRange)
	if !ok {
		t.Errorf("ISODateRange does not implement DateRange interface")
	}
}
