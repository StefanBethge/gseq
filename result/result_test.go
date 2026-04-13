package result

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/stefanbethge/gseq/option"
	"github.com/stefanbethge/gseq/slice"
)

func assertEqual(t *testing.T, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("mismatch\nwant: %v\n got: %v", want, got)
	}
}

func TestOkIsOk(t *testing.T) {
	r := Ok[int, string](42)
	if !r.IsOk() || r.IsErr() {
		t.Fatal("expected IsOk=true")
	}
	if r.Unwrap() != 42 {
		t.Fatalf("expected 42, got %v", r.Unwrap())
	}
}

func TestErrIsErr(t *testing.T) {
	r := Err[int, string]("oops")
	if r.IsOk() || !r.IsErr() {
		t.Fatal("expected IsErr=true")
	}
	if r.UnwrapErr() != "oops" {
		t.Fatalf("expected oops, got %v", r.UnwrapErr())
	}
}

func TestUnwrapOr(t *testing.T) {
	if Ok[int, string](7).UnwrapOr(0) != 7 {
		t.Fatal("expected 7")
	}
	if Err[int, string]("fail").UnwrapOr(99) != 99 {
		t.Fatal("expected fallback 99")
	}
}

func TestUnwrapPanicsOnErr(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on Unwrap of Err")
		}
	}()
	Err[int, string]("bad").Unwrap()
}

func TestUnwrapErrPanicsOnOk(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on UnwrapErr of Ok")
		}
	}()
	Ok[int, string](1).UnwrapErr()
}

func TestFromGoError(t *testing.T) {
	r := FromGoError(strconv.Atoi("42"))
	if r.IsErr() || r.Unwrap() != 42 {
		t.Fatalf("expected Ok(42), got %v", r)
	}
	r2 := FromGoError(strconv.Atoi("bad"))
	if r2.IsOk() {
		t.Fatal("expected Err for invalid input")
	}
}

func TestToOption(t *testing.T) {
	opt := Ok[int, string](5).ToOption()
	if opt.IsNone() || opt.Unwrap() != 5 {
		t.Fatalf("expected Some(5), got %v", opt)
	}
	if Err[int, string]("x").ToOption().IsSome() {
		t.Fatal("expected None for Err")
	}
}

func TestMap(t *testing.T) {
	r := Map(Ok[int, string](3), func(v int) int { return v * 2 })
	if r.IsErr() || r.Unwrap() != 6 {
		t.Fatalf("expected Ok(6), got %v", r)
	}
	r2 := Map(Err[int, string]("fail"), func(v int) int { return v * 2 })
	if r2.IsOk() || r2.UnwrapErr() != "fail" {
		t.Fatalf("expected Err(fail), got %v", r2)
	}
}

func TestMapTypeChange(t *testing.T) {
	r := Map(Ok[int, string](42), func(int) string { return "ok" })
	if r.IsErr() || r.Unwrap() != "ok" {
		t.Fatalf("expected Ok(ok), got %v", r)
	}
}

func TestFlatMap(t *testing.T) {
	safeDivide := func(n int) Result[int, string] {
		if n == 0 {
			return Err[int, string]("division by zero")
		}
		return Ok[int, string](100 / n)
	}
	r := FlatMap(Ok[int, string](4), safeDivide)
	if r.IsErr() || r.Unwrap() != 25 {
		t.Fatalf("expected Ok(25), got %v", r)
	}
	r2 := FlatMap(Ok[int, string](0), safeDivide)
	if r2.IsOk() || r2.UnwrapErr() != "division by zero" {
		t.Fatalf("expected Err, got %v", r2)
	}
	r3 := FlatMap(Err[int, string]("upstream"), safeDivide)
	if r3.IsOk() || r3.UnwrapErr() != "upstream" {
		t.Fatalf("expected propagated Err, got %v", r3)
	}
}

func TestMapErr(t *testing.T) {
	r := MapErr(Err[int, string]("oops"), func(e string) error { return errors.New(e) })
	if r.IsOk() || r.UnwrapErr().Error() != "oops" {
		t.Fatalf("unexpected: %v", r)
	}
	r2 := MapErr(Ok[int, string](5), func(e string) error { return errors.New(e) })
	if r2.IsErr() || r2.Unwrap() != 5 {
		t.Fatalf("Ok should pass through MapErr unchanged: %v", r2)
	}
}

