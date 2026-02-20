package temporal

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// localClientResult is used with fx.Annotate to tag the local client.
type localClientResult struct {
	fx.Out
	Client client.Client `name:"temporallocal"`
}

// LocalModule returns an fx.Option that provides a local (in-cluster) Temporal client
// tagged with name:"temporallocal". This is intended for GKE-based Temporal clusters
// that don't require TLS or authentication.
//
// The local client always uses NewLazyClient (doesn't block startup) and includes
// OTel tracing + metrics when available.
//
// It provides:
//   - client.Client `name:"temporallocal"` (lazy Temporal client)
//
// It requires:
//   - temporal.LocalConfig (must be provided by the application)
//   - *zap.Logger (from logger module)
//   - trace.Tracer (from tracing module)
//   - metric.Meter (from tracing module)
func LocalModule() fx.Option {
	return fx.Module("temporal-local",
		fx.Provide(provideLocalClient),
		fx.Invoke(registerLocalLifecycleHooks),
	)
}

// provideLocalClient creates a lazy Temporal client for local/in-cluster use.
func provideLocalClient(
	cfg LocalConfig,
	logger *zap.Logger,
	tracer trace.Tracer,
	meter metric.Meter,
) (localClientResult, error) {
	c, err := newClient(context.Background(), ClientParams{
		Config: &localConfigAdapter{cfg},
		Logger: logger,
		Tracer: tracer,
		Meter:  meter,
		Lazy:   true,
	})
	if err != nil {
		return localClientResult{}, err
	}
	return localClientResult{Client: c}, nil
}

// localLifecycleParams is used with fx.Annotate to inject the tagged local client.
type localLifecycleParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Client    client.Client `name:"temporallocal"`
}

// registerLocalLifecycleHooks registers shutdown hooks for the local client.
func registerLocalLifecycleHooks(p localLifecycleParams) {
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.Client.Close()
			return nil
		},
	})
}

// localConfigAdapter wraps a LocalConfig to satisfy the Config interface
// required by newClient. TLS methods return empty strings since local
// clusters don't use TLS.
type localConfigAdapter struct {
	LocalConfig
}

func (a *localConfigAdapter) GetTLSCert() string { return "" }
func (a *localConfigAdapter) GetTLSKey() string  { return "" }
