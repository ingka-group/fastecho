package daterange

// ISODateRange represents an ISODate range. The Timeframe is required to group the data by the specific date range.
type ISODateRange struct {
	BasicDateRange
	Timeframe Timeframe `query:"timeframe" validate:"required,oneof=day week month year" example:"week"`
} // @name ISODateRange

// GetBasicDateRange returns the from-to date range.
func (d ISODateRange) GetBasicDateRange() *BasicDateRange {
	return &d.BasicDateRange
}

// GetTimeframe returns the timeframe.
func (d ISODateRange) GetTimeframe() Timeframe {
	return d.Timeframe
}

// GetTimeColumns returns the time columns based which the data should be grouped by.
func (d ISODateRange) GetTimeColumns(alias ...bool) []string {
	// We can ignore the alias as we don't need it in this case
	if d.Timeframe == TimeframeDay {
		return []string{"date"}
	} else if d.Timeframe == TimeframeWeek {
		return []string{TimeframeYear.String(), TimeframeWeek.String()}
	} else if d.Timeframe == TimeframeMonth {
		return []string{TimeframeYear.String(), TimeframeMonth.String()}
	}

	return []string{TimeframeYear.String()}
}

// GetWhereClauseSQL returns the SQL where clause as a string.
func (d ISODateRange) GetWhereClauseSQL() string {
	fromYear, fromWeek := d.From.ISOWeek()
	toYear, toWeek := d.To.ISOWeek()

	opts := &sqlWhereClauseOpts{
		FromYear: fromYear,
		ToYear:   toYear,
		FromWeek: fromWeek,
		ToWeek:   toWeek,
	}

	return getWhereClauseSQL(d.From, d.To, d.Timeframe, opts, d.GetTimeColumns())
}
