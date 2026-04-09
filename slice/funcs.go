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
func MapParallelN[T, O any](s Slice[T], n int, fn func(T) O) Slice[O] {
	if len(s) < parallelThreshold {
		return Map(s, fn)
	}
	result := make(Slice[O], len(s))
	if len(s) == 0 {
		return result
	}
	if n > len(s) {
		n = len(s)
	}
	type job struct {
		i int
		v T
	}
	jobs := make(chan job, len(s))
	for i, v := range s {
		jobs <- job{i, v}
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			for j := range jobs {
				result[j.i] = fn(j.v)
			}
		}()
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
	if len(s) == 0 {
		return result
	}
	if n > len(s) {
		n = len(s)
	}
	type job struct {
		i int
		v T
	}
	jobs := make(chan job, len(s))
	for i, v := range s {
		jobs <- job{i, v}
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			for j := range jobs {
				result[j.i] = fn(j.i, j.v)
			}
		}()
	}
	wg.Wait()
	return result
}

// GroupBy groups elements by the key returned by fn.
func GroupBy[T any, K comparable](s Slice[T], fn func(T) K) map[K]Slice[T] {
	result := make(map[K]Slice[T])
	for _, v := range s {
		key := fn(v)
		result[key] = append(result[key], v)
	}
	return result
}

// KeyBy creates a map keyed by fn; last value wins on duplicate keys.
func KeyBy[T any, K comparable](s Slice[T], fn func(T) K) map[K]T {
	result := make(map[K]T)
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

// FlatMap applies fn to each element and flattens the result.
func FlatMap[T, O any](s Slice[T], fn func(T) Slice[O]) Slice[O] {
	result := make(Slice[O], 0, len(s))
	for _, v := range s {
		result = append(result, fn(v)...)
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

// Intersect returns elements present in both slices (keyed by fn).
func Intersect[T any, K comparable](a, b Slice[T], fn func(T) K) Slice[T] {
	keys := make(map[K]struct{})
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
	keys := make(map[K]struct{})
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
	var result Slice[T]
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
	var result Slice[O]
	for _, v := range s {
		if o := fn(v); o.IsSome() {
			result = append(result, o.Unwrap())
		}
	}
	return result
}
