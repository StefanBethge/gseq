package slice

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stefanbethge/gseq/option"
)

func TestMap(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	got := Map(s, func(v int) string {
		switch v {
		case 1:
			return "one"
		case 2:
			return "two"
		default:
			return "three"
		}
	})
	assertEqual(t, Slice[string]{"one", "two", "three"}, got)
}

func TestMapParallel(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	got := MapParallel(s, func(v int) int { return v * 2 })
	assertEqual(t, Slice[int]{2, 4, 6, 8, 10}, got)
}

func TestGroupBy(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5, 6}
	got := GroupBy(s, func(v int) string {
		if v%2 == 0 {
			return "even"
		}
		return "odd"
	})
	assertEqual(t, Slice[int]{1, 3, 5}, got["odd"])
	assertEqual(t, Slice[int]{2, 4, 6}, got["even"])
}

func TestKeyBy(t *testing.T) {
	type User struct{ ID int; Name string }
	s := Slice[User]{{1, "Alice"}, {2, "Bob"}}
	got := KeyBy(s, func(u User) int { return u.ID })
	if got[1].Name != "Alice" || got[2].Name != "Bob" {
		t.Fatalf("unexpected map: %v", got)
	}
}

func TestReduce(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	if Reduce(s, 0, func(acc, v int) int { return acc + v }) != 15 {
		t.Fatal("expected 15")
	}
}

func TestReduceString(t *testing.T) {
	s := Slice[string]{"a", "b", "c"}
	got := Reduce(s, "", func(acc, v string) string { return acc + v })
	if got != "abc" {
		t.Fatalf("expected abc, got %s", got)
	}
}

func TestFlatMap(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	got := FlatMap(s, func(v int) Slice[int] { return Slice[int]{v, v * 10} })
	assertEqual(t, Slice[int]{1, 10, 2, 20, 3, 30}, got)
}

func TestFlatten(t *testing.T) {
	s := Slice[Slice[int]]{{1, 2}, {3, 4}, {5}}
	assertEqual(t, Slice[int]{1, 2, 3, 4, 5}, Flatten(s))
}

func TestUniq(t *testing.T) {
	s := Slice[int]{1, 2, 2, 3, 1, 4}
	assertEqual(t, Slice[int]{1, 2, 3, 4}, Uniq(s, func(v int) int { return v }))
}

func TestIntersect(t *testing.T) {
	a := Slice[int]{1, 2, 3, 4}
	b := Slice[int]{2, 4, 6}
	assertEqual(t, Slice[int]{2, 4}, Intersect(a, b, func(v int) int { return v }))
}

func TestDifference(t *testing.T) {
	a := Slice[int]{1, 2, 3, 4}
	b := Slice[int]{2, 4}
	assertEqual(t, Slice[int]{1, 3}, Difference(a, b, func(v int) int { return v }))
}

func TestUnion(t *testing.T) {
	a := Slice[int]{1, 2, 3}
	b := Slice[int]{2, 3, 4}
	assertEqual(t, Slice[int]{1, 2, 3, 4}, Union(a, b, func(v int) int { return v }))
}

func TestZip(t *testing.T) {
	a := Slice[int]{1, 2, 3}
	b := Slice[string]{"a", "b", "c"}
	type pair struct{ N int; S string }
	got := Zip(a, b, func(n int, s string) pair { return pair{n, s} })
	assertEqual(t, Slice[pair]{{1, "a"}, {2, "b"}, {3, "c"}}, got)
}

func TestZipUnequalLength(t *testing.T) {
	a := Slice[int]{1, 2, 3}
	b := Slice[int]{10, 20}
	assertEqual(t, Slice[int]{11, 22}, Zip(a, b, func(x, y int) int { return x + y }))
}

func TestMapIndexed(t *testing.T) {
	s := Slice[string]{"a", "b", "c"}
	got := MapIndexed(s, func(i int, v string) string { return fmt.Sprintf("%d:%s", i, v) })
	assertEqual(t, Slice[string]{"0:a", "1:b", "2:c"}, got)
}

func TestMapParallelIndexed(t *testing.T) {
	s := Slice[int]{10, 20, 30}
	got := MapParallelIndexed(s, func(i int, v int) string { return fmt.Sprintf("%d=%d", i, v) })
	assertEqual(t, Slice[string]{"0=10", "1=20", "2=30"}, got)
}

func TestCompact(t *testing.T) {
	s := Slice[option.Option[int]]{
		option.Some(1), option.None[int](), option.Some(3), option.None[int](), option.Some(5),
	}
	assertEqual(t, Slice[int]{1, 3, 5}, Compact(s))
}

func TestCompactAllNone(t *testing.T) {
	s := Slice[option.Option[int]]{option.None[int](), option.None[int]()}
	if len(Compact(s)) != 0 {
		t.Fatal("expected empty")
	}
}

func TestTryMap(t *testing.T) {
	s := Slice[string]{"1", "bad", "3", "nope", "5"}
	got := TryMap(s, func(v string) option.Option[int] {
		n, err := strconv.Atoi(v)
		if err != nil {
			return option.None[int]()
		}
		return option.Some(n)
	})
	assertEqual(t, Slice[int]{1, 3, 5}, got)
}

func TestTryMapAllDropped(t *testing.T) {
	s := Slice[string]{"bad", "input"}
	got := TryMap(s, func(v string) option.Option[int] { return option.None[int]() })
	if len(got) != 0 {
		t.Fatal("expected empty")
	}
}
