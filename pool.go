package sync

import (
	"sync"
)

// Pool is a generic implementation of the standard library's sync.Pool.
type Pool[E any] struct {
	rwmutex *sync.RWMutex
	pool    *sync.Pool
}

// NewPool returns a new Pool with generator new.
func NewPool[E any](new func() E) *Pool[E] {
	pool := &Pool[E]{
		rwmutex: &sync.RWMutex{},
		pool:    &sync.Pool{},
	}
	if new != nil {
		pool.pool.New = func() any { return new() }
	}
	return pool
}

// Get selects an arbitrary item from the Pool, removes it from the Pool, and
// returns it to the caller. Get may choose to ignore the pool and treat it as
// empty. Callers should not assume any relation between values passed to Put
// and the values returned by Get.
//
// If Get would otherwise return nil and the generator is non-nil, Get returns
// the result of calling the generator.
func (pool *Pool[E]) Get() E {
	pool.rwmutex.RLock()
	defer pool.rwmutex.RUnlock()
	return pool.pool.Get().(E)
}

// Put adds x to the pool.
func (pool *Pool[E]) Put(x E) {
	pool.pool.Put(x)
}

// Update updates the pool's generator to new.
func (pool *Pool[E]) Update(new func() E) {
	pool.rwmutex.Lock()
	defer pool.rwmutex.Unlock()
	if new == nil {
		pool.pool.New = nil
	} else {
		pool.pool.New = func() any { return new() }
	}
}
