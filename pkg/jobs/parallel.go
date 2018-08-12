package jobs

import (
	"sync"
)

type Limiter struct {
	c chan bool
}

// If size is 0 the limiter is inactive
// If size is > 0, limiter is active and set to a max simultaneous of `size`
func NewLimiter(size int) *Limiter {
	var c chan bool
	if size > 0 {
		c = make(chan bool, size)
	}

	return &Limiter{c: c}
}
func (l *Limiter) Add() {
	if l.isActive() {
		l.c <- true
	}
}
func (l *Limiter) Done() {
	if l.isActive() {
		<-l.c
	}
}
func (l *Limiter) isActive() bool {
	return l.c != nil
}

// ParallelFunc is a type for parallelizeable functions that take
//
type ParallelFunc func(arg interface{}) error

type ParallelOpts struct {
	MaxParallel int
}

// Pass in a ParallaelFunc and an array of arguments, one for each call
func Parallel(executor ParallelFunc, args []interface{}, options *ParallelOpts) []error {
	var (
		errs    []error
		wg      sync.WaitGroup
		limiter *Limiter
	)

	limiter = NewLimiter(options.MaxParallel)

	for idx, _ := range args {
		limiter.Add()

		wg.Add(1)
		go func(arg interface{}) {
			defer wg.Done()
			defer limiter.Done()

			if err := executor(arg); err != nil {
				errs = append(errs, err)
			}
		}(args[idx])
	}

	wg.Wait()

	return errs
}
