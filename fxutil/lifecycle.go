package fxutil

import (
	"context"

	"go.uber.org/fx"
)

// OnStop is a helper function that registers a stop hook with the fx lifecycle.
// It simplifies the common pattern of registering cleanup functions.
//
// Example:
//
//	func provideDB(lc fx.Lifecycle) (*sql.DB, error) {
//	    db, err := sql.Open("postgres", dsn)
//	    if err != nil {
//	        return nil, err
//	    }
//	    fxutil.OnStop(lc, func(ctx context.Context) error {
//	        return db.Close()
//	    })
//	    return db, nil
//	}
func OnStop(lc fx.Lifecycle, stop func(ctx context.Context) error) {
	lc.Append(fx.Hook{
		OnStop: stop,
	})
}

// OnStart is a helper function that registers a start hook with the fx lifecycle.
//
// Example:
//
//	func provideServer(lc fx.Lifecycle, srv *http.Server) {
//	    fxutil.OnStart(lc, func(ctx context.Context) error {
//	        go srv.ListenAndServe()
//	        return nil
//	    })
//	}
func OnStart(lc fx.Lifecycle, start func(ctx context.Context) error) {
	lc.Append(fx.Hook{
		OnStart: start,
	})
}

// OnStartStop is a helper function that registers both start and stop hooks.
//
// Example:
//
//	func provideWorker(lc fx.Lifecycle, w *Worker) {
//	    fxutil.OnStartStop(lc,
//	        func(ctx context.Context) error {
//	            return w.Start(ctx)
//	        },
//	        func(ctx context.Context) error {
//	            return w.Stop(ctx)
//	        },
//	    )
//	}
func OnStartStop(lc fx.Lifecycle, start, stop func(ctx context.Context) error) {
	lc.Append(fx.Hook{
		OnStart: start,
		OnStop:  stop,
	})
}

// SimpleOnStop is a helper for stop functions that don't need context.
//
// Example:
//
//	fxutil.SimpleOnStop(lc, func() error {
//	    return db.Close()
//	})
func SimpleOnStop(lc fx.Lifecycle, stop func() error) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return stop()
		},
	})
}

// SimpleOnStart is a helper for start functions that don't need context.
func SimpleOnStart(lc fx.Lifecycle, start func() error) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return start()
		},
	})
}
