package mybad_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nlozgachev/mybad/v2"
)

var (
	benchErr = errors.New("benchmark error")
	benchOk  = mybad.Ok(42)
	benchBad = mybad.Err[int](benchErr)
	benchCtx = context.Background()
)

// Instantiation

func BenchmarkOk(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := mybad.Ok(i)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkFrom(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := mybad.From(i, nil)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkErr(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := mybad.Err[int](benchErr)
		if !r.IsErr() {
			b.Fatal()
		}
	}
}

// Methods

func BenchmarkResult_IsOk(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if !benchOk.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkResult_IsErr(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if !benchBad.IsErr() {
			b.Fatal()
		}
	}
}

func BenchmarkResult_Unpack(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		val, err := benchOk.Unpack()
		if val != 42 || err != nil {
			b.Fatal()
		}
	}
}

func BenchmarkResult_ValueOr(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if val := benchOk.ValueOr(0); val != 42 {
			b.Fatal()
		}
	}
}

func BenchmarkResult_ValueOrElse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		val := benchOk.ValueOrElse(func(err error) int { return 0 })
		if val != 42 {
			b.Fatal()
		}
	}
}

func BenchmarkResult_WrapErr(b *testing.B) {
	b.ReportAllocs()
	fn := func(err error) error { return err }
	for i := 0; i < b.N; i++ {
		r := benchOk.WrapErr(fn)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkResult_Peek(b *testing.B) {
	b.ReportAllocs()
	fn := func(n int) {}
	for i := 0; i < b.N; i++ {
		r := benchOk.Peek(fn)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkResult_PeekErr(b *testing.B) {
	b.ReportAllocs()
	fn := func(err error) {}
	for i := 0; i < b.N; i++ {
		r := benchOk.PeekErr(fn)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkResult_RecoverTry(b *testing.B) {
	b.ReportAllocs()
	fn := func(err error) (int, error) { return 99, nil }
	for i := 0; i < b.N; i++ {
		r := benchOk.RecoverTry(fn)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkResult_Ensure(b *testing.B) {
	b.ReportAllocs()
	fn := func(n int) error { return nil }
	for i := 0; i < b.N; i++ {
		r := benchOk.Ensure(fn)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkResult_Recover(b *testing.B) {
	b.ReportAllocs()
	fn := func(err error) int { return 99 }
	for i := 0; i < b.N; i++ {
		r := benchOk.Recover(fn)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkResult_Guard(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := benchOk.Guard(benchCtx)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkResult_Or(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := benchOk.Or(benchBad)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

// Chaining Operators

func BenchmarkTry(b *testing.B) {
	b.ReportAllocs()
	fn := func(n int) (int, error) { return n + 1, nil }
	for i := 0; i < b.N; i++ {
		r := mybad.Try(benchOk, fn)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkInto(b *testing.B) {
	b.ReportAllocs()
	fn := func(n int) int { return n + 1 }
	for i := 0; i < b.N; i++ {
		r := mybad.Into(benchOk, fn)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkAndThen(b *testing.B) {
	b.ReportAllocs()
	fn := func(n int) mybad.Result[int] { return mybad.Ok(n + 1) }
	for i := 0; i < b.N; i++ {
		r := mybad.AndThen(benchOk, fn)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkMatch(b *testing.B) {
	b.ReportAllocs()
	onOk := func(n int) int { return n }
	onErr := func(err error) int { return 0 }
	for i := 0; i < b.N; i++ {
		if val := mybad.Match(benchOk, onOk, onErr); val != 42 {
			b.Fatal()
		}
	}
}

// Accumulators & Collections

func BenchmarkCheck(b *testing.B) {
	b.ReportAllocs()
	rule1 := func(n int) error { return nil }
	rule2 := func(n int) error { return nil }
	for i := 0; i < b.N; i++ {
		r := mybad.Check(42, rule1, rule2)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkOr_variadic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := mybad.Or(benchBad, benchOk)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

// BenchmarkAllSeq accumulates values dynamically using AllSeq.
// NOTE: The minimal allocations (3 allocs, ~56B/op) observed in this benchmark
// are inherent to dynamic slice growth (append) inside the accumulator pattern.
// These allocations are expected and represent the standard baseline.
func BenchmarkAllSeq(b *testing.B) {
	b.ReportAllocs()
	seq := func(yield func(int, error) bool) {
		yield(1, nil)
		yield(2, nil)
		yield(3, nil)
	}
	for i := 0; i < b.N; i++ {
		r := mybad.AllSeq(seq)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

// BenchmarkPartitionSeq splits yielded items dynamically using PartitionSeq.
// NOTE: The minimal allocations (3 allocs, ~56B/op) observed in this benchmark
// are inherent to dynamic slice growth (append) inside the accumulator pattern.
// These allocations are expected and represent the standard baseline.
func BenchmarkPartitionSeq(b *testing.B) {
	b.ReportAllocs()
	seq := func(yield func(int, error) bool) {
		yield(1, nil)
		yield(2, nil)
		yield(3, nil)
	}
	for i := 0; i < b.N; i++ {
		vals, errs := mybad.PartitionSeq(seq)
		if len(vals) != 3 || len(errs) != 0 {
			b.Fatal()
		}
	}
}

// New API enhancement benchmarks

func BenchmarkFromBool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := mybad.FromBool(i, true, benchErr)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkResult_ErrIs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if benchOk.ErrIs(benchErr) {
			b.Fatal()
		}
	}
}

func BenchmarkResult_ErrAs(b *testing.B) {
	b.ReportAllocs()
	var target error
	for i := 0; i < b.N; i++ {
		if benchOk.ErrAs(&target) {
			b.Fatal()
		}
	}
}

func BenchmarkResult_RecoverIs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := benchOk.RecoverIs(benchErr, 99)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}

func BenchmarkFromFunc_Call(b *testing.B) {
	b.ReportAllocs()
	adapted := mybad.FromFunc(func(n int) (int, error) { return n + 1, nil })
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := adapted(i)
		if !r.IsOk() {
			b.Fatal()
		}
	}
}
