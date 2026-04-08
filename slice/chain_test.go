package slice

import (
	"fmt"
	"sort"
	"sync/atomic"
	"testing"
)

func TestFilter(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	assertEqual(t, Slice[int]{2, 4}, s.Filter(func(v int) bool { return v%2 == 0 }))
}

func TestFilterEmpty(t *testing.T) {
	s := Slice[int]{}
	if len(s.Filter(func(int) bool { return true })) != 0 {
		t.Fatal("expected empty")
	}
}

func TestReject(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	assertEqual(t, Slice[int]{1, 3, 5}, s.Reject(func(v int) bool { return v%2 == 0 }))
}

func TestLimit(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	assertEqual(t, Slice[int]{1, 2, 3}, s.Limit(3))
	assertEqual(t, s, s.Limit(10))
	if len(Slice[int]{}.Limit(3)) != 0 {
		t.Fatal("expected empty")
	}
}

func TestSkip(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	assertEqual(t, Slice[int]{3, 4, 5}, s.Skip(2))
	assertEqual(t, s, s.Skip(0))
	if len(s.Skip(10)) != 0 {
		t.Fatal("expected empty after skip beyond length")
	}
}

func TestReverse(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	assertEqual(t, Slice[int]{3, 2, 1}, s.Reverse())
	assertEqual(t, Slice[int]{1, 2, 3}, s)
}

func TestChunk(t *testing.T) {
	chunks := Slice[int]{1, 2, 3, 4, 5}.Chunk(2)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	assertEqual(t, Slice[int]{1, 2}, chunks[0])
	assertEqual(t, Slice[int]{3, 4}, chunks[1])
	assertEqual(t, Slice[int]{5}, chunks[2])
}

func TestChunkExact(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4}
	if len(s.Chunk(2)) != 2 {
		t.Fatal("expected 2 chunks")
	}
}

func TestWithout(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	eq := func(a, b int) bool { return a == b }
	assertEqual(t, Slice[int]{1, 3, 5}, s.Without(eq, 2, 4))
}

func TestSortBy(t *testing.T) {
	s := Slice[int]{3, 1, 4, 1, 5, 9, 2}
	got := s.SortBy(func(a, b int) bool { return a < b })
	assertEqual(t, Slice[int]{1, 1, 2, 3, 4, 5, 9}, got)
	assertEqual(t, Slice[int]{3, 1, 4, 1, 5, 9, 2}, s)
}

func TestSortByDescending(t *testing.T) {
	s := Slice[int]{3, 1, 4, 1, 5}
	assertEqual(t, Slice[int]{5, 4, 3, 1, 1}, s.SortBy(func(a, b int) bool { return a > b }))
}

func TestSortByStruct(t *testing.T) {
	type Person struct{ Name string; Age int }
	s := Slice[Person]{{"Bob", 30}, {"Alice", 25}, {"Carol", 35}}
	got := s.SortBy(func(a, b Person) bool { return a.Age < b.Age })
	assertEqual(t, Slice[Person]{{"Alice", 25}, {"Bob", 30}, {"Carol", 35}}, got)
}

func TestSort(t *testing.T) {
	s := Slice[int]{5, 2, 8, 1, 9}
	assertEqual(t, Slice[int]{1, 2, 5, 8, 9}, Sort(s))
	assertEqual(t, Slice[int]{5, 2, 8, 1, 9}, s)
}

func TestSortStrings(t *testing.T) {
	s := Slice[string]{"banana", "apple", "cherry"}
	assertEqual(t, Slice[string]{"apple", "banana", "cherry"}, Sort(s))
}

func TestWindow(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	windows := s.Window(3)
	if len(windows) != 3 {
		t.Fatalf("expected 3 windows, got %d", len(windows))
	}
	assertEqual(t, Slice[int]{1, 2, 3}, windows[0])
	assertEqual(t, Slice[int]{2, 3, 4}, windows[1])
	assertEqual(t, Slice[int]{3, 4, 5}, windows[2])
}

func TestWindowSizeEqualsLen(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	windows := s.Window(3)
	if len(windows) != 1 {
		t.Fatalf("expected 1 window, got %d", len(windows))
	}
}

func TestWindowTooLarge(t *testing.T) {
	s := Slice[int]{1, 2}
	if s.Window(5) != nil {
		t.Fatal("expected nil for window size > len")
	}
}

func TestPairwise(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4}
	pairs := s.Pairwise()
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}
	assertEqual(t, Slice[int]{1, 2}, pairs[0])
	assertEqual(t, Slice[int]{2, 3}, pairs[1])
	assertEqual(t, Slice[int]{3, 4}, pairs[2])
}

func TestEach(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	var collected []int
	s.Each(func(v int) { collected = append(collected, v) })
	assertEqual(t, []int{1, 2, 3}, collected)
}

func TestEachIndexed(t *testing.T) {
	s := Slice[string]{"a", "b", "c"}
	var got []string
	s.EachIndexed(func(i int, v string) { got = append(got, fmt.Sprintf("%d:%s", i, v)) })
	assertEqual(t, []string{"0:a", "1:b", "2:c"}, got)
}

func TestEachParallel(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	var count atomic.Int32
	s.EachParallel(func(int) { count.Add(1) })
	if int(count.Load()) != len(s) {
		t.Fatalf("expected %d calls, got %d", len(s), count.Load())
	}
}

