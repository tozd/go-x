package x_test

import (
	"context"
	"testing"

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
		v, err := v.LoadContext(context.Background())
		if assert.NoError(t, err) {
			assert.Equal(t, 1, v)
		}
		return nil
	})
	g.Go(func() error {
		v, err := v.LoadContext(context.Background())
		if assert.NoError(t, err) {
			assert.Equal(t, 1, v)
		}
		return nil
	})
	g.Go(func() error {
		v, err := v.LoadContext(context.Background())
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

func TestSyncVarContextCancel(t *testing.T) {
	t.Parallel()

	v := x.NewSyncVar[int]()

	ctx, cancel := context.WithCancel(context.Background())

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
