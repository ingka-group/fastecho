package daterange

import "github.com/ingka-group-digital/ocp-go-utils/date"

// IKEATimeframe represents the timeframe for an IKEADateRange.
type IKEATimeframe string

const (
	IKEATimeframeDay   IKEATimeframe = "day"
	IKEATimeframeWeek  IKEATimeframe = "week"
	IKEATimeframeMonth IKEATimeframe = "month"
	IKEATimeframeYear  IKEATimeframe = "year"
)

// GetTimeColumns returns the time columns based which the data should be grouped by.
// Returns the columns in the order they should be used in the GROUP BY clause.
// For timeframe `day` the column name for date is returned
// For timeframe `week` the column names for year and week are returned
// For timeframe `month` the column names for year and month are returned
// For timeframe `year` the column name for year is returned
func (t IKEATimeframe) GetTimeColumns() []string {
	if t == IKEATimeframeDay {
		return []string{"date"}
	} else if t == IKEATimeframeWeek {
		return []string{"ikea_year", "ikea_week"}
	} else if t == IKEATimeframeMonth {
		return []string{"iso_year", "iso_month"}
	}

	return []string{"iso_year"}
}

// IKEADateRange represents an IKEADate range. The Timeframe is required to group the data by the specific date range.
type IKEADateRange struct {
	DateRangeBasic
	Timeframe IKEATimeframe `query:"timeframe" validate:"required,oneof=day week month year" example:"week"`
} // @name IKEADateRange

func (d IKEADateRange) GetDateRangeBasic() *DateRangeBasic {
	return &d.DateRangeBasic
}

func (d IKEADateRange) GetTimeframe() Timeframe {
	return d.Timeframe
}

// GetWhereClause returns the where clause of the timeframe for the given date range.
// Assumptions:
// - The date columns are in the format of 'yyyy-mm-dd'.
func (t IKEATimeframe) GetWhereClause(from, to date.ISODate) (string, []interface{}) {
	ret := ""

	switch t {
	case IKEATimeframeDay:
		dateColumn := t.GetTimeColumns()[0]
		return dateColumn + " BETWEEN ? AND ?", []interface{}{from.String(), to.String()}

	case IKEATimeframeWeek:
		yearColumn, weekColumn := t.GetTimeColumns()[0], t.GetTimeColumns()[1]
		fromYear, fromWeek := date.IKEAWeek(from.Year(), int(from.Month()), from.Day())
		toYear, toWeek := date.IKEAWeek(to.Year(), int(to.Month()), to.Day())

		if fromYear == toYear {
			return "(" + yearColumn + " = ?) AND (" + weekColumn + " BETWEEN ? AND ?)", []interface{}{fromYear, fromWeek, toWeek}
		} else {
			// The date range spans multiple years

			// This avoids having to build query by looping through the years
			// If year > fromYear and year < toYear
			ret = "(" + yearColumn + " > ? AND " + yearColumn + " < ?) " +
				// If year is fromYear, then week should be >= fromWeek
				"OR (" + yearColumn + " = ? AND " + weekColumn + " >= ?) " +
				// If year is toYear, then week should be <= toWeek
				"OR (" + yearColumn + " = ? AND " + weekColumn + " <= ?)"

			return ret, []interface{}{fromYear, toYear, fromYear, fromWeek, toYear, toWeek}
		}

	// Same logic as for week, but with month instead of week
	// Duplicating the logic only to avoid non-readable code if we handle month and week together
	case IKEATimeframeMonth:
		yearColumn, monthColumn := t.GetTimeColumns()[0], t.GetTimeColumns()[1]
		fromYear, fromMonth := from.Year(), int(from.Month())
		toYear, toMonth := to.Year(), int(to.Month())

		// If the date range is within the same year, the query is simple
		if fromYear == toYear {
			return yearColumn + " = ? AND " + monthColumn + " BETWEEN ? AND ?", []interface{}{fromYear, fromMonth, toMonth}
		} else {
			// The date range spans multiple years

			// This avoids having to build query by looping through the years
			// If year > fromYear and year < toYear
			ret = "(" + yearColumn + " > ? AND " + yearColumn + " < ?) " +
				// If year is fromYear, then week should be >= fromMonth
				"OR (" + yearColumn + " = ? AND " + monthColumn + " >= ?) " +
				// If year is toYear, then week should be <= toMonth
				"OR (" + yearColumn + " = ? AND " + monthColumn + " <= ?) "

			return ret, []interface{}{fromYear, toYear, fromYear, fromMonth, toYear, toMonth}
		}

	}

	// case IKEATimeframeYear:
	yearColumn := t.GetTimeColumns()[0]
	return yearColumn + " BETWEEN ? AND ?", []interface{}{from.Year(), to.Year()}
}
