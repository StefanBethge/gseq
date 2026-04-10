package dict

import (
	"fmt"
	"testing"
)

const benchSize = 1000

func makeMap(n int) Map[int, int] {
	m := make(Map[int, int], n)
	for i := range n {
		m[i] = i * i
	}
	return m
}

// ─── Pick / Omit ──────────────────────────────────────────────────────────────

func BenchmarkPickSmall(b *testing.B) {
	m := makeMap(benchSize)
	keys := make([]int, 10)
	for i := range keys {
		keys[i] = i * 100
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		m.Pick(keys...)
	}
}

func BenchmarkPickHalf(b *testing.B) {
	m := makeMap(benchSize)
	keys := make([]int, benchSize/2)
	for i := range keys {
		keys[i] = i * 2
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		m.Pick(keys...)
	}
}

func BenchmarkOmitSmall(b *testing.B) {
	m := makeMap(benchSize)
	keys := make([]int, 10)
	for i := range keys {
		keys[i] = i * 100
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		m.Omit(keys...)
	}
}

func BenchmarkOmitHalf(b *testing.B) {
	m := makeMap(benchSize)
	keys := make([]int, benchSize/2)
	for i := range keys {
		keys[i] = i * 2
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		m.Omit(keys...)
	}
}

// ─── Merge / MergeWith ────────────────────────────────────────────────────────

func BenchmarkMergeNoConflict(b *testing.B) {
	// Disjoint key spaces
	a := makeMap(benchSize / 2)
	other := make(Map[int, int], benchSize/2)
	for i := benchSize / 2; i < benchSize; i++ {
		other[i] = i * i
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		a.Merge(other)
	}
}

func BenchmarkMergeFullConflict(b *testing.B) {
	// All keys overlap
	a := makeMap(benchSize)
	other := makeMap(benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		a.Merge(other)
	}
}

func BenchmarkMergeWithFullConflict(b *testing.B) {
	a := makeMap(benchSize)
	other := makeMap(benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		a.MergeWith(other, func(_ int, l, r int) int { return l + r })
	}
}

// Manual merge for comparison (right wins).
func BenchmarkMergeManual_ForComparison(b *testing.B) {
	a := makeMap(benchSize)
	other := makeMap(benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		result := make(Map[int, int], len(a)+len(other))
		for k, v := range a {
			result[k] = v
		}
		for k, v := range other {
			result[k] = v
		}
		_ = result
	}
}

// ─── MapKeys ──────────────────────────────────────────────────────────────────

func BenchmarkMapKeys(b *testing.B) {
	m := makeMap(benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		MapKeys(m, func(k, _ int) string { return fmt.Sprintf("%d", k) })
	}
}

// MapValues for comparison: same loop, value changes.
func BenchmarkMapValues_ForComparison(b *testing.B) {
	m := makeMap(benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		MapValues(m, func(_ int, v int) int { return v * 2 })
	}
}
