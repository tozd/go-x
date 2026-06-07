package x

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	"gitlab.com/tozd/go/errors"
)

// Counter is an int64 value which implements counter interface.
type Counter int64

// Increment increases counter by 1.
func (c *Counter) Increment() {
	atomic.AddInt64((*int64)(c), 1)
}

// Add increases counter by n.
func (c *Counter) Add(n int64) {
	atomic.AddInt64((*int64)(c), n)
}

// Count implements counter interface for Counter.
func (c *Counter) Count() int64 {
	return atomic.LoadInt64((*int64)(c))
}

// NewCounter returns a new Counter which holds the value c.
func NewCounter(c int64) *Counter {
	return (*Counter)(&c)
}

// CountingReader is an io.Reader proxy which counts the number of bytes
// it read and passed on.
type CountingReader struct {
	Reader io.Reader
	count  int64
}

// NewCountingReader returns a new CountingReader which reads
// from the reader and counts the bytes.
func NewCountingReader(reader io.Reader) *CountingReader {
	return &CountingReader{
		Reader: reader,
		count:  0,
	}
}

// Read implements io.Reader interface for CountingReader.
func (c *CountingReader) Read(p []byte) (int, error) {
	n, err := c.Reader.Read(p)
	atomic.AddInt64(&c.count, int64(n))
	if err == io.EOF {
		// See: https://github.com/golang/go/issues/39155
		return n, io.EOF
	}
	return n, errors.WithStack(err)
}

// Count implements counter interface for CountingReader.
//
// It returns the number of bytes read until now.
func (c *CountingReader) Count() int64 {
	return atomic.LoadInt64(&c.count)
}

type counter interface {
	Count() int64
}

// Progress describes current progress as reported by the counter.
type Progress struct {
	Count     int64
	Size      int64
	Started   time.Time
	Current   time.Time
	Elapsed   time.Duration
	remaining time.Duration
	estimated time.Time
}

// Percent returns the percentage of the progress.
func (p Progress) Percent() float64 {
	return float64(p.Count) / float64(p.Size) * 100.0 //nolint:mnd
}

// Remaining returns the remaining time to completion.
//
// It returns a negative duration when the remaining time cannot be estimated yet, i.e. before any
// progress has been made or right after size has changed. Estimated then returns the zero time.Time.
func (p Progress) Remaining() time.Duration {
	return p.remaining
}

// Estimated returns the estimated time of completion.
//
// It returns the zero time.Time when the time of completion cannot be estimated yet, i.e. before any
// progress has been made or right after size has changed. Remaining then returns a negative duration.
func (p Progress) Estimated() time.Time {
	return p.estimated
}

// Ticker at regular interval reports the progress.
type Ticker struct {
	C    <-chan Progress
	stop func()
}

// Stop stops the ticker and frees resources.
func (t *Ticker) Stop() {
	t.stop()
}

// notEstimated is reported as the remaining time while it cannot be estimated yet. It is negative so it
// is clearly an invalid value.
const notEstimated = -1 * time.Second

// NewTicker creates a new Ticker which at regular interval reports the
// progress as reported by counters count and size.
//
// The remaining and estimated completion times are computed from the rate of progress observed so far.
// Whenever size changes, that estimate is reset and recomputed only from the progress made after the
// change, so that a changing total (e.g. when additional work is discovered) does not skew the estimate.
// The reported Started and Elapsed, and thus Percent, always cover the whole run.
//
// counter interface is defined as:
//
//	type counter interface {
//		Count() int64
//	}
func NewTicker(ctx context.Context, count, size counter, interval time.Duration) *Ticker {
	ctx, cancel := context.WithCancel(ctx)
	started := time.Now()
	output := make(chan Progress)
	ticker := time.NewTicker(interval)
	go func() {
		defer cancel()
		defer close(output)
		defer ticker.Stop()
		// The remaining and estimated times are extrapolated from the progress made since baseStarted,
		// at which point baseCount had already been done. Both are reset to the current moment and count
		// whenever size changes so that the estimate reflects only the rate of the work after the change.
		baseStarted := started
		var baseCount int64
		prevSize := size.Count()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				c := count.Count()
				s := size.Count()
				if s != prevSize {
					baseStarted = now
					baseCount = c
					prevSize = s
				}
				remaining := notEstimated
				var estimated time.Time
				// We can extrapolate only once there is some progress since the baseline. Right after a size
				// change (or before any progress) there is none, so remaining stays the negative sentinel and
				// estimated the zero time.Time to signal that the estimate is not available yet.
				if c > baseCount {
					baseElapsed := now.Sub(baseStarted)
					remaining = time.Duration(float64(baseElapsed) * float64(s-c) / float64(c-baseCount))
					estimated = now.Add(remaining)
				}
				progress := Progress{
					Count:     c,
					Size:      s,
					Started:   started,
					Current:   now,
					Elapsed:   now.Sub(started),
					remaining: remaining,
					estimated: estimated,
				}
				select {
				case <-ctx.Done():
					return
				case output <- progress:
				}
			}
		}
	}()
	return &Ticker{
		C:    output,
		stop: cancel,
	}
}
