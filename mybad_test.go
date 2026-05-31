package mybad_test

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strconv"
	"testing"

	"github.com/nlozgachev/mybad"
)

var sentinel = errors.New("sentinel error")

// helpers

func double(n int) int        { return n * 2 }
func itoa(n int) string       { return strconv.Itoa(n) }
func fail(n int) (int, error) { return 0, sentinel }
func inc(n int) (int, error)  { return n + 1, nil }
func toString(n int) (string, error) {
	return itoa(n), nil
}

// Ok / From

func TestOk(t *testing.T) {
	r := mybad.Ok(42)
	if v, err := r.Unpack(); err != nil || v != 42 {
		t.Fatalf("got (%v, %v), want (42, nil)", v, err)
	}
}

func TestFrom_healthy(t *testing.T) {
	r := mybad.From(99, nil)
	if v, err := r.Unpack(); err != nil || v != 99 {
		t.Fatalf("got (%v, %v), want (99, nil)", v, err)
	}
}

func TestFrom_error(t *testing.T) {
	r := mybad.From(0, sentinel)
	if err := r.Err(); !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want sentinel", err)
	}
}

// Terminal methods

func TestIsOk_healthy(t *testing.T) {
	if !mybad.Ok(1).IsOk() {
		t.Fatal("expected IsOk to return true for healthy Result")
	}
}

func TestIsOk_error(t *testing.T) {
	if mybad.From(0, sentinel).IsOk() {
		t.Fatal("expected IsOk to return false for error Result")
	}
}

func TestIsErr_error(t *testing.T) {
	if !mybad.From(0, sentinel).IsErr() {
		t.Fatal("expected IsErr to return true for error Result")
	}
}

func TestIsErr_healthy(t *testing.T) {
	if mybad.Ok(1).IsErr() {
		t.Fatal("expected IsErr to return false for healthy Result")
	}
}

func TestErr_healthy(t *testing.T) {
	if err := mybad.Ok(1).Err(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestErr_error(t *testing.T) {
	r := mybad.From(0, sentinel)
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("expected sentinel, got %v", r.Err())
	}
}

func TestMust_healthy(t *testing.T) {
	if v := mybad.Ok(7).Must(); v != 7 {
		t.Fatalf("got %v, want 7", v)
	}
}

func TestMust_panics(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, got none")
		}
		if !errors.Is(v.(error), sentinel) {
			t.Fatalf("panic value: got %v, want sentinel", v)
		}
	}()
	mybad.From(0, sentinel).Must()
}

func TestUnpack(t *testing.T) {
	v, err := mybad.Ok(3).Unpack()
	if v != 3 || err != nil {
		t.Fatalf("got (%v, %v), want (3, nil)", v, err)
	}
}

func TestUnpack_error(t *testing.T) {
	v, err := mybad.From(0, sentinel).Unpack()
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel, got %v", err)
	}
	if v != 0 {
		t.Fatalf("expected zero value, got %v", v)
	}
}

// ValueOr / ValueOrElse

func TestValueOr_healthy(t *testing.T) {
	if v := mybad.Ok(5).ValueOr(99); v != 5 {
		t.Fatalf("got %v, want 5", v)
	}
}

func TestValueOr_error(t *testing.T) {
	if v := mybad.From(0, sentinel).ValueOr(99); v != 99 {
		t.Fatalf("got %v, want 99", v)
	}
}

func TestValueOrElse_healthy(t *testing.T) {
	called := false
	v := mybad.Ok(5).ValueOrElse(func(err error) int {
		called = true
		return 99
	})
	if called {
		t.Fatal("fn should not be called for healthy Result")
	}
	if v != 5 {
		t.Fatalf("got %v, want 5", v)
	}
}

func TestValueOrElse_error(t *testing.T) {
	v := mybad.From(0, sentinel).ValueOrElse(func(err error) int {
		if !errors.Is(err, sentinel) {
			t.Fatalf("fn received wrong error: %v", err)
		}
		return 99
	})
	if v != 99 {
		t.Fatalf("got %v, want 99", v)
	}
}

// Try

func TestTry_healthy_success(t *testing.T) {
	r := mybad.Try(mybad.Ok(1), inc)
	if v, err := r.Unpack(); err != nil || v != 2 {
		t.Fatalf("got (%v, %v), want (2, nil)", v, err)
	}
}

