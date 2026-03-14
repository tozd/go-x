package x

import (
	"sync"
)

// RecreatableChannel is a channel that can be recreated.
// When recreated, the previous channel is closed and a new one is created.
//
// The zero value for a RecreatableChannel is usable but without the first
// channel. Use Recreate to create the first channel.
//
// A RecreatableChannel must not be copied after first use.
type RecreatableChannel[T any] struct {
	lock sync.RWMutex
	ch   chan T
}

// Get returns the current channel.
func (c *RecreatableChannel[T]) Get() <-chan T {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.ch
}

// Recreate closes the existing channel and creates a new one.
// It returns the new channel for writing.
//
// Use it also to create the first channel.
func (c *RecreatableChannel[T]) Recreate(size int) chan<- T {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.ch != nil {
		close(c.ch)
	}
	c.ch = make(chan T, size)

	return c.ch
}
