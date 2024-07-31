package daterange

import (
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

func TestGetStringQuery(t *testing.T) {
	type args struct {
		from      date.ISODate
		to        date.ISODate
		timeframe Timeframe
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
				timeframe: TimeframeDay,
			},
			want: "date BETWEEN '2021-01-01' AND '2021-02-01'",
		},
		{
			name: "ok - week - same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 4, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeWeek,
			},
			want: "year = 2021 AND week BETWEEN 1 AND 5",
		},
		{
			name: "ok - week - week 53 same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeWeek,
			},
			want: "(year = 2020 AND week >= 53) OR (year = 2021 AND week <= 5)",
		},
		{
			name: "ok - week - succeeding years",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 12, 4, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 1, 4, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeWeek,
			},
			want: "(year = 2020 AND week >= 48) OR (year = 2021 AND week <= 1)",
		},
		{
			name: "ok - week - gap year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeWeek,
			},
			want: "((year = 2021 AND week >= 53 ) OR year = 2022 OR (year = 2023 AND week <= 1 ))",
		},
		{
			name: "ok - month - same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeMonth,
			},
			want: "year = 2021 AND month BETWEEN 1 AND 2",
		},
		{
			name: "ok - month - succeeding years",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 11, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeMonth,
			},
			want: "((year = 2021 AND month >= 11 ) OR (year = 2022 AND month <= 1 ))",
		},
		{
			name: "ok - month - gap years",
			args: args{
				from:      date.ISODate{Time: time.Date(2020, 11, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeMonth,
			},
			want: "((year = 2020 AND month >= 11 ) OR year = 2021 OR (year = 2022 AND month <= 1 ))",
		},
		{
			name: "ok - year - same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeYear,
			},
			want: "year = 2022",
		},
		{
			name: "ok - year - year range",
			args: args{
				from:      date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeYear,
			},
			want: "year >= 2022 AND year <= 2024",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStringQuery(tt.args.from, tt.args.to, tt.args.timeframe); got != tt.want {
				t.Errorf("GetStringQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddToGormQuery(t *testing.T) {
	type args struct {
		from      date.ISODate
		to        date.ISODate
		timeframe Timeframe
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
				timeframe: TimeframeDay,
			},
			want: `date BETWEEN "2021-01-01" AND "2021-02-01"`,
		},
		{
			name: "ok - week - same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 4, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeWeek,
			},
			want: "year = 2021 AND week BETWEEN 1 AND 5",
		},
		{
			name: "ok - week - week 53 same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeWeek,
			},
			want: "(year = 2020 AND week >= 53) OR (year = 2021 AND week <= 5)",
		},
		{
			name: "ok - week - succeeding years",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 12, 4, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 1, 4, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeWeek,
			},
			want: "(year = 2020 AND week >= 48) OR (year = 2021 AND week <= 1)",
		},
		{
			name: "ok - week - gap year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeWeek,
			},
			want: "((year = 2021 AND week >= 53) OR year = 2022 OR (year = 2023 AND week <= 1))",
		},
		{
			name: "ok - month - same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeMonth,
			},
			want: "year = 2021 AND month BETWEEN 1 AND 2",
		},
		{
			name: "ok - month - succeeding years",
			args: args{
				from:      date.ISODate{Time: time.Date(2021, 11, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeMonth,
			},
			want: "((year = 2021 AND month >= 11) OR (year = 2022 AND month <= 1))",
		},
		{
			name: "ok - month - gap years",
			args: args{
				from:      date.ISODate{Time: time.Date(2020, 11, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeMonth,
			},
			want: "((year = 2020 AND month >= 11) OR year = 2021 OR (year = 2022 AND month <= 1))",
		},
		{
			name: "ok - year - same year",
			args: args{
				from:      date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeYear,
			},
			want: "year = 2022",
		},
		{
			name: "ok - year - year range",
			args: args{
				from:      date.ISODate{Time: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
				to:        date.ISODate{Time: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
				timeframe: TimeframeYear,
			},
			want: "year BETWEEN 2022 AND 2024",
		},
	}

	// init gorm once
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
				stmt := AddToGormQuery(db.Table("test").Select("*"), tt.args.from, tt.args.to, tt.args.timeframe)
				return stmt.Find(&[]struct{}{})
			})

			sql = strings.Split(sql, "WHERE ")[1]
			if sql != tt.want {
				t.Errorf("Generated SQL = %v, want %v", sql, tt.want)
			}
		})
	}
}

func TestValidateISODateRange(t *testing.T) {
	type args struct {
		vt       *ValidTimeframes
		isoRange ISODateRange
	}
	tests := []struct {
		name       string
		args       args
		shouldFail bool
	}{
		{
			name: "day",
			args: args{
				vt: &ValidTimeframes{
					Day: true,
				},
				isoRange: ISODateRange{
					ISODateRangeBasic: ISODateRangeBasic{
						From: date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeDay,
				},
			},
			shouldFail: false,
		},
		{
			name: "week",
			args: args{
				vt: &ValidTimeframes{
					Week: true,
				},
				isoRange: ISODateRange{
					ISODateRangeBasic: ISODateRangeBasic{
						From: date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeWeek,
				},
			},
			shouldFail: false,
		},
		{
			name: "month",
			args: args{
				vt: &ValidTimeframes{
					Month: true,
				},
				isoRange: ISODateRange{
					ISODateRangeBasic: ISODateRangeBasic{
						From: date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeMonth,
				},
			},
			shouldFail: false,
		},
		{
			name: "year",
			args: args{
				vt: &ValidTimeframes{
					Year: true,
				},
				isoRange: ISODateRange{
					ISODateRangeBasic: ISODateRangeBasic{
						From: date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeYear,
				},
			},
			shouldFail: false,
		},
		{
			name: "to before from",
			args: args{
				vt: &ValidTimeframes{
					Day: true,
				},
				isoRange: ISODateRange{
					ISODateRangeBasic: ISODateRangeBasic{
						From: date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeDay,
				},
			},
			shouldFail: true,
		},
		{
			name: "from, to zero",
			args: args{
				vt: &ValidTimeframes{
					Day: true,
				},
				isoRange: ISODateRange{
					ISODateRangeBasic: ISODateRangeBasic{
						From: date.ISODate{Time: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeDay,
				},
			},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vdt := validator.New()
			f := ValidateISODateRange(tt.args.vt)
			vdt.RegisterStructValidation(f, ISODateRange{})

			err := vdt.Struct(tt.args.isoRange)
			if err != nil && !tt.shouldFail {
				t.Errorf("ValidateISODateRange() = %v, want %v", err, tt.shouldFail)
			}
		})
	}
}
