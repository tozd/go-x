package x_test

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/tozd/go/x"
)

func TestRatPrecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		n int64
		m int64
		k int
		r int
	}{
		{1, 3, 0, 1},
		{1, 6, 1, 1},
		{1, 7, 0, 6},
		{1, 9, 0, 1},
		{1, 28, 2, 6},
		{1, 67, 0, 33},
		{1, 81, 0, 9},
		{1, 96, 5, 1},
		{8, 13, 0, 6},
		{2, 14, 0, 6},
		{3, 30, 1, 0},
		{2, 3, 0, 1},
		{9, 11, 0, 2},
		{7, 12, 2, 1},
		{22, 7, 0, 6},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%d/%d", test.n, test.m), func(t *testing.T) {
			t.Parallel()

			l, q := x.RatPrecision(big.NewRat(test.n, test.m))
			assert.Equal(t, test.k, l)
			assert.Equal(t, test.r, q)
		})
	}
}

// Tests taken from tests for Rat.FloatPrec.
//
//nolint:godot
func TestRatPrecisionMore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		f    string
		prec int
		rep  int
	}{
		// examples from the issue #50489
		{"10/100", 1, 0},
		{"3/100", 2, 0},
		{"10", 0, 0},

		// more examples
		{"zero", 0, 0}, // test uninitialized zero value for Rat
		{"0", 0, 0},    // 0
		{"1", 0, 0},    // 1
		{"1/2", 1, 0},  // 0.5
		{"1/3", 0, 1},  // 0.(3)
		{"1/4", 2, 0},  // 0.25
		{"1/5", 1, 0},  // 0.2
		{"1/6", 1, 1},  // 0.1(6)
		{"1/7", 0, 6},  // 0.(142857)
		{"1/8", 3, 0},  // 0.125
		{"1/9", 0, 1},  // 0.(1)
		{"1/10", 1, 0}, // 0.1
		{"1/11", 0, 2}, // 0.(09)
		{"1/12", 2, 1}, // 0.08(3)
		{"1/13", 0, 6}, // 0.(076923)
		{"1/14", 1, 6}, // 0.0(714285)
		{"1/15", 1, 1}, // 0.0(6)
		{"1/16", 4, 0}, // 0.0625

		{"10/2", 0, 0},        // 5
		{"10/3", 0, 1},        // 3.(3)
		{"10/6", 0, 1},        // 1.(6)
		{"1/10000000", 7, 0},  // 0.0000001
		{"1/3125", 5, 0},      // "0.00032"
		{"1/1024", 10, 0},     // 0.0009765625
		{"1/304000", 7, 18},   // 0.0000032(894736842105263157)
		{"1/48828125", 11, 0}, // 0.00000002048
	}

	for _, tt := range tests {
		t.Run(tt.f, func(t *testing.T) {
			t.Parallel()

			var f big.Rat

			// check uninitialized zero value
			if tt.f != "zero" {
				_, ok := f.SetString(tt.f) //nolint:gosec
				require.True(t, ok, "invalid test case")
			}

			// results for f and -f must be the same
			for range 2 {
				prec, rep := x.RatPrecision(&f)
				assert.Equal(t, tt.prec, prec)
				assert.Equal(t, tt.rep, rep)

				// proceed with -f but don't add a "-" before a "0"
				if f.Sign() > 0 {
					f.Neg(&f)
				}
			}
		})
	}
}

func TestRatPrecisionString(t *testing.T) {
	t.Parallel()

	tests := []string{
		"123.34",
		"-2342343.2321234442",
		"235994.099923999900001",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			n, ok := new(big.Rat).SetString(test) //nolint:gosec
			require.True(t, ok)
			l, q := x.RatPrecision(n)
			assert.Equal(t, 0, q)
			assert.Equal(t, test, n.FloatString(l))
		})
	}
}

func BenchmarkRatPrecision(b *testing.B) {
	r := big.NewRat(1, 67)
	b.ResetTimer()
	for range b.N {
		x.RatPrecision(r)
	}
}

// Benchmark taken from tests for Rat.FloatPrec.
//
//nolint:godot
func BenchmarkFloatPrecExact(b *testing.B) {
	for _, n := range []int{1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6} {
		// d := 5^n
		d := big.NewInt(5)
		p := big.NewInt(int64(n))
		d.Exp(d, p, nil)

		// r := 1/d
		var r big.Rat
		r.SetFrac(big.NewInt(1), d)

		b.Run(strconv.Itoa(n), func(b *testing.B) {
			for range b.N {
				prec, rep := x.RatPrecision(&r)
				if prec != n || rep != 0 {
					b.Fatalf("got %d, %v; want %d, %v", prec, rep, n, 0)
				}
			}
		})
	}
}

// Benchmark taken from tests for Rat.FloatPrec.
//
//nolint:godot
func BenchmarkFloatPrecInexact(b *testing.B) {
	for _, n := range []int{1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6} {
		// d := 5^n + 1
		d := big.NewInt(5)
		p := big.NewInt(int64(n))
		d.Exp(d, p, nil)
		d.Add(d, big.NewInt(1))

		// r := 1/d
		var r big.Rat
		r.SetFrac(big.NewInt(1), d)

		b.Run(strconv.Itoa(n), func(b *testing.B) {
			for range b.N {
				_, rep := x.RatPrecision(&r)
				if rep == 0 {
					b.Fatalf("got unexpected zero rep")
				}
			}
		})
	}
}
