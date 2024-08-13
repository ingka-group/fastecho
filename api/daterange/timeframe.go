package daterange

import "github.com/ingka-group-digital/ocp-go-utils/date"

type Timeframe interface {
	// GetTimeColumns returns the time columns based which the data should be grouped by.
	// Returns the columns in the order they should be used in the GROUP BY clause.
	// For timeframe `day` the column name for date is returned
	// For timeframe `week` the column names for year and week are returned
	// For timeframe `month` the column names for year and month are returned
	// For timeframe `year` the column name for year is returned
	GetTimeColumns() []string

	// GetWhereClause returns the where clause with the values in placeholders for the given date range.
	// This helps avoid SQL injection.
	// Sample usage:
	// whereClause, values := timeframe.GetWhereClause(from, to)
	// gormDb.Where(whereClause, values...)
	GetWhereClause(from, to date.ISODate) (string, []interface{})
}

// ValidTimeframes accepts the valid timeframes.
type ValidTimeframes struct {
	Day   bool
	Week  bool
	Month bool
	Year  bool
}
