# mybad
![Go 1.23+](https://img.shields.io/badge/1.23+-0?style=for-the-badge&color=000&logo=Go&label&logoColor=fff&logoSize=auto)
[![Go Reference](https://img.shields.io/badge/reference-0?style=for-the-badge&color=000)](https://pkg.go.dev/github.com/nlozgachev/mybad/v2)
[![Latest release](https://img.shields.io/github/v/release/nlozgachev/mybad?style=for-the-badge&label=%20&color=000)](https://github.com/nlozgachev/mybad/releases)

```go
go get github.com/nlozgachev/mybad/v2@latest
```

Railway-oriented error handling for Go. Skip `if err != nil` boilerplate after every step. Errors propagate through the pipeline automatically; `Match` forces you to handle them at the boundary.

---

## The Problem

Multi-step logic where each step can fail quickly turns into noise:

```go
func handleRequest(userID string) Response {
    user, err := fetchUser(userID)
    if err != nil {
        return errResponse(err)
    }
    user, err = validateUser(user)
    if err != nil {
        return errResponse(err)
    }
    user, err = enrichUser(user)
    if err != nil {
        return errResponse(err)
    }
    dto := toDTO(user)
    dto, err = formatDTO(dto)
    if err != nil {
        return errResponse(err)
    }
    return okResponse(dto)
}
```

The error checks aren't incorrect—they are just noise. The actual business flow is buried under identical checking boilerplate.

---

## Core Concepts

The mental model is built on a single data carrier—`Result[T]`—which carries either a healthy value (`T`) or a standard `error`, never both.

To write your first pipeline, you only need four components:

1. **`Ok(v)` / `From(v, err)` / `Err[T](err)` (Constructors)**: Wrap raw values or standard Go dual-returns.
2. **`Try(r, fn)` (The Operator)**: Chains a fallible step (`func(T) (U, error)`). If the Result is in an error state, the step is skipped automatically.
3. **`Match(r, okFn, errFn)` (The Terminal)**: Collapses the pipeline into a concrete outcome, forcing you to handle both outcomes.

```go
// 1. Wrap a standard Go dual-return
r := mybad.From(strconv.Atoi(input))    // Result[int]

// 2. Chain fallible operations
r = mybad.Try(r, double)                // func(int) (int, error)
r = mybad.Try(r, clamp)                 // Skipped automatically if double failed

// 3. Handle both outcomes at the boundary
result := mybad.Match(r,
    func(n int) string { return fmt.Sprintf("result: %d", n) },
    func(err error) string { return fmt.Sprintf("failed: %s", err) },
)
```

If `strconv.Atoi` fails, `double` and `clamp` are never executed—`Match` receives the parsing error. The error propagates safely to the end of the chain.

---

## Building Pipelines

Real-world services require pure (infallible) transforms, steps that return Result types directly, and integration with legacy codebases.

#### Infallible Steps (`Into`)
Use `Into` to chain transformations that can never fail (`func(T) U`). It communicates intent clearly and avoids returning dummy `nil` errors:

```go
user  := mybad.Try(mybad.Ok(userID), fetchUser) // Result[User]
dto   := mybad.Into(user, toDTO)                // Result[UserDTO] (infallible)
```

#### Pre-existing Result Producers (`AndThen`)
If a step already returns a `Result[U]` directly (instead of a plain `(U, error)`), use `AndThen`:

```go
// fetchDetails returns Result[Details]
details := mybad.AndThen(user, fetchDetails)
```

#### Adapting Standard Functions (`FromFunc`)
Adapt standard Go functions returning `(U, error)` into pipeline-ready Result producers using `FromFunc`:

```go
// strconv.Atoi is func(string) (int, error)
parseInt := mybad.FromFunc(strconv.Atoi)

r := mybad.AndThen(mybad.Ok("42"), parseInt)
```

#### Fluent Context Wrapping (`WrapErr`)
Add structural context to an error inline within the pipeline using `WrapErr`. Wrap using `%w` to keep the underlying error chain accessible:

```go
r = r.WrapErr(func(err error) error {
    return fmt.Errorf("billing pipeline failed: %w", err)
})
```

---

## Domain Logic & Error Control

As business logic grows, you need validations, selective error recoveries, comma-ok integrations, and diagnostic hooks.

#### Post-Success Validation (`Ensure`)
Assert business rules on a healthy value. If a constraint is violated, the pipeline transitions to an error state:

```go
r = r.Ensure(func(u User) error {
    if u.Age < 18 {
        return errors.New("must be an adult")
    }
    return nil
})
```

#### Selective Error Recovery (`RecoverIs`)
Recover from a specific expected sentinel error (such as `sql.ErrNoRows` or cache misses) by providing a fallback value, while letting unexpected server errors bubble up:

```go
r = r.RecoverIs(sql.ErrNoRows, guestUser)
```

For general error recovery, use `RecoverTry` (fallible fallback) or `Recover` (infallible static fallback):

```go
r = r.RecoverTry(func(err error) (User, error) {
    if errors.Is(err, ErrNetwork) {
        return fetchFromCache() // Try alternative recovery
    }
    return User{}, err
})
```

#### Comma-Ok Integration (`FromBool`)
Wrap Go's comma-ok patterns (such as map lookups, channel receives, and type assertions) seamlessly:

```go
value, ok := cache.Get(key)
r := mybad.FromBool(value, ok, ErrCacheMiss) // Result[Value]
```

#### Fluent Error Assertions (`ErrIs` / `ErrAs`)
Inspect the error chain directly on the Result struct without importing the standard `errors` package:

```go
if r.ErrIs(sql.ErrNoRows) {
    // Handle record not found
}

var validationErr *ValidationError
if r.ErrAs(&validationErr) {
    log.Error("invalid fields", "fields", validationErr.Fields)
}
```

#### Inline Observation (`Peek` / `PeekErr`)
Execute side-effects (such as telemetry, logging, or debugging) at any point in the pipeline without modifying the Result:

```go
r = r.Peek(func(u User) {
    log.Info("fetched user", "id", u.ID)
}).PeekErr(func(err error) {
    log.Error("pipeline failed", "err", err)
})
```

---

## Collections & Streams

Handle batch checks, fallback chains, context cancellations, and high-performance range-over-func iterator streams.

#### Context Verification (`Guard`)
Stop execution if the surrounding `context.Context` is cancelled, transitioning the Result to an error containing `ctx.Err()`:

```go
r = r.Guard(ctx)
```

#### Multi-Constraint Validation (`Check`)
Validate a value against multiple constraints simultaneously. If any checks fail, it collects all non-nil errors and joins them using standard `errors.Join`:

```go
r := mybad.Check(user, validateAge, validateEmail, validatePermissions)
```

#### Batch Aggregations (`All` / `Partition`)
- **`All`**: Aggregates a slice of Results (`[]Result[T]` → `Result[[]T]`). Short-circuits on the first failure. Optimized to prevent heap allocations on early failures.
- **`Partition`**: Non-short-circuiting partition split separating success values and errors (`([]T, []error)`). Pre-allocates destination slices to eliminate growth reallocations.

#### Variadic Fallbacks (`Or`)
Evaluate a chain of alternative choices, returning the first healthy Result. If all choices fail, returns the last encountered error:

```go
r := mybad.Or(readLocalCache, readRedis, fetchFromAPI)
```

#### Zero-Allocation Iterator Streaming (`AllSeq` / `PartitionSeq`)
Consume Go 1.23 native range-over-func fallible streams (`iter.Seq2[T, error]`) directly without intermediate slice allocations:

```go
r := mybad.AllSeq(dbCursor.Iterate()) // Short-circuiting collection
```

---

## Safety Invariants & Guarantees

*   **Zero-Value Safety**: An uninitialized, zero-valued `Result[T]` (declared via `var r Result[T]`) is guaranteed to be a healthy success state containing the zero value of `T` and a `nil` error.
*   **Panic Integrity**: `mybad` never suppresses, catches, or alters panics occurring inside user callbacks (e.g. within `Try`, `Into`, `AndThen`). User panics bubble up naturally to the Go runtime.
*   **String Debug Ambiguity**: The `String()` method implements `fmt.Stringer`, rendering a result as `Ok(value)` or `Err(message)`. Under `Result[string]`, a value like `"hello"` prints as `Ok(hello)`. If you need exact quotation demarcations during debugging, check the value explicitly via `r.Unpack()`.

---

## Quick Reference Index

| Symbol / Construct | Category | Description |
| :--- | :--- | :--- |
| **`Ok(v)`** | Constructor | Wrap a value in a healthy Result |
| **`From(v, err)`** | Constructor | Wrap a `(value, error)` pair |
| **`Err[T](err)`** | Constructor | Wrap a known error; panics on nil |
| **`FromBool(v, ok, err)`** | Constructor | Wrap a `(value, ok)` pair with a fallback error |
| **`FromFunc(fn)`** | Constructor | Adapt a `func(T) (U, error)` into a Result producer |
| **`Try(r, fn)`** | Operator | Apply a fallible transform; skip if in error |
| **`Into(r, fn)`** | Operator | Apply an infallible transform; skip if in error |
| **`AndThen(r, fn)`** | Operator | Chain another Result producer; skip if in error |
| **`Match(r, okFn, errFn)`** | Terminal | Collapse Result; handle both outcomes |
| **`r.WrapErr(fn)`** | Method | Transform the error; no-op if healthy |
| **`r.ErrIs(target)`** | Method | Check if error matches target sentinel (`errors.Is`) |
| **`r.ErrAs(target)`** | Method | Unpack a matching error type in the chain (`errors.As`) |
| **`r.RecoverTry(fn)`** | Method | Attempt fallible recovery; no-op if healthy |
| **`r.Recover(fn)`** | Method | Infallibly recover to a static default; no-op if healthy |
| **`r.RecoverIs(tgt, val)`** | Method | Recover from specific sentinel with static fallback |
| **`r.Ensure(fn)`** | Method | Business validation constraint callback |
| **`r.Guard(ctx)`** | Method | Halt pipeline if context is cancelled |
| **`r.Peek(fn)`** | Method | Observe value; no-op if in error |
| **`r.PeekErr(fn)`** | Method | Observe error; no-op if healthy |
| **`r.Unpack()`** | Extraction | Returns `(value, error)` |
| **`r.ValueOr(default)`** | Extraction | Returns value or static default |
| **`r.ValueOrElse(fn)`** | Extraction | Returns value or callback default |
| **`r.Err()`** | Extraction | Returns error, `nil` if healthy |
| **`r.Must()`** | Extraction | Returns value, panics on error |
| **`r.Expect(msg)`** | Extraction | Returns value, panics with custom message on error |
| **`r.IsOk()`** | Assertion | Reports whether Result is healthy |
| **`r.IsErr()`** | Assertion | Reports whether Result is in error |
