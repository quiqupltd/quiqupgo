package fxutil

import (
	"context"
	"testing"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestApp creates an fx.App configured for testing.
// It uses fxtest.New under the hood, which provides automatic cleanup
// and better error handling for tests.
//
// Example:
//
//	func TestMyService(t *testing.T) {
//	    var svc *MyService
//	    app := fxutil.TestApp(t,
//	        tracing.testutil.NoopModule(),
//	        logger.testutil.NoopModule(),
//	        fx.Provide(NewMyService),
//	        fx.Populate(&svc),
//	    )
//	    app.RequireStart()
//	    defer app.RequireStop()
//	    // ... test svc
//	}
func TestApp(t testing.TB, opts ...fx.Option) *fxtest.App {
	return fxtest.New(t, opts...)
}

// TestAppStart creates and starts an fx.App for testing.
// It returns the app for manual cleanup if needed.
//
// Example:
//
//	func TestMyService(t *testing.T) {
//	    var svc *MyService
//	    app := fxutil.TestAppStart(t,
//	        tracing.testutil.NoopModule(),
//	        logger.testutil.NoopModule(),
//	        fx.Provide(NewMyService),
//	        fx.Populate(&svc),
//	    )
//	    defer app.RequireStop()
//	    // ... test svc
//	}
func TestAppStart(t testing.TB, opts ...fx.Option) *fxtest.App {
	app := fxtest.New(t, opts...)
	app.RequireStart()
	return app
}

// RunTestApp creates, starts, runs the test function, then stops the app.
// This is a convenience function for the common test pattern.
//
// Example:
//
//	func TestMyService(t *testing.T) {
//	    var svc *MyService
//	    fxutil.RunTestApp(t,
//	        []fx.Option{
//	            tracing.testutil.NoopModule(),
//	            logger.testutil.NoopModule(),
//	            fx.Provide(NewMyService),
//	            fx.Populate(&svc),
//	        },
//	        func() {
//	            // ... test svc
//	        },
//	    )
//	}
func RunTestApp(t testing.TB, opts []fx.Option, test func()) {
	app := fxtest.New(t, opts...)
	app.RequireStart()
	defer app.RequireStop()
	test()
}

// StartTimeout returns an fx.StartTimeout option with the specified duration.
// This is useful for tests that need custom startup timeouts.
func StartTimeout(d time.Duration) fx.Option {
	return fx.StartTimeout(d)
}

// StopTimeout returns an fx.StopTimeout option with the specified duration.
// This is useful for tests that need custom shutdown timeouts.
func StopTimeout(d time.Duration) fx.Option {
	return fx.StopTimeout(d)
}

// StartContext returns a context with the given timeout for app startup.
// This is useful when you need fine-grained control over the start context.
func StartContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// StopContext returns a context with the given timeout for app shutdown.
// This is useful when you need fine-grained control over the stop context.
func StopContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
