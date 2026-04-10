package dict

import (
	"testing"
)

// ─── Pick ─────────────────────────────────────────────────────────────────────

func TestPick(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2, "c": 3}
	got := d.Pick("a", "c")
	assertEqual(t, Map[string, int]{"a": 1, "c": 3}, got)
}

func TestPickMissingKey(t *testing.T) {
	d := Map[string, int]{"a": 1}
	got := d.Pick("a", "z")
	assertEqual(t, Map[string, int]{"a": 1}, got)
}

func TestPickEmpty(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2}
	got := d.Pick()
	if got.Len() != 0 {
		t.Fatalf("expected empty map, got %v", got)
	}
}

// ─── Omit ─────────────────────────────────────────────────────────────────────

func TestOmit(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2, "c": 3}
	got := d.Omit("b")
	assertEqual(t, Map[string, int]{"a": 1, "c": 3}, got)
}

func TestOmitMultiple(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2, "c": 3}
	got := d.Omit("a", "c")
	assertEqual(t, Map[string, int]{"b": 2}, got)
}

func TestOmitMissingKey(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2}
	got := d.Omit("z")
	if got.Len() != 2 || got["a"] != 1 || got["b"] != 2 {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestOmitEmpty(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2}
	got := d.Omit()
	if got.Len() != 2 {
		t.Fatalf("expected 2 entries, got %v", got)
	}
}

// ─── Merge ────────────────────────────────────────────────────────────────────

func TestMerge(t *testing.T) {
	a := Map[string, int]{"x": 1, "y": 2}
	b := Map[string, int]{"y": 99, "z": 3}
	got := a.Merge(b)
	// Other wins on conflict ("y" should be 99).
	if got["x"] != 1 || got["y"] != 99 || got["z"] != 3 {
		t.Fatalf("unexpected merge result: %v", got)
	}
}

func TestMergeDoesNotMutate(t *testing.T) {
	a := Map[string, int]{"a": 1}
	b := Map[string, int]{"b": 2}
	a.Merge(b)
	if a.Len() != 1 {
		t.Fatal("Merge must not mutate the receiver")
	}
}

func TestMergeEmptyMaps(t *testing.T) {
	empty := Map[string, int]{}
	assertEqual(t, Map[string, int]{"a": 1}, Map[string, int]{"a": 1}.Merge(empty))
	assertEqual(t, Map[string, int]{"a": 1}, empty.Merge(Map[string, int]{"a": 1}))
}

// ─── MergeWith ────────────────────────────────────────────────────────────────

func TestMergeWith(t *testing.T) {
	a := Map[string, int]{"x": 10, "y": 20}
	b := Map[string, int]{"y": 5, "z": 30}
	got := a.MergeWith(b, func(_ string, left, right int) int { return left + right })
	if got["x"] != 10 || got["y"] != 25 || got["z"] != 30 {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestMergeWithNoConflict(t *testing.T) {
	a := Map[string, int]{"a": 1}
	b := Map[string, int]{"b": 2}
	got := a.MergeWith(b, func(_ string, l, r int) int { return l + r })
	if got["a"] != 1 || got["b"] != 2 {
		t.Fatalf("unexpected result: %v", got)
	}
}

// ─── MapKeys ──────────────────────────────────────────────────────────────────

func TestMapKeys(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2}
	got := MapKeys(d, func(k string, _ int) string { return k + k })
	if got["aa"] != 1 || got["bb"] != 2 {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestMapKeysChangeType(t *testing.T) {
	d := Map[int, string]{1: "one", 2: "two"}
	got := MapKeys(d, func(k int, _ string) string {
		return []string{"", "one_key", "two_key"}[k]
	})
	if got["one_key"] != "one" || got["two_key"] != "two" {
		t.Fatalf("unexpected: %v", got)
	}
}
