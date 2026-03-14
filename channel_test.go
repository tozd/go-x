package x_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/tozd/go/x"
)

func TestRecreatableChannelZeroValue(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]
	assert.Nil(t, c.Get())
}

func TestRecreatableChannelFirstRecreate(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]

	ch := c.Recreate(0)
	assert.NotNil(t, ch)
	assert.NotNil(t, c.Get())
}

func TestRecreatableChannelSendReceive(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]

	writeCh := c.Recreate(1)
	writeCh <- 42

	got, ok := <-c.Get()
	assert.True(t, ok)
	assert.Equal(t, 42, got)
}

func TestRecreatableChannelRecreateClosesOld(t *testing.T) {
	t.Parallel()

	var c x.RecreatableChannel[int]

	_ = c.Recreate(0)
	oldCh := c.Get()

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
	got, ok := <-c.Get()
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

	readCh := c.Get()
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
			_ = c.Get()
		}()
	}

	for range 5 {
		_ = c.Recreate(0)
	}

	wg.Wait()
}
