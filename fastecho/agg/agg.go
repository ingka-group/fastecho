package agg

import (
	"slices"

	"golang.org/x/exp/constraints"

	"github.com/ingka-group-digital/ocp-go-utils/fp"
)

// Numeric is a constraint that permits integers or floats.
type Numeric interface {
	constraints.Integer | constraints.Float
}

// Median returns the middle number (50th percentile) of the numeric `input` slice.
func Median[T Numeric](input []T) float64 {
	vals := make([]T, len(input))
	copy(vals, input)
	slices.Sort(vals)

	length := len(vals)
	if length == 0 {
		return 0
	} else if length%2 == 0 {
		return float64(vals[length/2-1]+vals[length/2]) / 2.0
	} else {
		return float64(vals[length/2])
	}
}

// Avg returns the average of the numeric `input` slice
func Avg[T Numeric](input []T) float64 {
	return float64(Sum(input)) / float64(len(input))
}

// Sum returns the sum of all values in the numeric `input` slice.
func Sum[T Numeric](input []T) T {
	return fp.Reduce(input, Add)
}

// Min returns the smallest value in the numeric `input` slice.
func Min[T Numeric](input []T) T {
	return fp.ReduceFrom(input,
		Identity,
		func(current T, memo T) T {
			if current < memo {
				return current
			}

			return memo
		})
}

// Max returns the largest value in the numeric `input` slice.
func Max[T Numeric](input []T) T {
	return fp.ReduceFrom(input,
		Identity,
		func(current T, memo T) T {
			if current > memo {
				return current
			}

			return memo
		})
}

// Add adds two numbers. Useful for simplifying functional code (see [Sum])
func Add[T Numeric](lhs T, rhs T) T {
	return lhs + rhs
}

// Identity passes through the same value as it receives. (And you thought nothing could be more
// exciting than [Add]!)
func Identity[T any](input T) T {
	return input
}
