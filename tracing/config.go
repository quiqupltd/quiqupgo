package tracing

// Config defines the configuration interface for the tracing module.
// Implement this interface in your application to provide configuration.
type Config interface {
	// GetServiceName returns the name of the service for tracing identification.
	GetServiceName() string

	// GetEnvironmentName returns the deployment environment (e.g., "production", "staging", "development").
	GetEnvironmentName() string

	// GetOTLPEndpoint returns the OTLP collector HTTP endpoint (e.g., "otel-collector:4318").
	// Return empty string to disable tracing export.
	GetOTLPEndpoint() string

	// GetOTLPInsecure returns true to use HTTP instead of HTTPS for OTLP export.
	GetOTLPInsecure() bool

	// GetOTLPTLSCert returns the base64-encoded TLS certificate for OTLP export.
	// Return empty string if not using TLS or using system certificates.
	GetOTLPTLSCert() string

	// GetOTLPTLSKey returns the base64-encoded TLS key for OTLP export.
	// Return empty string if not using TLS or using system certificates.
	GetOTLPTLSKey() string

	// GetOTLPTLSCA returns the base64-encoded TLS CA certificate for OTLP export.
	// Return empty string if not using TLS or using system certificates.
	GetOTLPTLSCA() string
}

// StandardConfig is the default implementation of Config.
// Use this in your application if you don't need custom configuration logic.
type StandardConfig struct {
	ServiceName     string
	EnvironmentName string
	OTLPEndpoint    string
	OTLPInsecure    bool
	OTLPTLSCert     string
	OTLPTLSKey      string
	OTLPTLSCA       string
}

// GetServiceName returns the service name.
func (c *StandardConfig) GetServiceName() string {
	return c.ServiceName
}

// GetEnvironmentName returns the environment name.
func (c *StandardConfig) GetEnvironmentName() string {
	return c.EnvironmentName
}

// GetOTLPEndpoint returns the OTLP endpoint.
func (c *StandardConfig) GetOTLPEndpoint() string {
	return c.OTLPEndpoint
}

// GetOTLPInsecure returns whether to use insecure connection.
func (c *StandardConfig) GetOTLPInsecure() bool {
	return c.OTLPInsecure
}

// GetOTLPTLSCert returns the TLS certificate.
func (c *StandardConfig) GetOTLPTLSCert() string {
	return c.OTLPTLSCert
}

// GetOTLPTLSKey returns the TLS key.
func (c *StandardConfig) GetOTLPTLSKey() string {
	return c.OTLPTLSKey
}

// GetOTLPTLSCA returns the TLS CA certificate.
func (c *StandardConfig) GetOTLPTLSCA() string {
	return c.OTLPTLSCA
}

// Ensure StandardConfig implements Config.
var _ Config = (*StandardConfig)(nil)