func TestTry_healthy_fn_fails(t *testing.T) {
	r := mybad.Try(mybad.Ok(1), fail)
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("expected sentinel, got %v", r.Err())
	}
}

func TestTry_error_skips_fn(t *testing.T) {
	called := false
	r := mybad.Try(mybad.From(0, sentinel), func(n int) (int, error) {
		called = true
		return n, nil
	})
	if called {
		t.Fatal("fn should not be called when Result is in error state")
	}
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("expected sentinel, got %v", r.Err())
	}
}

func TestTry_type_change(t *testing.T) {
	r := mybad.Try(mybad.Ok(5), toString)
	if v, err := r.Unpack(); err != nil || v != "5" {
		t.Fatalf("got (%v, %v), want (\"5\", nil)", v, err)
	}
}

// Into

func TestInto_healthy(t *testing.T) {
	r := mybad.Into(mybad.Ok(3), double)
	if v, err := r.Unpack(); err != nil || v != 6 {
		t.Fatalf("got (%v, %v), want (6, nil)", v, err)
	}
}

func TestInto_error_skips_fn(t *testing.T) {
	called := false
	r := mybad.Into(mybad.From(0, sentinel), func(n int) int {
		called = true
		return n
	})
	if called {
		t.Fatal("fn should not be called when Result is in error state")
	}
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("expected sentinel, got %v", r.Err())
	}
}

func TestInto_type_change(t *testing.T) {
	r := mybad.Into(mybad.Ok(3), itoa)
	if v, err := r.Unpack(); err != nil || v != "3" {
		t.Fatalf("got (%v, %v), want (\"3\", nil)", v, err)
	}
}

// WrapErr

func TestWrapErr_healthy_noop(t *testing.T) {
	r := mybad.Ok(1).WrapErr(func(err error) error {
		t.Fatal("should not be called")
		return err
	})
	if r.Err() != nil {
		t.Fatalf("expected nil, got %v", r.Err())
	}
}

func TestWrapErr_wraps_error(t *testing.T) {
	wrapped := errors.New("wrapped")
	r := mybad.From(0, sentinel).WrapErr(func(err error) error {
		return wrapped
	})
	if !errors.Is(r.Err(), wrapped) {
		t.Fatalf("expected wrapped, got %v", r.Err())
	}
}

func TestWrapErr_preserves_chain(t *testing.T) {
	r := mybad.From(0, sentinel).WrapErr(func(err error) error {
		return fmt.Errorf("context: %w", err)
	})
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("expected sentinel to remain reachable via errors.Is, got %v", r.Err())
	}
}

// Peek

func TestPeek_healthy_calls_fn(t *testing.T) {
	called := false
	r := mybad.Ok(42).Peek(func(n int) {
		called = true
		if n != 42 {
			t.Fatalf("got %v, want 42", n)
		}
	})
	if !called {
		t.Fatal("fn should be called")
	}
	if v, _ := r.Unpack(); v != 42 {
		t.Fatalf("result value changed: got %v", v)
	}
}

func TestPeek_error_skips_fn(t *testing.T) {
	r := mybad.From(0, sentinel).Peek(func(n int) {
		t.Fatal("should not be called")
	})
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("expected sentinel to be preserved, got %v", r.Err())
	}
}

// PeekErr

func TestPeekErr_error_calls_fn(t *testing.T) {
	called := false
	r := mybad.From(0, sentinel).PeekErr(func(err error) {
		called = true
		if !errors.Is(err, sentinel) {
			t.Fatalf("got %v, want sentinel", err)
		}
	})
	if !called {
		t.Fatal("fn should be called")
	}
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("result error changed: got %v", r.Err())
	}
}

func TestPeekErr_healthy_skips_fn(t *testing.T) {
	mybad.Ok(1).PeekErr(func(err error) {
		t.Fatal("should not be called")
	})
}

// RecoverTry

func TestRecoverTry_healthy_noop(t *testing.T) {
	r := mybad.Ok(1).RecoverTry(func(err error) (int, error) {
		t.Fatal("should not be called")
		return 0, nil
	})
	if v, _ := r.Unpack(); v != 1 {
		t.Fatalf("got %v, want 1", v)
	}
}

func TestRecoverTry_recovers(t *testing.T) {
	r := mybad.From(0, sentinel).RecoverTry(func(err error) (int, error) {
		return 99, nil
	})
	if v, err := r.Unpack(); err != nil || v != 99 {
		t.Fatalf("got (%v, %v), want (99, nil)", v, err)
	}
}

