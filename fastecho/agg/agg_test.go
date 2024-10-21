package agg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Median(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		expected float64
	}{
		{
			name:     "empty slice returns zero",
			slice:    []int{},
			expected: 0,
		},
		{
			name:     "single element returns itself",
			slice:    []int{1},
			expected: 1,
		},
		{
			name:     "even-sized arrays return average of two middle values",
			slice:    []int{-50, 2, 3, 200},
			expected: 2.5,
		},
		{
			name:     "odd-sized arrays return the middle value",
			slice:    []int{3, 9, 10, 100, 500},
			expected: 10,
		},
		{
			name:     "unsorted arrays get sorted correctly",
			slice:    []int{100, 500, 10, 3, 9},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Median(tt.slice))
		})
	}
}

func Test_Avg(t *testing.T) {
	nums := []int{1, 2, 3, 100}
	assert.Equal(t, 26.5, Avg(nums))
}

func Test_Sum(t *testing.T) {
	nums := []int{1, 2, 3, 100}
	assert.Equal(t, 106, Sum(nums))
}

func Test_Add(t *testing.T) {
	// well, hopefully there's no need to test this thoroughly :P
	assert.Equal(t, 2, Add(1, 1))
}

func Test_Min(t *testing.T) {
	nums := []int{1, 2, 3, -100}
	assert.Equal(t, -100, Min(nums))
}

func Test_Max(t *testing.T) {
	nums := []int{1, 2, 3, -100}
	assert.Equal(t, 3, Max(nums))
}

func Test_Identity(t *testing.T) {
	assert.Equal(t, true, Identity(true))
}
