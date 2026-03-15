package x

import (
	"context"
	"sync"

	"gitlab.com/tozd/go/errors"
)

// RecreatableChannel is a channel that can be recreated.
// When recreated, the previous channel is closed and a new one is created.
//
// The zero value for a RecreatableChannel is usable but without the first
// channel. Use Recreate to create the first channel. Get blocks until
// the first channel is created.
//
// A RecreatableChannel must not be copied after first use.
type RecreatableChannel[T any] struct {
	mu   sync.Mutex
	cond *sync.Cond
	ch   chan T
}

func (c *RecreatableChannel[T]) init() {
	if c.cond == nil {
		c.cond = sync.NewCond(&c.mu)
	}
}

// Get returns the current channel.
//
// It blocks until the first channel is created with Recreate.
// If ctx is cancelled before that, it returns an error.
func (c *RecreatableChannel[T]) Get(ctx context.Context) (<-chan T, errors.E) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.init()

	// This is based on example for context.AfterFunc from the context package.
	// See comments there for explanation how it works and why.
	stop := context.AfterFunc(ctx, func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.cond.Broadcast()
	})
	defer stop()

	for c.ch == nil {
		c.cond.Wait()
		if ctx.Err() != nil {
			return nil, errors.WithStack(ctx.Err())
		}
	}

	return c.ch, nil
}

// Recreate closes the existing channel and creates a new one.
// It returns the new channel for writing.
//
// Use it also to create the first channel.
func (c *RecreatableChannel[T]) Recreate(size int) chan<- T {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.init()

	first := true
	if c.ch != nil {
		first = false
		close(c.ch)
	}

	c.ch = make(chan T, size)

	if first {
		c.cond.Broadcast()
	}

	return c.ch
}
