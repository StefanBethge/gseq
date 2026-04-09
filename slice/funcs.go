package slice

import (
	"runtime"
	"sync"

	"github.com/stefanbethge/gseq/option"
)

// parallelThreshold is the minimum slice length for which spawning a worker
// pool is worthwhile. Below this threshold all parallel functions fall back to
// their sequential equivalents to avoid goroutine and channel overhead.
const parallelThreshold = 256

// Map transforms every element from T to O.
func Map[T, O any](s Slice[T], fn func(T) O) Slice[O] {
	result := make(Slice[O], len(s))
	for i, v := range s {
		result[i] = fn(v)
	}
	return result
}

// MapIndexed is like Map but fn also receives the element's index.
func MapIndexed[T, O any](s Slice[T], fn func(int, T) O) Slice[O] {
	result := make(Slice[O], len(s))
	for i, v := range s {
		result[i] = fn(i, v)
	}
	return result
}

// MapParallel is like Map but runs fn concurrently using a worker pool sized
// to GOMAXPROCS, preserving order. For I/O-bound work that benefits from
// higher concurrency use MapParallelN to set the worker count explicitly.
func MapParallel[T, O any](s Slice[T], fn func(T) O) Slice[O] {
	return MapParallelN(s, runtime.GOMAXPROCS(0), fn)
}

// MapParallelN is like MapParallel but uses exactly n workers.
// Falls back to sequential Map when len(s) < parallelThreshold.
// Uses direct slice segmentation instead of channels to avoid per-item
// allocation and channel overhead.
func MapParallelN[T, O any](s Slice[T], n int, fn func(T) O) Slice[O] {
	if len(s) < parallelThreshold {
		return Map(s, fn)
	}
	result := make(Slice[O], len(s))
	if n > len(s) {
		n = len(s)
	}
	chunkSize := (len(s) + n - 1) / n
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		start := i * chunkSize
		if start >= len(s) {
			break
		}
		end := start + chunkSize
		if end > len(s) {
			end = len(s)
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				result[j] = fn(s[j])
			}
		}(start, end)
	}
	wg.Wait()
	return result
}

// MapParallelIndexed is like MapParallel but fn also receives the element's index.
// Useful for attaching position metadata to results or errors.
func MapParallelIndexed[T, O any](s Slice[T], fn func(int, T) O) Slice[O] {
	return MapParallelIndexedN(s, runtime.GOMAXPROCS(0), fn)
}

// MapParallelIndexedN is like MapParallelIndexed but uses exactly n workers.
// Falls back to sequential MapIndexed when len(s) < parallelThreshold.
func MapParallelIndexedN[T, O any](s Slice[T], n int, fn func(int, T) O) Slice[O] {
	if len(s) < parallelThreshold {
		return MapIndexed(s, fn)
	}
	result := make(Slice[O], len(s))
	if n > len(s) {
		n = len(s)
	}
	chunkSize := (len(s) + n - 1) / n
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		start := i * chunkSize
		if start >= len(s) {
			break
		}
		end := start + chunkSize
		if end > len(s) {
			end = len(s)
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				result[j] = fn(j, s[j])
			}
		}(start, end)
	}
	wg.Wait()
	return result
}

// FilterParallel is like Filter but runs fn concurrently using a worker pool
// sized to GOMAXPROCS, preserving element order.
func FilterParallel[T any](s Slice[T], fn func(T) bool) Slice[T] {
	return FilterParallelN(s, runtime.GOMAXPROCS(0), fn)
}

// FilterParallelN is like FilterParallel but uses exactly n workers.
// Falls back to sequential Filter when len(s) < parallelThreshold.
// Each worker filters its segment into a local slice; results are merged
// in order after all goroutines finish.
func FilterParallelN[T any](s Slice[T], n int, fn func(T) bool) Slice[T] {
	if len(s) < parallelThreshold {
		return s.Filter(fn)
	}
	if n > len(s) {
		n = len(s)
	}
	chunkSize := (len(s) + n - 1) / n
	partials := make([]Slice[T], n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		start := i * chunkSize
		if start >= len(s) {
			break
		}
		end := start + chunkSize
		if end > len(s) {
			end = len(s)
		}
		wg.Add(1)
		go func(i, start, end int) {
			defer wg.Done()
			local := make(Slice[T], 0, end-start)
			for j := start; j < end; j++ {
				if fn(s[j]) {
					local = append(local, s[j])
				}
			}
			partials[i] = local
		}(i, start, end)
	}
	wg.Wait()
	total := 0
	for _, p := range partials {
		total += len(p)
	}
	result := make(Slice[T], 0, total)
	for _, p := range partials {
		result = append(result, p...)
	}
	return result
}

// GroupBy groups elements by the key returned by fn.
func GroupBy[T any, K comparable](s Slice[T], fn func(T) K) map[K]Slice[T] {
	// Hint at the number of unique keys. Overestimates when cardinality is low,
	// but eliminates rehashing for high-cardinality inputs.
	result := make(map[K]Slice[T], len(s))
	for _, v := range s {
		key := fn(v)
		result[key] = append(result[key], v)
	}
	return result
}

// KeyBy creates a map keyed by fn; last value wins on duplicate keys.
func KeyBy[T any, K comparable](s Slice[T], fn func(T) K) map[K]T {
	result := make(map[K]T, len(s))
	for _, v := range s {
		result[fn(v)] = v
	}
	return result
}

