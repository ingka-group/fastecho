# agg

The `agg` package provides (very) basic statistics and grouping support. It is designed to be used
alongside the [fp package](../fp/README.md).

## Usage

`agg` provides [basic aggregations](#Basic-aggregations), as well as methods to [group time-series
data](#Time-aggregations).

## Example

Let's say we have a type representing a value on a given day.

```go
type DataPoint struct {
    timestamp time.Time
    Value float64
}

```

We want to **find out what is the median value of `Value` for each IKEA week in a given array of
these data points**.

### Preparation

The `fp` package provides [`fp.GroupBy`](../fp/README.md#GroupBy), which we can use to group by
arbitrary keys. While you can implement the grouping logic yourself using just that function, this
package provides pre-built groupers designed specifically for time-series data.

To get started with grouping and aggregating this data, we need to implement the
[`Dated`](#Time-aggregations) interface on our type, which returns the underlying `time.Time` value.

```go
func (d DataPoint) Date() time.Time {
    return timestamp
}
```

### Grouping

Let's move on to actual grouping logic. `ByIKEAWeek` is a grouper function that returns an
`IKEAWeek` type for a given `Dated` type. Combined with , you can use it to group your slice of
`Dated` values by IKEA week.

```go
dataPoints := []DataPoint {
    {timestamp: jan1, Value: 10},
    {timestamp: jan7, Value: 20},
    {timestamp: jan8, Value: 15},
}

grouped := fp.GroupBy(dataPoints, ByIKEAWeek)
// => map[agg.IKEAWeek][]dataPoints {
//  (2024-W01): {{timestamp: jan1, Value: 10}},
//  (2024-W02): {{timestamp: jan7, Value: 20}, {timestamp: jan8, Value: 15}},

```

Your data is now grouped in a map (hash-map/dictionary) simply by implementing `Dated` on your type
and using the appropriate grouper function. Additional grouper functions are available in the
[Reference section](#Time-aggregations).

### Aggregating

Now we need to extract the median value per IKEA week. For this, we can use
[`fp.Map`](../fp/README.md#Map), [`fp.Collect`](../fp/README.md#Collect), and the [basic
aggregations](#Basic-aggregations). It's also helpful to define a data type to represent the
aggregation itself.

```go
// Aggregated holds our weekly aggregation values to make this example more clear
type Aggregated struct {
    Week IKEAWeek `json:"week"`
    Value float64 `json:"value"`
}

// we are iterating through each group in this function and returning one `Aggregated` per week
medians := Collect(grouped, func(week IKEAWeek, dataPoints []DataPoint) Aggregated {
    // collect all `DataPoint.Value`s in a slice of floats
    values := fp.Map(dataPoints, func (d DataPoint) float64 { return d.Value })

    // calculate the median value of those values
    median := agg.Median(values)

    // return the week's aggregation
    return Aggregated {
        Week: week,
        Value: median
    }
})
// => []Aggregated{
//   {Week: (2024-W01), Value: 10},
//   {Week: (2024-W02), Value: 27.5},
// }
```

> [!NOTE]
> The order of the slice returned by `Collect` is not guaranteed to be stable -- if you need your
> slice sorted in a certain way, you should [sort it manually](https://pkg.go.dev/slices#SortFunc).

And that's all! To get the medians from the groups, we iterated over that map with `Collect` and
used `Map` to extract the raw value from the slice of data objects, then ran `Median` over that
slice to get the median value, and finally returned a new type containing the week and the
aggregation. The result is a slice of a type representing this aggregation that we can easily
forward elsewhere or use as an API response. All of this in about 10 lines of readable, clean and
reusable code.

## Reference

### Basic aggregations

Basic aggregations work on any numeric slice. If you have a slice of structs, you could use
[fp.Map](../fp/README.md#Map) to extract a slice of values you want to aggregate.

#### Add

`Add` adds two numbers (exciting!). It's used as a building block for higher-order functions because
in Go, `+` is a built-in instead of a method call on a numeric object.

#### Avg

`Avg` returns the mathematical average of a slice of numbers.


```go
nums := []int{1, 2, 3, 100}
Avg(nums) // => 26.5
```

#### Identity

`Identity` returns the given value (riveting!). Useful to simplify
[`ReduceFrom`](../fp/README.md#ReduceFrom) calls when initialiing the aggregator (memo) value to the
first value in the slice.

#### Max

`Max` returns the largest value in the slice.

```go
nums := []int{1, 2, 3, -100}
Max(nums) // => 3
```

#### Median

`Median` returns the middle number (50th percentile) in the slice. Even-sized slices return the
average of the two middle values.


```go
nums := []int{-50, 2, 3, 200}
Median(nums) // => 2.5
```


#### Min

`Min` returns the smallest value in the slice.

```go
nums := []int{1, 2, 3, -100}
Min(nums) // => -100
```


#### Sum

`Sum` returns the sum of a slice of numbers.

```go
nums := []int{1, 2, 3, 100}
Sum(nums) // => 106
```

### Time aggregations

These methods provide a standardized way for grouping and aggregating data by an arbitrary time
period. To get started, you will need to implement the `Dated` interface on your type:

```go
type Dated interface {
    Date() time.time
}

```

Several `Dated` types are provided with this package: `agg.Year`, `agg.YearMonth`, `agg.IKEAWeek`
and `date.ISODate`. If your data model uses any of those types, you can simply delegate the `Date()`
method to that field.


#### ByIKEAWeek

`ByIKEAWeek` is a grouper function that returns an `IKEAWeek` type for a given `Dated` type.
Combined with [`fp.GroupBy`](../fp/README.md#GroupBy), you can use it to group your slice of
`Dated` values by IKEA week.

See the [Example](#Example) for a demonstration.

#### ByISODate

`ByISODate` is a grouper function that returns an `ISODate` type for a given `Dated` type.
Combined with [`fp.GroupBy`](../fp/README.md#GroupBy), you can use it to group your slice of
`Dated` values by ISO date.

See the [Example](#Example) for a demonstration.

#### ByYearMonth

`ByISODate` is a grouper function that returns a `YearMonth` type for a given `Dated` type.
Combined with [`fp.GroupBy`](../fp/README.md#GroupBy), you can use it to group your slice of
`Dated` values by year and month.

See the [Example](#Example) for a demonstration.

#### ByYear

`ByYear` is a grouper function that returns the `Year` of a `Dated` type. Combined with
[`fp.GroupBy`](../fp/README.md#GroupBy), you can use it to group your slice of `Dated` values by
year.

See the [Example](#Example) for a demonstration.
