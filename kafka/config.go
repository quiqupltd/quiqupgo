package kafka

import "time"

// Config is the interface that applications must implement to configure the Kafka module.
// Applications can either implement this interface on their own config struct or use
// StandardConfig.
type Config interface {
	// GetBrokers returns the list of Kafka broker addresses.
	GetBrokers() []string

	// GetConsumerGroup returns the consumer group ID for this application.
	GetConsumerGroup() string

	// GetProducerTimeout returns the timeout for producing messages.
	// Return 0 to use the default (10 seconds).
	GetProducerTimeout() time.Duration

	// GetConsumerTimeout returns the timeout for consuming messages.
	// Return 0 to use the default (10 seconds).
	GetConsumerTimeout() time.Duration

	// GetEnableTracing returns whether OpenTelemetry tracing should be enabled.
	GetEnableTracing() bool

	// GetTLSEnabled returns whether TLS should be enabled for Kafka connections.
	GetTLSEnabled() bool

	// GetTLSCert returns the TLS certificate (PEM encoded).
	GetTLSCert() string

	// GetTLSKey returns the TLS private key (PEM encoded).
	GetTLSKey() string

	// GetTLSCA returns the TLS CA certificate (PEM encoded).
	GetTLSCA() string

	// GetSASLEnabled returns whether SASL authentication should be enabled.
	GetSASLEnabled() bool

	// GetSASLMechanism returns the SASL mechanism (e.g., "PLAIN", "SCRAM-SHA-256").
	GetSASLMechanism() string

	// GetSASLUsername returns the SASL username.
	GetSASLUsername() string

	// GetSASLPassword returns the SASL password.
	GetSASLPassword() string
}

// StandardConfig is a standard implementation of Config that applications can use.
type StandardConfig struct {
	// Brokers is the list of Kafka broker addresses.
	Brokers []string

	// ConsumerGroup is the consumer group ID.
	ConsumerGroup string

	// ProducerTimeout is the timeout for producing messages.
	// Defaults to 10 seconds if not set.
	ProducerTimeout time.Duration

	// ConsumerTimeout is the timeout for consuming messages.
	// Defaults to 10 seconds if not set.
	ConsumerTimeout time.Duration

	// EnableTracing enables OpenTelemetry tracing.
	// Defaults to true if not explicitly set.
	EnableTracing *bool

	// TLSEnabled enables TLS for Kafka connections.
	TLSEnabled bool

	// TLSCert is the TLS certificate (PEM encoded).
	TLSCert string

	// TLSKey is the TLS private key (PEM encoded).
	TLSKey string

	// TLSCA is the TLS CA certificate (PEM encoded).
	TLSCA string

	// SASLEnabled enables SASL authentication.
	SASLEnabled bool

	// SASLMechanism is the SASL mechanism (e.g., "PLAIN", "SCRAM-SHA-256").
	SASLMechanism string

	// SASLUsername is the SASL username.
	SASLUsername string

	// SASLPassword is the SASL password.
	SASLPassword string
}

// GetBrokers returns the list of Kafka broker addresses.
func (c *StandardConfig) GetBrokers() []string {
	if len(c.Brokers) == 0 {
		return []string{"localhost:9092"}
	}
	return c.Brokers
}

// GetConsumerGroup returns the consumer group ID.
func (c *StandardConfig) GetConsumerGroup() string {
	if c.ConsumerGroup == "" {
		return "default"
	}
	return c.ConsumerGroup
}

// GetProducerTimeout returns the timeout for producing messages.
func (c *StandardConfig) GetProducerTimeout() time.Duration {
	if c.ProducerTimeout == 0 {
		return 10 * time.Second
	}
	return c.ProducerTimeout
}

// GetConsumerTimeout returns the timeout for consuming messages.
func (c *StandardConfig) GetConsumerTimeout() time.Duration {
	if c.ConsumerTimeout == 0 {
		return 10 * time.Second
	}
	return c.ConsumerTimeout
}

// GetEnableTracing returns whether OpenTelemetry tracing should be enabled.
func (c *StandardConfig) GetEnableTracing() bool {
	if c.EnableTracing == nil {
		return true
	}
	return *c.EnableTracing
}

// GetTLSEnabled returns whether TLS should be enabled.
func (c *StandardConfig) GetTLSEnabled() bool {
	return c.TLSEnabled
}

// GetTLSCert returns the TLS certificate.
func (c *StandardConfig) GetTLSCert() string {
	return c.TLSCert
}

// GetTLSKey returns the TLS private key.
func (c *StandardConfig) GetTLSKey() string {
	return c.TLSKey
}

// GetTLSCA returns the TLS CA certificate.
func (c *StandardConfig) GetTLSCA() string {
	return c.TLSCA
}

// GetSASLEnabled returns whether SASL authentication should be enabled.
func (c *StandardConfig) GetSASLEnabled() bool {
	return c.SASLEnabled
}

// GetSASLMechanism returns the SASL mechanism.
func (c *StandardConfig) GetSASLMechanism() string {
	if c.SASLMechanism == "" {
		return "PLAIN"
	}
	return c.SASLMechanism
}

// GetSASLUsername returns the SASL username.
func (c *StandardConfig) GetSASLUsername() string {
	return c.SASLUsername
}

// GetSASLPassword returns the SASL password.
func (c *StandardConfig) GetSASLPassword() string {
	return c.SASLPassword
}

// Ensure StandardConfig implements Config.
var _ Config = (*StandardConfig)(nil)
