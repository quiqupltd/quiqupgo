package tracing

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// TracerName returns the standardized tracer name for this library.
func TracerName() string {
	return "github.com/quiqupltd/quiqupgo/tracing"
}

// GetResource creates an OpenTelemetry resource with service and deployment attributes.
func GetResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.GetServiceName()),
			semconv.DeploymentEnvironment(cfg.GetEnvironmentName()),
		),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithProcess(),
	)
}

// GetTLSConfig creates a TLS configuration from base64-encoded certificates.
// Returns nil if no TLS configuration is needed.
func GetTLSConfig(cfg Config) (*tls.Config, error) {
	certB64 := cfg.GetOTLPTLSCert()
	keyB64 := cfg.GetOTLPTLSKey()
	caB64 := cfg.GetOTLPTLSCA()

	// If no cert/key provided, return nil (use system defaults or insecure)
	if certB64 == "" && keyB64 == "" && caB64 == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Load client certificate if provided
	if certB64 != "" && keyB64 != "" {
		certPEM, err := base64.StdEncoding.DecodeString(certB64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode TLS certificate: %w", err)
		}

		keyPEM, err := base64.StdEncoding.DecodeString(keyB64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode TLS key: %w", err)
		}

		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS key pair: %w", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if caB64 != "" {
		caPEM, err := base64.StdEncoding.DecodeString(caB64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode TLS CA certificate: %w", err)
		}

		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("failed to parse TLS CA certificate")
		}

		tlsConfig.RootCAs = caPool
	}

	return tlsConfig, nil
}
