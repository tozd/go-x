package x_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"gitlab.com/tozd/go/x"
)

func TestSyncVar(t *testing.T) {
	t.Parallel()

	v := x.NewSyncVar[int]()

	g := errgroup.Group{}
	g.Go(func() error {
		assert.Equal(t, 1, v.Load())
		return nil
	})
	g.Go(func() error {
		assert.Equal(t, 1, v.Load())
		return nil
	})
	g.Go(func() error {
		assert.Equal(t, 1, v.Load())
		return nil
	})
	errE := v.Store(1)
	require.NoError(t, errE, "% -+#.1v", errE)
	err := g.Wait()
	require.NoError(t, err)
	errE = v.Store(1)
	assert.ErrorIs(t, errE, x.ErrSyncVarAlreadyStored)
}

func TestSyncVarContext(t *testing.T) {
	t.Parallel()

	v := x.NewSyncVar[int]()

	g := errgroup.Group{}
	g.Go(func() error {
		v, err := v.LoadContext(t.Context())
		if assert.NoError(t, err) {
			assert.Equal(t, 1, v)
		}
		return nil
	})
	g.Go(func() error {
		v, err := v.LoadContext(t.Context())
		if assert.NoError(t, err) {
			assert.Equal(t, 1, v)
		}
		return nil
	})
	g.Go(func() error {
		v, err := v.LoadContext(t.Context())
		if assert.NoError(t, err) {
			assert.Equal(t, 1, v)
		}
		return nil
	})
	errE := v.Store(1)
	require.NoError(t, errE, "% -+#.1v", errE)
	err := g.Wait()
	require.NoError(t, err)
	errE = v.Store(1)
	assert.ErrorIs(t, errE, x.ErrSyncVarAlreadyStored)
}

func TestSyncVarEnsureWait(t *testing.T) {
	t.Parallel()

	v := x.NewSyncVar[int]()
	result := make(chan int, 1)

	go func() {
		result <- v.Load()
	}()

	// Give the goroutine time to start and block in Wait before Store is called.
	time.Sleep(10 * time.Millisecond)

	errE := v.Store(99)
	require.NoError(t, errE, "% -+#.1v", errE)

	assert.Equal(t, 99, <-result)
}

func TestSyncVarLoadAfterStore(t *testing.T) {
	t.Parallel()

	v := x.NewSyncVar[int]()

	errE := v.Store(42)
	require.NoError(t, errE, "% -+#.1v", errE)

	// Load after store should return immediately without blocking.
	val := v.Load()
	assert.Equal(t, 42, val)
}

func TestSyncVarContextCancel(t *testing.T) {
	t.Parallel()

	v := x.NewSyncVar[int]()

	ctx, cancel := context.WithCancel(t.Context())

	g := errgroup.Group{}
	g.Go(func() error {
		v, err := v.LoadContext(ctx)
		if assert.ErrorIs(t, err, context.Canceled) {
			assert.Equal(t, 0, v)
		}
		return nil
	})
	g.Go(func() error {
		v, err := v.LoadContext(ctx)
		if assert.ErrorIs(t, err, context.Canceled) {
			assert.Equal(t, 0, v)
		}
		return nil
	})
	g.Go(func() error {
		v, err := v.LoadContext(ctx)
		if assert.ErrorIs(t, err, context.Canceled) {
			assert.Equal(t, 0, v)
		}
		return nil
	})

	cancel()

	err := g.Wait()
	assert.NoError(t, err)
}
