package slice

import "github.com/stefanbethge/gseq/option"

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
