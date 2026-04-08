package slice

import (
	"runtime"
	"testing"
	"time"
)

const benchSize = 10_000

// ─── allocation benchmarks ────────────────────────────────────────────────────

func BenchmarkFilter(b *testing.B) {
	s := Range(0, benchSize)
	b.ResetTimer()
	for range b.N {
		s.Filter(func(v int) bool { return v%2 == 0 })
	}
}

func BenchmarkFlatten(b *testing.B) {
	inner := Range(0, 100)
	s := Repeat(inner, 100) // 100 slices of 100 elements
	b.ResetTimer()
	for range b.N {
		Flatten(s)
	}
}

func BenchmarkRange(b *testing.B) {
	for range b.N {
		Range(0, benchSize)
	}
}

func BenchmarkRangeStep(b *testing.B) {
	for range b.N {
		RangeStep(0, benchSize, 2)
	}
}

func BenchmarkUniq(b *testing.B) {
	s := Range(0, benchSize)
	b.ResetTimer()
	for range b.N {
		Uniq(s, func(v int) int { return v })
	}
}

func BenchmarkPartition(b *testing.B) {
	s := Range(0, benchSize)
	b.ResetTimer()
	for range b.N {
		s.Partition(func(v int) bool { return v%2 == 0 })
	}
}

// ─── Map vs MapParallel ───────────────────────────────────────────────────────

// cpuWork simulates a small CPU-bound computation.
func cpuWork(v int) int {
	sum := 0
	for i := range v % 500 {
		sum += i
	}
	return sum
}

// ioWork simulates I/O latency (e.g. a database or HTTP call).
func ioWork(v int) int {
	time.Sleep(time.Millisecond)
	return v * 2
}

func BenchmarkMap_CPUBound(b *testing.B) {
	s := Range(0, 500)
	b.ResetTimer()
	for range b.N {
		Map(s, cpuWork)
	}
}

func BenchmarkMapParallel_CPUBound(b *testing.B) {
	s := Range(0, 500)
	b.ResetTimer()
	for range b.N {
		MapParallel(s, cpuWork)
	}
}

func BenchmarkMapParallelN_CPUBound(b *testing.B) {
	s := Range(0, 500)
	workers := runtime.GOMAXPROCS(0)
	b.ResetTimer()
	for range b.N {
		MapParallelN(s, workers, cpuWork)
	}
}

func BenchmarkMap_IOBound(b *testing.B) {
	s := Range(0, 20)
	b.ResetTimer()
	for range b.N {
		Map(s, ioWork)
	}
}

func BenchmarkMapParallel_IOBound(b *testing.B) {
	s := Range(0, 20)
	b.ResetTimer()
	for range b.N {
		MapParallel(s, ioWork)
	}
}

// MapParallelN with n=len(s) for I/O-bound: every item gets its own worker
func BenchmarkMapParallelN_IOBound_MaxWorkers(b *testing.B) {
	s := Range(0, 20)
	b.ResetTimer()
	for range b.N {
		MapParallelN(s, len(s), ioWork)
	}
}
