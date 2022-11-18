package timeutils

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

const (
	ISODateFmt = "2006-01-02" // yyyy-mm-dd
)

// ExtractDate removes completely the time part from a dateTime string
func ExtractDate(dt string) string {
	return strings.Split(dt, "T")[0]
}

// ToDate converts a dateTime string to time object
//
// The function first excludes the time from the computation
// and then assigns it again. The new assigned time will always be 00:00:00Z
func ToDate(dt string) (time.Time, error) {
	if len(dt) == 0 {
		return time.Time{}, errors.New("dateTime is empty")
	}

	// Remove T part from dateTime string if exists
	d := ExtractDate(dt)
	date, err := time.Parse(ISODateFmt, d)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date '%s': %s", dt, err.Error())
	}

	return date, nil
}

// DateToDDWWMMYYYY breaks the date given into day, week, month, year
func DateToDDWWMMYYYY(d time.Time) (int, int, int, int) {
	day := d.Day()
	_, week := date.IKEAWeek(d.Year(), int(d.Month()), d.Day())
	month := int(d.Month())
	year := d.Year()

	return day, week, month, year
}
