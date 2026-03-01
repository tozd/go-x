package x_test

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/tozd/go/x"
)

const (
	tickerInterval = 50 * time.Millisecond
)

func TestCountingReaderEOF(t *testing.T) {
	t.Parallel()

	cr := x.NewCountingReader(strings.NewReader("hello"))

	data, err := io.ReadAll(cr)
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), data)
	assert.Equal(t, int64(5), cr.Count())
}

func TestCounter(t *testing.T) {
	t.Parallel()

	c := x.NewCounter(0)
	assert.Equal(t, int64(0), c.Count())

	c.Increment()
	assert.Equal(t, int64(1), c.Count())

	c.Add(5)
	assert.Equal(t, int64(6), c.Count())
}

func TestTicker(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	r, w := io.Pipe()
	defer r.Close() //nolint:errcheck
	defer w.Close() //nolint:errcheck

	countingReader := x.NewCountingReader(r)

	go func() {
		_, _ = io.ReadAll(countingReader)
	}()

	ticker := x.NewTicker(ctx, countingReader, x.NewCounter(10), tickerInterval)
	require.NotNil(t, ticker)
	defer ticker.Stop()

	l := sync.Mutex{}
	progress := []x.Progress{}

	go func() {
		for p := range ticker.C {
			func() {
				l.Lock()
				defer l.Unlock()
				progress = append(progress, p)
			}()
		}
	}()

	time.Sleep(2 * tickerInterval)

	var p x.Progress
	func() {
		l.Lock()
		defer l.Unlock()
		require.NotEmpty(t, progress)
		p = progress[len(progress)-1]
	}()

	assert.Equal(t, int64(10), p.Size)
	assert.Equal(t, int64(0), p.Count)
	assert.Equal(t, 0.0, p.Percent()) //nolint:testifylint

	n, err := w.Write([]byte("abcd"))
	assert.Equal(t, 4, n)
	require.NoError(t, err)

	time.Sleep(2 * tickerInterval)

	func() {
		l.Lock()
		defer l.Unlock()
		require.NotEmpty(t, progress)
		p = progress[len(progress)-1]
	}()

	assert.Equal(t, int64(10), p.Size)
	assert.Equal(t, int64(4), p.Count)
	assert.Equal(t, 40.0, p.Percent()) //nolint:testifylint
	assert.Positive(t, p.Remaining())
	assert.False(t, p.Estimated().IsZero())

	cancel()

	// We give time for cancel to propagate.
	time.Sleep(2 * tickerInterval)

	var progressLen int
	func() {
		l.Lock()
		defer l.Unlock()
		progressLen = len(progress)
	}()

	// After this there should be no new progress added.
	time.Sleep(2 * tickerInterval)

	func() {
		l.Lock()
		defer l.Unlock()
		assert.Len(t, progress, progressLen)
	}()

	// Channel should be closed.
	select {
	case _, ok := <-ticker.C:
		if ok {
			require.Fail(t, "progress where there should be none")
		}
	default:
	}
}
