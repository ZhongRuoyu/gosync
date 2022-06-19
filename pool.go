package sync

import (
	"sync"
)

type Pool[E any] struct {
	rwmutex *sync.RWMutex
	pool    *sync.Pool
}

func NewPool[E any](new func() E) *Pool[E] {
	pool := &Pool[E]{
		rwmutex: &sync.RWMutex{},
		pool: &sync.Pool{
			New: func() any { return new() },
		},
	}
	return pool
}

func (pool *Pool[E]) Get() E {
	pool.rwmutex.RLock()
	defer pool.rwmutex.RUnlock()
	return pool.pool.Get().(E)
}

func (pool *Pool[E]) Put(x E) {
	pool.pool.Put(x)
}

func (pool *Pool[E]) Update(new func() E) {
	pool.rwmutex.Lock()
	defer pool.rwmutex.Unlock()
	pool.pool.New = func() any { return new() }
}
