package x_test

import (
	"fmt"
	"math/big"
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
		test := test
		t.Run(fmt.Sprintf("%d/%d", test.n, test.m), func(t *testing.T) {
			t.Parallel()

			l, q := x.RatPrecision(big.NewRat(test.n, test.m))
			assert.Equal(t, test.k, l)
			assert.Equal(t, test.r, q)
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
		test := test
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
	for n := 0; n < b.N; n++ {
		x.RatPrecision(r)
	}
}
