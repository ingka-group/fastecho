package daterange

// Timeframe represents the timeframe.
type Timeframe string

const (
	TimeframeDay   Timeframe = "day"
	TimeframeWeek  Timeframe = "week"
	TimeframeMonth Timeframe = "month"
	TimeframeYear  Timeframe = "year"
)

// String returns the string representation of the Timeframe.
func (t Timeframe) String() string {
	return string(t)
}

// ValidTimeframes accepts the valid timeframes.
type ValidTimeframes struct {
	Day   bool
	Week  bool
	Month bool
	Year  bool
}
