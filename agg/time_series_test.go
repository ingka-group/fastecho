package agg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ingka-group-digital/ocp-go-utils/date"
)

func Test_Year_Date(t *testing.T) {
	year := Year(2024)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), year.Date())
}

func Test_Year_String(t *testing.T) {
	year := Year(2024)
	assert.Equal(t, "2024", year.String())
}

func Test_YearMonth_Date(t *testing.T) {
	ym := YearMonth{Year: 2024, Month: 3}
	assert.Equal(t, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), ym.Date())
}

func Test_YearMonth_String(t *testing.T) {
	ym := YearMonth{Year: 2024, Month: 3}
	assert.Equal(t, "2024-03", ym.String())
}

func Test_IKEAWeek_Date(t *testing.T) {
	ym := IKEAWeek{Year: 2024, Week: 3}
	assert.Equal(t, time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), ym.Date())
}

func Test_IKEAWeek_String(t *testing.T) {
	ym := IKEAWeek{Year: 2024, Week: 3}
	assert.Equal(t, "2024-W03", ym.String())
}

type testData struct {
	Timestamp time.Time
	Value     int
}

func (td testData) Date() time.Time {
	return td.Timestamp
}

var sampleData = testData{Timestamp: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), Value: 20}

func Test_ByYear(t *testing.T) {
	year := ByYear(sampleData)
	assert.Equal(t, Year(2023), year)
}

func Test_ByYearMonth(t *testing.T) {
	ym := ByYearMonth(sampleData)
	assert.Equal(t, YearMonth{Year: 2023, Month: 12}, ym)
}

func Test_ByIKEAWeek(t *testing.T) {
	w := ByIKEAWeek(sampleData)
	assert.Equal(t, IKEAWeek{Year: 2024, Week: 1}, w)
}

func Test_ByISODate(t *testing.T) {
	w := ByISODate(sampleData)
	isoDate := date.FromTime(time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC))
	assert.Equal(t, isoDate, w)
}
