package option

import "testing"

func TestCoalesceFirstSome(t *testing.T) {
	got := Coalesce(None[int](), Some(1), Some(2))
	if !got.IsSome() || got.Unwrap() != 1 {
		t.Fatalf("expected Some(1), got %v", got)
	}
}

func TestCoalesceAllNone(t *testing.T) {
	got := Coalesce(None[string](), None[string]())
	if got.IsSome() {
		t.Fatal("expected None")
	}
}

func TestCoalesceEmpty(t *testing.T) {
	got := Coalesce[int]()
	if got.IsSome() {
		t.Fatal("expected None for empty input")
	}
}

func TestCoalesceSingleSome(t *testing.T) {
	got := Coalesce(Some(42))
	if !got.IsSome() || got.Unwrap() != 42 {
		t.Fatalf("expected Some(42), got %v", got)
	}
}

func TestCoalesceReturnsFirstNotSecond(t *testing.T) {
	got := Coalesce(Some(10), Some(20))
	if got.Unwrap() != 10 {
		t.Fatalf("expected 10, got %d", got.Unwrap())
	}
}
