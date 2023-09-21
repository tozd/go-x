package x_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, errE)
	err := g.Wait()
	assert.NoError(t, err)
	errE = v.Store(1)
	assert.ErrorIs(t, errE, x.ErrSyncVarAlreadyStored)
}
