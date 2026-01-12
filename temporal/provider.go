package temporal

import (
	"context"
	"crypto/tls"
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.uber.org/zap"
)

// NewClient creates a new Temporal client with the given configuration.
// It automatically configures:
//   - TLS if connecting to a remote server with certificates
//   - OpenTelemetry tracing interceptor
//   - Zap logger adapter
func NewClient(ctx context.Context, cfg Config, logger *zap.Logger, tracer trace.Tracer) (client.Client, error) {
	hostPort := cfg.GetHostPort()
	namespace := cfg.GetNamespace()

	// Build client options
	opts := client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
		Logger:    NewZapLoggerAdapter(logger.Named("temporal")),
	}

	// Add TLS configuration if not localhost and certs are provided
	if hostPort != "localhost:7233" && cfg.GetTLSCert() != "" && cfg.GetTLSKey() != "" {
		tlsCfg, err := getTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		opts.ConnectionOptions = client.ConnectionOptions{
			TLS: tlsCfg,
		}
	}

	// Add OpenTelemetry tracing interceptor if tracer is available
	if tracer != nil {
		tracerInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create tracing interceptor: %w", err)
		}
		opts.Interceptors = append(opts.Interceptors, tracerInterceptor)
	}

	// Create the client
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
