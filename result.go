package mybad

import (
	"context"
	"errors"
	"fmt"
)

// ErrNoChoices is returned when Or is called with no choices.
var ErrNoChoices = errors.New("no choices provided to Or")

// Result carries either a value of type T or an error, never both.
// Fields are unexported; use terminals to extract the value.
type Result[T any] struct {
	value T
	err   error
}

// Ok wraps a value in a healthy Result.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// From wraps a (value, error) pair, the natural shape of most Go functions.
// If err is non-nil the Result is in error state and value is ignored.
func From[T any](value T, err error) Result[T] {
	if err != nil {
		return Result[T]{err: err}
	}
	return Result[T]{value: value}
}

// FromBool wraps a (value, ok) pair, transforming the absence of a value (ok == false)
// into an unhealthy Result carrying the provided err.
// If ok is false and err is nil, it panics.
func FromBool[T any](value T, ok bool, err error) Result[T] {
	if !ok {
		return Err[T](err)
	}
	return Ok(value)
}

// Err wraps an error in an unhealthy Result of type T.
// Panics if err is nil.
func Err[T any](err error) Result[T] {
	if err == nil {
		panic("mybad: Err called with nil error")
	}
	return Result[T]{err: err}
}

// IsOk reports whether the Result is in a healthy (non-error) state.
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr reports whether the Result is in an error state.
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Err returns the error, or nil if the Result is healthy.
func (r Result[T]) Err() error {
	return r.err
}

// ErrIs reports whether the Result's error matches target (via errors.Is).
// Always reports false if the Result is healthy.
func (r Result[T]) ErrIs(target error) bool {
	if r.err == nil {
		return false
	}
	return errors.Is(r.err, target)
}

// ErrAs finds the first error in the Result's error chain that matches target (via errors.As).
// If found, it sets target to that error value and returns true.
// Always reports false if the Result is healthy.
func (r Result[T]) ErrAs(target any) bool {
	if r.err == nil {
		return false
	}
	return errors.As(r.err, target)
}

// Must returns the value, panicking if the Result is in error state.
// Intended for tests and cases the caller has proven cannot fail.
func (r Result[T]) Must() T {
	if r.err != nil {
		panic(r.err)
	}
	return r.value
}

// Unpack returns both the value and the error, surrendering back to Go conventions.
// In error state the returned value is the zero value of T; do not use it.
func (r Result[T]) Unpack() (T, error) {
	return r.value, r.err
}

// ValueOr returns the value if the Result is healthy, or defaultValue if it is in error state.
// The error is not surfaced; use Match or Unpack if you need it.
func (r Result[T]) ValueOr(defaultValue T) T {
	if r.err != nil {
		return defaultValue
	}
	return r.value
}

// ValueOrElse returns the value if the Result is healthy, or calls fn with the error
// and returns its result. fn is not called if the Result is healthy.
func (r Result[T]) ValueOrElse(fn func(error) T) T {
	if r.err != nil {
		return fn(r.err)
	}
	return r.value
}

// WrapErr transforms the error inside r.
// No-op if r is in healthy state.
// If fn returns a new error without wrapping the original (e.g. without fmt.Errorf("%w", err)),
// the original error will no longer be reachable via errors.Is or errors.As.
func (r Result[T]) WrapErr(fn func(error) error) Result[T] {
	if r.err == nil {
		return r
	}
	return Result[T]{err: fn(r.err)}
}

// Peek calls fn with the value inside r for observation.
// No-op if r is in error state. The Result is always returned unchanged.
func (r Result[T]) Peek(fn func(T)) Result[T] {
	if r.err == nil {
		fn(r.value)
	}
	return r
}

// PeekErr calls fn with the error inside r for observation.
// No-op if r is in healthy state. The Result is always returned unchanged.
func (r Result[T]) PeekErr(fn func(error)) Result[T] {
	if r.err != nil {
		fn(r.err)
	}
	return r
}

// RecoverTry attempts to recover from an error state by calling fn.
// No-op if r is healthy. If fn itself returns an error, the Result stays in error state.
func (r Result[T]) RecoverTry(fn func(error) (T, error)) Result[T] {
	if r.err == nil {
		return r
	}
	value, err := fn(r.err)
	if err != nil {
		return Result[T]{err: err}
	}
	return Result[T]{value: value}
}

// Expect returns the value, panicking with a custom message and the raw error if in error state.
// Intended for cases the caller has proven cannot fail.
func (r Result[T]) Expect(msg string) T {
	if r.err != nil {
		panic(fmt.Errorf("%s: %w", msg, r.err))
	}
	return r.value
}

// Ensure runs check on the value if Result is healthy.
// If check returns an error, the Result transitions to error state.
func (r Result[T]) Ensure(check func(T) error) Result[T] {
	if r.err != nil {
		return r
	}
	if err := check(r.value); err != nil {
		return Result[T]{err: err}
	}
	return r
}

// Recover maps an error state to a healthy Result using fn, which cannot fail.
// No-op if Result is healthy.
func (r Result[T]) Recover(fn func(error) T) Result[T] {
	if r.err == nil {
		return r
	}
	return Result[T]{value: fn(r.err)}
}

// RecoverIs recovers from a specific error matching target (via errors.Is) by
// transitioning to a healthy Result containing the fallback value.
// If the Result is healthy, or if the error does not match target, returns the Result unchanged.
func (r Result[T]) RecoverIs(target error, fallback T) Result[T] {
	if r.err == nil {
		return r
	}
	if errors.Is(r.err, target) {
		return Ok(fallback)
	}
	return r
}

// Guard short-circuits the pipeline if the context is cancelled, returning ctx.Err().
// If the Result is healthy and ctx is cancelled, it transitions to an error state with ctx.Err().
// Otherwise, returns the Result unchanged.
func (r Result[T]) Guard(ctx context.Context) Result[T] {
	if r.err != nil {
		return r
	}
	if err := ctx.Err(); err != nil {
		return Result[T]{err: err}
	}
	return r
}

// Or returns the current result if healthy, or the fallback choice if unhealthy.
func (r Result[T]) Or(fallback Result[T]) Result[T] {
	if r.err == nil {
		return r
	}
	return fallback
}

// String implements fmt.Stringer, returning a debug summary ("Ok(value)" or "Err(message)").
func (r Result[T]) String() string {
	if r.err != nil {
		return fmt.Sprintf("Err(%v)", r.err)
	}
	return fmt.Sprintf("Ok(%v)", r.value)
}
