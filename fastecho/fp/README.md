# fp

The `fp` package provides common functional programming tools such as `Map`, `Reduce` and `Filter`.
These simple functions can be composed together to form powerful and expressive data processing
pipelines in a relatively small amount of code.

> [!TIP]
> For a gentle introduction to these topics, I suggest [Reading 25: Map, Filter,
> Reduce](https://ocw.mit.edu/ans7870/6/6.005/s16/classes/25-map-filter-reduce/) from the Software
> Construction class on MIT.

The minimum supported Go version is **1.22**.

If you're still not sold on the concept, read on - otherwise, you can skip directly to
[Usage](#Usage) for the documentation.

## Why?

All the functions provided here can be easily replicated with traditional `for` loops. The benefits
of using these methods are in writing more concise code with less mental overhead, enabling reuse of
smaller components to build up to a larger data processing pipeline, and, in the future, enabling
parallelization (see [Iterators](#Iterators)).

For example, these two code snippets return the same result (the imperative version is how `Filter`
is implemented "under the hood"):

```go
nums := []int{-2, -1, 0, 1, 2}

// imperative
filtered := make([]int, len(nums))
var i int // prevent expensive reallocations on each pass
for _, n := range nums {
    if n > 0 { // <- this is where our "business logic" lives
        filtered[i] = n
        i += 1
    }
}
filtered = filtered[:i]

// functional
filtered = Filter(nums, func(n int) bool { return n > 0 })
```

However, in the imperative example, our "business logic" -- checking whether a number is bigger than
zero -- is buried in between performance optimizations, assignments and other busywork. Meanwhile,
in the functional example, the logic is front and center, and the result is ready to be chained into
another function.

Another benefit of the functional style is that we can extract the logic into a variable and reuse
it throughout our program.

```go
var GreaterThanZero = func(n int) bool {
    return n > 0
}

filtered = Filter(nums, GreaterThanZero)
```

And that makes it easy to build small blocks that combine together into something greater.

```go
type City struct { Name string; Temperature float64 }

var Add = func(i float64, m float64) float64 { return i + m }
var SumSlice = func(vals []float64) float64 { return Reduce(vals, Add) }
var AverageSlice = func(t []float64) float64 { return SumSlice(t) / float64(len(t)) }

var GetTemperature = func(c City) float64 { return c.Temperature }
var IsCityWarm = func(c City) bool { return GetTemperature(c) > 10 }

var GetWarmCities = func(c []City) []City { return Filter(c, IsCityWarm) }
var GetAllTemperatures = func(c []City) []float64 { return Map(c, GetTemperature) }
var AverageTemperature = func(c []City) float64 { return AverageSlice(GetAllTemperatures(c)) }

var cities []City // pretend these come from an external source or a mock
warmCities := GetWarmCities(cities)
fmt.Printf("Out of %d cities, %d are warm!\n", len(cities), len(warmCities))
fmt.Printf("Average temperature in all cities is %.2f 🥶\n", AverageTemperature(cities))
fmt.Printf("And in the warm cities, the average temperature is %.2f 😎\n", AverageTemperature(warmCities))
```

### Iterators

In a future version of Go, it [will be possible to build
iterators](https://go.dev/wiki/RangefuncExperiment), meaning that it will be possible to lazily
evaluate the results as needed, as well as simplify the function signatures to accept any iterable
value. This should also greatly improve the performance, especially when not consuming the entire
iterator, and memory use of these functions. You can
learn more about Go's planned iterator support [in this blog
post](https://bitfieldconsulting.com/golang/iterators).

### Verbosity

There are two main readability issues with this code: the inability to chain the operations (e.g.
`foo.Filter(...).Map(...)`, and the verbose closure declaration (`func(t []T) V { return ... }`).

Chainable operations are implemented in certain packages that provide these functions, but since Go
lacks the ability to take additional type arguments in methods, it is currently not possible to
implement both chainable methods _and_ the possibility of functions reducing different types. The Go
team [has their
reasoning](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#No-parameterized-methods).

Simplifying closure declarations to something more readable (e.g. JavaScript style `(foo) => foo +
1` or Rust style `|foo| foo + 1`, with automatically inferred types) has been an open proposal
[since 2017](https://github.com/golang/go/issues/21498), and while the discussion is still active,
it seems like making closures more readable is is not a priority for the Go team because the
community does not use closures as much, and the community does not use closures as much because it
is not a priority for the Go team to make them more readable.

### Performance

The functions provided in this package are developed using Go generics. A lot of care has been put
in to make them as performant as possible, and the provided benchmarks should provide a reference
for their speed. However, in contrast to practically all other compiled languages, the Go compiler
developers [chose to implement generics using runtime
maps](https://planetscale.com/blog/generics-can-make-your-go-code-slower), which trades faster
compile times for slower run times. In addition, all the operations are eagerly evaluated (at least
until iterators land in stable Go), which might hurt some use cases in which you do not want to
consume the entire iterator. If you absolutely need peak performance because you are dealing with
very large amounts of data consider using an imperative style, or a different language.

## Usage

All methods in this package are non-destructive, i.e. they will allocate a new data structure and
copy the values, leaving the old data structure unchanged. While mutating the original data
structure would be faster, it might lead to unexpected results and race conditions in concurrent
code.

### All

`All` returns `true` if the provided callback returns `true` for all values in the slice.

```go
input := []int{1, 2, 3}
All(input, func(i int) bool { return i < 100 }) // => true
All(input, func(i int) bool { return i == 1 }) // => false
```

### Any

`Any` returns `true` if the provided callback returns `true` for at least one value in the slice.

```go
input := []int{1, 2, 3}
Any(input, func(i int) bool { return i <= 2 }) // => true
Any(input, func(i int) bool { return i > 100 }) // => false
```

### Collect

`Collect` converts a Go map into a slice by calling the callback on each element of the map, then
collecting the return values of the callback into a slice.

Due to how maps in Go are implemented, the resulting slice is not guaranteed to be in any specific
or consistent order. If you need to guarantee a specific order, sort the slice afterwards.

The function signature of the callback is `func(key K, value V) T`, where `K` and `V` are key and
value types, and T is the value of the resulting slice.

```go
dict := map[string]string{
    "foo": "bar",
    "bar": "baz",
}
vals := Collect(input, func(_ string, v string) string {
    return v
}) // => either []string{"bar", "baz"} or []string{"baz", "bar"}
   // depending on how the Go runtime is feeling

```

### Compact

`Compact` returns a new slice filtered from all "zero values" of the slice's type.

Running `Compact` is the equivalent of running `var blank T; Filter(input, func(i T) bool { return i
!= blank })`. The result of `Compact` thus heavily depends on the semantics of the "zero value" for
your type, and might not be appropriate in all situations:

* for integers, `Compact` will filter out all zeroes,
* for strings, `Compact` will filter out all empty string,
* for bools, `Compact` will filter out all `false` values,
* for structs, `Compact` will filter out all values of struct `S` that can be representeda s `S{}`
  (i.e., with no initialized values).

In addition, `Compact` is constrained to `comparable` values.

```go
input := []int{0, 0, 1, 2, 0, 3}

Compact(input) // => []int{1, 2, 3}
```

### Flatten

`Flatten` returns a new slice that removes one layer of nesting in a slice. In other words, it will
turn  a two-dimensional slice into a one-dimensional slice, a three-dimensional slice into a
two-dimensional slice, and so on.

```go
input := [][]int{{1, 2}, {3, 4}, {5, 6}}
Flatten(input) // => []int{1, 2, 3, 4, 5, 6}
```

### Filter

`Filter` returns a new slice containing all elements of the `input` slice for which the provided
callback returns `true`.

The function signature of the callback is `func(value T) bool`, where T is the type of the slice.

```go
nums := []int{1, 2, 3, 4}
even := Filter(nums, func(i int) bool {
    return i % 2 == 0
}) // => []int{2, 4}
```

### GroupBy

`GroupBy` groups the input into a map by the result of the grouper.

Performance note: GroupBy will re-allocate each element of the input slice. This might be costly
for large objects or large arrays.

```go
input := []int{-2, -1, 0, 1, 2, 3}
GroupBy(input, func(i int) bool {
    return i > 0
}) // => map[bool][]int{false: {-2, -1, 0}, true: {1, 2, 3}}
```

### Map

`Map` returns a new slice with the results of running the callback once for each element of the
slice. The return type of the callback does not need to match the type of the slice - this enables
straightforward conversions between different types, such as an extraction of a specific value from
a struct.

The function signature of the callback is `func(value T) V`, where `T` is the type of the slice and
`V` is your return type.

```go
nums := []int{1, 2, 3, 4}
squares := Map(nums, func(val int) int {
    return val * val
}) // => []int{1, 4, 9, 16}
```

### MapI

`MapI` is a variant of `Map` that also provides the current index in the callback.

The function signature of the callback is `func(index int, value T) V`, where `T` is the type of the
slice and `V` is your return type.

### MapHash

`MapHash` is the equivalent of `Map` for Go maps (dictionaries, hash-maps).

The function signature of the callback is `func(value V) (T)`, where `V` is the value type of the
input map and `T` is the value type of the new map.

```go
nums := map[string]int{"foo": 1, "bar": 2, "baz": 3, "qux": 4}
squares := MapHash(input, func(v int) int {
    return v * v
}) // => { "foo": 1, "bar": 4, "baz": 9, "qux": 16 }
```

### MapHashK

`MapHash` is the equivalent of `MapI` for Go maps (dictionaries, hash-maps). It allows you to change
keys as well as values.

The function signature of the callback is `func(key K, value V) (NK, NV)`, where `K` and `V` are key
and value types of the input map, and `NK` and `NV` are key and value types of the new map.

```go
nums := map[string]int{"foo": 1, "bar": 2, "baz": 3, "qux": 4}
squares := MapHashK(input, func(k string, v int) (int, int) {
    return v, v * v
}) // => { 1: 1, 2: 4, 3: 9, 4: 16 }
```


### Median

`Median` returns the middle number (50th percentile) of a numeric slice.

```go
nums := []int{1, 2, 3}
Median(nums) // => 2
```

### Reduce

`Reduce` combines all elements of an input slice into one by applying a function to each element and
accumulating the result of the function (the accumulator or "memo" value) throughout each iteration.

```go
nums := []int{1, 2, 3, 4}
sum := Reduce(nums, func(i int, s int) int {
    return i + s
}); // => 10
```

In addition to `Reduce`, two other functions are provided.

#### ReduceWith

`ReduceWith` allows you to define a custom initial value of the accumulator.

```go
nums := []int{1, 2, 3, 4}
product := Reduce(nums, 1, func(i int, s int) int {
    return i * s
}); // => 24
```

#### ReduceFrom

`ReduceFrom` allows you to define a function that will be called to set the initial value, and the
argument of this callback will be the first element of the slice. The callback will then be called
for each subsequent element of the slice (but not the first one again). This also lets you
initialize the memo to a more complex value.

```go
nums := []int{1, 2, 3, 4, -1}
smallest := ReduceFrom(input,
    func(i int) int { return i }, // `memo` is set to the first entry in the slice
    func(v int, memo int) int {
        if v < memo {
            return v
        }
        return memo
    },
) // => -1
```

The function signature of every `Reduce` callback is `func(value T, memo V) V`, where T is the type
of the slice and V is the type of the return value.


### Zip

`Zip` merges the `lhs` and `rhs` slices by returning slices in which the first element comes from
the first slice, and the second element from the second slice. The merge stops once either of the
slices is exhausted.


```go
lhs := []int{1, 2, 3}
rhs := []int{4, 5, 6}

Zip(lhs, rhs) // => [][]int{ {1, 4}, {2, 5}, {3, 6} }
```
