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

type Interval interface {
	Dated
	// Add advances an interval by a provided amount. Amount can be negative, in which case the
	// interval goes back in time.
	//
	// Example: advancing the interval of `YearMonth{Year: 2024, Month: 1}` by 1 would return
	// `YearMonth{Year: 2024, Month: 2}`. Advancing it by -1 would return `YearMonth{Year: 2023,
	// Month: 12}`.
	Add(amount int) Interval

	// Start returns the first instant in the interval. It should be equivalent to `Date()` in most
	// situations.
	Start() time.Time

	// End returns the last day in the interval (at midnight).
	End() time.Time
}

// Year represents a single year.
type Year int

// Date returns January 1st of the given year.
func (y Year) Date() time.Time {
	return time.Date(int(y), 1, 1, 0, 0, 0, 0, time.UTC)
}

// String returns a textual representation of the year.
func (y Year) String() string {
	return fmt.Sprintf("%d", int(y))
}

// Add advances the year. A negative value will return a past year.
func (y Year) Add(amount int) Interval {
	return Year(int(y) + amount)
}

// Start returns the first day of the year.
func (y Year) Start() time.Time {
	return y.Date()
}

// End returns the last day of the year.
func (y Year) End() time.Time {
	return time.Date(int(y), 12, 31, 0, 0, 0, 0, time.UTC)
}

// YearMonth represents a month in a year.
type YearMonth struct {
	Year  int
	Month int
}

// Date returns the 1st day of the given month.
func (ym YearMonth) Date() time.Time {
	return time.Date(ym.Year, time.Month(ym.Month), 1, 0, 0, 0, 0, time.UTC)
}

func (ym YearMonth) String() string {
	return fmt.Sprintf("%d-%02d", ym.Year, ym.Month)
}

// Add advances the YearMonth by a set amount.
func (ym YearMonth) Add(amount int) Interval {
	new := ym.Date().AddDate(0, amount, 0)
	return YearMonth{
		Year:  new.Year(),
		Month: int(new.Month()),
	}
}

// Start returns the first day of the month.
func (ym YearMonth) Start() time.Time {
	return ym.Date()
}

// End returns the last day of the month.
func (ym YearMonth) End() time.Time {
	return ym.Date().AddDate(0, 1, -1)
}

// IKEAWeek represents a single IKEA week, as calculated by [date.IKEAWeek].
type IKEAWeek struct {
	Year int
	Week int
}

func (w IKEAWeek) Date() time.Time {
	return date.IKEAWeekFirstDay(w.Year, w.Week)
}

// String returns a textual representation of the IKEA week (e.g. `2024-W01`)
func (w IKEAWeek) String() string {
	return fmt.Sprintf("%d-W%02d", w.Year, w.Week)
}

// Add advances the IKEA Week by the specified amount of weeks.. The amount can be negative. This
// function will gracefully handle transitions across years.
func (w IKEAWeek) Add(amount int) Interval {
	new := w.Date().AddDate(0, 0, amount*7)
	return ByIKEAWeek(date.FromTime(new))
}

// Start returns the first day of the week.
func (w IKEAWeek) Start() time.Time {
	return w.Date()
}

// End returns the last day of the week.
func (w IKEAWeek) End() time.Time {
	return w.Date().AddDate(0, 0, 6)
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
