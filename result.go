package mybad

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

// Must returns the value, panicking if the Result is in error state.
// Intended for tests and cases the caller has proven cannot fail.
func (r Result[T]) Must() T {
	if r.err != nil {
		panic(r.err)
	}
	return r.value
}

// Unwrap returns both the value and the error, surrendering back to Go conventions.
// In error state the returned value is the zero value of T; do not use it.
func (r Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}
