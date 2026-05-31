package mybad_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nlozgachev/mybad/v2"
)

func ExampleOk() {
	r := mybad.Ok(42)
	fmt.Println(r.Must())
	// Output: 42
}

func ExampleFrom() {
	r := mybad.From(strconv.Atoi("7"))
	fmt.Println(r.Must())
	// Output: 7
}

func ExampleFrom_error() {
	r := mybad.From(strconv.Atoi("not-a-number"))
	fmt.Println(r.IsErr())
	// Output: true
}

func ExampleResult_IsOk() {
	fmt.Println(mybad.Ok(1).IsOk())
	fmt.Println(mybad.From(0, errors.New("oops")).IsOk())
	// Output:
	// true
	// false
}

func ExampleResult_IsErr() {
	fmt.Println(mybad.From(0, errors.New("oops")).IsErr())
	fmt.Println(mybad.Ok(1).IsErr())
	// Output:
	// true
	// false
}

func ExampleResult_Err() {
	r := mybad.From(0, errors.New("oops"))
	fmt.Println(r.Err())
	// Output: oops
}

func ExampleResult_Must() {
	fmt.Println(mybad.Ok("hello").Must())
	// Output: hello
}

func ExampleResult_Unpack() {
	v, err := mybad.Ok(3).Unpack()
	fmt.Println(v, err)
	// Output: 3 <nil>
}

func ExampleResult_ValueOr() {
	fmt.Println(mybad.Ok(5).ValueOr(99))
	fmt.Println(mybad.From(0, errors.New("oops")).ValueOr(99))
	// Output:
	// 5
	// 99
}

func ExampleResult_ValueOrElse() {
	v := mybad.From(0, errors.New("oops")).ValueOrElse(func(err error) int {
		return -1
	})
	fmt.Println(v)
	// Output: -1
}

func ExampleTry() {
	r := mybad.Try(mybad.Ok(1), func(n int) (int, error) {
		return n + 1, nil
	})
	fmt.Println(r.Must())
	// Output: 2
}

func ExampleInto() {
	r := mybad.Into(mybad.Ok(42), strconv.Itoa)
	fmt.Println(r.Must())
	// Output: 42
}

func ExampleResult_WrapErr() {
	r := mybad.From(0, errors.New("raw")).WrapErr(
		func(err error) error { return fmt.Errorf("context: %w", err) },
	)
	fmt.Println(r.Err())
	// Output: context: raw
}

func ExampleResult_Peek() {
	mybad.Ok(7).Peek(func(n int) {
		fmt.Println("value:", n)
	})
	// Output: value: 7
}

func ExampleResult_PeekErr() {
	mybad.From(0, errors.New("oops")).PeekErr(func(err error) {
		fmt.Println("error:", err)
	})
	// Output: error: oops
}

func ExampleResult_RecoverTry() {
	r := mybad.From(0, errors.New("oops")).RecoverTry(func(err error) (int, error) {
		return 99, nil
	})
	fmt.Println(r.Must())
	// Output: 99
}

func ExampleAndThen() {
	parse := func(s string) mybad.Result[int] {
		return mybad.From(strconv.Atoi(s))
	}

	divide := func(n int) mybad.Result[int] {
		if n == 0 {
			return mybad.From(0, errors.New("division by zero"))
		}
		return mybad.Ok(100 / n)
	}

	r := mybad.AndThen(mybad.Ok("4"), parse)
	r = mybad.AndThen(r, divide)

	fmt.Println(r.Must())
	// Output: 25
}

func ExampleResult_Expect() {
	v := mybad.Ok("config-value").Expect("hardcoded fallback is always safe")
	fmt.Println(v)
	// Output: config-value
}

func ExampleResult_Ensure() {
	validateAge := func(age int) error {
		if age < 18 {
			return errors.New("must be an adult")
		}
		return nil
	}

	r1 := mybad.Ok(25).Ensure(validateAge)
	fmt.Println("r1 is ok:", r1.IsOk())

	r2 := mybad.Ok(15).Ensure(validateAge)
	fmt.Println("r2 is ok:", r2.IsOk())
	fmt.Println("r2 error:", r2.Err())

	// Output:
	// r1 is ok: true
	// r2 is ok: false
	// r2 error: must be an adult
}

func ExampleResult_Recover() {
	// A pure recovery that guarantees a safe fallback value
	r := mybad.From(0, errors.New("not found")).Recover(func(err error) int {
		return 99 // Safe static fallback
	})
	fmt.Println(r.Must())
	// Output: 99
}