// Reduce reduces the slice to a single value.
func Reduce[T, O any](s Slice[T], initial O, fn func(O, T) O) O {
	acc := initial
	for _, v := range s {
		acc = fn(acc, v)
	}
	return acc
}

// ReduceParallel reduces s in parallel using fn sized to GOMAXPROCS workers.
// fn must be associative: the result of combining segments in any order must
// equal the sequential result. Typical uses: sum, max/min, merge.
// Falls back to sequential Reduce when len(s) < parallelThreshold.
func ReduceParallel[T any](s Slice[T], initial T, fn func(T, T) T) T {
	return ReduceParallelN(s, runtime.GOMAXPROCS(0), initial, fn)
}

// ReduceParallelN is like ReduceParallel but uses exactly n workers.
func ReduceParallelN[T any](s Slice[T], n int, initial T, fn func(T, T) T) T {
	if len(s) < parallelThreshold {
		return Reduce(s, initial, fn)
	}
	if n > len(s) {
		n = len(s)
	}
	chunkSize := (len(s) + n - 1) / n
	partials := make([]T, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		start := i * chunkSize
		if start >= len(s) {
			break
		}
		end := start + chunkSize
		if end > len(s) {
			end = len(s)
		}
		wg.Add(1)
		go func(i, start, end int) {
			defer wg.Done()
			acc := initial
			for j := start; j < end; j++ {
				acc = fn(acc, s[j])
			}
			partials[i] = acc
		}(i, start, end)
	}
	wg.Wait()
	acc := initial
	for _, p := range partials {
		acc = fn(acc, p)
	}
	return acc
}

// FlatMap applies fn to each element and flattens the result.
// The initial capacity is len(s) as a lower-bound estimate; use Flatten after
// collecting all inner slices when the total size is known up-front.
func FlatMap[T, O any](s Slice[T], fn func(T) Slice[O]) Slice[O] {
	// Two-pass: collect inner slices first so we can pre-calculate total size.
	parts := make([]Slice[O], len(s))
	total := 0
	for i, v := range s {
		parts[i] = fn(v)
		total += len(parts[i])
	}
	result := make(Slice[O], 0, total)
	for _, p := range parts {
		result = append(result, p...)
	}
	return result
}

// Flatten flattens a slice of slices: [][]T → []T.
func Flatten[T any](s Slice[Slice[T]]) Slice[T] {
	total := 0
	for _, inner := range s {
		total += len(inner)
	}
	result := make(Slice[T], 0, total)
	for _, inner := range s {
		result = append(result, inner...)
	}
	return result
}

// Uniq removes duplicates based on a key function.
func Uniq[T any, K comparable](s Slice[T], fn func(T) K) Slice[T] {
	seen := make(map[K]struct{}, len(s))
	result := make(Slice[T], 0, len(s))
	for _, v := range s {
		key := fn(v)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Exclude returns a copy of s without any element equal to any of elems.
// Unlike Without, it requires T to be comparable and runs in O(n + m) time
// instead of O(n × m), making it efficient for large exclusion sets.
func Exclude[T comparable](s Slice[T], elems ...T) Slice[T] {
	if len(elems) == 0 {
		return s
	}
	set := make(map[T]struct{}, len(elems))
	for _, e := range elems {
		set[e] = struct{}{}
	}
	return s.Filter(func(v T) bool {
		_, found := set[v]
		return !found
	})
}

// Intersect returns elements present in both slices (keyed by fn).
func Intersect[T any, K comparable](a, b Slice[T], fn func(T) K) Slice[T] {
	keys := make(map[K]struct{}, len(b))
	for _, v := range b {
		keys[fn(v)] = struct{}{}
	}
	return a.Filter(func(v T) bool {
		_, ok := keys[fn(v)]
		return ok
	})
}

// Difference returns elements from a that are NOT in b (keyed by fn).
func Difference[T any, K comparable](a, b Slice[T], fn func(T) K) Slice[T] {
	keys := make(map[K]struct{}, len(b))
	for _, v := range b {
		keys[fn(v)] = struct{}{}
	}
	return a.Filter(func(v T) bool {
		_, ok := keys[fn(v)]
		return !ok
	})
}

// Union merges two slices and removes duplicates (keyed by fn).
func Union[T any, K comparable](a, b Slice[T], fn func(T) K) Slice[T] {
	return Uniq(append(a, b...), fn)
}

// Zip combines two slices pairwise into a new slice using fn.
func Zip[A, B, O any](a Slice[A], b Slice[B], fn func(A, B) O) Slice[O] {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	result := make(Slice[O], n)
	for i := 0; i < n; i++ {
		result[i] = fn(a[i], b[i])
	}
	return result
}

// Compact extracts the values from a Slice of Options, discarding all None entries.
func Compact[T any](s Slice[option.Option[T]]) Slice[T] {
	result := make(Slice[T], 0, len(s))
	for _, o := range s {
		if o.IsSome() {
			result = append(result, o.Unwrap())
		}
	}
	return result
}

// TryMap applies fn to each element, collecting only the Some results.
// Elements for which fn returns None are silently dropped.
func TryMap[T, O any](s Slice[T], fn func(T) option.Option[O]) Slice[O] {
	result := make(Slice[O], 0, len(s))
	for _, v := range s {
		if o := fn(v); o.IsSome() {
			result = append(result, o.Unwrap())
		}
	}
	return result
}
