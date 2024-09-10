package daterange

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

// !NOTE
// The ISODateRange, IKEADateRange and BasicDateRange structs are used to represent the date and date range in the API.
// The functionality here is very opinionated. The goal is to standardize our APIs, make the more consistent on
// how we handle dates and date ranges. The functionality here offers validation based on the go-playground/validator.
// The validation occurs using `validate` struct tags but also with custom validation provided by the Validate* functions.
// In case you prefer to void using that, you can provide your own functions for validation on the API level.
//
// The idea behind the ISODateRange or IKEADateRange is that every incoming request on an API that has to do with time-series data,
// *must* specify `from` and `to` as `YYYY-MM-DD` input format. The Timeframe is used to specify the time aggregation.
// For example:
//
// If the Timeframe is set to `day`, the API will aggregate the data on a daily basis.
//
// If the Timeframe is set to `week`, the API will aggregate the data on a weekly basis. To achieve that,
// the API would extract the year and week number from the date and group the data by that.
//
// If the Timeframe is set to `month`, the API will aggregate the data on a monthly basis. To achieve that,
// the API would extract the year and month number from the date and group the data by that.
//
// If the Timeframe is set to `year`, the API will aggregate the data on a yearly basis. To achieve that,
// the API would extract the year from the date and group the data by that.
//
// Based on the Timeframe one can extract the database columns that need to be queried. This can be provided by the GetTimeColumns
// function which assumes that columns are named after the timeframe (week, month, year).
// For Timeframe `day` the column name is expected to be `date`
//
// Finally, the Timeframe is not used in the BasicDateRange struct as it is used to represent any time range without any specific
// time aggregation specified living it to the API to decide. For example, specific API calls might require a date range but
// by default they will aggregate the data on a weekly basis.
//
// The difference between the ISODateRange and the IKEADateRange lies only in the columns that are used for grouping the data.
// For IKEADateRange specific columns are required to be present and not the obvious week, month and year.
//

// DateRange represents a date range with a specific timeframe.
type DateRange interface {
	GetBasicDateRange() *BasicDateRange
	GetTimeframe() Timeframe

	// GetTimeColumns returns the time columns based which the data should be grouped by.
	// Returns the columns in the order they should be used in the GROUP BY clause.
	// For timeframe `day` the column name for date is returned
	// For timeframe `week` the column names for year and week are returned
	// For timeframe `month` the column names for year and month are returned
	// For timeframe `year` the column name for year is returned
	//
	// alias<bool> can be used to return the columns with an alias,
	// in case the function is used in a SELECT statement. The argument is optional
	// and by default is 'false'.
	GetTimeColumns(alias ...bool) []string

	// GetWhereClauseSQL returns the SQL where clause as a string.
	GetWhereClauseSQL() string
} // @name DateRange

// BasicDateRange represents the base parameters for a date range.
// It is called basic because it does not contain the Timeframe for data time-aggregations.
type BasicDateRange struct {
	From date.ISODate `query:"from" validate:"required" example:"2024-12-01"`
	To   date.ISODate `query:"to" validate:"required" example:"2024-12-10"`
} // @name BasicDateRange

// ValidateBasicDateRange validates the BasicDateRange struct.
//
// Example:
//
// validator.RegisterStructValidation(ValidateBasicDateRange, BasicDateRange{})
func ValidateBasicDateRange() validator.StructLevelFunc {
	return func(sl validator.StructLevel) {
		basicDateRange := sl.Current().Interface().(BasicDateRange)
		validateBasicDateRange(sl, &basicDateRange)
	}
}

func validateBasicDateRange(sl validator.StructLevel, dateRange *BasicDateRange) {
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
// ), < ISODateRange{} || IKEADateRange{} >)
func ValidateDateRange(vt *ValidTimeframes) validator.StructLevelFunc {
	return func(sl validator.StructLevel) {
		validateDateRange(sl, vt)
	}
}

func validateDateRange(sl validator.StructLevel, vt *ValidTimeframes) {
	dateRange := sl.Current().Interface().(DateRange)
	timeframe := dateRange.GetTimeframe()

	validateBasicDateRange(sl, dateRange.GetBasicDateRange())

	if timeframe == TimeframeDay && !vt.Day {
		sl.ReportError(
			timeframe, "Timeframe", "timeframe", "invalid-val-day", "",
		)
	} else if timeframe == TimeframeWeek && !vt.Week {
		sl.ReportError(
			timeframe, "Timeframe", "timeframe", "invalid-val-week", "",
		)
	} else if timeframe == TimeframeMonth && !vt.Month {
		sl.ReportError(
			timeframe, "Timeframe", "timeframe", "invalid-val-month", "",
		)
	} else if timeframe == TimeframeYear && !vt.Year {
		sl.ReportError(
			timeframe, "Timeframe", "timeframe", "invalid-val-year", "",
		)
	}
}

type sqlWhereClauseOpts struct {
	FromYear int
	ToYear   int
	FromWeek int
	ToWeek   int
}

func getWhereClauseSQL(from, to date.ISODate, t Timeframe, opts *sqlWhereClauseOpts, cols []string) string {
	switch t {
	case TimeframeDay:
		dateColumn := cols[0]
		return fmt.Sprintf("(%s BETWEEN '%s' AND '%s')", dateColumn, from.String(), to.String())
	case TimeframeWeek:
		yearColumn, weekColumn := cols[0], cols[1]

		if opts.FromYear == opts.ToYear {
			return fmt.Sprintf(`(%s = %v AND %s BETWEEN %v AND %v)`, yearColumn, opts.FromYear, weekColumn, opts.FromWeek, opts.ToWeek)
		} else {
			// The date range spans multiple years
			firstYearClause := fmt.Sprintf(`(%s = %v AND %s >= %v)`, yearColumn, opts.FromYear, weekColumn, opts.FromWeek)
			middleYearsClause := fmt.Sprintf(`(%s > %v AND %s < %v)`, yearColumn, opts.FromYear, yearColumn, opts.ToYear)
			lastYearClause := fmt.Sprintf(`(%s = %v AND %s <= %v)`, yearColumn, opts.ToYear, weekColumn, opts.ToWeek)

			return fmt.Sprintf(`(%s OR %s OR %s)`, middleYearsClause, firstYearClause, lastYearClause)
		}
	case TimeframeMonth:
		// NOTE
		// 	> Are we sure this is valid for IKEADateRange?
		yearColumn, monthColumn := cols[0], cols[1]
		fromYear, fromMonth := from.Year(), int(from.Month())
		toYear, toMonth := to.Year(), int(to.Month())

		if fromYear == toYear {
			return fmt.Sprintf(`(%s = %v AND %s BETWEEN %v AND %v)`, yearColumn, fromYear, monthColumn, fromMonth, toMonth)
		} else {
			// The date range spans multiple years
			firstYearClause := fmt.Sprintf(`(%s = %v AND %s >= %v)`, yearColumn, fromYear, monthColumn, fromMonth)
			middleYearsClause := fmt.Sprintf(`(%s > %v AND %s < %v)`, yearColumn, fromYear, yearColumn, toYear)
			lastYearClause := fmt.Sprintf(`(%s = %v AND %s <= %v)`, yearColumn, toYear, monthColumn, toMonth)

			return fmt.Sprintf(`(%s OR %s OR %s)`, middleYearsClause, firstYearClause, lastYearClause)
		}
	}

	yearColumn := cols[0]
	return fmt.Sprintf(`(%s BETWEEN %v AND %v)`, yearColumn, opts.FromYear, opts.ToYear)
}
