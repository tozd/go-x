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

// Increment increases counter by n.
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

func (p Progress) Percent() float64 {
	return float64(p.Count) / float64(p.Size) * 100.0 //nolint:mnd
}

func (p Progress) Remaining() time.Duration {
	return p.remaining
}

func (p Progress) Estimated() time.Time {
	return p.estimated
}

type Ticker struct {
	C    <-chan Progress
	stop func()
}

// Stop stops the ticker and frees resources.
func (t *Ticker) Stop() {
	t.stop()
}

// NewTicker creates a new Ticker which at regular interval reports the
// progress as reported by counters count and size.
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
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				c := count.Count()
				s := size.Count()
				elapsed := now.Sub(started)
				ratio := float64(c) / float64(s)
				total := time.Duration(float64(elapsed) / ratio)
				estimated := started.Add(total)
				progress := Progress{
					Count:     c,
					Size:      s,
					Started:   started,
					Current:   now,
					Elapsed:   elapsed,
					remaining: estimated.Sub(now),
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
