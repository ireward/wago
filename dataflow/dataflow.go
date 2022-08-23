package dataflow

import (
	"context"
)

// Dataflow represents a chain of functions where the next function is executed by the previous one by
// passing a commong object holding the shared state. The recursive nature of the calls causes acquired
// resources to be held until the full dataflow terminates.
type Dataflow[T any] struct {
	ctx       context.Context
	step      int
	max       int
	fns       []ChainedFn[T]
	successCb Callback[T]
	abortCb   Callback[T]
	errorCb   ErrorCallback[T]
}

// WithSuccessCb modifies the Dataflow to execute a callback after all the functions in the chain have
// been executed.
func (d *Dataflow[T]) WithSuccessCb(cb Callback[T]) *Dataflow[T] {
	d.successCb = cb
	return d
}

// WithAbortCb modifies the Dataflow to execute a callback after the dataflow has
// been aborted.
func (d *Dataflow[T]) WithAbortCb(cb Callback[T]) *Dataflow[T] {
	d.abortCb = cb
	return d
}

// WithErrorCb modifies the Dataflow to execute a callback after an error has been encountered.
func (d *Dataflow[T]) WithErrorCb(cb ErrorCallback[T]) *Dataflow[T] {
	d.errorCb = cb
	return d
}

// New instantiates a new dataflow.
func New[T any](ctx context.Context, fns ...ChainedFn[T]) *Dataflow[T] {
	return &Dataflow[T]{
		fns:  fns,
		max:  len(fns),
		step: -1,
		ctx:  ctx,
	}
}

// Run executes the Dataflow with the given argument. It aborts execution and
// returns an error if any of the functions returns an error.
func (d *Dataflow[T]) Run(arg T) error {
	var err error
	d.step++
	if d.step >= d.max {
		// trigger success callback
		if d.successCb != nil {
			d.successCb(arg)
			d.successCb = nil
		}
		return nil
	}
	for {
		select {
		case <-d.ctx.Done():
			// trigger abort callback
			if d.abortCb != nil {
				d.abortCb(arg)
				d.abortCb = nil
			}
			return err
		default:
			if err = d.fns[d.step](arg, d.Run); err != nil {
				// trigger error callback
				if d.errorCb != nil {
					d.errorCb(arg, err)
					d.errorCb = nil
				}
			}
			return err
		}
	}
}

// ChainedFn exposes the Dataflow as a ChainedFn without calling it.
func (d *Dataflow[T]) ChainedFn(arg T, next Next[T]) error {
	return d.append(func(arg T, done Next[T]) error {
		if next == nil {
			return done(arg)
		}
		if err := done(arg); err != nil {
			return err
		}
		return next(arg)
	}).Run(arg)
}

// append adds a new function to the Dataflow.
func (d *Dataflow[T]) append(fn ChainedFn[T]) *Dataflow[T] {
	d.fns = append(d.fns, fn)
	d.max++
	return d
}

var _ ChainedFn[int] = new(Dataflow[int]).ChainedFn

// ChainedFn represents the interface for callbacks used in a Dataflow.
type ChainedFn[T any] func(arg T, next Next[T]) error

// Next represents the interface for the next step in a Dataflow.
type Next[T any] func(arg T) error

// Callback represents the interface for the callback functions.
type Callback[T any] func(arg T)

// ErrorCallback represents the interface for the error callback functions.
type ErrorCallback[T any] func(arg T, err error)

// EmptyNext is an implementation of Next that does nothing.
func EmptyNext[T any](arg T) error {
	return nil
}
