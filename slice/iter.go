package slice

import "iter"

// Iter returns an iterator over slice elements (Go 1.23+).
func (s Slice[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// Iter2 returns an index-value iterator over slice elements (Go 1.23+).
func (s Slice[T]) Iter2() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range s {
			if !yield(i, v) {
				return
			}
		}
	}
}

// Collect builds a Slice from any iter.Seq[T].
func Collect[T any](seq iter.Seq[T]) Slice[T] {
	var result Slice[T]
	for v := range seq {
		result = append(result, v)
	}
	return result
}

// FilterIter returns a lazy iterator that only yields elements satisfying fn.
func FilterIter[T any](seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if fn(v) && !yield(v) {
				return
			}
		}
	}
}

// MapIter returns a lazy iterator that transforms each element.
func MapIter[T, O any](seq iter.Seq[T], fn func(T) O) iter.Seq[O] {
	return func(yield func(O) bool) {
		for v := range seq {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

// TakeIter returns a lazy iterator limited to the first n elements.
func TakeIter[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		i := 0
		for v := range seq {
			if i >= n {
				return
			}
			if !yield(v) {
				return
			}
			i++
		}
	}
}
