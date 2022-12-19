package timeutils

import (
	"time"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

// ToISODate converts a time object to a string representation of a date.ISODate (yyyy-mm-dd)
func ToISODate(d time.Time) string {
	return d.Format(date.ISODateFmt)
}

// DateToDDWWMMYYYY breaks the date given into day, week, month, year
func DateToDDWWMMYYYY(d time.Time) (int, int, int, int) {
	day := d.Day()
	_, week := date.IKEAWeek(d.Year(), int(d.Month()), d.Day())
	month := int(d.Month())
	year := d.Year()

	return day, week, month, year
}
