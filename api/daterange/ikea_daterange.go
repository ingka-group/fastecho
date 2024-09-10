package daterange

import (
	"github.com/ingka-group-digital/ocp-go-utils/date"
)

// IKEADateRange represents an IKEA date range. The Timeframe is required to group the data by the specific date range.
type IKEADateRange struct {
	BasicDateRange
	Timeframe Timeframe `query:"timeframe" validate:"required,oneof=day week month year" example:"week"`
} // @name IKEADateRange

// GetBasicDateRange returns the from-to date range.
func (d IKEADateRange) GetBasicDateRange() *BasicDateRange {
	return &d.BasicDateRange
}

// GetTimeframe returns the timeframe.
func (d IKEADateRange) GetTimeframe() Timeframe {
	return d.Timeframe
}

// GetTimeColumns returns the time columns based which the data should be grouped by.
func (d IKEADateRange) GetTimeColumns(alias ...bool) []string {
	useAlias := false
	if len(alias) > 0 {
		useAlias = alias[0]
	}

	if d.Timeframe == TimeframeDay {
		return []string{"date"}
	} else if d.Timeframe == TimeframeWeek {
		if useAlias {
			return []string{"ikea_year AS year", "ikea_week AS week"}
		}
		return []string{"ikea_year", "ikea_week"}
	} else if d.Timeframe == TimeframeMonth {
		if useAlias {
			return []string{"iso_year AS year", "iso_month AS month"}
		}
		return []string{"iso_year", "iso_month"}
	}

	if useAlias {
		return []string{"financial_year AS year"}
	}
	return []string{"financial_year"}
}

// GetWhereClauseSQL returns the SQL where clause as a string.
func (d IKEADateRange) GetWhereClauseSQL() string {
	fromYear, fromWeek := date.IKEAWeek(d.From.Year(), int(d.From.Month()), d.From.Day())
	toYear, toWeek := date.IKEAWeek(d.To.Year(), int(d.To.Month()), d.To.Day())

	opts := &sqlWhereClauseOpts{
		FromYear: fromYear,
		ToYear:   toYear,
		FromWeek: fromWeek,
		ToWeek:   toWeek,
	}

	if d.Timeframe == TimeframeYear {
		opts.FromYear = date.IKEAFinancialYear(d.From.Year(), int(d.From.Month()))
		opts.ToYear = date.IKEAFinancialYear(d.To.Year(), int(d.To.Month()))
	}

	return getWhereClauseSQL(d.From, d.To, d.Timeframe, opts, d.GetTimeColumns())
}
