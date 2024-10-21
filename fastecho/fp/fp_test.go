package fp_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"

	"github.com/ingka-group-digital/ocp-go-utils/agg"
	. "github.com/ingka-group-digital/ocp-go-utils/fp"
)

// -- Tests --

func Test_Any(t *testing.T) {
	input := []int{1, 2, 3}
	yes := Any(input, func(i int) bool { return i <= 2 })
	no := Any(input, func(i int) bool { return i > 100 })

	assert.True(t, yes)
	assert.False(t, no)
}

func Test_All(t *testing.T) {
	input := []int{1, 2, 3}
	yes := All(input, func(i int) bool { return i < 100 }) // => true
	no := All(input, func(i int) bool { return i == 1 })   // => false

	assert.True(t, yes)
	assert.False(t, no)
}

func Test_Collect(t *testing.T) {
	t.Run("simple conversion", func(t *testing.T) {
		input := map[string]string{
			"foo": "bar",
			"bar": "baz",
		}
		actual := Collect(input, func(_ string, v string) string { return v })
		// Using ElementsMatch instead of Equal as order is not guaranteed.
		assert.ElementsMatch(t, []string{"bar", "baz"}, actual)
	})

	t.Run("collect slice values", func(t *testing.T) {
		input := map[int][]benchStruct{
			10: {{N: 1}, {N: 2}},
			20: {{N: 3}, {N: 4}},
		}
		// Sum all the values of `n` for each element in the array, then multiply by the key.
		actual := Collect(input, func(k int, v []benchStruct) int {
			return k * Reduce(v, func(v benchStruct, memo int) int { return memo + int(v.N) })
		})
		// Using ElementsMatch instead of Equal as order is not guaranteed.
		assert.ElementsMatch(t, []int{30, 140}, actual)
	})
}

func Test_Compact(t *testing.T) {
	t.Run("test with primitive", func(t *testing.T) {
		input := []int{0, 0, 1, 2, 0, 3}
		assert.Equal(t, []int{1, 2, 3}, Compact(input))
	})

	t.Run("test with struct", func(t *testing.T) {
		input := []benchStruct{{N: 1}, {}, {N: 2}}
		assert.Equal(t, []benchStruct{{N: 1}, {N: 2}}, Compact(input))
	})
}

func Test_Filter(t *testing.T) {
	tests := []struct {
		name     string
		slice    []float64
		function func(f float64) bool
		expected []float64
	}{
		{
			name:  "empty slices do not get modified",
			slice: []float64{},
			function: func(f float64) bool {
				return true
			},
			expected: []float64{},
		},
		{
			name:  "filter method gets called",
			slice: []float64{0, 0, 3, 0, 1, 2, 0},
			function: func(f float64) bool {
				return f != 0
			},
			expected: []float64{3, 1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Filter(tt.slice, tt.function))
		})
	}
}

func Test_Flatten(t *testing.T) {
	t.Run("test flatten 2D array", func(t *testing.T) {
		input := [][]int{{1, 4}, {2, 5}, {3, 6}}
		output := Flatten(input)
		assert.Equal(t, []int{1, 4, 2, 5, 3, 6}, output)
	})

	t.Run("test flatten 3D array", func(t *testing.T) {
		input := [][][]int{{{1, 4}, {2, 5}}, {{3, 6}}}
		output := Flatten(input)
		assert.Equal(t, [][]int{{1, 4}, {2, 5}, {3, 6}}, output)
	})
}

func Test_GroupBy(t *testing.T) {
	t.Run("test group by primitive value", func(t *testing.T) {
		input := []int{-2, -1, 0, 1, 2, 3}
		output := GroupBy(input, func(i int) bool {
			return i > 0
		})
		assert.Equal(t, map[bool][]int{false: {-2, -1, 0}, true: {1, 2, 3}}, output)
	})

	t.Run("test group by complex value", func(t *testing.T) {
		// convenience variables
		jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		jan2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		dec31 := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
		december := month{year: 2023, month: 12}
		january := month{year: 2024, month: 1}

		// actual test
		input := []groupTest{
			{date: jan1, value: 100},
			{date: dec31, value: 1000},
			{date: jan2, value: 500},
		}

		actual := GroupBy(input, func(i groupTest) month {
			return monthFromDate(i.date)
		})
		expected := map[month][]groupTest{
			december: {{date: dec31, value: 1000}},
			january:  {{date: jan1, value: 100}, {date: jan2, value: 500}},
		}

		assert.Equal(t, expected, actual)
	})
}

