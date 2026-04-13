# gseq

A generic, functional data structure library for Go 1.23+, inspired by [samber/lo](https://github.com/samber/lo).

Four independent packages with a clear dependency chain:

```
option  ←  slice  ←  dict
   ↑           ↑
result  ────────
```

```go
import (
    "github.com/stefanbethge/gseq/option"
    "github.com/stefanbethge/gseq/result"
    "github.com/stefanbethge/gseq/slice"
    "github.com/stefanbethge/gseq/dict"
)
```

---

## `option` — Optional value

Replaces the `(T, bool)` pattern with an explicit `Option[T]` type.

```go
opt := option.Some(42)
opt := option.None[int]()

opt.IsSome()           // true
opt.IsNone()           // false
opt.Unwrap()           // 42  — panics if None
opt.UnwrapOr(0)        // 42  — safe fallback
opt.UnwrapOrElse(func() int { return computeDefault() }) // lazy fallback
v, ok := opt.Get()     // classic (T, bool) when needed
```

### Method chaining

```go
result := findUser(id).                          // Option[User]
    Filter(func(u User) bool { return u.Active }).
    Or(loadDefaultUser()).                        // fallback if None
    UnwrapOr(guestUser)
```

| Method | Description |
|---|---|
| `.Filter(fn)` | `Some(v)` → `None` if `fn(v)` is false |
| `.Or(other)` | returns itself if `Some`, otherwise `other` |
| `.OrElse(fn)` | lazy variant of `Or` — `fn` is only called when `None` |
| `.UnwrapOrElse(fn)` | lazy variant of `UnwrapOr` — `fn` is only called when `None` |

### Free functions

```go
// Map — transform the value without unwrapping
upper := option.Map(findName(id), strings.ToUpper) // Option[string]

// FlatMap — flatten a nested Option returned by fn
addr := option.FlatMap(findUser(id), func(u User) option.Option[Address] {
    return findAddress(u.AddressID)
})
```

---

## `slice` — Generic slice with method chaining

```go
s := slice.Slice[int]{3, 1, 4, 1, 5, 9, 2, 6}
```

### Transformations (return a new slice)

```go
s.Filter(func(v int) bool { return v > 3 })            // {4, 5, 9, 6}
s.Reject(func(v int) bool { return v > 3 })            // {3, 1, 1, 2}
s.Limit(3)                                             // {3, 1, 4}
s.Skip(2)                                              // {4, 1, 5, 9, 2, 6}
s.Reverse()                                            // {6, 2, 9, 5, 1, 4, 1, 3}
s.Shuffle()                                            // random order (copy)
s.Chunk(3)                                             // [{3,1,4}, {1,5,9}, {2,6}]
s.Without(func(a, b int) bool { return a == b }, 1, 9) // {3, 4, 5, 2, 6}
```

### Sorting

```go
// SortBy — arbitrary comparison function
s.SortBy(func(a, b int) bool { return a < b }) // ascending

// Sort — for all cmp.Ordered types (int, string, float…)
slice.Sort(s)
slice.Sort(slice.Slice[string]{"banana", "apple", "cherry"})
```

### Windows

```go
s := slice.Slice[int]{1, 2, 3, 4, 5}
s.Window(3)   // [{1,2,3}, {2,3,4}, {3,4,5}]
s.Pairwise()  // [{1,2}, {2,3}, {3,4}, {4,5}]
```

### Terminators (return a single value)

```go
s.First()                                           // option.Some(3)
s.Last()                                            // option.Some(6)
s.Nth(2)                                            // option.Some(4)
s.Find(func(v int) bool { return v > 4 })           // option.Some(5)
s.Sample()                                          // option.Some(<random>)
s.Samples(3)                                        // 3 random elements

s.Contains(func(v int) bool { return v == 9 })      // true
s.Every(func(v int) bool { return v > 0 })          // true
s.None(func(v int) bool { return v < 0 })           // true
s.Count(func(v int) bool { return v%2 == 0 })       // 3
s.IndexOf(func(v int) bool { return v == 5 })       // 4
s.Len()                                             // 8
s.IsEmpty()                                         // false

yes, no := s.Partition(func(v int) bool { return v%2 == 0 })
// yes={4,2,6}  no={3,1,1,5,9}
```

### Iteration

```go
s.Each(func(v int) { fmt.Println(v) })
s.EachIndexed(func(i int, v int) { fmt.Printf("[%d] %d\n", i, v) })

s.EachParallel(func(v int) { process(v) })               // parallel, order not guaranteed
s.EachParallelIndexed(func(i int, v int) { process(i, v) })
```

### Free functions

```go
// Map — type transformation T → O
names := slice.Map(users, func(u User) string { return u.Name })

// MapIndexed — like Map but fn also receives the element's index
tagged := slice.MapIndexed(users, func(i int, u User) string {
    return fmt.Sprintf("%d: %s", i, u.Name)
})

// MapParallel — like Map but parallel (order preserved)
scores := slice.MapParallel(ids, func(id int) int { return fetchScore(id) })

// MapParallelIndexed — parallel with index
scores := slice.MapParallelIndexed(ids, func(i int, id int) int { return fetchScore(i, id) })

// Reduce
total := slice.Reduce(prices, 0.0, func(acc, p float64) float64 { return acc + p })

// GroupBy
byDept := slice.GroupBy(employees, func(e Employee) string { return e.Department })
// map[string]Slice[Employee]

// KeyBy — map with a unique key
byID := slice.KeyBy(users, func(u User) int { return u.ID })
// map[int]User

// FlatMap
tags := slice.FlatMap(posts, func(p Post) slice.Slice[string] { return p.Tags })

// Flatten
slice.Flatten(slice.Slice[slice.Slice[int]]{{1, 2}, {3, 4}}) // {1, 2, 3, 4}

// Uniq — remove duplicates
slice.Uniq(s, func(v int) int { return v })

// Set operations
slice.Intersect(a, b, func(v int) int { return v })
slice.Difference(a, b, func(v int) int { return v })
slice.Union(a, b, func(v int) int { return v })

// Zip — combine two slices pairwise
slice.Zip(names, ages, func(name string, age int) string {
    return fmt.Sprintf("%s (%d)", name, age)
})

// Numeric
slice.Sum(slice.Slice[int]{1, 2, 3, 4, 5})                        // 15
slice.SumBy(products, func(p Product) float64 { return p.Price })
slice.Min(s)    // option.Some(1)
slice.Max(s)    // option.Some(9)
slice.MinBy(products, func(p Product) float64 { return p.Price })
slice.MaxBy(products, func(p Product) float64 { return p.Price })

// Range
slice.Range(0, 5)         // {0, 1, 2, 3, 4}
slice.RangeStep(0, 10, 2) // {0, 2, 4, 6, 8}
slice.Repeat("x", 4)      // {"x", "x", "x", "x"}
```

### Option integration

```go
// Compact — discard None values
opts := slice.Slice[option.Option[int]]{
    option.Some(1), option.None[int](), option.Some(3),
}
slice.Compact(opts) // {1, 3}

// TryMap — map and filter in one pass: None values are dropped
nums := slice.TryMap(inputs, func(s string) option.Option[int] {
    n, err := strconv.Atoi(s)
    if err != nil {
        return option.None[int]()
    }
    return option.Some(n)
})
```

### Iterator API (Go 1.23 `iter.Seq`)

```go
// Lazy pipeline — no intermediate allocations
result := slice.Collect(
    slice.TakeIter(
        slice.MapIter(
            slice.FilterIter(s.Iter(), func(v int) bool { return v%2 == 0 }),
            func(v int) int { return v * v },
        ),
        3,
    ),
)
// First 3 squares of even numbers

// Iter2 — index + value
for i, v := range s.Iter2() {
    fmt.Printf("[%d] = %d\n", i, v)
}

// Collect — iter.Seq[T] → Slice[T]
slice.Collect(someExternalIterator)
```

---

## `dict` — Generic map with method chaining

```go
d := dict.Map[string, int]{"alice": 90, "bob": 75, "carol": 88}
```

### Methods

```go
d.Filter(func(k string, v int) bool { return v >= 80 })
// {"alice": 90, "carol": 88}

d.Reject(func(k string, v int) bool { return v >= 80 })
// {"bob": 75}

d.Each(func(k string, v int) { fmt.Printf("%s: %d\n", k, v) })

d.Keys()    // slice.Slice[string] (order not guaranteed)
d.Values()  // slice.Slice[int]

d.Contains(func(k string, v int) bool { return v == 100 }) // false
d.Every(func(k string, v int) bool { return v > 50 })      // true
d.None(func(k string, v int) bool { return v < 0 })        // true
d.Len()                                                     // 3
d.ToMap()                                                   // map[string]int{...}

// Merge — other's values win on duplicate keys
merged := d.Merge(dict.Map[string, int]{"alice": 100, "dave": 70})

// MergeWith — custom resolution for duplicate keys (key, left, right)
merged := d.MergeWith(
    dict.Map[string, int]{"alice": 100, "dave": 70},
    func(_ string, existing, incoming int) int { return max(existing, incoming) },
)
```

### Free functions

```go
// MapValues — transform values
doubled := dict.MapValues(d, func(k string, v int) int { return v * 2 })

// Invert — swap keys and values
inv := dict.Invert(dict.Map[string, int]{"a": 1, "b": 2})
// dict.Map[int, string]{1: "a", 2: "b"}

// ToSlice — convert map to slice
pairs := dict.ToSlice(d, func(k string, v int) string {
    return fmt.Sprintf("%s=%d", k, v)
})

// FromSlice — build map from slice
byName := dict.FromSlice(users, func(u User) (string, User) {
    return u.Name, u
})
```

---

## `result` — Explicit error handling

`Result[T, E]` represents either a success value `Ok(T)` or a typed error `Err(E)`.

```go
r := result.Ok[int, string](42)
r := result.Err[int, string]("something went wrong")

// Wrap standard Go (value, error) pairs
r := result.FromGoError(strconv.Atoi(input))  // Result[int, error]
r := result.FromGoError(os.ReadFile("x.txt")) // Result[[]byte, error]
```

### Unwrapping

```go
r.IsOk()         // true / false
r.IsErr()        // false / true
r.Unwrap()       // value — panics on Err
r.UnwrapErr()    // error — panics on Ok
r.UnwrapOr(0)    // safe fallback
r.UnwrapOrElse(func() int { return computeDefault() }) // lazy fallback
r.ToOption()     // option.Option[T] — error is discarded
```

### Transformations

```go
// Map — transform value, propagate error unchanged
doubled := result.Map(result.FromGoError(strconv.Atoi("21")),
    func(n int) int { return n * 2 },
) // Ok(42)

// FlatMap — chain results
validated := result.FlatMap(
    result.FromGoError(strconv.Atoi(input)),
    func(n int) result.Result[int, error] {
        if n < 0 {
            return result.Err[int, error](errors.New("must be positive"))
        }
        return result.Ok[int, error](n)
    },
)

// MapErr — transform the error type
r := result.MapErr(dbResult, func(e dbError) string { return e.Message })
```

### Partition — collect errors from a batch run

`Partition` splits a `Slice[Result[T, E]]` into successes and errors, preserving order.
Use a custom error struct to attach context — the same `process` function works for both
sequential and parallel execution.

```go
type RecordError struct {
    Index  int
    Record Record
    Cause  error
}

process := func(i int, r Record) result.Result[Output, RecordError] {
    out, err := doWork(r)
    if err != nil {
        return result.Err[Output, RecordError](RecordError{Index: i, Record: r, Cause: err})
    }
    return result.Ok[Output, RecordError](out)
}

// Sequential
results := slice.MapIndexed(records, process)

// Parallel — swap one word, everything else stays the same
results := slice.MapParallelIndexed(records, process)

// Evaluate — identical for both
successes, failures := result.Partition(results)

fmt.Printf("✓ %d  ✗ %d\n", len(successes), len(failures))
failures.Each(func(e RecordError) {
    log.Printf("record[%d] %+v: %v", e.Index, e.Record, e.Cause)
})
```

### Bridging `result` and `option`

```go
// Result → Option (error is discarded)
opt := result.FromGoError(strconv.Atoi(s)).ToOption()

// Option → Result (supply the error value for the None case)
r := result.FromOption(findUser(id), errors.New("user not found")) // Result[User, error]
```

---

## Full example

```go
type User struct {
    ID     int
    Name   string
    Score  int
    Active bool
}

users := slice.Slice[User]{
    {1, "Alice", 95, true},
    {2, "Bob", 60, false},
    {3, "Carol", 88, true},
    {4, "Dave", 72, true},
}

// Top 2 active users by score, names only
top := slice.Map(
    users.
        Filter(func(u User) bool { return u.Active }).
        SortBy(func(a, b User) bool { return a.Score > b.Score }).
        Limit(2),
    func(u User) string { return u.Name },
)
// {"Alice", "Carol"}

// Average score of active users
active := users.Filter(func(u User) bool { return u.Active })
avg := float64(slice.SumBy(active, func(u User) int { return u.Score })) /
    float64(active.Len())
// 85.0

// Group into tiers
tiers := slice.GroupBy(users, func(u User) string {
    if u.Score >= 80 {
        return "top"
    }
    return "rest"
})
// {"top": [Alice, Carol], "rest": [Bob, Dave]}

// Parse and validate a list of raw IDs, silently drop invalid ones
ids := slice.TryMap(rawInputs, func(s string) option.Option[int] {
    return result.FromGoError(strconv.Atoi(s)).ToOption()
})
```
