package daterange

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/ingka-group-digital/ocp-go-utils/date"
	"github.com/ingka-group-digital/ocp-go-utils/timeutils"
)

// !NOTE
// The ISODateRange and ISODateRangeBasic structs are used to represent the date and date range in the API.
// The functionality here is very opinionated. The goal is to standardize our APIs, make the more consistent on
// how we handle dates and date ranges. The functionality here offers validation based on the go-playground/validator.
// The validation occurs using `validate` struct tags but also with custom validation provided by the Validate* functions.
// In case you prefer to void using that, you can provide your own functions for validation on the API level.
//
// The idea behind the ISODateRange* is that every incoming request on an API that has to do with time-series data,
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
// Finally, the Timeframe is not used in the ISODateRangeBasic struct as it is used to represent any time range without any specific
// time aggregation specified living it to the API to decide. For example, specific API calls might require a date range but
// by default they will aggregate the data on a weekly basis.

// Timeframe represents the timeframe for an ISODateRange.
type Timeframe string

const (
	TimeframeDay   Timeframe = "day"
	TimeframeWeek  Timeframe = "week"
	TimeframeMonth Timeframe = "month"
	TimeframeYear  Timeframe = "year"
)

// ValidTimeframes accepts the valid timeframes.
type ValidTimeframes struct {
	Day   bool
	Week  bool
	Month bool
	Year  bool
}

// String returns the string representation of the Timeframe.
func (t Timeframe) String() string {
	return string(t)
}

// GetTimeColumns returns the time columns based on the Timeframe.
//
// The function assumes that columns are named after the timeframe (week, month, year).
// For `day` timeframe, the column name is expected to be `date`.
func (t Timeframe) GetTimeColumns() []string {
	if t == TimeframeDay {
		return []string{"date"}
	} else if t == TimeframeWeek {
		return []string{TimeframeYear.String(), TimeframeWeek.String()}
	} else if t == TimeframeMonth {
		return []string{TimeframeYear.String(), TimeframeMonth.String()}
	}

	return []string{TimeframeYear.String()}
}

// ISODateRange represents an ISODate range. The Timeframe is required to group the data by the specific date range.
type ISODateRange struct {
	ISODateRangeBasic
	Timeframe Timeframe `query:"timeframe" validate:"required,oneof=day week month year" example:"week"`
} // @name ISODateRange

// ISODateRangeBasic represents the base parameters for a date range.
// It is called basic because it does not contain the Timeframe for data time-aggregations.
type ISODateRangeBasic struct {
	From date.ISODate `query:"from" validate:"required" example:"2024-12-01"`
	To   date.ISODate `query:"to" validate:"required" example:"2024-12-10"`
} // @name ISODateRangeBasic

// ValidateISODateRange validates an ISODateRange. To identify which timeframes are supported, a ValidTimeframes struct
// is expected to be passed.
//
// Example:
//
// vdt.RegisterStructValidation(ValidateISODateRange(
//
//	&ValidTimeframes{
//		Day:   false,
//		Week:  true,
//		Month: true,
//		Year:  true,
//	},
//
// ), ISODateRange{})
func ValidateISODateRange(vt *ValidTimeframes) validator.StructLevelFunc {
	return func(sl validator.StructLevel) {
		validateISODateRange(sl, vt)
	}
}