// Test_Example tests the example from the README with mock data.
func Test_Example(t *testing.T) {
	type City struct {
		Name        string
		Temperature float64
	}
	cities := []City{
		{Name: "Amsterdam", Temperature: 8.0},
		{Name: "Rome", Temperature: 14.5},
		{Name: "Athens", Temperature: 21.5},
		{Name: "Stockholm", Temperature: 2.0},
		{Name: "Trondheim", Temperature: -4.5},
		{Name: "San Francisco", Temperature: 18.0},
	}

	var Add = func(i float64, m float64) float64 { return i + m }
	var SumSlice = func(vals []float64) float64 { return Reduce(vals, Add) }
	var AverageSlice = func(t []float64) float64 { return SumSlice(t) / float64(len(t)) }

	var GetTemperature = func(c City) float64 { return c.Temperature }
	var IsCityWarm = func(c City) bool { return GetTemperature(c) > 10 }

	var GetWarmCities = func(c []City) []City { return Filter(c, IsCityWarm) }
	var GetAllTemperatures = func(c []City) []float64 { return Map(c, GetTemperature) }
	var AverageTemperature = func(c []City) float64 { return AverageSlice(GetAllTemperatures(c)) }

	warmCities := GetWarmCities(cities)

	assert.Equal(t, 6, len(cities))
	assert.Equal(t, 3, len(warmCities))
	assert.Equal(t, 10.0, math.Round(AverageTemperature(cities)))
	assert.Equal(t, 18.0, AverageTemperature(warmCities))
}

func Test_MapI(t *testing.T) {
	t.Run("test map into same type", func(t *testing.T) {
		input := []int64{1, 2, 3, 4}
		output := MapI(input, func(_ int, val int64) int64 {
			return val * val
		})
		assert.Equal(t, []int64{1, 4, 9, 16}, output)
	})

	t.Run("test map into different type", func(t *testing.T) {
		input := []int64{1, 2, 3, 4}
		output := MapI(input, func(_ int, val int64) string {
			return "meow"
		})
		assert.Equal(t, []string{"meow", "meow", "meow", "meow"}, output)
	})

	t.Run("test extract from struct", func(t *testing.T) {
		type test struct {
			num int64
		}
		input := []test{
			{num: 1}, {num: 2}, {num: 3},
		}
		output := MapI(input, func(_ int, val test) int64 {
			return val.num
		})

		assert.Equal(t, []int64{1, 2, 3}, output)
	})
}

func Test_Map(t *testing.T) {
	t.Run("test map into same type", func(t *testing.T) {
		input := []int64{1, 2, 3, 4}
		output := Map(input, func(val int64) int64 {
			return val * val
		})
		assert.Equal(t, []int64{1, 4, 9, 16}, output)
	})

	t.Run("test map into different type", func(t *testing.T) {
		input := []int64{1, 2, 3, 4}
		output := Map(input, func(val int64) string {
			return "meow"
		})
		assert.Equal(t, []string{"meow", "meow", "meow", "meow"}, output)
	})

	t.Run("test extract from struct", func(t *testing.T) {
		type test struct {
			num int64
		}
		input := []test{
			{num: 1}, {num: 2}, {num: 3},
		}
		output := Map(input, func(val test) int64 {
			return val.num
		})

		assert.Equal(t, []int64{1, 2, 3}, output)
	})
}

func Test_MapHash(t *testing.T) {
	t.Run("test map hash", func(t *testing.T) {
		input := map[string]int{"foo": 1, "bar": 2, "baz": 3, "qux": 4}
		expected := map[string]int{"foo": 1, "bar": 4, "baz": 9, "qux": 16}

		output := MapHash(input, func(v int) int {
			return v * v
		})
		assert.Equal(t, expected, output)
	})
}

func Test_MapHashK(t *testing.T) {
	t.Run("test map into same type", func(t *testing.T) {
		input := map[string]int{"foo": 1, "bar": 2, "baz": 3, "qux": 4}
		expected := map[string]int{"foo": 1, "bar": 4, "baz": 9, "qux": 16}

		output := MapHashK(input, func(k string, v int) (string, int) {
			return k, v * v
		})
		assert.Equal(t, expected, output)
	})

	t.Run("test map into different type", func(t *testing.T) {
		input := map[string]int{"foo": 1, "bar": 2, "baz": 3, "qux": 4}
		expected := map[int]string{1: "meow", 2: "meow", 3: "meow", 4: "meow"}

		output := MapHashK(input, func(_ string, v int) (int, string) {
			return v, "meow"
		})
		assert.Equal(t, expected, output)
	})
}

func Test_Reduce(t *testing.T) {
	t.Run("test reduce to sum", func(t *testing.T) {
		input := []int{1, 2, 3, 4}
		output := Reduce(input, func(v int, memo int) int {
			return v + memo // `memo` is 0 at the start
		})
		assert.Equal(t, 10, output)
	})

	t.Run("test Reduce with empty array", func(t *testing.T) {
		input := []int{}
		output := Reduce(input, func(_ int, _ int) int {
			return 999 // this should never be called
		})
		assert.Equal(t, 0, output) // memo was initialized to zero value and passed through
	})
}

