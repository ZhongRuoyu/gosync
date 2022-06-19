// Package filter provides a thread-safe, modifiable and extensible filter for
// channels.
package filter

import (
	"sync/atomic"
)

// Filter is a thread-safe modifiable filter for channels.
type Filter[E any] struct {
	in  <-chan E
	out chan E
	f   atomic.Value
}

// NewFilter returns a new Limiter that filters the channel in with f.
// Elements are examined with function f: if f returns true, then the element
// can pass through; otherwise, it will be filtered out and discarded.
func NewFilter[E any](in <-chan E, f func(E) bool) *Filter[E] {
	out := make(chan E, cap(in))
	filter := &Filter[E]{
		in:  in,
		out: out,
		f:   atomic.Value{},
	}
	filter.f.Store(f)
	filter.start()
	return filter
}

// New returns a new Limiter that allows all elements to pass through. It is
// particularly useful when chained with And, Or, Inverse.
func New[E any](in <-chan E) *Filter[E] {
	return NewFilter(
		in,
		func(E) bool { return true },
	)
}

// And modifies the filter in-place. f is a filtering condition, signaling
// whether an element is allowed to pass through or not. After the
// modification, only elements that satisfy both the previous conditions and
// the condition specified by f can pass through.
//
// The method returns the filter itself, making it easy to chain up multiple
// conditions.
func (filter *Filter[E]) And(f func(E) bool) *Filter[E] {
	oldF := filter.f.Load().(func(E) bool)
	filter.f.Store(func(element E) bool {
		return oldF(element) && f(element)
	})
	return filter
}

// Or modifies the filter in-place. f is a filtering condition, signaling
// whether an element is allowed to pass through or not. After the
// modification, only elements that satisfy either the previous conditions or
// the condition specified by f can pass through.
//
// The method returns the filter itself, making it easy to chain up multiple
// conditions.
func (filter *Filter[E]) Or(f func(E) bool) *Filter[E] {
	oldF := filter.f.Load().(func(E) bool)
	filter.f.Store(func(element E) bool {
		return oldF(element) || f(element)
	})
	return filter
}

// Or modifies the filter in-place. f is a filtering condition, signaling
// whether an element is allowed to pass through or not. After the
// modification, only elements that do not satisfy the previous conditions can
// pass through.
//
// The method returns the filter itself, making it easy to chain up multiple
// conditions.
func (filter *Filter[E]) Inverse() *Filter[E] {
	oldF := filter.f.Load().(func(E) bool)
	filter.f.Store(func(element E) bool {
		return !oldF(element)
	})
	return filter
}

// Out returns the filter's output channel.
func (filter *Filter[E]) Out() <-chan E {
	return filter.out
}

// Update updates the conditions of the filter. f is the new filtering
// condition, signaling whether an element is allowed to pass through or not.
func (filter *Filter[E]) Update(f func(E) bool) *Filter[E] {
	filter.f.Store(f)
	return filter
}

func (filter *Filter[E]) start() {
	go func() {
		for element := range filter.in {
			if filter.f.Load().(func(E) bool)(element) {
				filter.out <- element
			}
		}
		close(filter.out)
	}()
}