func validateISODateRange(sl validator.StructLevel, vt *ValidTimeframes) {
	dateRange := sl.Current().Interface().(ISODateRange)
	timeframe := dateRange.Timeframe

	validateISODateRangeBasic(sl, &dateRange.ISODateRangeBasic)

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

// ValidateISODateRangeBasic validates the ISODateRangeBasic struct.
//
// Example:
//
// validator.RegisterStructValidation(ValidateISODateRangeBasic, ISODateRangeBasic{})
func ValidateISODateRangeBasic() validator.StructLevelFunc {
	return func(sl validator.StructLevel) {
		dateRangeBasic := sl.Current().Interface().(ISODateRangeBasic)
		validateISODateRangeBasic(sl, &dateRangeBasic)
	}
}

func validateISODateRangeBasic(sl validator.StructLevel, dateRange *ISODateRangeBasic) {
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

// AddToGormQuery adds a date range to a GORM query based on the from/to and the timeframe.
func AddToGormQuery(db *gorm.DB, from, to date.ISODate, timeframe Timeframe) *gorm.DB {
	_, fromWeek, fromMonth, fromYear := timeutils.DateToDDWWMMYYYY(from.Date())
	_, toWeek, toMonth, toYear := timeutils.DateToDDWWMMYYYY(to.Date())

	if timeframe == TimeframeDay {
		db = db.Where("date BETWEEN ? AND ?", from, to)
	} else if timeframe == TimeframeWeek || timeframe == TimeframeMonth {
		var fromWeekOrMonth int
		var toWeekOrMonth int

		if timeframe == TimeframeWeek {
			fromWeekOrMonth = fromWeek
			toWeekOrMonth = toWeek
		} else {
			fromWeekOrMonth = fromMonth
			toWeekOrMonth = toMonth
		}

		// Here we consider that timeframe is `week` or `month`,
		// which is the same column names in the database table :)
		if fromYear == toYear {
			// special case for weeks when the year changes
			if timeframe == TimeframeWeek && fromWeek > toWeek {
				db = db.Where("(year = ? AND week >= ?) OR (year = ? AND week <= ?)", fromYear-1, fromWeek, toYear, toWeek)
			} else {
				// same year
				db = db.Where(
					fmt.Sprintf("year = ? AND %s BETWEEN ? AND ?", timeframe), fromYear, fromWeekOrMonth, toWeekOrMonth,
				)
			}
		} else {
			orStmt := *db.Session(&gorm.Session{})
			dateRange := orStmt.Where(
				fmt.Sprintf("year = ? AND %s >= ?", timeframe), fromYear, fromWeekOrMonth,
			)

			for year := fromYear + 1; year < toYear; year++ {
				dateRange = dateRange.Or("year = ?", year)
			}

			dateRange = dateRange.Or(
				fmt.Sprintf("year = ? AND %s <= ?", timeframe), toYear, toWeekOrMonth,
			)

			db = db.Or(dateRange)
		}
	} else if timeframe == TimeframeYear {
		if fromYear == toYear {
			db = db.Where("year = ?", fromYear)
		} else {
			db = db.Where("year BETWEEN ? AND ?", fromYear, toYear)
		}
	}

	return db
}

// GetStringQuery returns a query represented as a string based on the from/to and the timeframe.
func GetStringQuery(from, to date.ISODate, timeframe Timeframe) string {
	_, fromWeek, fromMonth, fromYear := timeutils.DateToDDWWMMYYYY(from.Date())
	_, toWeek, toMonth, toYear := timeutils.DateToDDWWMMYYYY(to.Date())

	var query string

	if timeframe == TimeframeWeek || timeframe == TimeframeMonth {
		var fromWeekOrMonth int
		var toWeekOrMonth int

		if timeframe == TimeframeWeek {
			fromWeekOrMonth = fromWeek
			toWeekOrMonth = toWeek
		} else {
			fromWeekOrMonth = fromMonth
			toWeekOrMonth = toMonth
		}

		// Here we consider that timeframe is `week` or `month`,
		// which is the same column names in the database table :)
		if fromYear == toYear {
			// special case for weeks when the year changes
			if timeframe == TimeframeWeek && fromWeek > toWeek {
				query += fmt.Sprintf(
					"(year = %d AND week >= %d) OR (year = %d AND week <= %d)", fromYear-1, fromWeek, toYear, toWeek,
				)
			} else {
				// same year
				query += fmt.Sprintf(
					"year = %d AND %s BETWEEN %d AND %d", fromYear, timeframe, fromWeekOrMonth, toWeekOrMonth,
				)
			}
		} else {
			// date range
			query += fmt.Sprintf(
				"((year = %d AND %s >= %d )", fromYear, timeframe, fromWeekOrMonth,
			)

			for year := fromYear + 1; year < toYear+1; year++ {
				if year != toYear {
					query += fmt.Sprintf(
						" OR year = %d", year,
					)
				}

				if year == toYear {
					query += fmt.Sprintf(
						" OR (year = %d AND %s <= %d )", year, timeframe, toWeekOrMonth,
					)
				}
			}

			query += ")"
		}
	} else if timeframe == TimeframeYear {
		if fromYear == toYear {
			query += fmt.Sprintf(
				"year = %d", fromYear,
			)
		} else {
			query += fmt.Sprintf(
				"year >= %d AND year <= %d", fromYear, toYear,
			)
		}
	} else if timeframe == TimeframeDay {
		query += fmt.Sprintf(
			"date BETWEEN '%s' AND '%s'", from, to,
		)
	}

	return query
}