func Test_ReduceWith(t *testing.T) {
	t.Run("test reduce to product", func(t *testing.T) {
		input := []int{1, 2, 3, 4}
		output := ReduceWith(input, 1, func(v int, memo int) int {
			return v * memo // `memo` is set to 1 at the start
		})
		assert.Equal(t, 24, output)
	})

	t.Run("test ReduceWith with empty array", func(t *testing.T) {
		input := []int{}
		output := ReduceWith(input, 1, func(_ int, _ int) int {
			return 999 // this should never be called
		})
		assert.Equal(t, 1, output) // memo was initialized to default value and passed through
	})
}

func Test_ReduceFrom(t *testing.T) {
	t.Run("test reduce to min", func(t *testing.T) {
		input := []int{1, 2, 3, 4, -1}
		output := ReduceFrom(input,
			func(i int) int { return i }, // `memo` is set to the first entry in the array
			func(v int, memo int) int {
				if v < memo {
					return v
				}
				return memo
			},
		)
		assert.Equal(t, -1, output)
	})

	t.Run("test reduce to max", func(t *testing.T) {
		input := []int{1, 2, 3, 4, -1}
		output := ReduceFrom(input,
			func(i int) int { return i },
			func(v int, memo int) int {
				if v > memo {
					return v
				}
				return memo
			},
		)
		assert.Equal(t, 4, output)
	})

	t.Run("test ReduceFrom with empty array", func(t *testing.T) {
		input := []int{}
		output := ReduceFrom(input,
			func(_ int) int {
				t.Fail() // this should never be called
				return 999
			},
			func(_ int, _ int) int {
				t.Fail() // this should never be called either
				return 999
			},
		)
		assert.Equal(t, 0, output) // memo was initialized to zero value and passed through
	})

	t.Run("test ReduceFrom with one value", func(t *testing.T) {
		input := []int{1}
		output := ReduceFrom(input,
			func(i int) int {
				return i
			},
			func(_ int, _ int) int {
				t.Fail() // this should never be called
				return 999
			},
		)
		assert.Equal(t, 1, output)
	})

	t.Run("test ReduceFrom only visits each value once", func(t *testing.T) {
		input := []int{1, 2, 3}
		output := ReduceFrom(input,
			func(i int) int {
				return i
			},
			func(i int, m int) int {
				return i + m
			},
		)
		assert.Equal(t, 6, output)
	})
}

func Test_Zip(t *testing.T) {
	t.Run("test zip with arrays of same length", func(t *testing.T) {
		lhs := []int{1, 2, 3}
		rhs := []int{4, 5, 6}

		output := Zip(lhs, rhs)
		assert.Equal(t, [][]int{{1, 4}, {2, 5}, {3, 6}}, output)
	})

	t.Run("test zip, lhs shorter", func(t *testing.T) {
		lhs := []int{1, 2}
		rhs := []int{4, 5, 6}

		output := Zip(lhs, rhs)
		assert.Equal(t, [][]int{{1, 4}, {2, 5}}, output)
	})

	t.Run("test zip, rhs shorter", func(t *testing.T) {
		lhs := []int{1, 2, 3}
		rhs := []int{4, 5}

		output := Zip(lhs, rhs)
		assert.Equal(t, [][]int{{1, 4}, {2, 5}}, output)
	})
}

// -- Benchmarks and supporting code --

type month struct {
	year  int
	month int
}

func (m month) Less(lhs, rhs month) bool {
	return (lhs.year < rhs.year && lhs.month < rhs.month)
}

func monthFromDate(date time.Time) month {
	return month{year: date.Year(), month: int(date.Month())}
}

type groupTest struct {
	date  time.Time
	value int
}

var filterBenchmarkData = func() []float64 {
	data := make([]float64, 1000) // return 1000 floats
	return Map(data, func(_ float64) float64 {
		return rand.Float64() * 5000 // return floats in range from 0.0 to 5000.0
	})
}()

type benchStruct struct {
	N float64 `json:"n"`
}

var mapBenchmarkData = func() []benchStruct {
	return Map(filterBenchmarkData, func(n float64) benchStruct {
		return benchStruct{N: n}
	})
}()

var testFilter = func(n float64) bool {
	return n != 0
}

var testMap = func(v benchStruct) float64 {
	return v.N
}

// Benchmark_Flatten_Append benchmarks an alternative implementation of Flatten.
func Benchmark_Flatten_Append(b *testing.B) {
	test := [][]int{{1, 4}, {2, 5}, {3, 6}}

	for n := 0; n < b.N; n++ {
		out := make([]int, 0)
		for _, v1 := range test {
			// disable SA4010 for next line as not using the result of `out` is intentional in the
			// benchmark
			//nolint:staticcheck
			out = append(out, v1...)
		}
	}

}

