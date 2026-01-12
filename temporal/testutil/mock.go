// Package testutil provides testing utilities for the temporal module.
package testutil

import (
	"github.com/quiqupltd/quiqupgo/temporal"
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
