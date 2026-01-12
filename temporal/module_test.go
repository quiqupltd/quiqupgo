package temporal_test

import (
	"context"
	"testing"

	"github.com/quiqupltd/quiqupgo/temporal"
	"github.com/quiqupltd/quiqupgo/temporal/testutil"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/api/enums/v1"
	"go.uber.org/zap"
)

func TestStandardConfig(t *testing.T) {
	cfg := &temporal.StandardConfig{
		HostPort:  "temporal:7233",
		Namespace: "my-namespace",
		TLSCert:   "cert-data",
		TLSKey:    "key-data",
	}

	assert.Equal(t, "temporal:7233", cfg.GetHostPort())
	assert.Equal(t, "my-namespace", cfg.GetNamespace())
	assert.Equal(t, "cert-data", cfg.GetTLSCert())
	assert.Equal(t, "key-data", cfg.GetTLSKey())
	assert.False(t, cfg.IsLocal())
}

func TestStandardConfig_Defaults(t *testing.T) {
	cfg := &temporal.StandardConfig{}

	assert.Equal(t, "localhost:7233", cfg.GetHostPort())
	assert.Equal(t, "default", cfg.GetNamespace())
	assert.Equal(t, "", cfg.GetTLSCert())
	assert.Equal(t, "", cfg.GetTLSKey())
	assert.True(t, cfg.IsLocal())
}

func TestNoopConfig(t *testing.T) {
	cfg := testutil.NewNoopConfig()

	assert.Equal(t, "localhost:7233", cfg.GetHostPort())
	assert.Equal(t, "default", cfg.GetNamespace())
	assert.Equal(t, "", cfg.GetTLSCert())
	assert.Equal(t, "", cfg.GetTLSKey())
}

func TestZapLoggerAdapter(t *testing.T) {
	zapLog := zap.NewNop()
	adapter := temporal.NewZapLoggerAdapter(zapLog)

	// Test all methods don't panic
	adapter.Debug("debug message", "key", "value")
	adapter.Info("info message", "key", "value")
	adapter.Warn("warn message", "key", "value")
	adapter.Error("error message", "key", "value")

	// Test With
	withAdapter := adapter.With("context", "value")
	assert.NotNil(t, withAdapter)
	withAdapter.Info("with context")
}

func TestZapLoggerAdapter_AllTypes(t *testing.T) {
	zapLog := zap.NewNop()
	adapter := temporal.NewZapLoggerAdapter(zapLog)

	// Test various value types in toZapFields
	adapter.Debug("test", "string", "value", "int", 42, "float", 3.14, "bool", true)
	adapter.Info("test", "nil", nil, "slice", []int{1, 2, 3})
}

func TestZapLoggerAdapter_NonStringKey(t *testing.T) {
	zapLog := zap.NewNop()
	adapter := temporal.NewZapLoggerAdapter(zapLog)

	// Test with non-string key (integer) - should be converted to string
	adapter.Debug("test with non-string key", 123, "value")
	adapter.Info("test", 456, "another value")
	adapter.Warn("test", struct{ Name string }{"test"}, "complex key")
	adapter.Error("test", []int{1, 2}, "slice as key")
}

func TestZapLoggerAdapter_OddKeyvals(t *testing.T) {
	zapLog := zap.NewNop()
	adapter := temporal.NewZapLoggerAdapter(zapLog)

	// Test with odd number of keyvals (should handle gracefully)
	adapter.Debug("test with odd keyvals", "key1", "value1", "dangling")
}

func TestWorkflowStatusHelpers(t *testing.T) {
	tests := []struct {
		name     string
		status   enums.WorkflowExecutionStatus
		running  bool
		complete bool
		failed   bool
		canceled bool
		termed   bool
		timedOut bool
	}{
		{
			name:    "running",
			status:  enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
			running: true,
		},
		{
			name:     "completed",
			status:   enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
			complete: true,
		},
		{
			name:   "failed",
			status: enums.WORKFLOW_EXECUTION_STATUS_FAILED,
			failed: true,
		},
		{
			name:     "canceled",
			status:   enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
			canceled: true,
		},
		{
			name:   "terminated",
			status: enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
			termed: true,
		},
		{
			name:     "timed_out",
			status:   enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
			timedOut: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.running, temporal.IsWorkflowRunning(tt.status))
			assert.Equal(t, tt.complete, temporal.IsWorkflowCompleted(tt.status))
			assert.Equal(t, tt.failed, temporal.IsWorkflowFailed(tt.status))
			assert.Equal(t, tt.canceled, temporal.IsWorkflowCanceled(tt.status))
			assert.Equal(t, tt.termed, temporal.IsWorkflowTerminated(tt.status))
			assert.Equal(t, tt.timedOut, temporal.IsWorkflowTimedOut(tt.status))
		})
	}
}

func TestNewClient_WithInvalidTLS(t *testing.T) {
	cfg := &temporal.StandardConfig{
		HostPort:  "temporal.example.com:7233", // Non-localhost
		Namespace: "default",
		TLSCert:   "-----BEGIN CERTIFICATE-----\ninvalid\n-----END CERTIFICATE-----",
		TLSKey:    "-----BEGIN PRIVATE KEY-----\ninvalid\n-----END PRIVATE KEY-----",
	}

	ctx := context.Background()
	_, err := temporal.NewClient(ctx, cfg, zap.NewNop(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TLS")
}

func TestNewClient_LocalhostSkipsTLS(t *testing.T) {
	cfg := &temporal.StandardConfig{
		HostPort:  "localhost:7233",
		Namespace: "default",
		TLSCert:   "will-be-ignored",
		TLSKey:    "will-be-ignored",
	}

	ctx := context.Background()
	// This will fail to connect but shouldn't fail on TLS
	_, err := temporal.NewClient(ctx, cfg, zap.NewNop(), nil)
	// The error should be about connection, not TLS
	if err != nil {
		assert.NotContains(t, err.Error(), "TLS")
	}
}

// Note: Integration tests for the actual Temporal client would require
// a running Temporal server and are better suited for integration test suites.
// Example integration test structure:
//
// func TestModule_Integration(t *testing.T) {
//     if testing.Short() {
//         t.Skip("skipping integration test")
//     }
//
//     var c client.Client
//     app := fx.New(
//         fx.NopLogger,
//         tracingtest.NoopModule(),
//         loggertest.NoopModule(),
//         fx.Provide(func() temporal.Config {
//             return &temporal.StandardConfig{
//                 HostPort:  "localhost:7233",
//                 Namespace: "default",
//             }
//         }),
//         temporal.Module(),
//         fx.Populate(&c),
//     )
//     // ... test with actual Temporal server
// }