// Benchmark_Flatten_PrecalculateLen benchmarks the actual implementation of Flatten.
func Benchmark_Flatten_PrecalculateLen(b *testing.B) {
	test := [][]int{{1, 4}, {2, 5}, {3, 6}}

	// actual implementation
	for n := 0; n < b.N; n++ {
		Flatten(test)
	}
}

// Benchmark_MedianFilter verifies that it is actually easier/faster to do this in one place than to
// introduce this calculation to each connector.
//
// Result: 17022 ns/op (it takes 0.02ms to filter and sort 1000 random float64s.
func Benchmark_MedianFilter(b *testing.B) {
	for n := 0; n < b.N; n++ {
		agg.Median(Filter(filterBenchmarkData, testFilter))
	}
}

// Benchmark_Map verifies the speed of extracting a scalar from a struct.
// Turns out we can do this pretty fast (1844 ns/op for an array of 1000 floats, i.e. we can do this
// 625336 times a second).
func Benchmark_Map(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Map(mapBenchmarkData, testMap)
	}
}

// Benchmark_GroupBy verifies the speed of grouping an array of structs into a map through a single
// if clause in which we group by a scalar. This is the fastest grouping case (11941 ns to group
// 1000 floats).
func Benchmark_GroupBy_Scalar(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GroupBy(mapBenchmarkData, func(s benchStruct) int {
			if s.N <= 1000 {
				return 1
			} else if s.N >= 1000 && s.N <= 4000 {
				return 2
			} else {
				return 3
			}
		})
	}
}

// Benchmark_GroupBy verifies the speed of grouping an array of structs into a map through a single
// if clause in which we group by a string. This is somewhat slower as we allocate a string in each
// iteration (14559 ns to group 1000 floats).
func Benchmark_GroupBy_String(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GroupBy(mapBenchmarkData, func(s benchStruct) string {
			if s.N <= 1000 {
				return "<= 1000"
			} else if s.N >= 1000 && s.N <= 4000 {
				return "1001 - 4000"
			} else {
				return ">= 4000"
			}
		})
	}
}

// Benchmark_GroupBy_Struct verifies the speed of grouping an array of structs into a map through a
// single if clause in which we group by a struct. This is the slowest case as we instantiate a
// struct on each pass, but we still get a respectable score of 24965 ns (0.02ms) to group 1000
// floats.
func Benchmark_GroupBy_Struct(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GroupBy(mapBenchmarkData, func(s benchStruct) time.Time {
			if s.N <= 1000 {
				return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			} else if s.N >= 1000 && s.N <= 4000 {
				return time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
			} else {
				return time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
			}
		})
	}
}

var collectBenchmarkData = GroupBy(mapBenchmarkData, func(s benchStruct) string {
	if s.N <= 1000 {
		return "<= 1000"
	} else if s.N >= 1000 && s.N <= 4000 {
		return "1000 - 4000"
	} else {
		return "> 4000"
	}
})

// Benchmark_Collect verifies the speed of converting a map into an array through a function that
// combines Map and Median. This is representative of a portion of our real code in which we use
// this to get medians for our actuals. Result: 12634 ns/op
func Benchmark_Collect(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Collect(collectBenchmarkData, func(_ string, s []benchStruct) float64 {
			return agg.Median(Map(s, testMap))
		})
	}
}

// Benchmark_GroupBy_Collect verifies the speed of grouping an array of structs into distinct
// buckets by a string, then converting that map into an aggregation array through a function that
// combines Map and Median.
// This is the synthetic benchmark most representative of the full processing pipeline.
// Result: 41025 ns/op (so, safe to assume the amount of time to process the data will be a fraction
// of the time it takes to fetch it from another microservice)
func Benchmark_GroupBy_Collect(b *testing.B) {
	for n := 0; n < b.N; n++ {
		grouped := GroupBy(mapBenchmarkData, func(s benchStruct) string {
			if s.N <= 1000 {
				return "<= 1000"
			} else if s.N >= 1000 && s.N <= 4000 {
				return "1000 - 4000"
			} else {
				return "> 4000"
			}
		})

		Collect(grouped, func(_ string, s []benchStruct) float64 {
			return agg.Median(Map(s, testMap)) * agg.Median(Map(s, testMap))
		})
	}
}

// Benchmark_SerializeDeserialize is a comparison benchmark to see how fast are these operations
// comparing to one serializing and deserializing roundtrip (what we'd expect as a bare minimum from
// offloading this to a separate microservice.
func Benchmark_SerializeDeserialize(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ser, _ := json.Marshal(mapBenchmarkData)
		var de []benchStruct
		_ = json.Unmarshal(ser, &de)
	}
}
