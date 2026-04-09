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

// TakeWhileIter returns a lazy iterator that yields elements as long as fn
// returns true, stopping at the first element for which fn returns false.
func TakeWhileIter[T any](seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if !fn(v) {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// DropIter returns a lazy iterator that skips the first n elements.
func DropIter[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		i := 0
		for v := range seq {
			if i < n {
				i++
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

// DropWhileIter returns a lazy iterator that skips elements as long as fn
// returns true, then yields all remaining elements unchanged.
func DropWhileIter[T any](seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		dropping := true
		for v := range seq {
			if dropping {
				if fn(v) {
					continue
				}
				dropping = false
			}
			if !yield(v) {
				return
			}
		}
	}
}

// ChainIter concatenates multiple iterators into a single iterator that
// exhausts each in order.
func ChainIter[T any](seqs ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, seq := range seqs {
			for v := range seq {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// ZipIter combines two iterators pairwise using fn. Stops when either
// iterator is exhausted.
func ZipIter[A, B, O any](a iter.Seq[A], b iter.Seq[B], fn func(A, B) O) iter.Seq[O] {
	return func(yield func(O) bool) {
		nextB, stopB := iter.Pull(b)
		defer stopB()
		for va := range a {
			vb, ok := nextB()
			if !ok {
				return
			}
			if !yield(fn(va, vb)) {
				return
			}
		}
	}
}

// FlatMapIter applies fn to each element and lazily flattens the resulting
// iterators into a single sequence — the monadic bind for iter.Seq.
func FlatMapIter[T, O any](seq iter.Seq[T], fn func(T) iter.Seq[O]) iter.Seq[O] {
	return func(yield func(O) bool) {
		for v := range seq {
			for o := range fn(v) {
				if !yield(o) {
					return
				}
			}
		}
	}
}
