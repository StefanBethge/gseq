package slice

import (
	"cmp"
	"math/rand"
	"runtime"
	"slices"
	"sync"

	"github.com/stefanbethge/gseq/option"
)

// ─── CHAINABLE METHODS ────────────────────────────────────────────────────────

// Filter keeps all elements for which fn returns true.
func (s Slice[T]) Filter(fn func(T) bool) Slice[T] {
	result := make(Slice[T], 0, len(s))
	for _, v := range s {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reject is the inverse of Filter — keeps elements where fn returns false.
func (s Slice[T]) Reject(fn func(T) bool) Slice[T] {
	return s.Filter(func(v T) bool { return !fn(v) })
}

// Limit returns the first n elements.
func (s Slice[T]) Limit(n int) Slice[T] {
	if n >= len(s) {
		return s
	}
	return s[:n]
}

// Skip skips the first n elements.
func (s Slice[T]) Skip(n int) Slice[T] {
	if n >= len(s) {
		return Slice[T]{}
	}
	return s[n:]
}

// Reverse returns a copy with elements in reversed order.
func (s Slice[T]) Reverse() Slice[T] {
	result := make(Slice[T], len(s))
	for i, v := range s {
		result[len(s)-1-i] = v
	}
	return result
}

// Shuffle returns a randomly shuffled copy.
func (s Slice[T]) Shuffle() Slice[T] {
	result := make(Slice[T], len(s))
	copy(result, s)
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return result
}

// Chunk splits the slice into groups of size n.
func (s Slice[T]) Chunk(n int) []Slice[T] {
	result := make([]Slice[T], 0, (len(s)+n-1)/n)
	for i := 0; i < len(s); i += n {
		end := i + n
		if end > len(s) {
			end = len(s)
		}
		result = append(result, s[i:end])
	}
	return result
}

// Without removes all elements for which eq(elem, e) is true for any e.
func (s Slice[T]) Without(eq func(T, T) bool, elems ...T) Slice[T] {
	return s.Filter(func(v T) bool {
		for _, e := range elems {
			if eq(v, e) {
				return false
			}
		}
		return true
	})
}

// SortBy returns a sorted copy using the provided less function.
func (s Slice[T]) SortBy(less func(a, b T) bool) Slice[T] {
	result := make(Slice[T], len(s))
	copy(result, s)
	slices.SortFunc(result, func(a, b T) int {
		if less(a, b) {
			return -1
		}
		if less(b, a) {
			return 1
		}
		return 0
	})
	return result
}

// Sort returns a sorted copy of a slice of ordered values.
func Sort[T cmp.Ordered](s Slice[T]) Slice[T] {
	result := make(Slice[T], len(s))
	copy(result, s)
	slices.Sort(result)
	return result
}

// Window returns all contiguous sub-slices of the given size.
func (s Slice[T]) Window(size int) []Slice[T] {
	if size <= 0 || size > len(s) {
		return nil
	}
	result := make([]Slice[T], len(s)-size+1)
	for i := range result {
		result[i] = s[i : i+size]
	}
	return result
}

// Pairwise is shorthand for Window(2).
func (s Slice[T]) Pairwise() []Slice[T] {
	return s.Window(2)
}

// Each calls fn for every element.
func (s Slice[T]) Each(fn func(T)) {
	for _, v := range s {
		fn(v)
	}
}

// EachIndexed calls fn(index, value) for every element.
func (s Slice[T]) EachIndexed(fn func(int, T)) {
	for i, v := range s {
		fn(i, v)
	}
}

// EachParallel calls fn for every element using a worker pool sized to
// GOMAXPROCS. For I/O-bound work that benefits from higher concurrency use
// EachParallelN to set the worker count explicitly.
func (s Slice[T]) EachParallel(fn func(T)) {
	s.EachParallelN(runtime.GOMAXPROCS(0), fn)
}

// EachParallelN calls fn for every element using a pool of n workers.
// Falls back to sequential Each when len(s) < parallelThreshold.
func (s Slice[T]) EachParallelN(n int, fn func(T)) {
	if len(s) < parallelThreshold {
		s.Each(fn)
		return
	}
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
				fn(s[j])
			}
		}(start, end)
	}
	wg.Wait()
}

