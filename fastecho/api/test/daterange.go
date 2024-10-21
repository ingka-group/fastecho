package test

import (
	"time"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

// NOTE!
// The following functions are used to generate the start date of a given week, month or year.
// They are very useful for testing purposes on the API level when you need to pass a date range for specific week.
//
// For example, if we want to pass a range:
// 	From: 2024 w50
// 	To: 2024 w51
// We would have to compute somehow the start date of the week 50 and 51 of 2024.
//
// With the functions below, this can be done as follows:
// 	from := WeekToISODate(2024, 50)
// 	to := WeekToISODate(2024, 51)
// Apart from the computation, the functions are also useful for readability and maintainability.
//
// The same applies for months and years.

// WeekToISODate accepts a year and week and generates the start date for the given week in the format of YYYY-MM-DD.
func WeekToISODate(year, week int) date.ISODate {
	from := date.IKEAWeekFirstDay(year, week)
	return date.FromTime(from)
}

// WeekToISODateStr is similar to WeekToISODate but returns the date as a string.
func WeekToISODateStr(year, week int) string {
	from := WeekToISODate(year, week)
	return from.String()
}

// MonthToISODate accepts a year and month and generates the start date for the given month in the format of YYYY-MM-DD.
func MonthToISODate(year, month int) date.ISODate {
	// The start of the month is simply the first day of the month
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return date.FromTime(from)
}

// MonthToISODateStr is similar to MonthToISODate but returns the date as a string.
func MonthToISODateStr(year, month int) string {
	from := MonthToISODate(year, month)
	return from.String()
}

// YearToISODate accepts a year and generates the start date for the given year in the format of YYYY-MM-DD.
func YearToISODate(year int) date.ISODate {
	// The start of the month is simply the first day of the month
	from := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	return date.FromTime(from)
}

// YearToISODateStr is similar to YearToISODate but returns the date as a string.
func YearToISODateStr(year int) string {
	from := YearToISODate(year)
	return from.String()
}
