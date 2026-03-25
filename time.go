package x

import "time"

// TimeToFloat64 returns the float64 representation of a time.Time.
func TimeToFloat64(t time.Time) float64 {
	return float64(t.Unix()) + float64(t.Nanosecond())/1e9
}
