package slice

import "testing"

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
