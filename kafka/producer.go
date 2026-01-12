package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Producer is an interface for publishing messages to Kafka.
type Producer interface {
	// Publish sends a message to the specified topic.
	// The context is used for tracing and cancellation.
	Publish(ctx context.Context, topic string, key, value []byte) error

	// PublishBatch sends multiple messages to the specified topic.
	PublishBatch(ctx context.Context, topic string, messages []Message) error

	// Close closes the producer and releases resources.
	Close() error
}

// Message represents a message to be published.
type Message struct {
	Key     []byte
	Value   []byte
	Headers map[string]string
}

// KafkaProducer is a Kafka-based implementation of Producer.
type KafkaProducer struct {
	cfg    Config
	tracer trace.Tracer
	logger *zap.Logger
	dialer *kafka.Dialer
}

// NewProducer creates a new Kafka producer.
func NewProducer(cfg Config, tracer trace.Tracer, logger *zap.Logger) (*KafkaProducer, error) {
	dialer := &kafka.Dialer{
		Timeout: cfg.GetProducerTimeout(),
	}

	// Configure TLS if enabled
	if cfg.GetTLSEnabled() {
		tlsCfg, err := buildTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		dialer.TLS = tlsCfg
	}

	// Configure SASL if enabled
	if cfg.GetSASLEnabled() {
		mechanism, err := buildSASLMechanism(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build SASL mechanism: %w", err)
		}
		dialer.SASLMechanism = mechanism
	}

	return &KafkaProducer{
		cfg:    cfg,
		tracer: tracer,
		logger: logger,
		dialer: dialer,
	}, nil
}

// Publish sends a message to the specified topic.
func (p *KafkaProducer) Publish(ctx context.Context, topic string, key, value []byte) error {
	return p.PublishBatch(ctx, topic, []Message{{Key: key, Value: value}})
}

// PublishBatch sends multiple messages to the specified topic.
func (p *KafkaProducer) PublishBatch(ctx context.Context, topic string, messages []Message) error {
	// Start tracing span if enabled
	if p.cfg.GetEnableTracing() && p.tracer != nil {
		var span trace.Span
		ctx, span = p.tracer.Start(ctx, "kafka.produce",
			trace.WithSpanKind(trace.SpanKindProducer),
			trace.WithAttributes(
				attribute.String("messaging.system", "kafka"),
				attribute.String("messaging.destination", topic),
				attribute.Int("messaging.batch_size", len(messages)),
			),
		)
		defer span.End()
	}

	// Create writer for this batch
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(p.cfg.GetBrokers()...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		Transport: &kafka.Transport{
			TLS:  p.dialer.TLS,
			SASL: p.dialer.SASLMechanism,
		},
	}
	defer func() {
		if err := writer.Close(); err != nil {
			p.logger.Warn("failed to close kafka writer", zap.Error(err))
		}
	}()

	// Convert messages
	kafkaMessages := make([]kafka.Message, len(messages))
	for i, msg := range messages {
		headers := make([]kafka.Header, 0, len(msg.Headers))
		for k, v := range msg.Headers {
			headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
		}

		// Inject trace context into headers if tracing is enabled
		if p.cfg.GetEnableTracing() && p.tracer != nil {
			headers = injectTraceContext(ctx, headers)
		}

		kafkaMessages[i] = kafka.Message{
			Key:     msg.Key,
			Value:   msg.Value,
			Headers: headers,
		}
	}

	if err := writer.WriteMessages(ctx, kafkaMessages...); err != nil {
		p.logger.Error("failed to publish messages",
			zap.String("topic", topic),
			zap.Int("count", len(messages)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish messages: %w", err)
	}

	p.logger.Debug("published messages",
		zap.String("topic", topic),
		zap.Int("count", len(messages)),
	)

	return nil
}

// Close closes the producer.
func (p *KafkaProducer) Close() error {
	// No persistent connections to close in this implementation
	return nil
}

// injectTraceContext injects the trace context into Kafka headers.
func injectTraceContext(ctx context.Context, headers []kafka.Header) []kafka.Header {
	carrier := &kafkaHeaderCarrier{headers: headers}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier.headers
}

// kafkaHeaderCarrier adapts Kafka headers to the propagation.TextMapCarrier interface.
type kafkaHeaderCarrier struct {
	headers []kafka.Header
}

func (c *kafkaHeaderCarrier) Get(key string) string {
	for _, h := range c.headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *kafkaHeaderCarrier) Set(key, value string) {
	for i, h := range c.headers {
		if h.Key == key {
			c.headers[i].Value = []byte(value)
			return
		}
	}
	c.headers = append(c.headers, kafka.Header{Key: key, Value: []byte(value)})
}

func (c *kafkaHeaderCarrier) Keys() []string {
	keys := make([]string, len(c.headers))
	for i, h := range c.headers {
		keys[i] = h.Key
	}
	return keys
}

// buildTLSConfig creates a TLS configuration from the provided config.
func buildTLSConfig(cfg Config) (*tls.Config, error) {
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if cfg.GetTLSCert() != "" && cfg.GetTLSKey() != "" {
		cert, err := tls.X509KeyPair([]byte(cfg.GetTLSCert()), []byte(cfg.GetTLSKey()))
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS key pair: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	if cfg.GetTLSCA() != "" {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(cfg.GetTLSCA())) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsCfg.RootCAs = caCertPool
	}

	return tlsCfg, nil
}

// buildSASLMechanism creates a SASL mechanism from the provided config.
func buildSASLMechanism(cfg Config) (sasl.Mechanism, error) {
	switch cfg.GetSASLMechanism() {
	case "PLAIN":
		return plain.Mechanism{
			Username: cfg.GetSASLUsername(),
			Password: cfg.GetSASLPassword(),
		}, nil
	case "SCRAM-SHA-256":
		mechanism, err := scram.Mechanism(scram.SHA256, cfg.GetSASLUsername(), cfg.GetSASLPassword())
		if err != nil {
			return nil, fmt.Errorf("failed to create SCRAM-SHA-256 mechanism: %w", err)
		}
		return mechanism, nil
	case "SCRAM-SHA-512":
		mechanism, err := scram.Mechanism(scram.SHA512, cfg.GetSASLUsername(), cfg.GetSASLPassword())
		if err != nil {
			return nil, fmt.Errorf("failed to create SCRAM-SHA-512 mechanism: %w", err)
		}
		return mechanism, nil
	default:
		return nil, fmt.Errorf("unsupported SASL mechanism: %s", cfg.GetSASLMechanism())
	}
}

// Ensure KafkaProducer implements Producer.
var _ Producer = (*KafkaProducer)(nil)
