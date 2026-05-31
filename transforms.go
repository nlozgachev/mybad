package mybad

import (
	"errors"
	"iter"
)

// Try applies a fallible function to the value inside r.
// If r is in error state, it is returned unchanged.
// If fn returns an error, the Result transitions to error state.
// fn may change the type: func(T) (U, error).
func Try[T, U any](r Result[T], fn func(T) (U, error)) Result[U] {
	if r.err != nil {
		return Result[U]{err: r.err}
	}
	return From(fn(r.value))
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

// AndThen applies a function that returns a Result[U] to the value inside r.
// If r is in error state, it is returned unchanged.
// fn may change the type: func(T) Result[U].
func AndThen[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.err != nil {
		return Result[U]{err: r.err}
	}
	return fn(r.value)
}

// FromFunc adapts a standard Go function returning (U, error)
// into a function that returns a Result[U].
func FromFunc[T, U any](fn func(T) (U, error)) func(T) Result[U] {
	return func(val T) Result[U] {
		return From(fn(val))
	}
}

// Match collapses a Result into a single value R by applying onOk to the value
// or onErr to the error. Always returns a concrete value, never an unhandled state.
func Match[T, R any](r Result[T], onOk func(T) R, onErr func(error) R) R {
	if r.err != nil {
		return onErr(r.err)
	}
	return onOk(r.value)
}

// Check runs multiple validation checks on a value.
// If any check fails, it collects and joins all non-nil errors using errors.Join.
// Returns a healthy Result containing the value if all checks pass.
func Check[T any](value T, rules ...func(T) error) Result[T] {
	var errs []error
	for _, rule := range rules {
		if err := rule(value); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return Err[T](errors.Join(errs...))
	}
	return Ok(value)
}

// All performs short-circuiting aggregation over a slice of Results.
// If any Result is unhealthy, returns the first encountered error.
// Otherwise, returns a healthy Result containing a slice of all values.
func All[T any](results []Result[T]) Result[[]T] {
	for _, r := range results {
		if r.err != nil {
			return Err[[]T](r.err)
		}
	}
	values := make([]T, len(results))
	for i, r := range results {
		values[i] = r.value
	}
	return Ok(values)
}

// Partition splits a slice of Results into a slice of healthy values and a slice of errors.
// Does not short-circuit.
func Partition[T any](results []Result[T]) (values []T, errs []error) {
	var numErr int
	for _, r := range results {
		if r.err != nil {
			numErr++
		}
	}
	values = make([]T, 0, len(results)-numErr)
	errs = make([]error, 0, numErr)
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
		} else {
			values = append(values, r.value)
		}
	}
	return values, errs
}

// Or returns the first healthy Result from the choices fallback chain.
// If choices is empty, returns Err[T](ErrNoChoices).
// If all choices are unhealthy, returns a Result carrying the last error.
func Or[T any](choices ...Result[T]) Result[T] {
	if len(choices) == 0 {
		return Err[T](ErrNoChoices)
	}
	var lastErr error
	for _, r := range choices {
		if r.err == nil {
			return r
		}
		lastErr = r.err
	}
	return Err[T](lastErr)
}

// AllSeq performs short-circuiting aggregation over a Go 1.23 native fallible tuple iterator stream.
// If any item yielded contains a non-nil error, returns the first encountered error.
// Otherwise, returns a healthy Result containing a slice of all successfully yielded values.
func AllSeq[T any](seq iter.Seq2[T, error]) Result[[]T] {
	var values []T
	var err error
	seq(func(v T, e error) bool {
		if e != nil {
			err = e
			return false
		}
		values = append(values, v)
		return true
	})
	if err != nil {
		return Err[[]T](err)
	}
	return Ok(values)
}

// PartitionSeq splits a Go 1.23 native fallible tuple iterator stream into standard slices of values and errors.
// Does not short-circuit.
func PartitionSeq[T any](seq iter.Seq2[T, error]) (values []T, errs []error) {
	seq(func(v T, e error) bool {
		if e != nil {
			errs = append(errs, e)
		} else {
			values = append(values, v)
		}
		return true
	})
	return values, errs
}
