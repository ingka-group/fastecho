package date

import (
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
