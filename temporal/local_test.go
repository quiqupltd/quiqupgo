package temporal_test

import (
	"testing"

	loggertest "github.com/quiqupltd/quiqupgo/logger/testutil"
	"github.com/quiqupltd/quiqupgo/temporal"
	"github.com/quiqupltd/quiqupgo/temporal/testutil"
	tracingtest "github.com/quiqupltd/quiqupgo/tracing/testutil"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestStandardLocalConfig(t *testing.T) {
	cfg := &temporal.StandardLocalConfig{
		HostPort:  "temporal-local:7233",
		Namespace: "my-namespace",
	}

	assert.Equal(t, "temporal-local:7233", cfg.GetHostPort())
	assert.Equal(t, "my-namespace", cfg.GetNamespace())
}

func TestStandardLocalConfig_Defaults(t *testing.T) {
	cfg := &temporal.StandardLocalConfig{}

	assert.Equal(t, "localhost:7233", cfg.GetHostPort())
	assert.Equal(t, "default", cfg.GetNamespace())
}

func TestNoopLocalConfig(t *testing.T) {
	cfg := testutil.NewNoopLocalConfig()

	assert.Equal(t, "localhost:7233", cfg.GetHostPort())
	assert.Equal(t, "default", cfg.GetNamespace())
}

func TestLocalModule_FxWiring(t *testing.T) {
	// Verify the local module wires up correctly with fx.
	// Uses a lazy client so it won't actually connect.
	type localClientParam struct {
		fx.In
		Client client.Client `name:"temporallocal"`
	}

	var result localClientParam

	app := fxtest.New(t,
		tracingtest.NoopModule(),
		loggertest.NoopModule(),
		fx.Provide(func() temporal.LocalConfig {
			return &temporal.StandardLocalConfig{
				HostPort:  "localhost:7233",
				Namespace: "default",
			}
		}),
		temporal.LocalModule(),
		fx.Populate(&result),
	)

	app.RequireStart()
	defer app.RequireStop()

	assert.NotNil(t, result.Client)
}

func TestNoopLocalModule_FxWiring(t *testing.T) {
	type localClientParam struct {
		fx.In
		Client client.Client `name:"temporallocal"`
	}

	var result localClientParam

	app := fxtest.New(t,
		testutil.NoopLocalModule(),
		fx.Populate(&result),
	)

	app.RequireStart()
	defer app.RequireStop()

	// NoopLocalModule provides nil client
	assert.Nil(t, result.Client)
}

func TestDualClientModule_FxWiring(t *testing.T) {
	// Verify both cloud and local clients can coexist in the same fx container.
	type dualClients struct {
		fx.In
		Cloud client.Client
		Local client.Client `name:"temporallocal"`
	}

	var result dualClients

	app := fxtest.New(t,
		tracingtest.NoopModule(),
		loggertest.NoopModule(),
		fx.Provide(func() temporal.Config {
			return &temporal.StandardConfig{
				HostPort:  "localhost:7233",
				Namespace: "default",
			}
		}),
		fx.Provide(func() temporal.LocalConfig {
			return &temporal.StandardLocalConfig{
				HostPort:  "localhost:7233",
				Namespace: "default",
			}
		}),
		temporal.Module(temporal.WithLazyClient()),
		temporal.LocalModule(),
		fx.Populate(&result),
	)

	app.RequireStart()
	defer app.RequireStop()

	assert.NotNil(t, result.Cloud)
	assert.NotNil(t, result.Local)
}
