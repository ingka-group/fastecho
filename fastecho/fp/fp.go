// Package fp  provides common functional programming tools such as `Map`, `Reduce` and `Filter`.
// These simple functions can be composed together to form powerful and expressive data processing
// pipelines in a relatively small amount of code.
package fp

// All returns true if the provided callback returns true for all values in the slice.
//
// The function signature of the callback is `func(value T) bool`, where T is the type of the slice.
func All[T any](input []T, evaluator func(T) bool) bool {
	for _, v := range input {
		if !evaluator(v) {
			return false
		}
	}

	return true
}

// Any returns true if the provided callback returns true for at least one value in the slice.
//
// The function signature of the callback is `func(value T) bool`, where T is the type of the slice.
func Any[T any](input []T, evaluator func(T) bool) bool {
	for _, v := range input {
		if evaluator(v) {
			return true
		}
	}

	return false
}

// Collect converts a map into a slice by running the callback on each element of the map. The
// result is not guaranteed to be in any specific, or consistent, order.
//
// The first argument of the callback is the key of the current map element and the second argument
// is the value of that element.
func Collect[K comparable, V any, T any](input map[K]V, callback func(K, V) T) []T {
	out := make([]T, len(input))
	var i int
	for k, v := range input {
		out[i] = callback(k, v)
		i += 1
	}
	return out
}

// Compact returns a new slice filtered from all "zero values" of the slice's type.
func Compact[T comparable](input []T) []T {
	var zeroval T
	return Filter(input, func(i T) bool { return i != zeroval })
}

// Flatten returns a new slice that removes one layer of nesting in a slice. In other words, it will
// turn a three-dimensional slice into a two-dimensional slice, or a two-dimensional slice into a
// one-dimensional slice.
func Flatten[T any](input [][]T) []T {
	// Calculate the length of the output slice so that we can pre-allocate it instead of
	// re-allocating each time we append. This increases the speed by ~15% (see benchmarks).
	l := Reduce(Map(input,
		func(v []T) int { return len(v) }),
		func(i int, m int) int { return i + m },
	)

	out := make([]T, l)
	i := 0

	for _, v1 := range input {
		for _, v2 := range v1 {
			out[i] = v2
			i += 1
		}
	}

	return out
}

// Filter returns a new slice containing all elements of `input` for which the provided `filter`
// function returns a true value.
//
// The function signature of the callback is `func(value T) bool`, where T is the type of the slice.
func Filter[T any](input []T, filter func(T) bool) []T {
	copy := make([]T, len(input))
	n := 0
	for _, v := range input {
		if filter(v) {
			copy[n] = v
			n++
		}
	}
	return copy[:n]
}

// GroupBy groups (partitions) the input into a map by the result of the grouper.
func GroupBy[T any, G comparable](input []T, grouper func(T) G) map[G][]T {
	grouped := make(map[G][]T)

	for _, v := range input {
		group := grouper(v)
		grouped[group] = append(grouped[group], v)
	}

	return grouped
}

// Map returns a new slice with the results of running the callback once for every element in the
// slice.
//
// The function signature of the callback is `func(value T) V`, where `T` is the type of
// the slice and `V` is your return type.
func Map[T any, V any](input []T, callback func(T) V) []V {
	copy := make([]V, len(input))
	for i, v := range input {
		copy[i] = callback(v)
	}
	return copy
}

// MapHash is the equivalent of Map for Go maps (dictionary, hash-map).
func MapHash[K comparable, V any, T any](input map[K]V, callback func(V) T) map[K]T {
	out := make(map[K]T)
	for k, v := range input {
		out[k] = callback(v)
	}
	return out
}

// MapHashK is the equivalent of MapI for Go maps (dictionary, hash-map). It allows you to alter
// keys as well as values.
func MapHashK[K comparable, V any, NK comparable, NV any](input map[K]V, callback func(K, V) (NK, NV)) map[NK]NV {
	out := make(map[NK]NV)
	for k, v := range input {
		nk, nv := callback(k, v)
		out[nk] = nv
	}
	return out
}

// MapI is the same as [Map], except the callback also provides the current index.
//
// The function signature of the callback is `func(index int, value T) V`, where `T` is the type of
// the slice and `V` is your return type.
func MapI[T any, V any](input []T, callback func(int, T) V) []V {
	copy := make([]V, len(input))
	for i, v := range input {
		copy[i] = callback(i, v)
	}
	return copy
}

// Reduce combines all elements of input by applying a binary operation specified by the reducer.
//
// The value of `memo` is uninitialized at the first iteration (i.e., it is set to Go's magic "zero
// value".)
func Reduce[T any, V any](input []T, reducer func(T, V) V) V {
	var memo V
	return ReduceWith(input, memo, reducer)
}

// ReduceFrom combines all elements of input by applying a binary operation specified by the
// reducer, starting from the value returned by the `initial` function.
//
// The value of `memo` is initialized to the result of the `initial` function, which will receive
// the first element of the `input` array.
//
// If there are no elements in the `input` array, `initial` will not be called.
func ReduceFrom[T any, V any](input []T, initial func(T) V, reducer func(T, V) V) V {
	if len(input) == 0 {
		var memo V
		return memo
	}

	memo := initial(input[0])

	if len(input) == 1 {
		return memo
	}

	return ReduceWith(input[1:], memo, reducer)
}

// ReduceWith combines all elements of input by applying a binary operation specified by the
// reducer, starting from a known value.
//
// The value of `memo` is initialized to the provided value at the first iteration.
func ReduceWith[T any, V any](input []T, memo V, reducer func(T, V) V) V {
	for _, v := range input {
		memo = reducer(v, memo)
	}
	return memo
}

// Zip merges the `lhs` and `rhs` slices by returning slices in which the first element comes from
// the first slice, and the second element from the second slice. The merge stops once either of the
// slices is exhausted.
func Zip[T any](lhs []T, rhs []T) [][]T {
	n := min(len(lhs), len(rhs))
	out := make([][]T, n)

	for i := 0; i < n; i++ {
		out[i] = []T{lhs[i], rhs[i]}
	}

	return out
}
