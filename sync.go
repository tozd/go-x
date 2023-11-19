package x

import (
	"context"
	"sync"

	"gitlab.com/tozd/go/errors"
)

var ErrSyncVarAlreadyStored = errors.Base("already stored")

// SyncVar allows multiple goroutines to wait for a value
// to be available while one other goroutine stored the value.
//
// It is useful if you do not know in advance which goroutine
// will be the one (and only one) which stores the value while
// all goroutines need the value.
//
// The zero value for a SyncVar is not usable. Use NewSyncVar.
type SyncVar[T any] struct {
	lock *sync.RWMutex
	cond *sync.Cond
	v    *T
}

// Load returns the value stored in the SyncVar. It blocks
// until the value is stored.
func (w *SyncVar[T]) Load() T { //nolint:ireturn
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	for w.v == nil {
		w.cond.Wait()
	}
	return *w.v
}

// LoadContext is similar to Load, but it stops waiting if ctx gets cancelled,
// returning an error in that case.
func (w *SyncVar[T]) LoadContext(ctx context.Context) (T, errors.E) { //nolint:ireturn
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	// This is based on example for context.AfterFunc from the context package.
	// See comments there for explanation how it works and why.
	stop := context.AfterFunc(ctx, func() {
		w.cond.L.Lock()
		defer w.cond.L.Unlock()
		w.cond.Broadcast()
	})
	defer stop()

	for w.v == nil {
		w.cond.Wait()
		if ctx.Err() != nil {
			return *new(T), errors.WithStack(ctx.Err())
		}
	}

	return *w.v, nil
}

// Store stores the value in the SyncVar and unblocks
// any prior calls to Load.
// It can be called only once.
func (w *SyncVar[T]) Store(v T) errors.E {
	w.lock.Lock()
	defer w.lock.Unlock()
	if w.v != nil {
		return errors.WithStack(ErrSyncVarAlreadyStored)
	}
	w.v = &v
	w.cond.Broadcast()
	return nil
}

// NewSyncVar creates a new SyncVar.
func NewSyncVar[T any]() *SyncVar[T] {
	l := &sync.RWMutex{}
	return &SyncVar[T]{
		lock: l,
		cond: sync.NewCond(l.RLocker()),
		v:    nil,
	}
}
