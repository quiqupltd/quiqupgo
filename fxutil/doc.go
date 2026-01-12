// Package fxutil provides shared utilities for uber/fx modules.
//
// It includes helpers for lifecycle management, testing, and common patterns
// used across multiple modules.
//
// Example usage for testing:
//
//	func TestMyService(t *testing.T) {
//	    var svc *MyService
//	    app := fxutil.TestApp(t,
//	        tracingtest.NoopModule(),
//	        loggertest.NoopModule(),
//	        fx.Provide(NewMyService),
//	        fx.Populate(&svc),
//	    )
//	    require.NoError(t, app.Start(context.Background()))
//	    // ... test svc
//	}
package fxutil
