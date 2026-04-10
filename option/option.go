// Package option provides a generic Option[T] type representing an optional value.
package option

// Option represents an optional value: either Some(value) or None.
type Option[T any] struct {
	value T
	valid bool
}

// Some wraps a value in an Option.
func Some[T any](v T) Option[T] { return Option[T]{value: v, valid: true} }

// None returns an empty Option.
func None[T any]() Option[T] { return Option[T]{} }

// IsSome returns true if the Option contains a value.
func (o Option[T]) IsSome() bool { return o.valid }

// IsNone returns true if the Option is empty.
func (o Option[T]) IsNone() bool { return !o.valid }

// Unwrap returns the value, panicking if the Option is None.
func (o Option[T]) Unwrap() T {
	if !o.valid {
		panic("option: Unwrap called on None")
	}
	return o.value
}

// UnwrapOr returns the value, or fallback if the Option is None.
func (o Option[T]) UnwrapOr(fallback T) T {
	if !o.valid {
		return fallback
	}
	return o.value
}

// Get returns the underlying (value, ok) pair for use in if-assignments.
func (o Option[T]) Get() (T, bool) {
	return o.value, o.valid
}

// Filter keeps Some(v) only if fn(v) is true, otherwise returns None.
func (o Option[T]) Filter(fn func(T) bool) Option[T] {
	if o.valid && fn(o.value) {
		return o
	}
	// Return the zero Option directly — avoids a heap allocation vs None[T]().
	var zero Option[T]
	return zero
}

// Or returns the Option itself if it is Some, otherwise returns other.
func (o Option[T]) Or(other Option[T]) Option[T] {
	if o.valid {
		return o
	}
	return other
}

// OrElse is the lazy variant of Or: calls fn only if the Option is None.
func (o Option[T]) OrElse(fn func() Option[T]) Option[T] {
	if o.valid {
		return o
	}
	return fn()
}

// ─── FREE FUNCTIONS ───────────────────────────────────────────────────────────

// Map transforms the value inside an Option using fn.
// Returns None if the Option is None.
func Map[T, U any](o Option[T], fn func(T) U) Option[U] {
	if o.IsNone() {
		return None[U]()
	}
	return Some(fn(o.value))
}

// FlatMap applies fn to the value inside an Option, flattening the result.
// Returns None if the Option is None or fn returns None.
func FlatMap[T, U any](o Option[T], fn func(T) Option[U]) Option[U] {
	if o.IsNone() {
		return None[U]()
	}
	return fn(o.value)
}

// Coalesce returns the first Some value among the given options.
// Returns None if all options are None or no options are provided.
func Coalesce[T any](opts ...Option[T]) Option[T] {
	for _, o := range opts {
		if o.IsSome() {
			return o
		}
	}
	return None[T]()
}
