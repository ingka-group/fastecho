package agg

import (
	"testing"
	"time"

	"github.com/ingka-group-digital/ocp-go-utils/date"

	"github.com/stretchr/testify/assert"

	_ "github.com/ingka-group-digital/ocp-go-utils/fp"
)

func Test_Year_Date(t *testing.T) {
	year := Year(2024)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), year.Date())
}

func Test_Year_String(t *testing.T) {
	year := Year(2024)
	assert.Equal(t, "2024", year.String())
}

func Test_Year_Add(t *testing.T) {
	year := Year(2024)
	assert.Equal(t, year.Add(1), Year(2025))
}

func Test_Year_Start(t *testing.T) {
	year := Year(2024)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), year.Start())
}

func Test_Year_End(t *testing.T) {
	year := Year(2024)
	assert.Equal(t, time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), year.End())
}

func Test_YearMonth_Date(t *testing.T) {
	ym := YearMonth{Year: 2024, Month: 3}
	assert.Equal(t, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), ym.Date())
}

func Test_YearMonth_String(t *testing.T) {
	ym := YearMonth{Year: 2024, Month: 3}
	assert.Equal(t, "2024-03", ym.String())
}

func TestYearMonth_Add(t *testing.T) {
	tests := []struct {
		name   string
		year   int
		month  int
		amount int
		want   Interval
	}{
		{
			name:   "add month in same year",
			year:   2024,
			month:  6,
			amount: 1,
			want:   YearMonth{Year: 2024, Month: 7},
		},
		{
			name:   "subtract month in same year",
			year:   2024,
			month:  6,
			amount: -1,
			want:   YearMonth{Year: 2024, Month: 5},
		},
		{
			name:   "add months across years",
			year:   2024,
			month:  6,
			amount: 10,
			want:   YearMonth{Year: 2025, Month: 4},
		},
		{
			name:   "subtract months across years",
			year:   2024,
			month:  6,
			amount: -10,
			want:   YearMonth{Year: 2023, Month: 8},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ym := YearMonth{
				Year:  tt.year,
				Month: tt.month,
			}
			assert.Equal(t, tt.want, ym.Add(tt.amount))
		})
	}
}

func Test_YearMonth_Start(t *testing.T) {
	ym := YearMonth{Year: 2024, Month: 3}
	assert.Equal(t, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), ym.Start())
}

func Test_YearMonth_End(t *testing.T) {
	t.Run("31 days", func(t *testing.T) {
		ym := YearMonth{Year: 2024, Month: 3}
		assert.Equal(t, time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC), ym.End())
	})

	t.Run("30 days", func(t *testing.T) {
		ym := YearMonth{Year: 2024, Month: 4}
		assert.Equal(t, time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC), ym.End())
	})

	t.Run("29 days", func(t *testing.T) {
		ym := YearMonth{Year: 2024, Month: 2}
		assert.Equal(t, time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), ym.End())
	})

	t.Run("28 days", func(t *testing.T) {
		ym := YearMonth{Year: 2023, Month: 2}
		assert.Equal(t, time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC), ym.End())
	})
}

func Test_IKEAWeek_Date(t *testing.T) {
	ym := IKEAWeek{Year: 2024, Week: 3}
	assert.Equal(t, time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), ym.Date())
}

func Test_IKEAWeek_String(t *testing.T) {
	ym := IKEAWeek{Year: 2024, Week: 3}
	assert.Equal(t, "2024-W03", ym.String())
}

func TestIKEAWeek_Add(t *testing.T) {
	tests := []struct {
		name   string
		year   int
		week   int
		amount int
		want   Interval
	}{
		{
			name:   "add week in same year",
			year:   2024,
			week:   6,
			amount: 5,
			want:   IKEAWeek{Year: 2024, Week: 11},
		},
		{
			name:   "subtract week in same year",
			year:   2024,
			week:   6,
			amount: -5,
			want:   IKEAWeek{Year: 2024, Week: 1},
		},
		{
			name:   "add weeks across years",
			year:   2024,
			week:   6,
			amount: 80,
			want:   IKEAWeek{Year: 2025, Week: 34},
		},
		{
			name:   "subtract weeks across years",
			year:   2024,
			week:   6,
			amount: -10,
			want:   IKEAWeek{Year: 2023, Week: 48},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := IKEAWeek{
				Year: tt.year,
				Week: tt.week,
			}
			assert.Equal(t, tt.want, w.Add(tt.amount))
		})
	}
}

func Test_IKEAWeek_Start(t *testing.T) {
	t.Run("same month", func(t *testing.T) {
		ym := IKEAWeek{Year: 2024, Week: 3}
		assert.Equal(t, time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), ym.Start())
	})

	t.Run("across months", func(t *testing.T) {
		ym := IKEAWeek{Year: 2024, Week: 14}
		assert.Equal(t, time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC), ym.Start())
	})

	t.Run("end of year", func(t *testing.T) {
		ym := IKEAWeek{Year: 2024, Week: 52}
		assert.Equal(t, time.Date(2024, 12, 22, 0, 0, 0, 0, time.UTC), ym.Start())
	})

	t.Run("across years", func(t *testing.T) {
		ym := IKEAWeek{Year: 2025, Week: 1}
		assert.Equal(t, time.Date(2024, 12, 29, 0, 0, 0, 0, time.UTC), ym.Start())
	})

}

func Test_IKEAWeek_End(t *testing.T) {
	t.Run("same month", func(t *testing.T) {
		ym := IKEAWeek{Year: 2024, Week: 3}
		assert.Equal(t, time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), ym.End())
	})

	t.Run("across months", func(t *testing.T) {
		ym := IKEAWeek{Year: 2024, Week: 14}
		assert.Equal(t, time.Date(2024, 4, 6, 0, 0, 0, 0, time.UTC), ym.End())
	})

	t.Run("across years", func(t *testing.T) {
		ym := IKEAWeek{Year: 2021, Week: 52}
		assert.Equal(t, time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC), ym.End())
	})
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
