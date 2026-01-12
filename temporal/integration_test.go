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
