package mybad_test

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/nlozgachev/mybad"
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

func ExampleResult_Unwrap() {
	v, err := mybad.Ok(3).Unwrap()
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

func ExampleWrapErr() {
	r := mybad.WrapErr(
		mybad.From(0, errors.New("raw")),
		func(err error) error { return fmt.Errorf("context: %w", err) },
	)
	fmt.Println(r.Err())
	// Output: context: raw
}

func ExamplePeek() {
	mybad.Peek(mybad.Ok(7), func(n int) {
		fmt.Println("value:", n)
	})
	// Output: value: 7
}

func ExamplePeekErr() {
	mybad.PeekErr(mybad.From(0, errors.New("oops")), func(err error) {
		fmt.Println("error:", err)
	})
	// Output: error: oops
}

func ExampleOrElse() {
	r := mybad.OrElse(mybad.From(0, errors.New("oops")), func(err error) (int, error) {
		return 99, nil
	})
	fmt.Println(r.Must())
	// Output: 99
}

func ExampleMatch() {
	result := mybad.Match(mybad.Ok(10),
		func(n int) string { return fmt.Sprintf("ok: %d", n) },
		func(err error) string { return "err" },
	)
	fmt.Println(result)
	// Output: ok: 10
}
