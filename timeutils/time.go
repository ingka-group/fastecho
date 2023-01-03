package timeutils

import (
	"time"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

// ToISODateStr converts a time object to a string representation of date.ISODateFmt (yyyy-mm-dd)
func ToISODateStr(d time.Time) string {
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
