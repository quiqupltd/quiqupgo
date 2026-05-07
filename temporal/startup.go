package temporal

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// defaultBackoffs defines the exponential backoff delays between startup retry attempts.
var defaultBackoffs = []time.Duration{
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
	16 * time.Second,
	30 * time.Second,
}

// WorkerStatus exposes the readiness state of a Temporal worker for health checks.
// Safe for concurrent use.
type WorkerStatus struct {
	ready atomic.Bool
}

// IsReady returns true when the worker has successfully connected to Temporal.
func (s *WorkerStatus) IsReady() bool {
	return s.ready.Load()
}

// SetReady sets the worker readiness state.
func (s *WorkerStatus) SetReady(ready bool) {
	s.ready.Store(ready)
}

// Starter is implemented by any type that has a Start method returning an error.
// go.temporal.io/sdk/worker.Worker satisfies this interface.
type Starter interface {
	Start() error
}

// StartWithRetry attempts to start a Temporal worker with exponential backoff.
// On success, it sets status to ready. On exhaustion, it returns an error.
//
// The backoff schedule is: 2s, 4s, 8s, 16s, 30s (5 attempts total).
// Use the returned error to abort the process — the caller should propagate it
// to fx so the application exits and K8s can restart the pod.
//
// Example usage in an fx lifecycle hook:
//
//	status := &temporal.WorkerStatus{}
//	lc.Append(fx.Hook{
//	    OnStart: func(ctx context.Context) error {
//	        return temporal.StartWithRetry(ctx, w, status, logger)
//	    },
//	})
func StartWithRetry(ctx context.Context, w Starter, status *WorkerStatus, logger *zap.Logger) error {
	return StartWithRetryBackoffs(ctx, w, status, logger, defaultBackoffs)
}

// StartWithRetryBackoffs is like StartWithRetry but allows custom backoff durations.
// Useful for testing with shorter intervals.
func StartWithRetryBackoffs(
	ctx context.Context,
	w Starter,
	status *WorkerStatus,
	logger *zap.Logger,
	backoffs []time.Duration,
) error {
	var lastErr error
	for attempt, backoff := range backoffs {
		if err := w.Start(); err != nil {
			lastErr = err
			logger.Warn("temporal worker start failed, retrying",
				zap.Int("attempt", attempt+1),
				zap.Int("max_attempts", len(backoffs)),
				zap.Duration("backoff", backoff),
				zap.Error(err),
			)

			// Don't wait after the last attempt
			if attempt == len(backoffs)-1 {
				break
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf("worker start cancelled: %w", ctx.Err())
			case <-time.After(backoff):
				continue
			}
		}

		status.SetReady(true)
		logger.Info("temporal worker started")
		return nil
	}

	return fmt.Errorf("temporal worker failed to start after %d attempts: %w", len(backoffs), lastErr)
}
