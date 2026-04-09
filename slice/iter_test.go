package slice

import (
	"iter"
	"testing"
)

func TestIter(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	var collected []int
	for v := range s.Iter() {
		collected = append(collected, v)
	}
	assertEqual(t, []int{1, 2, 3, 4, 5}, collected)
}

func TestIterEarlyStop(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	var collected []int
	for v := range s.Iter() {
		collected = append(collected, v)
		if v == 3 {
			break
		}
	}
	assertEqual(t, []int{1, 2, 3}, collected)
}

func TestIter2(t *testing.T) {
	s := Slice[string]{"a", "b", "c"}
	var indices []int
	var values []string
	for i, v := range s.Iter2() {
		indices = append(indices, i)
		values = append(values, v)
	}
	assertEqual(t, []int{0, 1, 2}, indices)
	assertEqual(t, []string{"a", "b", "c"}, values)
}

func TestCollect(t *testing.T) {
	s := Slice[int]{10, 20, 30}
	assertEqual(t, Slice[int]{10, 20, 30}, Collect(s.Iter()))
}

func TestFilterIter(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	got := Collect(FilterIter(s.Iter(), func(v int) bool { return v%2 == 0 }))
	assertEqual(t, Slice[int]{2, 4}, got)
}

func TestMapIter(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	got := Collect(MapIter(s.Iter(), func(v int) int { return v * 10 }))
	assertEqual(t, Slice[int]{10, 20, 30}, got)
}

func TestTakeIter(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	assertEqual(t, Slice[int]{1, 2, 3}, Collect(TakeIter(s.Iter(), 3)))
}

func TestTakeIterBeyondLen(t *testing.T) {
	s := Slice[int]{1, 2}
	assertEqual(t, Slice[int]{1, 2}, Collect(TakeIter(s.Iter(), 10)))
}

func TestIterChaining(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	got := Collect(
		TakeIter(
			MapIter(
				FilterIter(s.Iter(), func(v int) bool { return v%2 == 0 }),
				func(v int) int { return v * v },
			),
			3,
		),
	)
	assertEqual(t, Slice[int]{4, 16, 36}, got)
}

func TestTakeWhileIter(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	got := Collect(TakeWhileIter(s.Iter(), func(v int) bool { return v < 4 }))
	assertEqual(t, Slice[int]{1, 2, 3}, got)
}

func TestTakeWhileIterAll(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	got := Collect(TakeWhileIter(s.Iter(), func(v int) bool { return true }))
	assertEqual(t, Slice[int]{1, 2, 3}, got)
}

func TestTakeWhileIterNone(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	got := Collect(TakeWhileIter(s.Iter(), func(v int) bool { return false }))
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestDropIter(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	got := Collect(DropIter(s.Iter(), 2))
	assertEqual(t, Slice[int]{3, 4, 5}, got)
}

func TestDropIterBeyondLen(t *testing.T) {
	s := Slice[int]{1, 2}
	got := Collect(DropIter(s.Iter(), 10))
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestDropWhileIter(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	got := Collect(DropWhileIter(s.Iter(), func(v int) bool { return v < 3 }))
	assertEqual(t, Slice[int]{3, 4, 5}, got)
}

func TestDropWhileIterKeepsAfterFirstFalse(t *testing.T) {
	// 1→drop, 2→drop, 3→keep, 2→keep (doesn't re-apply predicate)
	s := Slice[int]{1, 2, 3, 2, 1}
	got := Collect(DropWhileIter(s.Iter(), func(v int) bool { return v < 3 }))
	assertEqual(t, Slice[int]{3, 2, 1}, got)
}

func TestChainIter(t *testing.T) {
	a := Slice[int]{1, 2, 3}
	b := Slice[int]{4, 5}
	c := Slice[int]{6}
	got := Collect(ChainIter(a.Iter(), b.Iter(), c.Iter()))
	assertEqual(t, Slice[int]{1, 2, 3, 4, 5, 6}, got)
}

func TestChainIterEmpty(t *testing.T) {
	got := Collect(ChainIter(Slice[int]{}.Iter(), Slice[int]{1}.Iter()))
	assertEqual(t, Slice[int]{1}, got)
}

func TestZipIter(t *testing.T) {
	a := Slice[int]{1, 2, 3}
	b := Slice[string]{"a", "b", "c"}
	type pair struct{ N int; S string }
	got := Collect(ZipIter(a.Iter(), b.Iter(), func(n int, s string) pair { return pair{n, s} }))
	assertEqual(t, Slice[pair]{{1, "a"}, {2, "b"}, {3, "c"}}, got)
}

func TestZipIterUnequalLength(t *testing.T) {
	a := Slice[int]{1, 2, 3}
	b := Slice[int]{10, 20}
	got := Collect(ZipIter(a.Iter(), b.Iter(), func(x, y int) int { return x + y }))
	assertEqual(t, Slice[int]{11, 22}, got)
}

func TestFlatMapIter(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	got := Collect(FlatMapIter(s.Iter(), func(v int) iter.Seq[int] {
		return Slice[int]{v, v * 10}.Iter()
	}))
	assertEqual(t, Slice[int]{1, 10, 2, 20, 3, 30}, got)
}

func TestFlatMapIterEarlyStop(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	got := Collect(TakeIter(FlatMapIter(s.Iter(), func(v int) iter.Seq[int] {
		return Slice[int]{v, v * 10}.Iter()
	}), 3))
	assertEqual(t, Slice[int]{1, 10, 2}, got)
}