func ExampleFromFunc() {
	// FromFunc adapts standard strconv.Atoi func(string) (int, error)
	// into a mybad pipeline-friendly component: func(string) Result[int]
	liftedAtoi := mybad.FromFunc(strconv.Atoi)

	r := liftedAtoi("42")
	fmt.Println(r.Must())
	// Output: 42
}

func ExampleResult_String() {
	r1 := mybad.Ok(42)
	r2 := mybad.From(0, errors.New("oops"))

	fmt.Println(r1)
	fmt.Println(r2)
	// Output:
	// Ok(42)
	// Err(oops)
}

type CustomError struct {
	Code string
}

func (e CustomError) Error() string {
	return fmt.Sprintf("code: %s", e.Code)
}

func ExampleResult_errorsInterop() {
	sentinelErr := errors.New("unauthorized")

	// Create an error result containing a wrapped CustomError
	r := mybad.From(0, fmt.Errorf("context: %w", CustomError{Code: "AUTH_FAILED"}))

	// Check using errors.As
	var cErr CustomError
	if errors.As(r.Err(), &cErr) {
		fmt.Println("Found custom error code:", cErr.Code)
	}

	// Create another error result containing wrapped sentinel
	r2 := mybad.From(0, fmt.Errorf("pipeline block: %w", sentinelErr))

	// Check using errors.Is
	if errors.Is(r2.Err(), sentinelErr) {
		fmt.Println("Found sentinel error: unauthorized")
	}

	// Output:
	// Found custom error code: AUTH_FAILED
	// Found sentinel error: unauthorized
}

func ExampleMatch() {
	result := mybad.Match(mybad.Ok(10),
		func(n int) string { return fmt.Sprintf("ok: %d", n) },
		func(err error) string { return "err" },
	)
	fmt.Println(result)
	// Output: ok: 10
}

func ExampleErr() {
	r := mybad.Err[int](errors.New("custom error"))
	fmt.Println(r.IsErr(), r.Err())
	// Output: true custom error
}

func ExampleResult_Guard() {
	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel()

	r1 := mybad.Ok(42).Guard(context.Background())
	r2 := mybad.Ok(42).Guard(ctxCancel)

	fmt.Println(r1.IsOk(), r1.Must())
	fmt.Println(r2.IsErr(), r2.Err())
	// Output:
	// true 42
	// true context canceled
}

func ExampleResult_Or() {
	r1 := mybad.Ok(123).Or(mybad.Ok(999))
	r2 := mybad.Err[int](errors.New("fail")).Or(mybad.Ok(999))

	fmt.Println(r1.Must())
	fmt.Println(r2.Must())
	// Output:
	// 123
	// 999
}

func ExampleCheck() {
	rule1 := func(s string) error {
		if len(s) < 3 {
			return errors.New("too short")
		}
		return nil
	}
	rule2 := func(s string) error {
		if !strconv.CanBackquote(s) { // simplified check
			return errors.New("invalid chars")
		}
		return nil
	}

	r := mybad.Check("go", rule1, rule2)
	fmt.Println(r.IsErr(), r.Err())
	// Output: true too short
}

func ExampleAll() {
	results := []mybad.Result[int]{
		mybad.Ok(10),
		mybad.Ok(20),
	}
	r := mybad.All(results)
	fmt.Println(r.Must())
	// Output: [10 20]
}

func ExamplePartition() {
	results := []mybad.Result[int]{
		mybad.Ok(10),
		mybad.Err[int](errors.New("fail")),
		mybad.Ok(20),
	}
	vals, errs := mybad.Partition(results)
	fmt.Println(vals)
	fmt.Println(errs)
	// Output:
	// [10 20]
	// [fail]
}

func ExampleOr() {
	r := mybad.Or(
		mybad.Err[int](errors.New("err1")),
		mybad.Ok(55),
		mybad.Ok(99),
	)
	fmt.Println(r.Must())
	// Output: 55
}

func ExampleAllSeq() {
	seq := func(yield func(int, error) bool) {
		yield(10, nil)
		yield(20, nil)
	}
	r := mybad.AllSeq(seq)
	fmt.Println(r.Must())
	// Output: [10 20]
}

