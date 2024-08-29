package daterange

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

// TestValidateDateRange tests the validation of a date range.
// Here we test only one implementation of date range - ISODateRange,
// since in both the implementation of DateRange and the validation is the same.
func TestValidateDateRange(t *testing.T) {
	type args struct {
		vt       *ValidTimeframes
		isoRange ISODateRange
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "day",
			args: args{
				vt: &ValidTimeframes{
					Day: true,
				},
				isoRange: ISODateRange{
					BasicDateRange: BasicDateRange{
						From: date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeDay,
				},
			},
		},
		{
			name: "week",
			args: args{
				vt: &ValidTimeframes{
					Week: true,
				},
				isoRange: ISODateRange{
					BasicDateRange: BasicDateRange{
						From: date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeWeek,
				},
			},
		},
		{
			name: "month",
			args: args{
				vt: &ValidTimeframes{
					Month: true,
				},
				isoRange: ISODateRange{
					BasicDateRange: BasicDateRange{
						From: date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeMonth,
				},
			},
		},
		{
			name: "year",
			args: args{
				vt: &ValidTimeframes{
					Year: true,
				},
				isoRange: ISODateRange{
					BasicDateRange: BasicDateRange{
						From: date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeYear,
				},
			},
		},
		{
			name: "to before from",
			args: args{
				vt: &ValidTimeframes{
					Day: true,
				},
				isoRange: ISODateRange{
					BasicDateRange: BasicDateRange{
						From: date.ISODate{Time: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeDay,
				},
			},
			wantErr: true,
		},
		{
			name: "from, to zero",
			args: args{
				vt: &ValidTimeframes{
					Day: true,
				},
				isoRange: ISODateRange{
					BasicDateRange: BasicDateRange{
						From: date.ISODate{Time: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)},
						To:   date.ISODate{Time: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)},
					},
					Timeframe: TimeframeDay,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vdt := validator.New()
			f := ValidateDateRange(tt.args.vt)
			vdt.RegisterStructValidation(f, ISODateRange{})

			err := vdt.Struct(tt.args.isoRange)
			if err != nil && !tt.wantErr {
				t.Errorf("ValidateDateRange() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
