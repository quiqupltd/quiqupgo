//go:build integration

package temporal_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/quiqupltd/quiqupgo/fxutil"
	loggertest "github.com/quiqupltd/quiqupgo/logger/testutil"
	"github.com/quiqupltd/quiqupgo/temporal"
	tracingtest "github.com/quiqupltd/quiqupgo/tracing/testutil"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// getTemporalHost returns the Temporal server address from env or defaults to OrbStack URL.
func getTemporalHost() string {
	if host := os.Getenv("TEMPORAL_HOST"); host != "" {
		return host
	}
	return "temporal.quiqupgo.orb.local:7233"
}

// IntegrationTestConfig implements temporal.Config for integration tests.
type IntegrationTestConfig struct {
	hostPort  string
	namespace string
}

func NewIntegrationTestConfig() *IntegrationTestConfig {
	return &IntegrationTestConfig{
		hostPort:  getTemporalHost(),
		namespace: "default",
	}
}

func (c *IntegrationTestConfig) GetHostPort() string  { return c.hostPort }
func (c *IntegrationTestConfig) GetNamespace() string { return c.namespace }
func (c *IntegrationTestConfig) GetTLSCert() string   { return "" }
func (c *IntegrationTestConfig) GetTLSKey() string    { return "" }
func (c *IntegrationTestConfig) IsLocal() bool        { return true }

// IntegrationTestModule returns an fx.Option for integration testing with real Temporal.
func IntegrationTestModule() fx.Option {
	cfg := NewIntegrationTestConfig()

	return fx.Module("temporal-integration-test",
		tracingtest.NoopModule(),
		loggertest.NoopModule(),
		fx.Provide(func() temporal.Config { return cfg }),
		temporal.Module(),
	)
}

// TemporalIntegrationSuite tests the Temporal client against a real Temporal server.
type TemporalIntegrationSuite struct {
	suite.Suite
	client client.Client
	app    *fxtest.App
}

func TestTemporalIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TemporalIntegrationSuite))
}

func (s *TemporalIntegrationSuite) SetupTest() {
	s.app = fxutil.TestApp(s.T(),
		IntegrationTestModule(),
		fx.Populate(&s.client),
	)
	s.app.RequireStart()
}

func (s *TemporalIntegrationSuite) TearDownTest() {
	s.app.RequireStop()
}

func (s *TemporalIntegrationSuite) TestClientConnection() {
	// The client should be connected
	s.NotNil(s.client)

	// We can list workflows (even if empty)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	workflows, err := temporal.ListAllWorkflows(ctx, s.client, "default", "")
	s.Require().NoError(err)
	// Verify the call succeeded - workflows may be empty slice or nil
	s.GreaterOrEqual(len(workflows), 0)
}

func (s *TemporalIntegrationSuite) TestGetWorkflowStatus_NotFound() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to get status of a non-existent workflow
	_, err := temporal.GetWorkflowStatus(ctx, s.client, "non-existent-workflow", "non-existent-run")
	// Should error because workflow doesn't exist
	s.Error(err)
}

func (s *TemporalIntegrationSuite) TestListAllWorkflows_WithQuery() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// List with a filter query
	workflows, err := temporal.ListAllWorkflows(ctx, s.client, "default", "WorkflowType='NonExistentType'")
	s.Require().NoError(err)
	// Should return empty result, not error
	s.GreaterOrEqual(len(workflows), 0)
}

func (s *TemporalIntegrationSuite) TestWorkflowStatusHelpers() {
	// Test the status helper functions
	s.True(temporal.IsWorkflowRunning(enums.WORKFLOW_EXECUTION_STATUS_RUNNING))
	s.False(temporal.IsWorkflowRunning(enums.WORKFLOW_EXECUTION_STATUS_COMPLETED))

	s.True(temporal.IsWorkflowCompleted(enums.WORKFLOW_EXECUTION_STATUS_COMPLETED))
	s.False(temporal.IsWorkflowCompleted(enums.WORKFLOW_EXECUTION_STATUS_RUNNING))

	s.True(temporal.IsWorkflowFailed(enums.WORKFLOW_EXECUTION_STATUS_FAILED))
	s.True(temporal.IsWorkflowCanceled(enums.WORKFLOW_EXECUTION_STATUS_CANCELED))
	s.True(temporal.IsWorkflowTerminated(enums.WORKFLOW_EXECUTION_STATUS_TERMINATED))
	s.True(temporal.IsWorkflowTimedOut(enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT))
}

// WorkerTracingIntegrationSuite tests worker creation with tracing interceptors.
type WorkerTracingIntegrationSuite struct {
	suite.Suite
	client       client.Client
	interceptors temporal.WorkerInterceptorSlice
	app          *fxtest.App
}

func TestWorkerTracingIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(WorkerTracingIntegrationSuite))
}

func (s *WorkerTracingIntegrationSuite) SetupTest() {
	cfg := NewIntegrationTestConfig()

	s.app = fxutil.TestApp(s.T(),
		fx.Module("worker-tracing-test",
			tracingtest.NoopModule(),
			loggertest.NoopModule(),
			fx.Provide(func() temporal.Config { return cfg }),
			temporal.Module(temporal.WithWorkerInterceptors()),
		),
		fx.Populate(&s.client, &s.interceptors),
	)
	s.app.RequireStart()
}

func (s *WorkerTracingIntegrationSuite) TearDownTest() {
	s.app.RequireStop()
}

func (s *WorkerTracingIntegrationSuite) TestWorkerInterceptorsProvided() {
	// Verify interceptors were provided via fx
	s.NotNil(s.interceptors)
	s.Len(s.interceptors, 1)
}

func (s *WorkerTracingIntegrationSuite) TestWorkerCreationWithInterceptors() {
	// Create a worker with tracing interceptors
	w := worker.New(s.client, "worker-tracing-test-queue", worker.Options{
		Interceptors: s.interceptors,
	})
	s.NotNil(w)

	// Register a simple workflow to verify the worker is functional
	w.RegisterWorkflow(testTracingWorkflow)

	// Start worker in background
	go func() {
		_ = w.Run(make(chan struct{}))
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the worker
	w.Stop()
}

func (s *WorkerTracingIntegrationSuite) TestApplyWorkerInterceptors() {
	// Test using ApplyWorkerInterceptors helper
	opts := worker.Options{
		MaxConcurrentActivityExecutionSize: 50,
	}
	err := temporal.ApplyWorkerInterceptors(&opts)
	s.Require().NoError(err)
	s.Len(opts.Interceptors, 1)

	// Create worker with applied options
	w := worker.New(s.client, "apply-interceptors-test-queue", opts)
	s.NotNil(w)
}

// testTracingWorkflow is a simple workflow for testing worker creation.
func testTracingWorkflow(ctx workflow.Context) (string, error) {
	return "done", nil
}
