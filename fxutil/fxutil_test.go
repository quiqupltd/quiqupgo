package fxutil_test

import (
	"context"
	"testing"
	"time"

	"github.com/quiqupltd/quiqupgo/fxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestOnStop(t *testing.T) {
	var stopped bool

	app := fxutil.TestApp(t,
		fx.Invoke(func(lc fx.Lifecycle) {
			fxutil.OnStop(lc, func(ctx context.Context) error {
				stopped = true
				return nil
			})
		}),
	)

	app.RequireStart()
	assert.False(t, stopped)

	app.RequireStop()
	assert.True(t, stopped)
}

func TestOnStart(t *testing.T) {
	var started bool

	app := fxutil.TestApp(t,
		fx.Invoke(func(lc fx.Lifecycle) {
			fxutil.OnStart(lc, func(ctx context.Context) error {
				started = true
				return nil
			})
		}),
	)

	assert.False(t, started)
	app.RequireStart()
	assert.True(t, started)
}

func TestOnStartStop(t *testing.T) {
	var started, stopped bool

	app := fxutil.TestApp(t,
		fx.Invoke(func(lc fx.Lifecycle) {
			fxutil.OnStartStop(lc,
				func(ctx context.Context) error {
					started = true
					return nil
				},
				func(ctx context.Context) error {
					stopped = true
					return nil
				},
			)
		}),
	)

	assert.False(t, started)
	assert.False(t, stopped)

	app.RequireStart()
	assert.True(t, started)
	assert.False(t, stopped)

	app.RequireStop()
	assert.True(t, started)
	assert.True(t, stopped)
}

func TestSimpleOnStop(t *testing.T) {
	var stopped bool

	app := fxutil.TestApp(t,
		fx.Invoke(func(lc fx.Lifecycle) {
			fxutil.SimpleOnStop(lc, func() error {
				stopped = true
				return nil
			})
		}),
	)

	app.RequireStart()
	assert.False(t, stopped)

	app.RequireStop()
	assert.True(t, stopped)
}

func TestSimpleOnStart(t *testing.T) {
	var started bool

	app := fxutil.TestApp(t,
		fx.Invoke(func(lc fx.Lifecycle) {
			fxutil.SimpleOnStart(lc, func() error {
				started = true
				return nil
			})
		}),
	)

	assert.False(t, started)
	app.RequireStart()
	assert.True(t, started)
}

func TestTestApp(t *testing.T) {
	var value string

	app := fxutil.TestApp(t,
		fx.Provide(func() string { return "hello" }),
		fx.Populate(&value),
	)

	require.NoError(t, app.Err())
	app.RequireStart()
	assert.Equal(t, "hello", value)
	app.RequireStop()
}

func TestTestAppStart(t *testing.T) {
	var value int

	app := fxutil.TestAppStart(t,
		fx.Provide(func() int { return 42 }),
		fx.Populate(&value),
	)

	assert.Equal(t, 42, value)
	app.RequireStop()
}

func TestRunTestApp(t *testing.T) {
	var value string
	testRan := false

	fxutil.RunTestApp(t,
		[]fx.Option{
			fx.Provide(func() string { return "test-value" }),
			fx.Populate(&value),
		},
		func() {
			testRan = true
			assert.Equal(t, "test-value", value)
		},
	)

	assert.True(t, testRan)
}

func TestStartTimeout(t *testing.T) {
	opt := fxutil.StartTimeout(5 * time.Second)
	assert.NotNil(t, opt)
}

func TestStopTimeout(t *testing.T) {
	opt := fxutil.StopTimeout(5 * time.Second)
	assert.NotNil(t, opt)
}

func TestStartContext(t *testing.T) {
	ctx, cancel := fxutil.StartContext(1 * time.Second)
	defer cancel()

	assert.NotNil(t, ctx)
	deadline, ok := ctx.Deadline()
	assert.True(t, ok)
	assert.True(t, deadline.After(time.Now()))
}

func TestStopContext(t *testing.T) {
	ctx, cancel := fxutil.StopContext(1 * time.Second)
	defer cancel()

	assert.NotNil(t, ctx)
	deadline, ok := ctx.Deadline()
	assert.True(t, ok)
	assert.True(t, deadline.After(time.Now()))
}
