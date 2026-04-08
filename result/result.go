// Package result provides a generic Result[T, E] type for explicit error handling.
package result

import (
	"github.com/stefanbethge/gseq/option"
	"github.com/stefanbethge/gseq/slice"
)

// Result represents either a successful value Ok(T) or an error Err(E).
type Result[T, E any] struct {
	value T
	err   E
	ok    bool
}

// Ok wraps a success value in a Result.
func Ok[T, E any](v T) Result[T, E] { return Result[T, E]{value: v, ok: true} }

// Err wraps an error value in a Result.
func Err[T, E any](e E) Result[T, E] { return Result[T, E]{err: e} }

// FromGoError wraps a standard Go (value, error) pair into a Result.
// If err is nil the result is Ok(v), otherwise Err(err).
func FromGoError[T any](v T, err error) Result[T, error] {
	if err != nil {
		return Err[T, error](err)
	}
	return Ok[T, error](v)
}

// IsOk returns true if the Result contains a success value.
func (r Result[T, E]) IsOk() bool { return r.ok }

// IsErr returns true if the Result contains an error.
func (r Result[T, E]) IsErr() bool { return !r.ok }

// Unwrap returns the success value, panicking if the Result is Err.
func (r Result[T, E]) Unwrap() T {
	if !r.ok {
		panic("result: Unwrap called on Err")
	}
	return r.value
}

// UnwrapErr returns the error value, panicking if the Result is Ok.
func (r Result[T, E]) UnwrapErr() E {
	if r.ok {
		panic("result: UnwrapErr called on Ok")
	}
	return r.err
}

// UnwrapOr returns the success value, or fallback if the Result is Err.
func (r Result[T, E]) UnwrapOr(fallback T) T {
	if !r.ok {
		return fallback
	}
	return r.value
}

// ToOption converts the Result to an Option, discarding any error.
func (r Result[T, E]) ToOption() option.Option[T] {
	if r.ok {
		return option.Some(r.value)
	}
	return option.None[T]()
}

// ─── FREE FUNCTIONS ───────────────────────────────────────────────────────────

// Map transforms the success value of a Result using fn.
// Propagates Err unchanged.
func Map[T, U, E any](r Result[T, E], fn func(T) U) Result[U, E] {
	if r.IsErr() {
		return Err[U, E](r.err)
	}
	return Ok[U, E](fn(r.value))
}

// FlatMap applies fn to the success value, flattening the nested Result.
// Propagates Err unchanged.
func FlatMap[T, U, E any](r Result[T, E], fn func(T) Result[U, E]) Result[U, E] {
	if r.IsErr() {
		return Err[U, E](r.err)
	}
	return fn(r.value)
}

// MapErr transforms the error value of a Result using fn.
// Propagates Ok unchanged.
func MapErr[T, E, F any](r Result[T, E], fn func(E) F) Result[T, F] {
	if r.IsOk() {
		return Ok[T, F](r.value)
	}
	return Err[T, F](fn(r.err))
}

// Partition splits a slice of Results into a slice of success values and a
// slice of errors. Order within each output slice matches the input order.
func Partition[T, E any](s slice.Slice[Result[T, E]]) (slice.Slice[T], slice.Slice[E]) {
	var oks slice.Slice[T]
	var errs slice.Slice[E]
	for _, r := range s {
		if r.ok {
			oks = append(oks, r.value)
		} else {
			errs = append(errs, r.err)
		}
	}
	return oks, errs
}