func TestUnwrapOrElse(t *testing.T) {
	called := false
	got := Ok[int, string](7).UnwrapOrElse(func() int {
		called = true
		return 99
	})
	if got != 7 || called {
		t.Fatal("UnwrapOrElse should not call fn when Ok")
	}
	got = Err[int, string]("fail").UnwrapOrElse(func() int { return 42 })
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestFromOption(t *testing.T) {
	r := FromOption[int, string](option.Some(5), "missing")
	if r.IsErr() || r.Unwrap() != 5 {
		t.Fatalf("expected Ok(5), got %v", r)
	}
	r2 := FromOption[int, string](option.None[int](), "missing")
	if r2.IsOk() || r2.UnwrapErr() != "missing" {
		t.Fatalf("expected Err(missing), got %v", r2)
	}
}

func TestChaining(t *testing.T) {
	// Real-world pattern: parse → validate → transform
	parse := func(s string) Result[int, string] {
		n, err := strconv.Atoi(s)
		if err != nil {
			return Err[int, string]("not a number: " + s)
		}
		return Ok[int, string](n)
	}
	positive := func(n int) Result[int, string] {
		if n <= 0 {
			return Err[int, string]("must be positive")
		}
		return Ok[int, string](n)
	}

	r := FlatMap(parse("42"), positive)
	if r.IsErr() || r.Unwrap() != 42 {
		t.Fatalf("expected Ok(42), got %v", r)
	}
	r2 := FlatMap(parse("abc"), positive)
	if r2.IsOk() {
		t.Fatal("expected Err for non-numeric input")
	}
	r3 := FlatMap(parse("-5"), positive)
	if r3.IsOk() || r3.UnwrapErr() != "must be positive" {
		t.Fatalf("expected validation Err, got %v", r3)
	}
}

func TestPartition(t *testing.T) {
	s := slice.Slice[Result[int, string]]{
		Ok[int, string](1),
		Err[int, string]("e1"),
		Ok[int, string](3),
		Err[int, string]("e2"),
		Ok[int, string](5),
	}
	oks, errs := Partition(s)
	assertEqual(t, slice.Slice[int]{1, 3, 5}, oks)
	assertEqual(t, slice.Slice[string]{"e1", "e2"}, errs)
}

func TestPartitionAllOk(t *testing.T) {
	s := slice.Slice[Result[int, string]]{Ok[int, string](1), Ok[int, string](2)}
	oks, errs := Partition(s)
	assertEqual(t, slice.Slice[int]{1, 2}, oks)
	if len(errs) != 0 {
		t.Fatal("expected no errors")
	}
}

func TestPartitionAllErr(t *testing.T) {
	s := slice.Slice[Result[int, string]]{Err[int, string]("a"), Err[int, string]("b")}
	oks, errs := Partition(s)
	if len(oks) != 0 {
		t.Fatal("expected no successes")
	}
	assertEqual(t, slice.Slice[string]{"a", "b"}, errs)
}

func TestPartitionParallelWorkflow(t *testing.T) {
	// Simulate processing 100 records in parallel where some fail.
	// Records with ID divisible by 20 are treated as failures.
	ids := make(slice.Slice[int], 100)
	for i := range ids {
		ids[i] = i + 1
	}

	process := func(id int) Result[string, error] {
		if id%20 == 0 {
			return Err[string, error](fmt.Errorf("record %d: processing failed", id))
		}
		return Ok[string, error](fmt.Sprintf("record-%d", id))
	}

	results := slice.MapParallel(ids, process)
	successes, failures := Partition(results)

	if len(successes) != 95 {
		t.Fatalf("expected 95 successes, got %d", len(successes))
	}
	if len(failures) != 5 {
		t.Fatalf("expected 5 failures, got %d", len(failures))
	}
	for _, err := range failures {
		if err == nil {
			t.Fatal("expected non-nil error")
		}
	}
}