func TestEachParallelIndexed(t *testing.T) {
	s := Slice[int]{10, 20, 30}
	var count atomic.Int32
	s.EachParallelIndexed(func(i int, v int) { count.Add(1) })
	if int(count.Load()) != len(s) {
		t.Fatalf("expected %d calls, got %d", len(s), count.Load())
	}
}

func TestShuffle(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	shuffled := s.Shuffle()
	if len(shuffled) != len(s) {
		t.Fatalf("length mismatch after shuffle")
	}
	assertEqual(t, Slice[int]{1, 2, 3, 4, 5}, s)
	sorted := shuffled.ToSlice()
	sort.Ints(sorted)
	assertEqual(t, []int{1, 2, 3, 4, 5}, sorted)
}

func TestFirst(t *testing.T) {
	s := Slice[int]{10, 20}
	opt := s.First()
	if opt.IsNone() || opt.Unwrap() != 10 {
		t.Fatalf("expected Some(10), got %v", opt)
	}
	empty := Slice[int]{}
	if empty.First().IsSome() {
		t.Fatal("expected None for empty slice")
	}
}

func TestLast(t *testing.T) {
	s := Slice[int]{10, 20}
	opt := s.Last()
	if opt.IsNone() || opt.Unwrap() != 20 {
		t.Fatalf("expected Some(20), got %v", opt)
	}
	empty := Slice[int]{}
	if empty.Last().IsSome() {
		t.Fatal("expected None for empty slice")
	}
}

func TestNth(t *testing.T) {
	s := Slice[string]{"a", "b", "c"}
	if v := s.Nth(1).Unwrap(); v != "b" {
		t.Fatalf("expected b, got %v", v)
	}
	if s.Nth(-1).IsSome() {
		t.Fatal("expected None for negative index")
	}
	if s.Nth(10).IsSome() {
		t.Fatal("expected None for out-of-bounds")
	}
}

func TestSample(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	opt := s.Sample()
	if opt.IsNone() {
		t.Fatal("expected Some")
	}
	v := opt.Unwrap()
	if !s.Contains(func(x int) bool { return x == v }) {
		t.Fatalf("sample %d not in slice", v)
	}
	empty := Slice[int]{}
	if empty.Sample().IsSome() {
		t.Fatal("expected None for empty slice")
	}
}

func TestSamples(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	got := s.Samples(3)
	if len(got) != 3 {
		t.Fatalf("expected 3 samples, got %d", len(got))
	}
	for _, v := range got {
		if !s.Contains(func(x int) bool { return x == v }) {
			t.Fatalf("sample %d not in original slice", v)
		}
	}
}

func TestContains(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	if !s.Contains(func(v int) bool { return v == 2 }) {
		t.Fatal("expected true")
	}
	if s.Contains(func(v int) bool { return v == 99 }) {
		t.Fatal("expected false")
	}
}

func TestEvery(t *testing.T) {
	allEven := Slice[int]{2, 4, 6}
	mixed := Slice[int]{2, 3, 6}
	if !allEven.Every(func(v int) bool { return v%2 == 0 }) {
		t.Fatal("expected true")
	}
	if mixed.Every(func(v int) bool { return v%2 == 0 }) {
		t.Fatal("expected false")
	}
}

func TestNoneMethod(t *testing.T) {
	allOdd := Slice[int]{1, 3, 5}
	hasEven := Slice[int]{1, 2, 5}
	if !allOdd.None(func(v int) bool { return v%2 == 0 }) {
		t.Fatal("expected true")
	}
	if hasEven.None(func(v int) bool { return v%2 == 0 }) {
		t.Fatal("expected false")
	}
}

func TestCount(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	if s.Count(func(v int) bool { return v%2 == 0 }) != 2 {
		t.Fatal("expected count 2")
	}
}

func TestIndexOf(t *testing.T) {
	s := Slice[string]{"a", "b", "c"}
	if s.IndexOf(func(v string) bool { return v == "b" }) != 1 {
		t.Fatal("expected index 1")
	}
	if s.IndexOf(func(v string) bool { return v == "z" }) != -1 {
		t.Fatal("expected -1")
	}
}

func TestFind(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	opt := s.Find(func(v int) bool { return v > 1 })
	if opt.IsNone() || opt.Unwrap() != 2 {
		t.Fatalf("expected Some(2), got %v", opt)
	}
	if s.Find(func(v int) bool { return v > 99 }).IsSome() {
		t.Fatal("expected None")
	}
}

func TestPartition(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	yes, no := s.Partition(func(v int) bool { return v%2 == 0 })
	assertEqual(t, Slice[int]{2, 4}, yes)
	assertEqual(t, Slice[int]{1, 3, 5}, no)
}

func TestLen(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	if s.Len() != 3 {
		t.Fatal("expected Len 3")
	}
}

func TestIsEmpty(t *testing.T) {
	full := Slice[int]{1}
	empty := Slice[int]{}
	if full.IsEmpty() {
		t.Fatal("expected not empty")
	}
	if !empty.IsEmpty() {
		t.Fatal("expected empty")
	}
}

func TestToSlice(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	assertEqual(t, []int{1, 2, 3}, s.ToSlice())
}
