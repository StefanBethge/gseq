// Package slice provides a generic, chainable Slice type with functional operations.
package slice

// Number encompasses all numeric types for arithmetic operations.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Slice is a generic slice type with method chaining support.
type Slice[T any] []T
