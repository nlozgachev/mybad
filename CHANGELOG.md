# Changelog

All notable changes to the `mybad` library will be documented in this file.

---

## [2.0.0] - 2026-05-31

`v2.0.0` is a major release introducing breaking API renames to achieve perfect naming ergonomics, fixing a core invariant bug in `Try`, adding support for Go 1.23 native range-over-func iterators, and adding context and collection utilities.

### ⚠️ Breaking Changes & Migration Guide

#### 1. Method Rename: `OrElse` → `RecoverTry`
*   **Why**: The "or-shaped" name cluster was crowded (`OrElse`, `Or` method, `Or` variadic). `RecoverTry` pairs logically with the pure `Recover` method and groups them under `Recover...` in IDE autocomplete.
*   **Migration**:
    ```diff
    -r = r.OrElse(func(err error) (User, error) { ... })
    +r = r.RecoverTry(func(err error) (User, error) { ... })
    ```

#### 2. Method Rename: `Unwrap` → `Unpack`
*   **Why**: Avoids conceptual confusion with standard Go's `errors.Unwrap` interface. Better describes dismantling the `Result` into a Go raw `(T, error)` tuple.
*   **Migration**:
    ```diff
    -val, err := r.Unwrap()
    +val, err := r.Unpack()
    ```

---

### 🐛 Bug Fixes & Invariant Protections

*   **`Try` Invariant Violation Fixed**: Currently, if the mapping callback `fn` in `Try(r, fn)` returns an error alongside a partial value, the returned Result retained a non-zero value alongside the non-nil error, violating the strict "never both" invariant of `Result[T]`. This has been fixed: `Try` now returns `From(v, e)`, ensuring the value channel is cleanly zero-initialized in error state.

---

### 🚀 New Features & Enhancements

#### 1. Go 1.23 Ergonomic Methods (Migrated to Methods)
To enable fluent chaining, non-type-changing operations have been migrated from package-level functions to methods:
*   **`r.WrapErr(fn)`**: Transforms the error inside a Result; no-op if healthy.
*   **`r.Peek(fn)` / `r.PeekErr(fn)`**: Observes the value or error without modifying pipeline state.

#### 2. Fluent Assertion, Validation, and Pure Recovery Methods
*   **`r.Expect(msg)`**: Safe value extractor that panics with a clear developer message on failure, easing logs/crash analysis.
*   **`r.Ensure(check)`**: Post-success validation check. Transitions a healthy Result to an error state if the rule fails.
*   **`r.Recover(fn)`**: Pure infallible recovery that transforms an error state back to success without returning an error.
*   **`r.RecoverIs(target, fallback)`**: Declarative recovery from a specific expected sentinel error (such as `sql.ErrNoRows`), while letting unexpected system failures bubble up.
*   **`r.ErrIs(target)`**: Fluent inspection method checking if the Result's error matches a sentinel (via standard `errors.Is`).
*   **`r.ErrAs(target)`**: Fluent inspection method matching and extracting custom error types from the Result's error chain (via standard `errors.As`).
*   **`r.String()`**: Implements `fmt.Stringer` for clean debug representations: `Ok(value)` / `Err(message)`.

#### 3. Chaining & Adapter Functions
*   **`AndThen(r, fn)`**: Package-level transform to cleanly chain functions returning `Result[U]` directly.
*   **`FromFunc(fn)`**: Adapts standard Go functions returning `(U, error)` into `Result[U]` producers.
*   **`FromBool(val, ok, err)`**: Directly bridges standard Go comma-ok patterns (maps, channel receives, type assertions) into pipelines.

#### 4. Context & Symmetrical Constructors
*   **`Err[T](err)`**: Symmetrical error constructor. Catches programmer misuse early by panicking on `nil` errors with `mybad: Err called with nil error`.
*   **`r.Guard(ctx)`**: Restricts step execution by context cancellation status.
*   **`r.Or(fallback)`**: Symmetrical fallback method on a Result.

#### 5. Collection & Validation Accumulators
*   **`Check[T](value, rules...)`**: Validates a value against multiple rules, joining all non-nil errors using standard `errors.Join`.
*   **`All[T]([]Result[T])`**: Short-circuiting slice aggregation: `[]Result[T]` → `Result[[]T]`. Optimized to prevent heap allocations on early failure paths.
*   **`Partition[T]([]Result[T])`**: Non-short-circuiting partition split returning standard `([]T, []error)` slices. Optimized to pre-allocate exact capacities, completely avoiding slice growth reallocations.
*   **`Or[T](choices...)`**: Variadic fallback evaluation. Returns the first healthy Result, surfaces the last error if all fail, and returns `Err[T](ErrNoChoices)` if called with no arguments.

#### 6. Zero-Allocation Go 1.23 Iterator Streaming
Added native streaming adapters consuming native fallible range-over-func sequences:
*   **`AllSeq[T](iter.Seq2[T, error]) Result[[]T]`**: Aggregates a fallible stream. Short-circuits on the first error.
*   **`PartitionSeq[T](iter.Seq2[T, error]) ([]T, []error)`**: Splits fallible streams into healthy values and standard errors without short-circuiting.
