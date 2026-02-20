// Package testutil provides testing utilities for the temporal module.
package testutil

import (
	"github.com/quiqupltd/quiqupgo/temporal"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
)

// NoopConfig is a test configuration for the temporal module.
type NoopConfig struct {
	HostPort  string
	Namespace string
}

// NewNoopConfig creates a NoopConfig with test defaults.
func NewNoopConfig() *NoopConfig {
	return &NoopConfig{
		HostPort:  "localhost:7233",
		Namespace: "default",
	}
}

func (c *NoopConfig) GetHostPort() string  { return c.HostPort }
func (c *NoopConfig) GetNamespace() string { return c.Namespace }
func (c *NoopConfig) GetTLSCert() string   { return "" }
func (c *NoopConfig) GetTLSKey() string    { return "" }

// Ensure NoopConfig implements Config.
var _ temporal.Config = (*NoopConfig)(nil)

// NoopLocalConfig is a test configuration for the local temporal module.
type NoopLocalConfig struct {
	HostPort  string
	Namespace string
}

// NewNoopLocalConfig creates a NoopLocalConfig with test defaults.
func NewNoopLocalConfig() *NoopLocalConfig {
	return &NoopLocalConfig{
		HostPort:  "localhost:7233",
		Namespace: "default",
	}
}

func (c *NoopLocalConfig) GetHostPort() string  { return c.HostPort }
func (c *NoopLocalConfig) GetNamespace() string { return c.Namespace }

// Ensure NoopLocalConfig implements LocalConfig.
var _ temporal.LocalConfig = (*NoopLocalConfig)(nil)

// noopLocalClientResult provides a nil local client for testing.
type noopLocalClientResult struct {
	fx.Out
	Client client.Client `name:"temporallocal"`
}

// NoopLocalModule provides a nil local Temporal client tagged with name:"temporallocal"
// for testing. Use this when your code depends on a local client but you don't need
// a real Temporal connection.
func NoopLocalModule() fx.Option {
	return fx.Module("temporal-local-test",
		fx.Provide(func() noopLocalClientResult {
			return noopLocalClientResult{Client: nil}
		}),
	)
}

// Note: For testing Temporal workflows and activities, use the testsuite package
// from the Temporal SDK directly:
//
//	import "go.temporal.io/sdk/testsuite"
//
//	func TestWorkflow(t *testing.T) {
//	    testSuite := &testsuite.WorkflowTestSuite{}
//	    env := testSuite.NewTestWorkflowEnvironment()
//	    // ... test your workflow
//	}
//
// The testsuite provides comprehensive mocking capabilities including:
// - MockActivity for mocking activity implementations
// - MockWorkflow for mocking child workflows
// - Time control and assertion helpers
// - Signal and query testing
