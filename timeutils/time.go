package timeutils

import (
	"time"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

const (
	ISODateFmt = "2006-01-02" // yyyy-mm-dd
)

// ToISODate converts a time object to a string representation of an ISODate (yyyy-mm-dd)
func ToISODate(d time.Time) string {
	return d.Format(ISODateFmt)
}

// DateToDDWWMMYYYY breaks the date given into day, week, month, year
func DateToDDWWMMYYYY(d time.Time) (int, int, int, int) {
	day := d.Day()
	_, week := date.IKEAWeek(d.Year(), int(d.Month()), d.Day())
	month := int(d.Month())
	year := d.Year()

	return day, week, month, year
}
