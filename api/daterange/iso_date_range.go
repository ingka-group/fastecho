package daterange

import (
	"fmt"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

// !NOTE
// The ISODateRange and DateRangeBasic structs are used to represent the date and date range in the API.
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
// Finally, the Timeframe is not used in the DateRangeBasic struct as it is used to represent any time range without any specific
// time aggregation specified living it to the API to decide. For example, specific API calls might require a date range but
// by default they will aggregate the data on a weekly basis.

// ISOTimeframe represents the timeframe for an ISODateRange.
type ISOTimeframe string

const (
	ISOTimeframeDay   ISOTimeframe = "day"
	ISOTimeframeWeek  ISOTimeframe = "week"
	ISOTimeframeMonth ISOTimeframe = "month"
	ISOTimeframeYear  ISOTimeframe = "year"
)

// String returns the string representation of the Timeframe.
func (t ISOTimeframe) String() string {
	return string(t)
}

// GetTimeColumns returns the time columns based on the Timeframe.
//
// The function assumes that columns are named after the timeframe (week, month, year).
// For `day` timeframe, the column name is expected to be `date`.
func (t ISOTimeframe) GetTimeColumns() []string {
	if t == ISOTimeframeDay {
		return []string{"date"}
	} else if t == ISOTimeframeWeek {
		return []string{ISOTimeframeYear.String(), ISOTimeframeWeek.String()}
	} else if t == ISOTimeframeMonth {
		return []string{ISOTimeframeYear.String(), ISOTimeframeMonth.String()}
	}

	return []string{ISOTimeframeYear.String()}
}

func (t ISOTimeframe) GetWhereClause(from, to date.ISODate) string {
	switch t {
	case ISOTimeframeDay:
		dateColumn := t.GetTimeColumns()[0]
		return fmt.Sprintf("%s BETWEEN '%s' AND '%s'", dateColumn, from.String(), to.String())
	case ISOTimeframeWeek:
		yearColumn, weekColumn := t.GetTimeColumns()[0], t.GetTimeColumns()[1]
		fromYear, fromWeek := from.ISOWeek()
		toYear, toWeek := to.ISOWeek()

		if fromYear == toYear {
			return fmt.Sprintf(`%s = %v AND %s BETWEEN %v AND %v`, yearColumn, fromYear, weekColumn, fromWeek, toWeek)
		} else {
			// The date range spans multiple years
			firstYearClause := fmt.Sprintf(`(%s = %v AND %s >= %v)`, yearColumn, fromYear, weekColumn, fromWeek)
			middleYearsClause := fmt.Sprintf(`(%s > %v AND %s < %v)`, yearColumn, fromYear, yearColumn, toYear)
			lastYearClause := fmt.Sprintf(`(%s = %v AND %s <= %v)`, yearColumn, toYear, weekColumn, toWeek)

			return fmt.Sprintf(`%s OR %s OR %s`, middleYearsClause, firstYearClause, lastYearClause)
		}
	case ISOTimeframeMonth:
		yearColumn, monthColumn := t.GetTimeColumns()[0], t.GetTimeColumns()[1]
		fromMonth, fromYear := int(from.Month()), from.Year()
		toMonth, toYear := int(to.Month()), to.Year()

		if fromYear == toYear {
			return fmt.Sprintf(`%s = %v AND %s BETWEEN %v AND %v`, yearColumn, fromYear, monthColumn, fromMonth, toMonth)
		} else {
			// The date range spans multiple years
			firstYearClause := fmt.Sprintf(`(%s = %v AND %s >= %v)`, yearColumn, fromYear, monthColumn, fromMonth)
			middleYearsClause := fmt.Sprintf(`(%s > %v AND %s < %v)`, yearColumn, fromYear, yearColumn, toYear)
			lastYearClause := fmt.Sprintf(`(%s = %v AND %s <= %v)`, yearColumn, toYear, monthColumn, toMonth)

			return fmt.Sprintf(`%s OR %s OR %s`, middleYearsClause, firstYearClause, lastYearClause)
		}
	}

	// case IKEATimeframeYear:
	yearColumn := t.GetTimeColumns()[0]
	return fmt.Sprintf(`%s BETWEEN %v AND %v`, yearColumn, from.Year(), to.Year())
}

// ISODateRange represents an ISODate range. The Timeframe is required to group the data by the specific date range.
type ISODateRange struct {
	DateRangeBasic
	Timeframe ISOTimeframe `query:"timeframe" validate:"required,oneof=day week month year" example:"week"`
} // @name ISODateRange

func (d ISODateRange) GetDateRangeBasic() *DateRangeBasic {
	return &d.DateRangeBasic
}

func (d ISODateRange) GetTimeframe() Timeframe {
	return d.Timeframe
}
