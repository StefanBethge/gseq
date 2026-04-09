// Package dict provides a generic, chainable Map type for key-value operations.
package dict

import (
	"runtime"
	"sync"

	"github.com/stefanbethge/gseq/slice"
)

// parallelThreshold mirrors the slice package: below this entry count, parallel
// operations fall back to their sequential equivalents.
const parallelThreshold = 256

// Map is a generic map type with chainable operations.
type Map[K comparable, V any] map[K]V

// Filter returns a new Map containing only entries for which fn returns true.
func (m Map[K, V]) Filter(fn func(K, V) bool) Map[K, V] {
	result := make(Map[K, V])
	for k, v := range m {
		if fn(k, v) {
			result[k] = v
		}
	}
	return result
}

// Reject is the inverse of Filter.
func (m Map[K, V]) Reject(fn func(K, V) bool) Map[K, V] {
	return m.Filter(func(k K, v V) bool { return !fn(k, v) })
}

// Each calls fn for every key-value pair.
func (m Map[K, V]) Each(fn func(K, V)) {
	for k, v := range m {
		fn(k, v)
	}
}

// EachParallel calls fn for every entry concurrently using a worker pool sized
// to GOMAXPROCS. Falls back to sequential Each for small maps.
func (m Map[K, V]) EachParallel(fn func(K, V)) {
	m.EachParallelN(runtime.GOMAXPROCS(0), fn)
}

// EachParallelN calls fn for every entry using exactly n workers.
// Falls back to sequential Each when len(m) < parallelThreshold.
func (m Map[K, V]) EachParallelN(n int, fn func(K, V)) {
	if len(m) < parallelThreshold {
		m.Each(fn)
		return
	}
	type entry struct{ k K; v V }
	entries := make([]entry, 0, len(m))
	for k, v := range m {
		entries = append(entries, entry{k, v})
	}
	if n > len(entries) {
		n = len(entries)
	}
	chunkSize := (len(entries) + n - 1) / n
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		start := i * chunkSize
		if start >= len(entries) {
			break
		}
		end := start + chunkSize
		if end > len(entries) {
			end = len(entries)
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				fn(entries[j].k, entries[j].v)
			}
		}(start, end)
	}
	wg.Wait()
}

// Keys returns all keys as a Slice (order not guaranteed).
func (m Map[K, V]) Keys() slice.Slice[K] {
	result := make(slice.Slice[K], 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// Values returns all values as a Slice (order not guaranteed).
func (m Map[K, V]) Values() slice.Slice[V] {
	result := make(slice.Slice[V], 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

// Contains returns true if at least one entry satisfies fn.
func (m Map[K, V]) Contains(fn func(K, V) bool) bool {
	for k, v := range m {
		if fn(k, v) {
			return true
		}
	}
	return false
}

// Every returns true if ALL entries satisfy fn.
func (m Map[K, V]) Every(fn func(K, V) bool) bool {
	for k, v := range m {
		if !fn(k, v) {
			return false
		}
	}
	return true
}

// Len returns the number of entries.
func (m Map[K, V]) Len() int { return len(m) }

// ToMap converts back to a plain Go map.
func (m Map[K, V]) ToMap() map[K]V { return map[K]V(m) }

// ─── FREE FUNCTIONS ───────────────────────────────────────────────────────────

// MapValues transforms each value using fn, producing a new Map.
func MapValues[K comparable, V, W any](m Map[K, V], fn func(K, V) W) Map[K, W] {
	result := make(Map[K, W], len(m))
	for k, v := range m {
		result[k] = fn(k, v)
	}
	return result
}

// MapValuesParallel is like MapValues but runs fn concurrently using a worker
// pool sized to GOMAXPROCS. Falls back to sequential MapValues for small maps.
func MapValuesParallel[K comparable, V, W any](m Map[K, V], fn func(K, V) W) Map[K, W] {
	return MapValuesParallelN(m, runtime.GOMAXPROCS(0), fn)
}

// MapValuesParallelN is like MapValuesParallel but uses exactly n workers.
// Falls back to sequential MapValues when len(m) < parallelThreshold.
func MapValuesParallelN[K comparable, V, W any](m Map[K, V], n int, fn func(K, V) W) Map[K, W] {
	if len(m) < parallelThreshold {
		return MapValues(m, fn)
	}
	type entry struct{ k K; v V }
	type result struct {
		k K
		w W
	}
	entries := make([]entry, 0, len(m))
	for k, v := range m {
		entries = append(entries, entry{k, v})
	}
	if n > len(entries) {
		n = len(entries)
	}
	results := make([]result, len(entries))
	chunkSize := (len(entries) + n - 1) / n
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		start := i * chunkSize
		if start >= len(entries) {
			break
		}
		end := start + chunkSize
		if end > len(entries) {
			end = len(entries)
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				results[j] = result{entries[j].k, fn(entries[j].k, entries[j].v)}
			}
		}(start, end)
	}
	wg.Wait()
	out := make(Map[K, W], len(m))
	for _, r := range results {
		out[r.k] = r.w
	}
	return out
}

// Invert swaps keys and values. Last key wins on duplicate values.
func Invert[K, V comparable](m Map[K, V]) Map[V, K] {
	result := make(Map[V, K], len(m))
	for k, v := range m {
		result[v] = k
	}
	return result
}

// ToSlice converts a Map to a Slice by applying fn to each entry.
func ToSlice[K comparable, V, T any](m Map[K, V], fn func(K, V) T) slice.Slice[T] {
	result := make(slice.Slice[T], 0, len(m))
	for k, v := range m {
		result = append(result, fn(k, v))
	}
	return result
}

// FromSlice builds a Map from a Slice by applying fn to each element.
func FromSlice[T any, K comparable, V any](s slice.Slice[T], fn func(T) (K, V)) Map[K, V] {
	result := make(Map[K, V], len(s))
	for _, v := range s {
		k, val := fn(v)
		result[k] = val
	}
	return result
}
