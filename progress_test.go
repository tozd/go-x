package x_test

import (
	"context"
	"io"
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

func TestTicker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	countingReader := x.NewCountingReader(r)

	go func() {
		_, _ = io.ReadAll(countingReader)
	}()

	ticker := x.NewTicker(ctx, countingReader, 10, tickerInterval)
	require.NotNil(t, ticker)
	defer ticker.Stop()

	l := sync.RWMutex{}
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
	assert.Equal(t, 0.0, p.Percent())

	n, err := w.Write([]byte("abcd"))
	assert.Equal(t, 4, n)
	assert.NoError(t, err)

	time.Sleep(2 * tickerInterval)

	func() {
		l.Lock()
		defer l.Unlock()
		require.NotEmpty(t, progress)
		p = progress[len(progress)-1]
	}()

	assert.Equal(t, int64(10), p.Size)
	assert.Equal(t, int64(4), p.Count)
	assert.Equal(t, 40.0, p.Percent())

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
		assert.Equal(t, progressLen, len(progress))
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
