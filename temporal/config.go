package temporal

// Config defines the configuration interface for the temporal module.
// Implement this interface in your application to provide configuration.
type Config interface {
	// GetHostPort returns the Temporal server host:port (e.g., "localhost:7233").
	GetHostPort() string

	// GetNamespace returns the Temporal namespace to use.
	GetNamespace() string

	// GetTLSCert returns the PEM-encoded TLS certificate for mTLS.
	// Return empty string if not using TLS.
	GetTLSCert() string

	// GetTLSKey returns the PEM-encoded TLS key for mTLS.
	// Return empty string if not using TLS.
	GetTLSKey() string
}

// StandardConfig is the default implementation of Config.
// Use this in your application if you don't need custom configuration logic.
type StandardConfig struct {
	HostPort  string
	Namespace string
	TLSCert   string
	TLSKey    string
}

// GetHostPort returns the Temporal host:port.
func (c *StandardConfig) GetHostPort() string {
	if c.HostPort == "" {
		return "localhost:7233"
	}
	return c.HostPort
}

// GetNamespace returns the Temporal namespace.
func (c *StandardConfig) GetNamespace() string {
	if c.Namespace == "" {
		return "default"
	}
	return c.Namespace
}

// GetTLSCert returns the TLS certificate.
func (c *StandardConfig) GetTLSCert() string {
	return c.TLSCert
}

// GetTLSKey returns the TLS key.
func (c *StandardConfig) GetTLSKey() string {
	return c.TLSKey
}

// IsLocal returns true if connecting to localhost.
func (c *StandardConfig) IsLocal() bool {
	return c.HostPort == "" || c.HostPort == "localhost:7233"
}

// Ensure StandardConfig implements Config.
var _ Config = (*StandardConfig)(nil)
