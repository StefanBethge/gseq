package slice

import (
	"runtime"
	"sync"

	"github.com/stefanbethge/gseq/option"
)

// Repeat creates a slice with elem repeated n times.
func Repeat[T any](elem T, n int) Slice[T] {
	result := make(Slice[T], n)
	for i := range result {
		result[i] = elem
	}
	return result
}

// Range creates a slice from start to end (exclusive).
func Range[T Number](start, end T) Slice[T] {
	if end <= start {
		return nil
	}
	result := make(Slice[T], 0, int(end-start))
	for i := start; i < end; i++ {
		result = append(result, i)
	}
	return result
}

// RangeStep is like Range but with a custom step size.
func RangeStep[T Number](start, end, step T) Slice[T] {
	if end <= start || step <= 0 {
		return nil
	}
	result := make(Slice[T], 0, int((end-start)/step)+1)
	for i := start; i < end; i += step {
		result = append(result, i)
	}
	return result
}

// Sum returns the sum of all elements.
func Sum[T Number](s Slice[T]) T {
	var total T
	for _, v := range s {
		total += v
	}
	return total
}

// SumParallel is like Sum but runs concurrently using a worker pool sized to
// GOMAXPROCS. Each worker sums its segment into a local variable; partial sums
// are combined sequentially after all workers finish. No mutex required.
// Falls back to sequential Sum when len(s) < parallelThreshold.
func SumParallel[T Number](s Slice[T]) T {
	return SumParallelN(s, runtime.GOMAXPROCS(0))
}

// SumParallelN is like SumParallel but uses exactly n workers.
func SumParallelN[T Number](s Slice[T], n int) T {
	if len(s) < parallelThreshold {
		return Sum(s)
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
			var local T
			for j := start; j < end; j++ {
				local += s[j]
			}
			partials[i] = local
		}(i, start, end)
	}
	wg.Wait()
	var total T
	for _, p := range partials {
		total += p
	}
	return total
}

// Min returns the smallest element, or None if the slice is empty.
func Min[T Number](s Slice[T]) option.Option[T] {
	if len(s) == 0 {
		return option.None[T]()
	}
	m := s[0]
	for _, v := range s[1:] {
		if v < m {
			m = v
		}
	}
	return option.Some(m)
}

// Max returns the largest element, or None if the slice is empty.
func Max[T Number](s Slice[T]) option.Option[T] {
	if len(s) == 0 {
		return option.None[T]()
	}
	m := s[0]
	for _, v := range s[1:] {
		if v > m {
			m = v
		}
	}
	return option.Some(m)
}

// MinBy returns the element for which fn yields the smallest value, or None if empty.
func MinBy[T any, N Number](s Slice[T], fn func(T) N) option.Option[T] {
	if len(s) == 0 {
		return option.None[T]()
	}
	best, bestVal := s[0], fn(s[0])
	for _, v := range s[1:] {
		if val := fn(v); val < bestVal {
			best, bestVal = v, val
		}
	}
	return option.Some(best)
}

// MaxBy returns the element for which fn yields the largest value, or None if empty.
func MaxBy[T any, N Number](s Slice[T], fn func(T) N) option.Option[T] {
	if len(s) == 0 {
		return option.None[T]()
	}
	best, bestVal := s[0], fn(s[0])
	for _, v := range s[1:] {
		if val := fn(v); val > bestVal {
			best, bestVal = v, val
		}
	}
	return option.Some(best)
}

// SumBy sums a derived value from each element.
func SumBy[T any, N Number](s Slice[T], fn func(T) N) N {
	var total N
	for _, v := range s {
		total += fn(v)
	}
	return total
}

// SumByParallel is like SumBy but runs concurrently using GOMAXPROCS workers.
// Falls back to sequential SumBy when len(s) < parallelThreshold.
func SumByParallel[T any, N Number](s Slice[T], fn func(T) N) N {
	return SumByParallelN(s, runtime.GOMAXPROCS(0), fn)
}

// SumByParallelN is like SumByParallel but uses exactly n workers.
func SumByParallelN[T any, N Number](s Slice[T], n int, fn func(T) N) N {
	if len(s) < parallelThreshold {
		return SumBy(s, fn)
	}
	if n > len(s) {
		n = len(s)
	}
	chunkSize := (len(s) + n - 1) / n
	partials := make([]N, n)
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
			var local N
			for j := start; j < end; j++ {
				local += fn(s[j])
			}
			partials[i] = local
		}(i, start, end)
	}
	wg.Wait()
	var total N
	for _, p := range partials {
		total += p
	}
	return total
}
