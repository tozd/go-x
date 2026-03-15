package x_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/tozd/go/x"
)

func TestRecreatableChannelGetBlocks(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]

	type result struct {
		ch  <-chan int
		err error
	}
	got := make(chan result, 1)
	go func() {
		ch, err := c.Get(t.Context())
		got <- result{ch, err}
	}()

	// Get should be blocking; nothing in got yet.
	select {
	case <-got:
		t.Fatal("Get returned before Recreate was called")
	default:
	}

	ch := c.Recreate(0)
	assert.NotNil(t, ch)

	// Now Get should unblock.
	r := <-got
	require.NoError(t, r.err)
	assert.NotNil(t, r.ch)
}

func TestRecreatableChannelGetContextCanceled(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]

	ctx, cancel := context.WithCancel(t.Context())

	type result struct {
		ch  <-chan int
		err error
	}
	got := make(chan result, 1)
	go func() {
		ch, err := c.Get(ctx)
		got <- result{ch, err}
	}()

	// Get should be blocking; nothing in got yet.
	select {
	case <-got:
		t.Fatal("Get returned before cancel was called")
	default:
	}

	cancel()

	r := <-got
	assert.Nil(t, r.ch)
	require.Error(t, r.err)
	assert.ErrorIs(t, r.err, context.Canceled)
}

func TestRecreatableChannelFirstRecreate(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]

	ch := c.Recreate(0)
	assert.NotNil(t, ch)
	got, err := c.Get(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, got)
}

func TestRecreatableChannelSendReceive(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]

	writeCh := c.Recreate(1)
	writeCh <- 42

	readCh, err := c.Get(t.Context())
	require.NoError(t, err)
	got, ok := <-readCh
	assert.True(t, ok)
	assert.Equal(t, 42, got)
}

func TestRecreatableChannelRecreateClosesOld(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]

	_ = c.Recreate(0)
	oldCh, err := c.Get(t.Context())
	require.NoError(t, err)

	// Recreate closes the old channel.
	_ = c.Recreate(0)

	// Reads from the old channel should return zero value and false (closed).
	val, ok := <-oldCh
	assert.False(t, ok)
	assert.Equal(t, 0, val)
}

func TestRecreatableChannelGetAfterRecreate(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]

	_ = c.Recreate(0)
	ch2 := c.Recreate(1)

	// Get returns the latest channel: sending on ch2 is receivable via Get.
	ch2 <- 99
	readCh, err := c.Get(t.Context())
	require.NoError(t, err)
	got, ok := <-readCh
	assert.True(t, ok)
	assert.Equal(t, 99, got)
}

func TestRecreatableChannelBuffered(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]

	writeCh := c.Recreate(3)

	// Buffered channel allows sending without a receiver ready.
	writeCh <- 1
	writeCh <- 2
	writeCh <- 3

	readCh, err := c.Get(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 1, <-readCh)
	assert.Equal(t, 2, <-readCh)
	assert.Equal(t, 3, <-readCh)
}

func TestRecreatableChannelConcurrentGetRecreate(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]
	_ = c.Recreate(0)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			// Concurrent Gets should not race with Recreates.
			_, _ = c.Get(t.Context())
		}()
	}

	for range 5 {
		_ = c.Recreate(0)
	}

	wg.Wait()
}
