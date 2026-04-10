package slice

import (
	"errors"
	"testing"
)

// ─── TakeWhile ────────────────────────────────────────────────────────────────

func TestTakeWhile(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	assertEqual(t, Slice[int]{1, 2}, s.TakeWhile(func(v int) bool { return v < 3 }))
}

func TestTakeWhileAll(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	assertEqual(t, Slice[int]{1, 2, 3}, s.TakeWhile(func(v int) bool { return v > 0 }))
}

func TestTakeWhileNone(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	got := s.TakeWhile(func(v int) bool { return v > 10 })
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestTakeWhileEmpty(t *testing.T) {
	got := Slice[int]{}.TakeWhile(func(v int) bool { return true })
	if len(got) != 0 {
		t.Fatal("expected empty slice")
	}
}

// ─── DropWhile ────────────────────────────────────────────────────────────────

func TestDropWhile(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4, 5}
	assertEqual(t, Slice[int]{3, 4, 5}, s.DropWhile(func(v int) bool { return v < 3 }))
}

func TestDropWhileAll(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	got := s.DropWhile(func(v int) bool { return v > 0 })
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestDropWhileNone(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	assertEqual(t, s, s.DropWhile(func(v int) bool { return v > 10 }))
}

func TestDropWhileEmpty(t *testing.T) {
	got := Slice[int]{}.DropWhile(func(v int) bool { return true })
	if len(got) != 0 {
		t.Fatal("expected empty slice")
	}
}

// ─── ChunkBy ─────────────────────────────────────────────────────────────────

func TestChunkBy(t *testing.T) {
	s := Slice[int]{1, 1, 2, 2, 3, 1, 1}
	chunks := ChunkBy(s, func(v int) int { return v })
	if len(chunks) != 4 {
		t.Fatalf("expected 4 chunks, got %d", len(chunks))
	}
	assertEqual(t, Slice[int]{1, 1}, chunks[0])
	assertEqual(t, Slice[int]{2, 2}, chunks[1])
	assertEqual(t, Slice[int]{3}, chunks[2])
	assertEqual(t, Slice[int]{1, 1}, chunks[3])
}

func TestChunkByOddEven(t *testing.T) {
	s := Slice[int]{1, 3, 2, 4, 5}
	parity := func(v int) string {
		if v%2 == 0 {
			return "even"
		}
		return "odd"
	}
	chunks := ChunkBy(s, parity)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d: %v", len(chunks), chunks)
	}
	assertEqual(t, Slice[int]{1, 3}, chunks[0])
	assertEqual(t, Slice[int]{2, 4}, chunks[1])
	assertEqual(t, Slice[int]{5}, chunks[2])
}

func TestChunkByEmpty(t *testing.T) {
	if ChunkBy(Slice[int]{}, func(v int) int { return v }) != nil {
		t.Fatal("expected nil for empty slice")
	}
}

func TestChunkBySingle(t *testing.T) {
	chunks := ChunkBy(Slice[int]{42}, func(v int) int { return v })
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	assertEqual(t, Slice[int]{42}, chunks[0])
}

// ─── Scan ─────────────────────────────────────────────────────────────────────

func TestScan(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4}
	got := Scan(s, 0, func(acc, v int) int { return acc + v })
	assertEqual(t, Slice[int]{0, 1, 3, 6, 10}, got)
}

func TestScanEmpty(t *testing.T) {
	got := Scan(Slice[int]{}, 99, func(acc, v int) int { return acc + v })
	assertEqual(t, Slice[int]{99}, got)
}

func TestScanString(t *testing.T) {
	s := Slice[string]{"a", "b", "c"}
	got := Scan(s, "", func(acc, v string) string { return acc + v })
	assertEqual(t, Slice[string]{"", "a", "ab", "abc"}, got)
}

// ─── Associate ────────────────────────────────────────────────────────────────

func TestAssociate(t *testing.T) {
	type User struct{ ID int; Name string }
	s := Slice[User]{{1, "Alice"}, {2, "Bob"}}
	got := Associate(s, func(u User) (int, string) { return u.ID, u.Name })
	if got[1] != "Alice" || got[2] != "Bob" {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestAssociateTransformsValue(t *testing.T) {
	s := Slice[int]{1, 2, 3}
	got := Associate(s, func(v int) (string, int) {
		return []string{"one", "two", "three"}[v-1], v * v
	})
	if got["one"] != 1 || got["two"] != 4 || got["three"] != 9 {
		t.Fatalf("unexpected result: %v", got)
	}
}

// ─── Interleave ───────────────────────────────────────────────────────────────

func TestInterleave(t *testing.T) {
	a := Slice[int]{1, 2, 3}
	b := Slice[int]{4, 5, 6}
	assertEqual(t, Slice[int]{1, 4, 2, 5, 3, 6}, Interleave(a, b))
}

func TestInterleaveUnequalLengths(t *testing.T) {
	a := Slice[int]{1, 2}
	b := Slice[int]{3, 4, 5}
	assertEqual(t, Slice[int]{1, 3, 2, 4, 5}, Interleave(a, b))
}

func TestInterleaveThreeSlices(t *testing.T) {
	a := Slice[int]{1}
	b := Slice[int]{2}
	c := Slice[int]{3}
	assertEqual(t, Slice[int]{1, 2, 3}, Interleave(a, b, c))
}

func TestInterleaveEmpty(t *testing.T) {
	got := Interleave(Slice[int]{}, Slice[int]{})
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestInterleaveNoArgs(t *testing.T) {
	got := Interleave[int]()
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

// ─── TryReduce ────────────────────────────────────────────────────────────────

func TestTryReduce(t *testing.T) {
	s := Slice[int]{1, 2, 3, 4}
	sum, err := TryReduce(s, 0, func(acc, v int) (int, error) {
		return acc + v, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum != 10 {
		t.Fatalf("expected 10, got %d", sum)
	}
}

func TestTryReduceStopsOnError(t *testing.T) {
	s := Slice[int]{1, 2, 0, 4}
	errDiv := errors.New("division by zero")
	_, err := TryReduce(s, 0, func(acc, v int) (int, error) {
		if v == 0 {
			return acc, errDiv
		}
		return acc + 100/v, nil
	})
	if !errors.Is(err, errDiv) {
		t.Fatalf("expected errDiv, got %v", err)
	}
}

func TestTryReduceEmpty(t *testing.T) {
	got, err := TryReduce(Slice[int]{}, 42, func(acc, v int) (int, error) {
		return acc + v, nil
	})
	if err != nil || got != 42 {
		t.Fatalf("expected (42, nil), got (%d, %v)", got, err)
	}
}
