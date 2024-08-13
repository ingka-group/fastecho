package daterange

import (
	"github.com/go-playground/validator/v10"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

// DateRange represents a date range with a specific timeframe.
type DateRange interface {
	GetDateRangeBasic() *DateRangeBasic
	GetTimeframe() Timeframe
} // @name DateRange

// DateRangeBasic represents the base parameters for a date range.
// It is called basic because it does not contain the Timeframe for data time-aggregations.
// TODO: Rename this to DateRangeBasic.
// Do this after informing the current consumers
type DateRangeBasic struct {
	From date.ISODate `query:"from" validate:"required" example:"2024-12-01"`
	To   date.ISODate `query:"to" validate:"required" example:"2024-12-10"`
} // @name DateRangeBasic

// ValidateDateRangeBasic validates the DateRangeBasic struct.
//
// Example:
//
// validator.RegisterStructValidation(ValidateDateRangeBasic, DateRangeBasic{})
func ValidateDateRangeBasic() validator.StructLevelFunc {
	return func(sl validator.StructLevel) {
		dateRangeBasic := sl.Current().Interface().(DateRangeBasic)
		validateDateRangeBasic(sl, &dateRangeBasic)
	}
}

func validateDateRangeBasic(sl validator.StructLevel, dateRange *DateRangeBasic) {
	if dateRange.From.IsZero() {
		sl.ReportError(
			dateRange.From, "From", "from", "from-zero", "",
		)
	}

	if dateRange.To.IsZero() {
		sl.ReportError(
			dateRange.From, "To", "to", "to-zero", "",
		)
	}

	if dateRange.From.Date().After(dateRange.To.Date()) {
		sl.ReportError(
			dateRange.From, "From", "from", "from-after-to", "",
		)
	}
}

// ValidateDateRange validates a DateRange implementation. To identify which timeframes are supported, a ValidTimeframes struct
// is expected to be passed.
//
// Example:
//
// vdt.RegisterStructValidation(ValidateDateRange(
//
//	&ValidTimeframes{
//		Day:   false,
//		Week:  true,
//		Month: true,
//		Year:  true,
//	},
//
// ), ISODateRange{})
func ValidateDateRange(vt *ValidTimeframes) validator.StructLevelFunc {
	return func(sl validator.StructLevel) {
		validateDateRange(sl, vt)
	}
}

func validateDateRange(sl validator.StructLevel, vt *ValidTimeframes) {
	dateRange := sl.Current().Interface().(DateRange)
	timeframe := dateRange.GetTimeframe()

	validateDateRangeBasic(sl, dateRange.GetDateRangeBasic())

	if timeframe == ISOTimeframeDay && !vt.Day {
		sl.ReportError(
			timeframe, "Timeframe", "timeframe", "invalid-val-day", "",
		)
	} else if timeframe == ISOTimeframeWeek && !vt.Week {
		sl.ReportError(
			timeframe, "Timeframe", "timeframe", "invalid-val-week", "",
		)
	} else if timeframe == ISOTimeframeMonth && !vt.Month {
		sl.ReportError(
			timeframe, "Timeframe", "timeframe", "invalid-val-month", "",
		)
	} else if timeframe == ISOTimeframeYear && !vt.Year {
		sl.ReportError(
			timeframe, "Timeframe", "timeframe", "invalid-val-year", "",
		)
	}
}
