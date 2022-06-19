// Package rate provides a thread-safe modifiable rate limiter.
package rate

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// Limiter is a rate limiter. It is a thread-safe modifiable version of
// golang.org/x/time/rate.Limiter, providing a similar interface.
type Limiter struct {
	frequency atomic.Value
	limiter   *rate.Limiter
}

// NewLimiter returns a new Limiter that allows events occuring at the given
// maximum frequency.
func NewLimiter(frequency float64) (*Limiter, error) {
	if frequency < 0.0 {
		return nil, errors.New("invalid argument to NewLimiter")
	}
	lim := &Limiter{
		frequency: atomic.Value{},
		limiter:   rate.NewLimiter(rate.Limit(frequency), int(frequency)),
	}
	lim.frequency.Store(frequency)
	return lim, nil
}

// Allow is shorthand for AllowN(time.Now(), 1).
func (lim *Limiter) Allow() bool {
	return lim.limiter.Allow()
}

// AllowN reports whether n events may happen at time now.
func (lim *Limiter) AllowN(now time.Time, n int) bool {
	return lim.limiter.AllowN(now, n)
}

// Burst returns the maximum burst size. Burst is the maximum number of tokens
// that can be consumed in a single call to Allow, Reserve, or Wait, so higher
// Burst values allow more events to happen at once.
func (lim *Limiter) Burst() int {
	return lim.limiter.Burst()
}

// Limit returns the maximum overall event rate.
func (lim *Limiter) Limit() rate.Limit {
	return lim.limiter.Limit()
}

// SetFrequency sets a new frequency for the limiter.
func (lim *Limiter) SetFrequency(newFrequency float64) error {
	if newFrequency < 0.0 {
		return errors.New("invalid argument to SetFrequency")
	}
	lim.frequency.Store(newFrequency)
	lim.limiter.SetLimit(rate.Limit(newFrequency))
	lim.limiter.SetBurst(int(newFrequency))
	return nil
}

// Wait is shorthand for WaitN(ctx, 1).
func (lim *Limiter) Wait(ctx context.Context) (err error) {
	err = lim.limiter.Wait(ctx)
	return
}

// WaitN blocks until lim permits n events to happen. It returns an error if n
// exceeds the Limiter's burst size, the Context is canceled, or the expected
// wait time exceeds the Context's Deadline.
func (lim *Limiter) WaitN(ctx context.Context, n int) (err error) {
	err = lim.limiter.WaitN(ctx, n)
	return
}
