package x

import (
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
