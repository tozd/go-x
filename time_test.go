package x_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/tozd/go/x"
)

func TestTimeToFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value    time.Time
		expected float64
	}{
		{time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC), 0},
		{time.Date(1970, time.January, 1, 0, 0, 1, 500000000, time.UTC), 1.5},
		{time.Date(1969, time.December, 31, 23, 59, 59, 500000000, time.UTC), -0.5},
		{time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), 946684800},
	}
	for _, tt := range tests {
		assert.InDelta(t, tt.expected, x.TimeToFloat64(tt.value), 0, "value %v", tt.value)
	}
}

func TestTimeFromFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value    float64
		expected time.Time
	}{
		{0, time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{1.5, time.Date(1970, time.January, 1, 0, 0, 1, 500000000, time.UTC)},
		{-0.5, time.Date(1969, time.December, 31, 23, 59, 59, 500000000, time.UTC)},
		{-1.25, time.Date(1969, time.December, 31, 23, 59, 58, 750000000, time.UTC)},
		{946684800, time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		got := x.TimeFromFloat64(tt.value)
		assert.True(t, tt.expected.Equal(got), "value %v: got %v", tt.value, got.UTC())
	}
}

func TestTimeFromFloat64Large(t *testing.T) {
	t.Parallel()

	tests := []time.Time{
		time.Date(2262, time.April, 12, 0, 0, 0, 0, time.UTC),
		time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1_000_000, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1_000_000_000, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(-10000, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(-1_000_000_000, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	for _, tt := range tests {
		got := x.TimeFromFloat64(x.TimeToFloat64(tt))
		assert.True(t, tt.Equal(got), "value %v: got %v", tt, got.UTC())
		assert.Equal(t, 0, got.Nanosecond(), "value %v", tt)
	}
}