func ExamplePartitionSeq() {
	seq := func(yield func(int, error) bool) {
		yield(10, nil)
		yield(0, errors.New("oops"))
		yield(20, nil)
	}
	vals, errs := mybad.PartitionSeq(seq)
	fmt.Println(vals)
	fmt.Println(errs)
	// Output:
	// [10 20]
	// [oops]
}

func ExampleErr_panic() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panicked:", r)
		}
	}()
	mybad.Err[int](nil)
	// Output: Panicked: mybad: Err called with nil error
}

type TransactionReq struct {
	UserID string
	Amount int
}
type User struct {
	ID      string
	Balance int
}
type Receipt struct {
	TransactionID string
	AmountPaid    int
	Remaining     int
}

func parseRequest(input string) (TransactionReq, error) {
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return TransactionReq{}, errors.New("malformed raw request")
	}
	amount, err := strconv.Atoi(parts[1])
	if err != nil {
		return TransactionReq{}, fmt.Errorf("invalid amount format: %w", err)
	}
	return TransactionReq{UserID: parts[0], Amount: amount}, nil
}

func fetchUser(userID string) mybad.Result[User] {
	if strings.ToLower(userID) == "blocked-user" {
		return mybad.From(User{}, errors.New("user account is frozen"))
	}
	return mybad.Ok(User{ID: userID, Balance: 500})
}

func processPayment(user User, amount int) (Receipt, error) {
	if user.Balance < amount {
		return Receipt{}, errors.New("insufficient balance")
	}
	return Receipt{
		TransactionID: fmt.Sprintf("TXN-%s-999", user.ID),
		AmountPaid:    amount,
		Remaining:     user.Balance - amount,
	}, nil
}

func processTransaction(rawInput string) string {
	parse := mybad.FromFunc(parseRequest)

	pipeline := mybad.Into(
		parse(rawInput).Ensure(func(req TransactionReq) error {
			if req.Amount <= 0 {
				return errors.New("transaction amount must be positive")
			}
			return nil
		}),
		func(req TransactionReq) TransactionReq {
			req.UserID = strings.TrimSpace(strings.ToUpper(req.UserID))
			return req
		},
	)

	user := mybad.AndThen(pipeline, func(req TransactionReq) mybad.Result[User] {
		return fetchUser(req.UserID)
	})

	amountToPay := pipeline.ValueOr(TransactionReq{}).Amount

	receipt := mybad.Try(user, func(u User) (Receipt, error) {
		return processPayment(u, amountToPay)
	}).
		WrapErr(func(err error) error {
			return fmt.Errorf("billing error: %w", err)
		}).
		RecoverTry(func(err error) (Receipt, error) {
			if strings.Contains(err.Error(), "insufficient balance") {
				u := user.Must()
				return Receipt{
					TransactionID: fmt.Sprintf("TXN-%s-OVERDRAFT", u.ID),
					AmountPaid:    amountToPay,
					Remaining:     u.Balance - amountToPay,
				}, nil
			}
			return Receipt{}, err
		}).
		Peek(func(r Receipt) {
			fmt.Printf("[TELEMETRY] Payment processed. TxnID: %s\n", r.TransactionID)
		}).
		PeekErr(func(err error) {
			fmt.Printf("[ALERT] System failed to process transaction: %v\n", err)
		})

	return mybad.Match(receipt,
		func(r Receipt) string {
			return fmt.Sprintf("Success! Receipt ID: %s. Balance remaining: $%d", r.TransactionID, r.Remaining)
		},
		func(err error) string {
			return fmt.Sprintf("Failure: %v", err)
		},
	)
}

func Example_transactionPipeline() {
	// Successful standard transaction
	fmt.Println(processTransaction("alice:100"))
	fmt.Println("---")
	// Overdraft transaction (recovered cleanly)
	fmt.Println(processTransaction("bob:600"))
	// ---
	// Failing frozen user transaction
	fmt.Println("---")
	fmt.Println(processTransaction("blocked-user:50"))

	// Output:
	// [TELEMETRY] Payment processed. TxnID: TXN-ALICE-999
	// Success! Receipt ID: TXN-ALICE-999. Balance remaining: $400
	// ---
	// [TELEMETRY] Payment processed. TxnID: TXN-BOB-OVERDRAFT
	// Success! Receipt ID: TXN-BOB-OVERDRAFT. Balance remaining: $-100
	// ---
	// [ALERT] System failed to process transaction: billing error: user account is frozen
	// Failure: billing error: user account is frozen
}
