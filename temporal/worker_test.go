package temporal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestWorkerInterceptors(t *testing.T) {
	interceptors, err := WorkerInterceptors()
	require.NoError(t, err)
	assert.Len(t, interceptors, 1)
	assert.NotNil(t, interceptors[0])
}

func TestApplyWorkerInterceptors(t *testing.T) {
	t.Run("applies to empty options", func(t *testing.T) {
		opts := worker.Options{}
		err := ApplyWorkerInterceptors(&opts)
		require.NoError(t, err)
		assert.Len(t, opts.Interceptors, 1)
	})

	t.Run("appends to existing interceptors", func(t *testing.T) {
		// Create initial interceptors
		existing, err := WorkerInterceptors()
		require.NoError(t, err)

		opts := worker.Options{
			Interceptors: existing,
		}
		err = ApplyWorkerInterceptors(&opts)
		require.NoError(t, err)
		assert.Len(t, opts.Interceptors, 2)
	})
}

func TestWorkerInterceptorsModule(t *testing.T) {
	var interceptors WorkerInterceptorSlice

	app := fxtest.New(t,
		WorkerInterceptorsModule(),
		fx.Populate(&interceptors),
	)

	app.RequireStart()
	defer app.RequireStop()

	assert.Len(t, interceptors, 1)
	assert.NotNil(t, interceptors[0])
}

func TestProvideWorkerInterceptors(t *testing.T) {
	interceptors, err := provideWorkerInterceptors()
	require.NoError(t, err)
	assert.Len(t, interceptors, 1)
}

func TestWorkerInterceptors_WithWorkerNew(t *testing.T) {
	// This test verifies that worker.New accepts our interceptors without error.
	// Note: worker.New doesn't connect immediately, so this works without a server.

	// Create a client connected to localhost (may or may not be running)
	logger, _ := zap.NewDevelopment()
	c, err := NewClient(
		t.Context(),
		&StandardConfig{HostPort: "localhost:7233", Namespace: "default"},
		logger,
		nil, // no tracer
	)
	if err != nil {
		t.Skip("skipping test - cannot create client")
	}
	defer c.Close()

	// Get our interceptors
	interceptors, err := WorkerInterceptors()
	require.NoError(t, err)

	// Create worker with our interceptors - this should succeed
	w := worker.New(c, "test-task-queue", worker.Options{
		Interceptors: interceptors,
	})
	assert.NotNil(t, w)
}

func TestApplyWorkerInterceptors_WithWorkerNew(t *testing.T) {
	// Create a client
	logger, _ := zap.NewDevelopment()
	c, err := NewClient(
		t.Context(),
		&StandardConfig{HostPort: "localhost:7233", Namespace: "default"},
		logger,
		nil,
	)
	if err != nil {
		t.Skip("skipping test - cannot create client")
	}
	defer c.Close()

	// Apply interceptors to options
	opts := worker.Options{
		MaxConcurrentActivityExecutionSize: 50,
	}
	err = ApplyWorkerInterceptors(&opts)
	require.NoError(t, err)

	// Create worker with applied options
	w := worker.New(c, "test-task-queue", opts)
	assert.NotNil(t, w)
}
