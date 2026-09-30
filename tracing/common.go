package tracing

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// TracerName returns the standardized tracer name for this library.
func TracerName() string {
	return "github.com/quiqupltd/quiqupgo/tracing"
}

// GetResource creates an OpenTelemetry resource with service and deployment attributes.
// Note: We use specific process detectors instead of resource.WithProcess() to avoid
// calling os/user.Current() which fails in minimal containers without CGO or $USER set.
func GetResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(cfg.GetServiceName()),
		semconv.DeploymentEnvironment(cfg.GetEnvironmentName()),
	}

	// An empty instance id is worse than none: it exports a label every replica
	// shares, which looks like identity while providing none.
	if id := serviceInstanceID(); id != "" {
		attrs = append(attrs, semconv.ServiceInstanceID(id))
	}

	return resource.New(ctx,
		resource.WithAttributes(attrs...),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		// Use specific process detectors to avoid os/user.Current() dependency
		resource.WithProcessPID(),
		resource.WithProcessExecutableName(),
		resource.WithProcessExecutablePath(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeVersion(),
		resource.WithProcessRuntimeDescription(),
	)
}

// serviceInstanceID identifies the individual process behind a service, which is
// what keeps one replica's telemetry from being merged with another's.
//
// Why this matters more than it looks. Metrics leave here over OTLP, and on the
// Prometheus side an OTLP *resource* attribute lands in `target_info`, not on
// the series. Two replicas sharing a service name therefore write into ONE
// series whose value alternates between two independent counters. Prometheus
// reads each drop as a counter reset and counts the whole new value as an
// increase, so rate() reports a figure with no relationship to reality — and
// keeps reporting one even when nothing is incrementing at all, which is the
// failure mode that makes a dashboard lie rather than merely go blank.
//
// Setting this is NECESSARY BUT NOT SUFFICIENT. The ingest path must also
// promote the attribute onto the series (Mimir's
// -distributor.otel-promote-resource-attributes, or an Alloy transform).
// Without that half this changes what is in target_info and nothing else, so
// do not read a deploy of this alone as a fix.
//
// POD_NAME is preferred because it is explicit, set from the downward API. The
// fallback is correct in Kubernetes too, where the hostname is already the pod
// name; outside Kubernetes it is the machine name, which is still the right
// grain for "which process emitted this".
func serviceInstanceID() string {
	if pod := os.Getenv("POD_NAME"); pod != "" {
		return pod
	}

	if host, err := os.Hostname(); err == nil {
		return host
	}

	return ""
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
