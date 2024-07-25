package date

import (
	"time"
)

// IKEAWeekFirstDay returns the first day of the week in the IKEA week numbering system.
func IKEAWeekFirstDay(y, w int) time.Time {
	firstSunday := time.Date(y, 1, 4, 0, 0, 0, 0, time.UTC)

	// first week is the one with jan4 in it
	if firstSunday.Weekday() != time.Sunday && w == 1 {
		// first day of the week is before jan4
		for firstSunday.Weekday() != time.Sunday {
			firstSunday = firstSunday.AddDate(0, 0, -1)
		}
		return firstSunday
	}

	// iterate to first sunday of year after jan4
	for firstSunday.Weekday() != time.Sunday {
		firstSunday = firstSunday.AddDate(0, 0, +1)
	}
	// loop every 7 days until we reach week number
	firstDayWeek := time.Date(y, firstSunday.Month(), firstSunday.Day(), 0, 0, 0, 0, time.UTC)
	_, ikeaWeek := IKEAWeek(firstDayWeek.Year(), int(firstDayWeek.Month()), firstDayWeek.Day())
	for w != ikeaWeek {
		firstDayWeek = firstDayWeek.AddDate(0, 0, 7)
		_, ikeaWeek = IKEAWeek(firstDayWeek.Year(), int(firstDayWeek.Month()), firstDayWeek.Day())
	}

	return firstDayWeek
}

// IKEAWeek returns the year and week number in which the given date (specified by year, month, day) occurs,
// according to the IKEA week numbering system. This follows ISO 8601, but with the exception that weeks start on Sunday.
// This leads for some years to the possibility of creating a nonexisting week number week 53 eg. 2004/2015/2026, etc.
func IKEAWeek(y, m, day int) (weekYear, week int) {
	d := time.Date(y, time.Month(m), day, 0, 0, 0, 0, time.UTC)
	year, week := d.ISOWeek()
	// if it's sunday we move it to next week because weeks need to start on sundays
	if d.Weekday() == time.Sunday {
		// shift week
		if week < 52 {
			week++
		} else if week == 52 {
			// next week can be either 53 or 1, we find out by looking at the week the next day is in
			nextDay := d.Add(time.Hour * 24)
			nextYearWeek, nextweek := nextDay.ISOWeek()
			week = nextweek
			year = nextYearWeek
		} else if week == 53 {
			// reset year and week
			week = 1
			year++
		}
	}
	// for the years where jan 4th is the first day of the yearweek, we created week 53 out of thin air
	// this means that all the previous calculation that incremented week by 1 is not needed, so we need to decrement it
	jan4th := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC) // IMPORTANT we use here year instead of d.Year() because eg. 2005-01-01 is saturday in week 53 of 2004
	if jan4th.Weekday() == time.Sunday || (d.Year() != year && d.Weekday() == time.Sunday && d.Month() == time.January) {
		// this happens to all weeks during a year which has jan 4th as the first day of the yearweek
		// and also to the first week of the year after
		week--
	}
	// in december or january we may need to create week 53 if jan 4th is on a sunday
	if d.Month() == time.December || d.Month() == time.January {
		// I'm doing it more verbose than just jan4 = time.Date(d.Year()+(year-d.Year()), 1, 4, 0, 0, 0, 0, time.UTC)
		var jan4 time.Time
		if d.Month() == time.December {
			jan4 = time.Date(d.Year()+1, 1, 4, 0, 0, 0, 0, time.UTC)
		} else if d.Month() == time.January {
			jan4 = time.Date(d.Year(), 1, 4, 0, 0, 0, 0, time.UTC)
		}
		// if current date is in the week before jan 4 when it falls on a sunday we create week 53
		if d.After(jan4.AddDate(0, 0, -8)) && d.Before(jan4) && jan4.Weekday() == time.Sunday {
			week = 53
			if d.Year() == jan4.Year() {
				year = d.Year() - 1
			} else {
				year = d.Year()
			}
		}
	}
	return year, week
}