func TestRecoverTry_fn_fails(t *testing.T) {
	other := errors.New("other")
	r := mybad.From(0, sentinel).RecoverTry(func(err error) (int, error) {
		return 0, other
	})
	if !errors.Is(r.Err(), other) {
		t.Fatalf("expected other, got %v", r.Err())
	}
}

// Match

func TestMatch_healthy(t *testing.T) {
	result := mybad.Match(mybad.Ok(10),
		func(n int) string { return "ok" },
		func(err error) string { return "err" },
	)
	if result != "ok" {
		t.Fatalf("got %q, want \"ok\"", result)
	}
}

func TestMatch_error(t *testing.T) {
	result := mybad.Match(mybad.From(0, sentinel),
		func(n int) string { return "ok" },
		func(err error) string { return "err" },
	)
	if result != "err" {
		t.Fatalf("got %q, want \"err\"", result)
	}
}

// Full pipeline

func TestPipeline(t *testing.T) {
	user := mybad.Try(mybad.Ok(1), inc) // 1 -> 2
	user = mybad.Try(user, inc)         // 2 -> 3
	label := mybad.Into(user, itoa)     // 3 -> "3"

	result := mybad.Match(label,
		func(s string) string { return "ok:" + s },
		func(err error) string { return "err" },
	)
	if result != "ok:3" {
		t.Fatalf("got %q, want \"ok:3\"", result)
	}
}

func TestPipeline_short_circuits(t *testing.T) {
	calls := 0
	count := func(n int) (int, error) {
		calls++
		return n, nil
	}

	r := mybad.Try(mybad.From(0, sentinel), count)
	mybad.Try(r, count)

	if calls != 0 {
		t.Fatalf("expected 0 calls after error, got %d", calls)
	}
}

// AndThen

func TestAndThen_healthy_success(t *testing.T) {
	fn := func(n int) mybad.Result[int] {
		return mybad.Ok(n + 1)
	}
	r := mybad.AndThen(mybad.Ok(1), fn)
	if v, err := r.Unpack(); err != nil || v != 2 {
		t.Fatalf("got (%v, %v), want (2, nil)", v, err)
	}
}

func TestAndThen_healthy_fn_fails(t *testing.T) {
	fn := func(n int) mybad.Result[int] {
		return mybad.From(0, sentinel)
	}
	r := mybad.AndThen(mybad.Ok(1), fn)
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("expected sentinel error, got %v", r.Err())
	}
}

func TestAndThen_error_skips_fn(t *testing.T) {
	called := false
	fn := func(n int) mybad.Result[int] {
		called = true
		return mybad.Ok(n)
	}
	r := mybad.AndThen(mybad.From(0, sentinel), fn)
	if called {
		t.Fatal("fn should not be called when Result is in error state")
	}
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("expected sentinel error, got %v", r.Err())
	}
}

func TestAndThen_type_change(t *testing.T) {
	fn := func(n int) mybad.Result[string] {
		return mybad.Ok(itoa(n))
	}
	r := mybad.AndThen(mybad.Ok(5), fn)
	if v, err := r.Unpack(); err != nil || v != "5" {
		t.Fatalf("got (%v, %v), want (\"5\", nil)", v, err)
	}
}

// Expect

func TestExpect_healthy(t *testing.T) {
	v := mybad.Ok("success").Expect("proved safe")
	if v != "success" {
		t.Fatalf("got %q, want \"success\"", v)
	}
}

func TestExpect_panics(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, got none")
		}
		err, ok := v.(error)
		if !ok {
			t.Fatalf("panic value is not error: %T", v)
		}
		if !errors.Is(err, sentinel) {
			t.Fatalf("panic error: got %v, want wrapped sentinel", err)
		}
	}()
	mybad.From(0, sentinel).Expect("it failed")
}

// Ensure

func TestEnsure_healthy_pass(t *testing.T) {
	r := mybad.Ok(42).Ensure(func(n int) error {
		if n < 0 {
			return errors.New("negative")
		}
		return nil
	})
	if v, err := r.Unpack(); err != nil || v != 42 {
		t.Fatalf("got (%v, %v), want (42, nil)", v, err)
	}
}

