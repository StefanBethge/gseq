package option

import "testing"

func TestSomeIsSome(t *testing.T) {
	opt := Some(42)
	if !opt.IsSome() || opt.IsNone() {
		t.Fatal("expected IsSome=true")
	}
	if opt.Unwrap() != 42 {
		t.Fatalf("expected 42, got %v", opt.Unwrap())
	}
}

func TestNoneIsNone(t *testing.T) {
	opt := None[int]()
	if opt.IsSome() || !opt.IsNone() {
		t.Fatal("expected IsNone=true")
	}
}

func TestUnwrapOr(t *testing.T) {
	if Some(7).UnwrapOr(0) != 7 {
		t.Fatal("expected 7")
	}
	if None[int]().UnwrapOr(99) != 99 {
		t.Fatal("expected fallback 99")
	}
}

func TestGet(t *testing.T) {
	v, ok := Some("hello").Get()
	if !ok || v != "hello" {
		t.Fatalf("expected (hello, true), got (%v, %v)", v, ok)
	}
	_, ok = None[string]().Get()
	if ok {
		t.Fatal("expected false")
	}
}

func TestUnwrapPanicsOnNone(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on Unwrap of None")
		}
	}()
	None[int]().Unwrap()
}

func TestMap(t *testing.T) {
	doubled := Map(Some(5), func(v int) int { return v * 2 })
	if doubled.IsNone() || doubled.Unwrap() != 10 {
		t.Fatalf("expected Some(10), got %v", doubled)
	}
	if Map(None[int](), func(v int) int { return v * 2 }).IsSome() {
		t.Fatal("expected None")
	}
}

func TestMapTypeChange(t *testing.T) {
	opt := Map(Some(42), func(int) string { return "ok" })
	if opt.IsNone() || opt.Unwrap() != "ok" {
		t.Fatalf("expected Some(ok), got %v", opt)
	}
}

func TestFilter(t *testing.T) {
	isEven := func(v int) bool { return v%2 == 0 }
	if Some(4).Filter(isEven).IsNone() {
		t.Fatal("expected Some(4) to pass filter")
	}
	if Some(3).Filter(isEven).IsSome() {
		t.Fatal("expected Some(3) to fail filter")
	}
	if None[int]().Filter(isEven).IsSome() {
		t.Fatal("expected None to remain None")
	}
}

func TestOr(t *testing.T) {
	got := Some(1).Or(Some(2))
	if got.Unwrap() != 1 {
		t.Fatal("Some.Or should return self")
	}
	got = None[int]().Or(Some(2))
	if got.Unwrap() != 2 {
		t.Fatal("None.Or should return other")
	}
	if None[int]().Or(None[int]()).IsSome() {
		t.Fatal("None.Or(None) should remain None")
	}
}

func TestOrElse(t *testing.T) {
	called := false
	got := Some(1).OrElse(func() Option[int] {
		called = true
		return Some(2)
	})
	if got.Unwrap() != 1 || called {
		t.Fatal("OrElse should not call fn when Some")
	}
	got = None[int]().OrElse(func() Option[int] { return Some(99) })
	if got.Unwrap() != 99 {
		t.Fatal("OrElse should call fn when None")
	}
}

func TestFlatMap(t *testing.T) {
	safeSqrt := func(n int) Option[int] {
		if n < 0 {
			return None[int]()
		}
		return Some(n * n)
	}
	if FlatMap(Some(4), safeSqrt).Unwrap() != 16 {
		t.Fatal("expected Some(16)")
	}
	if FlatMap(Some(-1), safeSqrt).IsSome() {
		t.Fatal("expected None when fn returns None")
	}
	if FlatMap(None[int](), safeSqrt).IsSome() {
		t.Fatal("expected None when input is None")
	}
}

func TestChaining(t *testing.T) {
	// Or → Filter → Map → UnwrapOr pipeline
	result := None[int]().
		Or(Some(3)).
		Filter(func(v int) bool { return v > 2 }).
		UnwrapOr(0)
	if result != 3 {
		t.Fatalf("expected 3, got %d", result)
	}

	result = None[int]().
		Or(Some(1)).
		Filter(func(v int) bool { return v > 2 }).
		UnwrapOr(99)
	if result != 99 {
		t.Fatalf("expected fallback 99, got %d", result)
	}
}
