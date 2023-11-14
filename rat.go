package x

import (
	"math/big"
)

var (
	zeroInt = big.NewInt(0)  //nolint:gochecknoglobals
	oneInt  = big.NewInt(1)  //nolint:gochecknoglobals
	twoInt  = big.NewInt(2)  //nolint:gomnd,gochecknoglobals
	fiveInt = big.NewInt(5)  //nolint:gomnd,gochecknoglobals
	tenInt  = big.NewInt(10) //nolint:gomnd,gochecknoglobals
)

// RatPrecision computes for rat the number of non-repeating digits on the right
// of the decimal point and the number of repeating digits which cyclicly follow.
//
// It can be used with Rat.FloatString to convert a number to full precision
// representation, when there are no repeating digits.
//
// This is similar to Rat.FloatPrec but returns also the number of repeating digits
// following the non-repeating digits. Rat.FloatPrec is also much faster.
func RatPrecision(rat *big.Rat) (int, int) {
	// Go assures that rat is normalized.
	m := new(big.Int).Set(rat.Denom())

	q := new(big.Int)
	r := new(big.Int)

	k := 0
	for {
		q.QuoRem(m, twoInt, r)
		if r.Cmp(zeroInt) == 0 {
			m, q = q, m
			k++
		} else {
			break
		}
	}

	l := 0
	for {
		q.QuoRem(m, fiveInt, r)
		if r.Cmp(zeroInt) == 0 {
			m, q = q, m
			l++
		} else {
			break
		}
	}

	j := 0
	if m.Cmp(oneInt) != 0 {
		q.SetInt64(1)
		for {
			q.Mul(q, tenInt)
			q.Mod(q, m)
			j++
			if q.Cmp(oneInt) == 0 {
				break
			}
		}
	}

	if k > l {
		return k, j
	}
	return l, j
}
