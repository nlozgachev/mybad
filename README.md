# mybad
![Go 1.18+](https://img.shields.io/badge/1.18+-0?style=for-the-badge&color=000&logo=Go&label&logoColor=fff&logoSize=auto)[![Go Reference](https://img.shields.io/badge/reference-0?style=for-the-badge&color=000)](https://pkg.go.dev/github.com/nlozgachev/mybad)

```go
go get github.com/nlozgachev/mybad@latest
```

Skip `if err != nil` after every step. Errors propagate through the chain automatically; `Match` forces you to handle them at the end.


## The problem

Multi-step logic that can fail at any point ends up like this:

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

The error checks aren't wrong — they're just noise. The actual logic is buried under identical boilerplate.


## The idea

`Result[T]` is a value that holds either a success (`T`) or an error — never both. You pass it through a chain of steps, and each step only runs if the previous one succeeded. The first failure short-circuits everything that follows: the error skips the remaining steps and travels to the end of the chain untouched.

At the end, `Match` forces you to handle both outcomes before you can get a value out. There's no way to accidentally ignore an error.

This pattern is known as railway-oriented programming.

```go
// after
func handleRequest(userID string) Response {
    user := mybad.Try(mybad.Ok(userID), fetchUser)
    user  = mybad.Try(user, validateUser)
    user  = mybad.Try(user, enrichUser)
    dto  := mybad.Into(user, toDTO)
    dto   = mybad.Try(dto, formatDTO)
    return mybad.Match(dto, okResponse, errResponse)
}
```


## A minimal example

Four functions are enough to get started: `Ok`, `From`, `Try`, and `Match`.

```go
// Ok wraps a value you already have.
// From wraps a (value, error) pair — the shape most Go functions return.
r := mybad.From(strconv.Atoi(input))   // Result[int]

// Try applies a fallible function to the value inside.
// If the Result is already in error state, the function is never called.
r = mybad.Try(r, double)               // func(int) (int, error)
r = mybad.Try(r, clamp)               // skipped entirely if double failed

// Match collapses the Result into a concrete value.
// You must handle both the success and error cases.
result := mybad.Match(r,
    func(n int) string { return fmt.Sprintf("result: %d", n) },
    func(err error) string { return fmt.Sprintf("error: %s", err) },
)
```

If `strconv.Atoi` fails, `double` and `clamp` are never called — `Match` receives the parse error. If `double` fails, `clamp` is never called — `Match` receives that error. Only if every step succeeds does the value reach the `onOk` branch.


## Building pipelines

`Try` applies a fallible function — `func(T) (U, error)`. `Into` applies one that never fails — `func(T) U`. Both pass errors through unchanged, and both support type-changing steps.

```go
user  := mybad.Try(mybad.Ok(userID), fetchUser)    // Result[string] → Result[User]
user   = mybad.Try(user, validateUser)
dto   := mybad.Into(user, toDTO)                    // Result[User] → Result[UserDTO], never fails
dto    = mybad.Try(dto, formatDTO)

response := mybad.Match(dto, okResponse, errResponse)
```

Use `Into` anywhere a step cannot return an error — it communicates intent clearly and removes the need for a dummy `nil` return.


## Working with errors

**Add context** to an error without touching the happy path:

```go
r = mybad.WrapErr(r, func(err error) error {
    return fmt.Errorf("user lookup: %w", err)
})
```

Always wrap with `%w` to keep the original error reachable via `errors.Is` and `errors.As`.

**Recover** from a known error — return a fallback value, or let it propagate:

```go
r = mybad.OrElse(r, func(err error) (User, error) {
    if errors.Is(err, ErrNotFound) {
        return guestUser, nil   // recover: switch to the happy path
    }
    return User{}, err          // propagate: stay in error state
})
```

**Observe** without changing anything — useful for logging at any point in the chain:

```go
r = mybad.Peek(r, func(u User) {
    log.Info("user fetched", "id", u.ID)
})
r = mybad.PeekErr(r, func(err error) {
    log.Error("pipeline failed", "err", err)
})
```

`WrapErr`, `OrElse`, and `PeekErr` are no-ops when the Result is healthy. `Peek` is a no-op when it is in error state. Neither side of the chain interferes with the other.


## Getting a value out

**`Match`** is the primary exit. It handles both outcomes and always returns a concrete value:

```go
response := mybad.Match(dto,
    func(d UserDTO) Response { return okResponse(d) },
    func(err error) Response { return errResponse(err) },
)
```

The rest are escape hatches for specific situations:

```go
// Unwrap: both value and error, back to Go conventions.
// In error state the returned value is the zero value of T; do not use it.
value, err := r.Unwrap()

// ValueOr: value or a static default — use when you don't need the error.
value := r.ValueOr(defaultValue)

// ValueOrElse: value or a default computed from the error.
value := r.ValueOrElse(func(err error) T { return fallback })

// Err: just the error, nil if healthy.
if err := r.Err(); err != nil { ... }

// Must: just the value, panics on error.
// Intended for tests and cases the caller has proven cannot fail.
value := r.Must()

// IsOk / IsErr: check state without extracting anything.
if r.IsOk() { ... }
if r.IsErr() { ... }
```

## Quick reference

|                         |                                                                    |
| ----------------------- | ------------------------------------------------------------------ |
| `Ok(v)`                 | Wrap a value in a healthy Result                                   |
| `From(v, err)`          | Wrap a (value, error) pair                                         |
| `Try(r, fn)`            | Apply a fallible `func(T) (U, error)`; skip if r is in error state |
| `Into(r, fn)`           | Apply a pure `func(T) U`; skip if r is in error state              |
| `WrapErr(r, fn)`        | Transform the error; no-op if healthy                              |
| `OrElse(r, fn)`         | Attempt recovery from error; no-op if healthy                      |
| `Peek(r, fn)`           | Observe the value; no-op if in error state                         |
| `PeekErr(r, fn)`        | Observe the error; no-op if healthy                                |
| `Match(r, onOk, onErr)` | Collapse Result — handle both branches, always returns a value     |
| `r.Unwrap()`            | Returns `(value, error)` — back to Go conventions                  |
| `r.ValueOr(default)`    | Returns value or a static default                                  |
| `r.ValueOrElse(fn)`     | Returns value or `fn(err)`                                         |
| `r.Err()`               | Returns the error, nil if healthy                                  |
| `r.Must()`              | Returns the value, panics if in error state                        |
| `r.IsOk()`              | Reports whether the Result is healthy                              |
| `r.IsErr()`             | Reports whether the Result is in error state                       |
