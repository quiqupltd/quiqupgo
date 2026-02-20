package temporal

import (
	"context"
	"crypto/tls"
	"fmt"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.uber.org/zap"
)

// ClientParams holds parameters for creating a Temporal client.
type ClientParams struct {
	Config Config
	Logger *zap.Logger
	Tracer trace.Tracer
	Meter  metric.Meter
	Lazy   bool
}

// NewClient creates a new Temporal client with the given configuration.
// It automatically configures:
//   - TLS if connecting to a remote server with certificates
//   - OpenTelemetry tracing interceptor
//   - OpenTelemetry metrics handler (when meter is provided)
//   - Zap logger adapter
func NewClient(ctx context.Context, cfg Config, logger *zap.Logger, tracer trace.Tracer, meter metric.Meter) (client.Client, error) {
	return newClient(ctx, ClientParams{
		Config: cfg,
		Logger: logger,
		Tracer: tracer,
		Meter:  meter,
	})
}

// newClient is the internal constructor that supports all client options.
func newClient(_ context.Context, p ClientParams) (client.Client, error) {
	hostPort := p.Config.GetHostPort()
	namespace := p.Config.GetNamespace()

	// Build client options
	opts := client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
		Logger:    NewZapLoggerAdapter(p.Logger.Named("temporal")),
	}

	// Add TLS configuration if not localhost and certs are provided
	if hostPort != "localhost:7233" && p.Config.GetTLSCert() != "" && p.Config.GetTLSKey() != "" {
		tlsCfg, err := getTLSConfig(p.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		opts.ConnectionOptions = client.ConnectionOptions{
			TLS: tlsCfg,
		}
	}

	// Add OpenTelemetry tracing interceptor if tracer is available
	if p.Tracer != nil {
		tracerInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create tracing interceptor: %w", err)
		}
		opts.Interceptors = append(opts.Interceptors, tracerInterceptor)
	}

	// Add OpenTelemetry metrics handler if meter is available
	if p.Meter != nil {
		opts.MetricsHandler = opentelemetry.NewMetricsHandler(opentelemetry.MetricsHandlerOptions{
			Meter: p.Meter,
		})
	}

	// Create the client (lazy or eager)
	if p.Lazy {
		c, err := client.NewLazyClient(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to create lazy Temporal client: %w", err)
		}
		return c, nil
	}

	c, err := client.Dial(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	return c, nil
}

// getTLSConfig creates a TLS configuration from the provided certificates.
func getTLSConfig(cfg Config) (*tls.Config, error) {
	cert, err := tls.X509KeyPair([]byte(cfg.GetTLSCert()), []byte(cfg.GetTLSKey()))
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS key pair: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}
