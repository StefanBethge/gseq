package slice

import "testing"

func TestRepeat(t *testing.T) {
	assertEqual(t, Slice[string]{"x", "x", "x", "x"}, Repeat("x", 4))
}

func TestRepeatZero(t *testing.T) {
	if len(Repeat(1, 0)) != 0 {
		t.Fatal("expected empty")
	}
}

func TestRange(t *testing.T) {
	assertEqual(t, Slice[int]{0, 1, 2, 3, 4}, Range(0, 5))
	assertEqual(t, Slice[int]{2, 3, 4}, Range(2, 5))
}

func TestRangeEmpty(t *testing.T) {
	if len(Range(5, 5)) != 0 {
		t.Fatal("expected empty range")
	}
}

func TestRangeStep(t *testing.T) {
	assertEqual(t, Slice[int]{0, 2, 4, 6, 8}, RangeStep(0, 10, 2))
}

func TestRangeStepFloat(t *testing.T) {
	got := RangeStep(0.0, 1.0, 0.5)
	if len(got) != 2 || got[0] != 0.0 || got[1] != 0.5 {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestSum(t *testing.T) {
	if Sum(Slice[int]{1, 2, 3, 4, 5}) != 15 {
		t.Fatal("expected 15")
	}
	if Sum(Slice[int]{}) != 0 {
		t.Fatal("expected 0 for empty")
	}
}

func TestMin(t *testing.T) {
	opt := Min(Slice[int]{3, 1, 4, 1, 5})
	if opt.IsNone() || opt.Unwrap() != 1 {
		t.Fatalf("expected Some(1), got %v", opt)
	}
	if Min(Slice[int]{}).IsSome() {
		t.Fatal("expected None for empty")
	}
}

func TestMax(t *testing.T) {
	opt := Max(Slice[int]{3, 1, 4, 1, 5})
	if opt.IsNone() || opt.Unwrap() != 5 {
		t.Fatalf("expected Some(5), got %v", opt)
	}
	if Max(Slice[int]{}).IsSome() {
		t.Fatal("expected None for empty")
	}
}

func TestMinBy(t *testing.T) {
	type Item struct{ Name string; Score int }
	s := Slice[Item]{{"a", 3}, {"b", 1}, {"c", 2}}
	opt := MinBy(s, func(i Item) int { return i.Score })
	if opt.IsNone() || opt.Unwrap().Name != "b" {
		t.Fatalf("expected Some({b,1}), got %v", opt)
	}
	if MinBy(Slice[Item]{}, func(i Item) int { return i.Score }).IsSome() {
		t.Fatal("expected None for empty")
	}
}

func TestMaxBy(t *testing.T) {
	type Item struct{ Name string; Score int }
	s := Slice[Item]{{"a", 3}, {"b", 1}, {"c", 5}}
	opt := MaxBy(s, func(i Item) int { return i.Score })
	if opt.IsNone() || opt.Unwrap().Name != "c" {
		t.Fatalf("expected Some({c,5}), got %v", opt)
	}
}

func TestSumBy(t *testing.T) {
	type Item struct{ Price float64 }
	s := Slice[Item]{{1.5}, {2.5}, {3.0}}
	got := SumBy(s, func(i Item) float64 { return i.Price })
	if got != 7.0 {
		t.Fatalf("expected 7.0, got %v", got)
	}
}

func TestSumParallel(t *testing.T) {
	s := Range(0, parallelThreshold+100)
	want := Sum(s)
	got := SumParallel(s)
	if got != want {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestSumParallelN(t *testing.T) {
	s := Range(1, parallelThreshold+1)
	want := Sum(s)
	got := SumParallelN(s, 4)
	if got != want {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestSumParallelSmall(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	if SumParallel(s) != 15 {
		t.Fatal("expected 15")
	}
}

func TestSumByParallel(t *testing.T) {
	type Item struct{ V int }
	s := make(Slice[Item], parallelThreshold+50)
	for i := range s {
		s[i] = Item{i + 1}
	}
	want := SumBy(s, func(x Item) int { return x.V })
	got := SumByParallel(s, func(x Item) int { return x.V })
	if got != want {
		t.Fatalf("want %d, got %d", want, got)
	}
}
