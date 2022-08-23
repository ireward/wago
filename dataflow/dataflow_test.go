package dataflow

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataFlow_Basic(t *testing.T) {
	ctx := context.Background()
	x := func(arg int, next Next[int]) error {
		return next(arg + 1)
	}
	y := func(arg int, next Next[int]) error {
		return next(arg + 1)
	}
	z := func(arg int, next Next[int]) error {
		return errors.New("fatal error occured")
	}
	result := 0
	d1 := New(ctx, x, y, func(arg int, next Next[int]) error {
		result = arg
		return next(arg)
	}).Run(1)
	// make sure no error was returned
	assert.NoError(t, d1)
	assert.Equal(t, 3, result)

	d2 := New(ctx, x, y, z).Run(1)
	fmt.Println(d2)
	assert.Error(t, d2)
}

func TestDataFlow_WithCallbacks(t *testing.T) {
	ctx, cancle := context.WithCancel(context.Background())
	x := func(arg int, next Next[int]) error {
		return next(arg + 1)
	}
	y := func(arg int, next Next[int]) error {
		return next(arg + 1)
	}

	var calledSuccess bool
	d := New(ctx, x, y).WithSuccessCb(func(arg int) {
		calledSuccess = true
	})

	d.Run(1)
	assert.Equal(t, true, calledSuccess)

	var calledAbort bool
	d1 := New(ctx, d.ChainedFn).WithAbortCb(func(arg int) {
		calledAbort = true
	})
	cancle()
	d1.Run(2)
	assert.Equal(t, true, calledAbort)
}