// EachParallelIndexed calls fn(index, value) for every element using a worker
// pool sized to GOMAXPROCS.
func (s Slice[T]) EachParallelIndexed(fn func(int, T)) {
	s.EachParallelIndexedN(runtime.GOMAXPROCS(0), fn)
}

// EachParallelIndexedN calls fn(index, value) for every element using a pool
// of n workers. Falls back to sequential EachIndexed when len(s) < parallelThreshold.
func (s Slice[T]) EachParallelIndexedN(n int, fn func(int, T)) {
	if len(s) < parallelThreshold {
		s.EachIndexed(fn)
		return
	}
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
				fn(j, s[j])
			}
		}(start, end)
	}
	wg.Wait()
}

// ─── TERMINATORS ──────────────────────────────────────────────────────────────

// First returns the first element, or None if the slice is empty.
func (s Slice[T]) First() option.Option[T] {
	if len(s) == 0 {
		return option.None[T]()
	}
	return option.Some(s[0])
}

// Last returns the last element, or None if the slice is empty.
func (s Slice[T]) Last() option.Option[T] {
	if len(s) == 0 {
		return option.None[T]()
	}
	return option.Some(s[len(s)-1])
}

// Nth returns the element at position n, or None if out of bounds.
func (s Slice[T]) Nth(n int) option.Option[T] {
	if n < 0 || n >= len(s) {
		return option.None[T]()
	}
	return option.Some(s[n])
}

// Sample returns a random element, or None if the slice is empty.
func (s Slice[T]) Sample() option.Option[T] {
	if len(s) == 0 {
		return option.None[T]()
	}
	return option.Some(s[rand.Intn(len(s))])
}

// Samples returns n random elements without repetition.
func (s Slice[T]) Samples(n int) Slice[T] {
	return s.Shuffle().Limit(n)
}

// Contains returns true if at least one element satisfies fn.
func (s Slice[T]) Contains(fn func(T) bool) bool {
	for _, v := range s {
		if fn(v) {
			return true
		}
	}
	return false
}

// Every returns true if ALL elements satisfy fn.
func (s Slice[T]) Every(fn func(T) bool) bool {
	for _, v := range s {
		if !fn(v) {
			return false
		}
	}
	return true
}

// None returns true if NO element satisfies fn.
func (s Slice[T]) None(fn func(T) bool) bool {
	return !s.Contains(fn)
}

// Count counts how many elements satisfy fn.
func (s Slice[T]) Count(fn func(T) bool) int {
	n := 0
	for _, v := range s {
		if fn(v) {
			n++
		}
	}
	return n
}

// IndexOf returns the index of the first element satisfying fn, or -1 if not found.
func (s Slice[T]) IndexOf(fn func(T) bool) int {
	for i, v := range s {
		if fn(v) {
			return i
		}
	}
	return -1
}

// Find returns the first element satisfying fn, or None if not found.
func (s Slice[T]) Find(fn func(T) bool) option.Option[T] {
	for _, v := range s {
		if fn(v) {
			return option.Some(v)
		}
	}
	return option.None[T]()
}

// Partition splits into two slices: (fn=true, fn=false).
func (s Slice[T]) Partition(fn func(T) bool) (Slice[T], Slice[T]) {
	// Pre-allocate len(s) for each side. Worst case one side holds all elements;
	// this avoids reallocation regardless of how skewed the split is.
	yes := make(Slice[T], 0, len(s))
	no := make(Slice[T], 0, len(s))
	for _, v := range s {
		if fn(v) {
			yes = append(yes, v)
		} else {
			no = append(no, v)
		}
	}
	return yes, no
}

// Len returns the number of elements.
func (s Slice[T]) Len() int { return len(s) }

// IsEmpty returns true if the slice has no elements.
func (s Slice[T]) IsEmpty() bool { return len(s) == 0 }

// ToSlice converts back to a plain Go slice.
func (s Slice[T]) ToSlice() []T { return []T(s) }