func TestEnsure_healthy_fail(t *testing.T) {
	customErr := errors.New("too small")
	r := mybad.Ok(5).Ensure(func(n int) error {
		if n < 10 {
			return customErr
		}
		return nil
	})
	if !errors.Is(r.Err(), customErr) {
		t.Fatalf("expected customErr, got %v", r.Err())
	}
}

func TestEnsure_error_bypass(t *testing.T) {
	called := false
	r := mybad.From(0, sentinel).Ensure(func(n int) error {
		called = true
		return nil
	})
	if called {
		t.Fatal("predicate check should be bypassed on error Result")
	}
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("expected sentinel, got %v", r.Err())
	}
}

// Recover

func TestRecover_healthy_bypass(t *testing.T) {
	called := false
	r := mybad.Ok(10).Recover(func(err error) int {
		called = true
		return 99
	})
	if called {
		t.Fatal("recover callback should be bypassed on healthy Result")
	}
	if v := r.Must(); v != 10 {
		t.Fatalf("got %v, want 10", v)
	}
}

func TestRecover_error_success(t *testing.T) {
	r := mybad.From(0, sentinel).Recover(func(err error) int {
		if errors.Is(err, sentinel) {
			return 99
		}
		return -1
	})
	if v, err := r.Unpack(); err != nil || v != 99 {
		t.Fatalf("got (%v, %v), want (99, nil)", v, err)
	}
}

// FromFunc

func TestFromFunc_success(t *testing.T) {
	fn := func(s string) (int, error) {
		return strconv.Atoi(s)
	}
	lifted := mybad.FromFunc(fn)
	r := lifted("123")
	if v, err := r.Unpack(); err != nil || v != 123 {
		t.Fatalf("got (%v, %v), want (123, nil)", v, err)
	}
}

func TestFromFunc_fail(t *testing.T) {
	fn := func(s string) (int, error) {
		return strconv.Atoi(s)
	}
	lifted := mybad.FromFunc(fn)
	r := lifted("not-a-number")
	if r.Err() == nil {
		t.Fatal("expected error, got nil")
	}
}

// String

func TestString_ok(t *testing.T) {
	r := mybad.Ok(123)
	s := r.String()
	if s != "Ok(123)" {
		t.Fatalf("got %q, want \"Ok(123)\"", s)
	}
}

func TestString_err(t *testing.T) {
	r := mybad.From(0, sentinel)
	s := r.String()
	if s != "Err(sentinel error)" {
		t.Fatalf("got %q, want \"Err(sentinel error)\"", s)
	}
}

// v2 Backlog Features

func TestErr_constructor(t *testing.T) {
	err := errors.New("my error")
	r := mybad.Err[int](err)
	if !r.IsErr() {
		t.Fatal("expected Result to be in error state")
	}
	if !errors.Is(r.Err(), err) {
		t.Fatalf("got %v, want %v", r.Err(), err)
	}

	// Verify panic on nil error
	defer func() {
		val := recover()
		if val == nil {
			t.Fatal("expected panic on Err(nil), got none")
		}
		if s, ok := val.(string); !ok || s != "mybad: Err called with nil error" {
			t.Fatalf("unexpected panic value: %v", val)
		}
	}()
	mybad.Err[int](nil)
}

func TestGuard(t *testing.T) {
	// Setup cancelled and non-cancelled contexts
	ctxOK := context.Background()
	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel()

	// 1. Healthy result, non-cancelled context -> no-op
	r1 := mybad.Ok(42).Guard(ctxOK)
	if v, err := r1.Unpack(); err != nil || v != 42 {
		t.Fatalf("expected (42, nil), got (%v, %v)", v, err)
	}

	// 2. Healthy result, cancelled context -> transitions to context error
	r2 := mybad.Ok(42).Guard(ctxCancel)
	if !errors.Is(r2.Err(), context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", r2.Err())
	}

	// 3. Unhealthy result, non-cancelled context -> returns unchanged
	r3 := mybad.From(0, sentinel).Guard(ctxOK)
	if !errors.Is(r3.Err(), sentinel) {
		t.Fatalf("expected sentinel, got %v", r3.Err())
	}

	// 4. Unhealthy result, cancelled context -> returns unchanged (retains original error)
	r4 := mybad.From(0, sentinel).Guard(ctxCancel)
	if !errors.Is(r4.Err(), sentinel) {
		t.Fatalf("expected sentinel to take precedence, got %v", r4.Err())
	}
}

