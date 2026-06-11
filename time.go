package x

import (
	"math"
	"time"
)

// TimeToFloat64 returns the float64 representation of a time.Time,
// as seconds since the Unix epoch.
func TimeToFloat64(t time.Time) float64 {
	return float64(t.Unix()) + float64(t.Nanosecond())/1e9
}

// TimeFromFloat64 converts float64 seconds since the Unix epoch
// to time.Time.
func TimeFromFloat64(t float64) time.Time {
	sec := math.Trunc(t)
	return time.Unix(int64(sec), int64((t-sec)*1e9)) //nolint:mnd
}
