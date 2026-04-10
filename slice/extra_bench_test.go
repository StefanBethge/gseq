package slice

import (
	"errors"
	"testing"
)

// ─── TakeWhile / DropWhile ────────────────────────────────────────────────────

func BenchmarkTakeWhileHalf(b *testing.B) {
	s := Range(0, benchSize) // takes first half
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		s.TakeWhile(func(v int) bool { return v < benchSize/2 })
	}
}

func BenchmarkTakeWhileAll(b *testing.B) {
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		s.TakeWhile(func(v int) bool { return v < benchSize })
	}
}

func BenchmarkTakeWhileNone(b *testing.B) {
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		s.TakeWhile(func(v int) bool { return v < 0 })
	}
}

// Compare: Filter achieves the same as TakeWhile-half but scans the full slice.
func BenchmarkFilterHalf_ForComparison(b *testing.B) {
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		s.Filter(func(v int) bool { return v < benchSize/2 })
	}
}

func BenchmarkDropWhileHalf(b *testing.B) {
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		s.DropWhile(func(v int) bool { return v < benchSize/2 })
	}
}

func BenchmarkDropWhileAll(b *testing.B) {
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		s.DropWhile(func(v int) bool { return v < benchSize })
	}
}

// ─── ChunkBy ──────────────────────────────────────────────────────────────────

func BenchmarkChunkByLowCardinality(b *testing.B) {
	// 2 distinct keys → long runs, few chunks
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		ChunkBy(s, func(v int) int { return v % 2 })
	}
}

func BenchmarkChunkByHighCardinality(b *testing.B) {
	// Every element is its own chunk (worst case allocation count)
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		ChunkBy(s, func(v int) int { return v })
	}
}

// GroupBy for comparison: same key function, different semantic (order-insensitive).
func BenchmarkGroupBy_ForComparison(b *testing.B) {
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		GroupBy(s, func(v int) int { return v % 2 })
	}
}

// ─── Scan ─────────────────────────────────────────────────────────────────────

func BenchmarkScan(b *testing.B) {
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Scan(s, 0, func(acc, v int) int { return acc + v })
	}
}

// Reduce for comparison: same work, no intermediate slice.
func BenchmarkReduce_ForComparison(b *testing.B) {
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Reduce(s, 0, func(acc, v int) int { return acc + v })
	}
}

// ─── Associate ────────────────────────────────────────────────────────────────

func BenchmarkAssociate(b *testing.B) {
	type kv struct{ k, v int }
	s := Map(Range(0, benchSize), func(i int) kv { return kv{i, i * i} })
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Associate(s, func(p kv) (int, int) { return p.k, p.v })
	}
}

// KeyBy for comparison: similar map-building, value == element.
func BenchmarkKeyBy_ForComparison(b *testing.B) {
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		KeyBy(s, func(v int) int { return v })
	}
}

// ─── Interleave ───────────────────────────────────────────────────────────────

func BenchmarkInterleave2(b *testing.B) {
	a := Range(0, benchSize/2)
	c := Range(benchSize/2, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Interleave(a, c)
	}
}

func BenchmarkInterleave4(b *testing.B) {
	quarter := benchSize / 4
	slices := make([]Slice[int], 4)
	for i := range slices {
		slices[i] = Range(i*quarter, (i+1)*quarter)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Interleave(slices...)
	}
}

func BenchmarkInterleave8(b *testing.B) {
	eighth := benchSize / 8
	slices := make([]Slice[int], 8)
	for i := range slices {
		slices[i] = Range(i*eighth, (i+1)*eighth)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Interleave(slices...)
	}
}

// ─── TryReduce ────────────────────────────────────────────────────────────────

func BenchmarkTryReduceNoError(b *testing.B) {
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		TryReduce(s, 0, func(acc, v int) (int, error) { return acc + v, nil })
	}
}

// Reduce for comparison: no error return path overhead.
func BenchmarkReduce_ForTryComparison(b *testing.B) {
	s := Range(0, benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Reduce(s, 0, func(acc, v int) int { return acc + v })
	}
}

func BenchmarkTryReduceEarlyExit(b *testing.B) {
	// Error at index benchSize/2 — exercises early-exit path.
	s := Range(0, benchSize)
	half := benchSize / 2
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		TryReduce(s, 0, func(acc, v int) (int, error) {
			if v == half {
				return acc, errors.New("stop")
			}
			return acc + v, nil
		})
	}
}