func TestOr_method(t *testing.T) {
	r1 := mybad.Ok(5).Or(mybad.Ok(99))
	if v, err := r1.Unpack(); err != nil || v != 5 {
		t.Fatalf("expected 5, got (%v, %v)", v, err)
	}

	r2 := mybad.From(0, sentinel).Or(mybad.Ok(99))
	if v, err := r2.Unpack(); err != nil || v != 99 {
		t.Fatalf("expected 99, got (%v, %v)", v, err)
	}
}

func TestCheck(t *testing.T) {
	errSmall := errors.New("too small")
	errEven := errors.New("must be even")
	rule1 := func(n int) error {
		if n < 10 {
			return errSmall
		}
		return nil
	}
	rule2 := func(n int) error {
		if n%2 != 0 {
			return errEven
		}
		return nil
	}

	// All rules pass
	r1 := mybad.Check(12, rule1, rule2)
	if v, err := r1.Unpack(); err != nil || v != 12 {
		t.Fatalf("expected 12, got (%v, %v)", v, err)
	}

	// Single rule fails
	r2 := mybad.Check(8, rule1, rule2)
	if !errors.Is(r2.Err(), errSmall) {
		t.Fatalf("unexpected error: %v", r2.Err())
	}

	// Multiple rules fail
	r3 := mybad.Check(7, rule1, rule2)
	if r3.Err() == nil {
		t.Fatal("expected aggregated errors, got nil")
	}
	if !errors.Is(r3.Err(), errSmall) || !errors.Is(r3.Err(), errEven) {
		t.Fatalf("expected both errors to be joined, got: %q", r3.Err().Error())
	}
}

func TestAll(t *testing.T) {
	// All healthy
	resultsOK := []mybad.Result[int]{mybad.Ok(1), mybad.Ok(2), mybad.Ok(3)}
	r1 := mybad.All(resultsOK)
	if v, err := r1.Unpack(); err != nil || len(v) != 3 || v[0] != 1 || v[1] != 2 || v[2] != 3 {
		t.Fatalf("expected [1 2 3], got (%v, %v)", v, err)
	}

	// With error -> short-circuits
	resultsErr := []mybad.Result[int]{mybad.Ok(1), mybad.From(0, sentinel), mybad.Ok(3)}
	r2 := mybad.All(resultsErr)
	if !errors.Is(r2.Err(), sentinel) {
		t.Fatalf("expected sentinel, got %v", r2.Err())
	}
}

func TestPartition(t *testing.T) {
	results := []mybad.Result[int]{
		mybad.Ok(1),
		mybad.From(0, sentinel),
		mybad.Ok(3),
		mybad.From(0, errors.New("other")),
	}
	values, errs := mybad.Partition(results)
	if len(values) != 2 || values[0] != 1 || values[1] != 3 {
		t.Fatalf("unexpected partitioned values: %v", values)
	}
	if len(errs) != 2 || !errors.Is(errs[0], sentinel) || errs[1].Error() != "other" {
		t.Fatalf("unexpected partitioned errors: %v", errs)
	}
}

func TestOr_variadic(t *testing.T) {
	// Empty choices -> ErrNoChoices
	r0 := mybad.Or[int]()
	if !errors.Is(r0.Err(), mybad.ErrNoChoices) {
		t.Fatalf("expected ErrNoChoices, got %v", r0.Err())
	}

	// First healthy wins
	r1 := mybad.Or(mybad.From(0, sentinel), mybad.Ok(42), mybad.Ok(99))
	if v, err := r1.Unpack(); err != nil || v != 42 {
		t.Fatalf("expected 42, got (%v, %v)", v, err)
	}

	// All unhealthy -> last error wins
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	r2 := mybad.Or(mybad.Err[int](err1), mybad.Err[int](err2))
	if !errors.Is(r2.Err(), err2) {
		t.Fatalf("expected last error (%v), got %v", err2, r2.Err())
	}
}

// helper for iterators
func seqOf[T any](items ...mybad.Result[T]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for _, r := range items {
			if !yield(r.Unpack()) {
				return
			}
		}
	}
}

