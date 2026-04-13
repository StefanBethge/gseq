package dict

import (
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/stefanbethge/gseq/slice"
)

func assertEqual(t *testing.T, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("mismatch\nwant: %v\n got: %v", want, got)
	}
}

func TestFilter(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2, "c": 3}
	got := d.Filter(func(_ string, v int) bool { return v > 1 })
	if got.Len() != 2 || got["a"] != 0 || got["b"] != 2 || got["c"] != 3 {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestReject(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2, "c": 3}
	got := d.Reject(func(_ string, v int) bool { return v > 1 })
	if got.Len() != 1 || got["a"] != 1 {
		t.Fatalf("expected only a:1, got %v", got)
	}
}

func TestEach(t *testing.T) {
	d := Map[string, int]{"x": 10, "y": 20}
	sum := 0
	d.Each(func(_ string, v int) { sum += v })
	if sum != 30 {
		t.Fatalf("expected sum 30, got %d", sum)
	}
}

func TestKeys(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2, "c": 3}
	keys := d.Keys().ToSlice()
	sort.Strings(keys)
	assertEqual(t, []string{"a", "b", "c"}, keys)
}

func TestValues(t *testing.T) {
	d := Map[string, int]{"x": 1, "y": 2, "z": 3}
	vals := d.Values().ToSlice()
	sort.Ints(vals)
	assertEqual(t, []int{1, 2, 3}, vals)
}

func TestContains(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2}
	if !d.Contains(func(_ string, v int) bool { return v == 2 }) {
		t.Fatal("expected true")
	}
	if d.Contains(func(_ string, v int) bool { return v == 99 }) {
		t.Fatal("expected false")
	}
}

func TestNone(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2}
	if !d.None(func(_ string, v int) bool { return v > 10 }) {
		t.Fatal("expected true: no value > 10")
	}
	if d.None(func(_ string, v int) bool { return v == 2 }) {
		t.Fatal("expected false: value 2 exists")
	}
	empty := Map[string, int]{}
	if !empty.None(func(_ string, _ int) bool { return true }) {
		t.Fatal("empty map: None should always return true")
	}
}

func TestEvery(t *testing.T) {
	d := Map[string, int]{"a": 2, "b": 4}
	if !d.Every(func(_ string, v int) bool { return v%2 == 0 }) {
		t.Fatal("expected true")
	}
	d2 := Map[string, int]{"a": 2, "b": 3}
	if d2.Every(func(_ string, v int) bool { return v%2 == 0 }) {
		t.Fatal("expected false")
	}
}

func TestLen(t *testing.T) {
	d := Map[int, string]{1: "a", 2: "b"}
	if d.Len() != 2 {
		t.Fatal("expected Len 2")
	}
}

func TestToMap(t *testing.T) {
	d := Map[string, int]{"a": 1}
	plain := d.ToMap()
	if plain["a"] != 1 {
		t.Fatal("unexpected ToMap result")
	}
}

func TestMapValues(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2}
	got := MapValues(d, func(k string, v int) string { return k + "=" + string(rune('0'+v)) })
	if got["a"] != "a=1" || got["b"] != "b=2" {
		t.Fatalf("unexpected MapValues result: %v", got)
	}
}

func TestInvert(t *testing.T) {
	d := Map[string, int]{"a": 1, "b": 2}
	inv := Invert(d)
	if inv[1] != "a" || inv[2] != "b" {
		t.Fatalf("unexpected Invert result: %v", inv)
	}
}

func TestToSlice(t *testing.T) {
	d := Map[string, int]{"a": 1}
	got := ToSlice(d, func(k string, _ int) string { return k })
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestFromSlice(t *testing.T) {
	s := slice.Slice[string]{"alice", "bob"}
	d := FromSlice(s, func(name string) (string, int) { return name, len(name) })
	if d["alice"] != 5 || d["bob"] != 3 {
		t.Fatalf("unexpected result: %v", d)
	}
}

func TestEachParallel(t *testing.T) {
	// Build a map large enough to trigger parallel path.
	d := make(Map[int, int], parallelThreshold+1)
	for i := range parallelThreshold + 1 {
		d[i] = i
	}
	var mu sync.Mutex
	sum := 0
	d.EachParallel(func(_ int, v int) {
		mu.Lock()
		sum += v
		mu.Unlock()
	})
	want := (parallelThreshold) * (parallelThreshold + 1) / 2
	if sum != want {
		t.Fatalf("want sum %d, got %d", want, sum)
	}
}

func TestEachParallelSmall(t *testing.T) {
	// Small map → sequential fallback, no goroutines needed.
	d := Map[string, int]{"a": 1, "b": 2}
	sum := 0
	d.EachParallel(func(_ string, v int) { sum += v })
	assertEqual(t, 3, sum)
}

func TestMapValuesParallel(t *testing.T) {
	d := make(Map[int, int], parallelThreshold+1)
	for i := range parallelThreshold + 1 {
		d[i] = i
	}
	got := MapValuesParallel(d, func(_ int, v int) int { return v * 2 })
	for k, v := range d {
		if got[k] != v*2 {
			t.Fatalf("key %d: want %d got %d", k, v*2, got[k])
		}
	}
}

func TestMapValuesParallelSmall(t *testing.T) {
	d := Map[string, int]{"x": 3, "y": 4}
	got := MapValuesParallel(d, func(_ string, v int) int { return v * 10 })
	assertEqual(t, 30, got["x"])
	assertEqual(t, 40, got["y"])
}
