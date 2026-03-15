package mybad_test

import (
	"errors"
	"fmt"
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
	if v, err := r.Unwrap(); err != nil || v != 42 {
		t.Fatalf("got (%v, %v), want (42, nil)", v, err)
	}
}

func TestFrom_healthy(t *testing.T) {
	r := mybad.From(99, nil)
	if v, err := r.Unwrap(); err != nil || v != 99 {
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

func TestUnwrap(t *testing.T) {
	v, err := mybad.Ok(3).Unwrap()
	if v != 3 || err != nil {
		t.Fatalf("got (%v, %v), want (3, nil)", v, err)
	}
}

func TestUnwrap_error(t *testing.T) {
	v, err := mybad.From(0, sentinel).Unwrap()
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel, got %v", err)
	}
	if v != 0 {
		t.Fatalf("expected zero value, got %v", v)
	}
}

// Try

func TestTry_healthy_success(t *testing.T) {
	r := mybad.Try(mybad.Ok(1), inc)
	if v, err := r.Unwrap(); err != nil || v != 2 {
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
	if v, err := r.Unwrap(); err != nil || v != "5" {
		t.Fatalf("got (%v, %v), want (\"5\", nil)", v, err)
	}
}

// Into

func TestInto_healthy(t *testing.T) {
	r := mybad.Into(mybad.Ok(3), double)
	if v, err := r.Unwrap(); err != nil || v != 6 {
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
	if v, err := r.Unwrap(); err != nil || v != "3" {
		t.Fatalf("got (%v, %v), want (\"3\", nil)", v, err)
	}
}

// WrapErr

func TestWrapErr_healthy_noop(t *testing.T) {
	r := mybad.WrapErr(mybad.Ok(1), func(err error) error {
		t.Fatal("should not be called")
		return err
	})
	if r.Err() != nil {
		t.Fatalf("expected nil, got %v", r.Err())
	}
}

func TestWrapErr_wraps_error(t *testing.T) {
	wrapped := errors.New("wrapped")
	r := mybad.WrapErr(mybad.From(0, sentinel), func(err error) error {
		return wrapped
	})
	if !errors.Is(r.Err(), wrapped) {
		t.Fatalf("expected wrapped, got %v", r.Err())
	}
}

func TestWrapErr_preserves_chain(t *testing.T) {
	r := mybad.WrapErr(mybad.From(0, sentinel), func(err error) error {
		return fmt.Errorf("context: %w", err)
	})
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("expected sentinel to remain reachable via errors.Is, got %v", r.Err())
	}
}

// Peek

func TestPeek_healthy_calls_fn(t *testing.T) {
	called := false
	r := mybad.Peek(mybad.Ok(42), func(n int) {
		called = true
		if n != 42 {
			t.Fatalf("got %v, want 42", n)
		}
	})
	if !called {
		t.Fatal("fn should be called")
	}
	if v, _ := r.Unwrap(); v != 42 {
		t.Fatalf("result value changed: got %v", v)
	}
}

func TestPeek_error_skips_fn(t *testing.T) {
	r := mybad.Peek(mybad.From(0, sentinel), func(n int) {
		t.Fatal("should not be called")
	})
	if !errors.Is(r.Err(), sentinel) {
		t.Fatalf("expected sentinel to be preserved, got %v", r.Err())
	}
}

// PeekErr

func TestPeekErr_error_calls_fn(t *testing.T) {
	called := false
	r := mybad.PeekErr(mybad.From(0, sentinel), func(err error) {
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
	mybad.PeekErr(mybad.Ok(1), func(err error) {
		t.Fatal("should not be called")
	})
}

// OrElse

func TestOrElse_healthy_noop(t *testing.T) {
	r := mybad.OrElse(mybad.Ok(1), func(err error) (int, error) {
		t.Fatal("should not be called")
		return 0, nil
	})
	if v, _ := r.Unwrap(); v != 1 {
		t.Fatalf("got %v, want 1", v)
	}
}

func TestOrElse_recovers(t *testing.T) {
	r := mybad.OrElse(mybad.From(0, sentinel), func(err error) (int, error) {
		return 99, nil
	})
	if v, err := r.Unwrap(); err != nil || v != 99 {
		t.Fatalf("got (%v, %v), want (99, nil)", v, err)
	}
}

func TestOrElse_fn_fails(t *testing.T) {
	other := errors.New("other")
	r := mybad.OrElse(mybad.From(0, sentinel), func(err error) (int, error) {
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
	user := mybad.Try(mybad.Ok(1), inc)  // 1 -> 2
	user = mybad.Try(user, inc)               // 2 -> 3
	label := mybad.Into(user, itoa)           // 3 -> "3"

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