func TestAllSeq(t *testing.T) {
	// All healthy
	seqOK := seqOf(mybad.Ok(1), mybad.Ok(2), mybad.Ok(3))
	r1 := mybad.AllSeq(seqOK)
	if v, err := r1.Unpack(); err != nil || len(v) != 3 || v[0] != 1 || v[1] != 2 || v[2] != 3 {
		t.Fatalf("expected [1 2 3], got (%v, %v)", v, err)
	}

	// With error -> short-circuits
	seqErr := seqOf(mybad.Ok(1), mybad.From(0, sentinel), mybad.Ok(3))
	r2 := mybad.AllSeq(seqErr)
	if !errors.Is(r2.Err(), sentinel) {
		t.Fatalf("expected sentinel, got %v", r2.Err())
	}
}

func TestPartitionSeq(t *testing.T) {
	seq := seqOf(
		mybad.Ok(1),
		mybad.From(0, sentinel),
		mybad.Ok(3),
		mybad.From(0, errors.New("other")),
	)
	values, errs := mybad.PartitionSeq(seq)
	if len(values) != 2 || values[0] != 1 || values[1] != 3 {
		t.Fatalf("unexpected partitioned values: %v", values)
	}
	if len(errs) != 2 || !errors.Is(errs[0], sentinel) || errs[1].Error() != "other" {
		t.Fatalf("unexpected partitioned errors: %v", errs)
	}
}

func TestString_stringAmbiguity(t *testing.T) {
	r := mybad.Ok("hello")
	s := r.String()
	if s != "Ok(hello)" {
		t.Fatalf("got %q, want \"Ok(hello)\"", s)
	}
}

// New API enhancements

type customDiagnosticErr struct {
	Code string
}

func (e customDiagnosticErr) Error() string {
	return e.Code
}

func TestFromBool(t *testing.T) {
	// ok = true
	r1 := mybad.FromBool(42, true, sentinel)
	if v, err := r1.Unpack(); err != nil || v != 42 {
		t.Fatalf("expected (42, nil), got (%v, %v)", v, err)
	}

	// ok = false
	r2 := mybad.FromBool(0, false, sentinel)
	if !errors.Is(r2.Err(), sentinel) {
		t.Fatalf("expected sentinel, got %v", r2.Err())
	}

	// ok = false, err = nil -> panic
	defer func() {
		val := recover()
		if val == nil {
			t.Fatal("expected panic on FromBool(0, false, nil), got none")
		}
	}()
	mybad.FromBool(0, false, nil)
}

func TestErrIs(t *testing.T) {
	rOk := mybad.Ok(42)
	rErr := mybad.From(0, fmt.Errorf("wrap: %w", sentinel))

	if rOk.ErrIs(sentinel) {
		t.Fatal("healthy result should always report false for ErrIs")
	}

	if !rErr.ErrIs(sentinel) {
		t.Fatal("unhealthy result should report true for matching sentinel")
	}

	if rErr.ErrIs(errors.New("other")) {
		t.Fatal("unhealthy result should report false for non-matching sentinel")
	}
}

func TestErrAs(t *testing.T) {
	rOk := mybad.Ok(42)
	rErr := mybad.From(0, fmt.Errorf("wrap: %w", customDiagnosticErr{Code: "FAIL_CODE"}))

	var target customDiagnosticErr
	if rOk.ErrAs(&target) {
		t.Fatal("healthy result should always report false for ErrAs")
	}

	if !rErr.ErrAs(&target) || target.Code != "FAIL_CODE" {
		t.Fatalf("unhealthy result should match and unpack CustomDiagnosticErr, target: %v", target)
	}

	var otherErr *strconv.NumError
	if rErr.ErrAs(&otherErr) {
		t.Fatal("unhealthy result should report false for non-matching type")
	}
}

func TestRecoverIs(t *testing.T) {
	errOther := errors.New("other")

	// Match recovery
	r1 := mybad.From(0, sentinel).RecoverIs(sentinel, 99)
	if v, err := r1.Unpack(); err != nil || v != 99 {
		t.Fatalf("expected recovered Result containing 99, got (%v, %v)", v, err)
	}

	// Mismatch recovery
	r2 := mybad.From(0, errOther).RecoverIs(sentinel, 99)
	if !errors.Is(r2.Err(), errOther) {
		t.Fatalf("expected Result carrying errOther, got %v", r2.Err())
	}

	// Healthy Result recovery -> no-op
	r3 := mybad.Ok(42).RecoverIs(sentinel, 99)
	if v, err := r3.Unpack(); err != nil || v != 42 {
		t.Fatalf("expected Result carrying 42, got (%v, %v)", v, err)
	}
}
