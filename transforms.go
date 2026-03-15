package mybad

// Try applies a fallible function to the value inside r.
// If r is in error state, it is returned unchanged.
// If fn returns an error, the Result transitions to error state.
// fn may change the type: func(T) (U, error).
func Try[T, U any](r Result[T], fn func(T) (U, error)) Result[U] {
	if r.err != nil {
		return Result[U]{err: r.err}
	}
	value, err := fn(r.value)
	if err != nil {
		return Result[U]{err: err}
	}
	return Result[U]{value: value}
}

// Into applies a pure function to the value inside r.
// If r is in error state, it is returned unchanged.
// fn never fails and may change the type: func(T) U.
func Into[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.err != nil {
		return Result[U]{err: r.err}
	}
	return Result[U]{value: fn(r.value)}
}

// WrapErr transforms the error inside r.
// No-op if r is in healthy state.
// If fn returns a new error without wrapping the original (e.g. without fmt.Errorf("%w", err)),
// the original error will no longer be reachable via errors.Is or errors.As.
func WrapErr[T any](r Result[T], fn func(error) error) Result[T] {
	if r.err == nil {
		return r
	}
	return Result[T]{err: fn(r.err)}
}

// Peek calls fn with the value inside r for observation.
// No-op if r is in error state. The Result is always returned unchanged.
func Peek[T any](r Result[T], fn func(T)) Result[T] {
	if r.err == nil {
		fn(r.value)
	}
	return r
}

// PeekErr calls fn with the error inside r for observation.
// No-op if r is in healthy state. The Result is always returned unchanged.
func PeekErr[T any](r Result[T], fn func(error)) Result[T] {
	if r.err != nil {
		fn(r.err)
	}
	return r
}

// OrElse attempts to recover from an error state by calling fn.
// No-op if r is healthy. If fn itself returns an error, the Result stays in error state.
func OrElse[T any](r Result[T], fn func(error) (T, error)) Result[T] {
	if r.err == nil {
		return r
	}
	value, err := fn(r.err)
	if err != nil {
		return Result[T]{err: err}
	}
	return Result[T]{value: value}
}

// Match collapses a Result into a single value R by applying onOk to the value
// or onErr to the error. Always returns a concrete value, never an unhandled state.
func Match[T, R any](r Result[T], onOk func(T) R, onErr func(error) R) R {
	if r.err != nil {
		return onErr(r.err)
	}
	return onOk(r.value)
}
