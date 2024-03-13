package agg

import (
	"fmt"
	"time"

	"github.com/ingka-group-digital/ocp-go-utils/date"

	_ "github.com/ingka-group-digital/ocp-go-utils/fp" // for documentation
)

// Dated is an interface used to represent a type associated to a certain point in time. It is not
// necessary that a type uses the full precision of the [time.Time] struct.
//
// This interface facilitates aggregations (grouping, statistics etc.) on data in a time series.
//
// Implementations of Dated should provide the highest precision possible while leaving the other
// fields set to their default values. Users of Dated are not guaranteed to act on the entire
// underlying object.
type Dated interface {
	Date() time.Time
}

// Year represents a single year.
type Year int

// Returns January 1st of the given year.
func (y Year) Date() time.Time {
	return time.Date(int(y), 1, 1, 0, 0, 0, 0, time.UTC)
}

func (y Year) String() string {
	return fmt.Sprintf("%d", int(y))
}

// YearMonth represents a month in a year.
type YearMonth struct {
	Year  int
	Month int
}

// Returns the 1st day of the given month.
func (ym YearMonth) Date() time.Time {
	return time.Date(ym.Year, time.Month(ym.Month), 1, 0, 0, 0, 0, time.UTC)
}

func (ym YearMonth) String() string {
	return fmt.Sprintf("%d-%02d", ym.Year, ym.Month)
}

// IKEAWeek represents a single IKEA week, as calculated by [date.IKEAWeek].
type IKEAWeek struct {
	Year int
	Week int
}

func (w IKEAWeek) Date() time.Time {
	return date.IKEAWeekFirstDay(w.Year, w.Week)
}

func (w IKEAWeek) String() string {
	return fmt.Sprintf("%d-W%02d", w.Year, w.Week)
}

// TimeAggregation is a constraint that permits any of the built-in time series groups provided by
// the grouper functions in this package.
type TimeAggregation interface {
	Year | YearMonth | IKEAWeek | date.ISODate
}

// ByYear is a grouper function that groups a [Dated] by year.
//
// See [ByIKEAWeek] for an example of grouper usage.
func ByYear[D Dated](v D) Year {
	return Year(v.Date().Year())
}

// ByYearMonth is a grouper function that groups a [Dated] by year and month.
//
// See [ByIKEAWeek] for an example of grouper usage.
func ByYearMonth[D Dated](v D) YearMonth {
	return YearMonth{Year: v.Date().Year(), Month: int(v.Date().Month())}
}

// ByIKEAWeek is a grouper function that groups a [Dated] by IKEA week.
//
// You can use this function as a [fp.GroupBy] callback to group any slice of [Dated] types into a
// map with [IKEAWeek] structs as keys:
//
//	type dataPoint struct { value int; ts time.Time }
//	func (d dataPoint) Date() time.Time { return d.ts }
//
//	var dataPoints []dataPoint // sourced from somewhere
//	fp.GroupBy(dataPoints, ByIKEAWeek)
//
// You can then further aggregate the result by combining [fp.MapHash] with other aggregate
// functions.
func ByIKEAWeek[D Dated](v D) IKEAWeek {
	year, week := date.IKEAWeek(v.Date().Year(), int(v.Date().Month()), v.Date().Day())
	return IKEAWeek{Year: year, Week: week}
}

// ByISODate is a grouper function that groups a [Dated] by [date.ISODate]
func ByISODate[D Dated](v D) date.ISODate {
	return date.FromTime(v.Date())
}
