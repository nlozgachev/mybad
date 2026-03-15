# mybad
![Go 1.18+](https://img.shields.io/badge/1.18+-0?style=for-the-badge&color=000&logo=Go&label&logoColor=fff&logoSize=auto)[![Go Reference](https://img.shields.io/badge/reference-0?style=for-the-badge&color=000)](https://pkg.go.dev/github.com/nlozgachev/mybad)


```go
go get github.com/nlozgachev/mybad@latest
```

Railway-oriented error handling for Go. Instead of checking errors after every step, you compose a pipeline of typed transforms. The first error short-circuits everything that follows; `Match` forces you to handle both outcomes before you can get a value out.


## Before / after

```go
// before
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


## Starting a chain

```go
// Wrap a value you already have
r := mybad.Ok(userID)

// Wrap a (value, error) pair, the natural shape of most Go functions
r := mybad.From(fetchUser(userID))

```


## Stepping through

`Try` applies a fallible function. `Into` applies one that cannot fail. Both are no-ops if the `Result` is already in error state; the rest of the chain is skipped automatically.

```go
user := mybad.Try(mybad.Ok(userID), fetchUser)   // fetchUser: func(string) (User, error)
user  = mybad.Try(user, validateUser)
dto  := mybad.Into(user, toDTO)                   // toDTO: func(User) UserDTO, never fails
dto   = mybad.Try(dto, formatDTO)
```

Both functions support type-changing steps:

```go
// Result[int] → Result[string]
label := mybad.Into(mybad.Ok(42), strconv.Itoa)

// Result[string] → Result[User]
user := mybad.Try(mybad.Ok(userID), fetchUser)
```


## Working with the error track

```go
// Add context to an error without affecting the happy path
r = mybad.WrapErr(r, func(err error) error {
    return fmt.Errorf("config lookup: %w", err)
})

// Attempt recovery: return a fallback or let the error propagate
r = mybad.OrElse(r, func(err error) (User, error) {
    if errors.Is(err, ErrNotFound) {
        return guestUser, nil
    }
    return User{}, err
})

// Observe the value or the error without changing anything
r = mybad.Peek(r, func(u User) { log.Info("fetched user", u.ID) })
r = mybad.PeekErr(r, func(err error) { log.Error("pipeline failed", err) })
```

`WrapErr`, `OrElse`, and `PeekErr` are no-ops when the `Result` is healthy. `Peek` is a no-op when it is in error state.


## Getting out

```go
// Match: provide a handler for each branch, always returns a concrete value
response := mybad.Match(dto,
    func(d UserDTO) Response { return okResponse(d) },
    func(err error) Response { return errResponse(err) },
)

// Unwrap: both value and error, back to Go conventions
// In error state the returned value is the zero value of T.
value, err := r.Unwrap()

// Err: just the error, nil if healthy
if err := r.Err(); err != nil { ... }

// Must: just the value, panics on error
// Intended for tests and cases proven safe by earlier logic.
value := r.Must()

// IsOk / IsErr: check state without extracting anything
if r.IsOk() { ... }
if r.IsErr() { ... }
```

