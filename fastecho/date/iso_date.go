package date

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const (
	ISODateFmt = "2006-01-02" // yyyy-mm-dd
)

// ISODate wraps a time.Time object
type ISODate struct {
	time.Time
}

// UnmarshalJSON converts an incoming date string, e.g. 2022-01-01, in a time.Time object
//
// Golang calls this function by default in the background. However, if we use the default
// function, unmarshalling will fail because the 'yyyy-mm-dd' format cannot be mapped to a
// time.Time. To avoid this situation, we need to provide a custom unmarshaller, by overriding
// the original function 'UnmarshalJSON' so that Golang knows what to do when it encounters
// a type of ISODate which tries to be mapped to a time.Time
func (d *ISODate) UnmarshalJSON(b []byte) (err error) {
	raw := string(b)
	t, err := time.Parse(ISODateFmt, raw[1:len(raw)-1])
	if err != nil {
		return err
	}

	d.Time = t
	return
}

// MarshalJSON converts a time.Time object to an ISODate
func (d *ISODate) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", d.Time.Format(ISODateFmt))
	return []byte(formatted), nil
}

// UnmarshalParam converts an incoming date string, e.g. 2022-01-01, in a time.Time object
//
// The function is similar to UnmarshalJSON but this can be used in order for query parameters
// and form values: https://echo.labstack.com/guide/request/#custom-binder
func (d *ISODate) UnmarshalParam(p string) (err error) {
	t, err := time.Parse(ISODateFmt, p)
	if err != nil {
		return err
	}

	d.Time = t
	return
}

// Value returns the time.Time object of the ISODate struct (gorm)
func (d ISODate) Value() (driver.Value, error) {
	return d.Time, nil
}

// Scan assigns a value from a database driver (gorm)
func (d *ISODate) Scan(value interface{}) error {
	d.Time = value.(time.Time)
	return nil
}

// Date returns the wrapped time.Time object of the ISODate struct for interoperability with the
// [agg.Dated] aggregation methods.
func (d ISODate) Date() time.Time {
	return d.Time
}

// String returns the formatted representation of the wrapped time.Time struct as the ISO date
// format (YYYY-MM-DD).
func (d ISODate) String() string {
	return d.Time.Format(ISODateFmt)
}

// FromTime returns an ISODate for a given [time.Time] struct.
func FromTime(time time.Time) ISODate {
	return ISODate{time}
}

// FromStr returns an ISODate for a given string that is based on the ISODateFmt.
func FromStr(str string) (*ISODate, error) {
	dateVal, err := time.Parse(ISODateFmt, str)
	if err != nil {
		return nil, err
	}

	res := FromTime(dateVal)
	return &res, nil
}
