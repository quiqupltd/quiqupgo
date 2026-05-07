package temporal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockStarter struct {
	startErr    error
	startCalls  int
	failNTimes  int
	startCalled chan struct{}
}

func (m *mockStarter) Start() error {
	m.startCalls++
	if m.startCalled != nil {
		m.startCalled <- struct{}{}
	}
	if m.startCalls <= m.failNTimes {
		return m.startErr
	}
	return nil
}

var testBackoffs = []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond}

func TestStartWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	w := &mockStarter{}
	status := &WorkerStatus{}
	logger := zap.NewNop()

	err := StartWithRetryBackoffs(context.Background(), w, status, logger, testBackoffs)

	require.NoError(t, err)
	assert.True(t, status.IsReady())
	assert.Equal(t, 1, w.startCalls)
}

func TestStartWithRetry_SuccessAfterRetries(t *testing.T) {
	w := &mockStarter{
		startErr:   errors.New("connection refused"),
		failNTimes: 2,
	}
	status := &WorkerStatus{}
	logger := zap.NewNop()

	err := StartWithRetryBackoffs(context.Background(), w, status, logger, testBackoffs)

	require.NoError(t, err)
	assert.True(t, status.IsReady())
	assert.Equal(t, 3, w.startCalls)
}

func TestStartWithRetry_ExhaustsRetries(t *testing.T) {
	w := &mockStarter{
		startErr:   errors.New("connection refused"),
		failNTimes: 100, // always fail
	}
	status := &WorkerStatus{}
	logger := zap.NewNop()

	err := StartWithRetryBackoffs(context.Background(), w, status, logger, testBackoffs)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start after 3 attempts")
	assert.False(t, status.IsReady())
	assert.Equal(t, 3, w.startCalls)
}

func TestStartWithRetry_ContextCancelled(t *testing.T) {
	w := &mockStarter{
		startErr:    errors.New("connection refused"),
		failNTimes:  100,
		startCalled: make(chan struct{}, 10),
	}
	status := &WorkerStatus{}
	logger := zap.NewNop()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-w.startCalled
		cancel()
	}()

	err := StartWithRetryBackoffs(ctx, w, status, logger, []time.Duration{time.Hour, time.Hour})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
	assert.False(t, status.IsReady())
}

func TestWorkerStatus_DefaultNotReady(t *testing.T) {
	status := &WorkerStatus{}
	assert.False(t, status.IsReady())
}

func TestWorkerStatus_SetReady(t *testing.T) {
	status := &WorkerStatus{}
	status.SetReady(true)
	assert.True(t, status.IsReady())
	status.SetReady(false)
	assert.False(t, status.IsReady())
}
