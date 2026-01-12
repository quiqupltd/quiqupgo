package logger

// Config defines the configuration interface for the logger module.
// Implement this interface in your application to provide configuration.
type Config interface {
	// GetServiceName returns the name of the service for log identification.
	GetServiceName() string

	// GetEnvironment returns the deployment environment.
	// Use "development" or "local" for human-readable console output.
	// Any other value results in JSON structured logging for production.
	GetEnvironment() string
}

// StandardConfig is the default implementation of Config.
// Use this in your application if you don't need custom configuration logic.
type StandardConfig struct {
	ServiceName string
	Environment string
}

// GetServiceName returns the service name.
func (c *StandardConfig) GetServiceName() string {
	return c.ServiceName
}

// GetEnvironment returns the environment.
func (c *StandardConfig) GetEnvironment() string {
	return c.Environment
}

// IsDevelopment returns true if the environment is a development environment.
func (c *StandardConfig) IsDevelopment() bool {
	env := c.GetEnvironment()
	return env == "development" || env == "local" || env == "dev"
}

// Ensure StandardConfig implements Config.
var _ Config = (*StandardConfig)(nil)
