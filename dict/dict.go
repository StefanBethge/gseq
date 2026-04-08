// Package dict provides a generic, chainable Map type for key-value operations.
package dict

import "github.com/stefanbethge/gseq/slice"

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
