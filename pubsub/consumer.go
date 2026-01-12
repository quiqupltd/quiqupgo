package pubsub

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ConsumerMessage represents a received message.
type ConsumerMessage struct {
	Topic     string
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
	Headers   map[string]string
}

// MessageHandler is a function that handles a consumed message.
// Return an error to indicate processing failure (message will not be committed).
type MessageHandler func(ctx context.Context, msg ConsumerMessage) error

// Consumer is an interface for consuming messages from Kafka.
type Consumer interface {
	// Subscribe subscribes to the specified topics.
	Subscribe(ctx context.Context, topics []string, handler MessageHandler) error

	// Close closes the consumer and releases resources.
	Close() error
}

// KafkaConsumer is a Kafka-based implementation of Consumer.
type KafkaConsumer struct {
	cfg     Config
	tracer  trace.Tracer
	logger  *zap.Logger
	readers []*kafka.Reader
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(cfg Config, tracer trace.Tracer, logger *zap.Logger) (*KafkaConsumer, error) {
	return &KafkaConsumer{
		cfg:     cfg,
		tracer:  tracer,
		logger:  logger,
		readers: make([]*kafka.Reader, 0),
	}, nil
}

// Subscribe subscribes to the specified topics and calls the handler for each message.
// This method blocks until the context is cancelled or an error occurs.
func (c *KafkaConsumer) Subscribe(ctx context.Context, topics []string, handler MessageHandler) error {
	// Create a reader for each topic
	for _, topic := range topics {
		reader := c.createReader(topic)
		c.readers = append(c.readers, reader)

		// Start consuming in a goroutine
		go c.consumeTopic(ctx, reader, topic, handler)
	}

	// Wait for context cancellation
	<-ctx.Done()
	return ctx.Err()
}

// createReader creates a Kafka reader for the specified topic.
func (c *KafkaConsumer) createReader(topic string) *kafka.Reader {
	dialer := &kafka.Dialer{
		Timeout: c.cfg.GetConsumerTimeout(),
	}

	// Configure TLS if enabled
	if c.cfg.GetTLSEnabled() {
		tlsCfg, err := buildTLSConfig(c.cfg)
		if err != nil {
			c.logger.Error("failed to build TLS config", zap.Error(err))
		} else {
			dialer.TLS = tlsCfg
		}
	}

	// Configure SASL if enabled
	if c.cfg.GetSASLEnabled() {
		mechanism, err := buildSASLMechanism(c.cfg)
		if err != nil {
			c.logger.Error("failed to build SASL mechanism", zap.Error(err))
		} else {
			dialer.SASLMechanism = mechanism
		}
	}

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.cfg.GetBrokers(),
		GroupID: c.cfg.GetConsumerGroup(),
		Topic:   topic,
		Dialer:  dialer,
	})
}

// consumeTopic consumes messages from a single topic.
func (c *KafkaConsumer) consumeTopic(ctx context.Context, reader *kafka.Reader, topic string, handler MessageHandler) {
	c.logger.Info("starting consumer", zap.String("topic", topic), zap.String("group", c.cfg.GetConsumerGroup()))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("stopping consumer", zap.String("topic", topic))
			return
		default:
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) || ctx.Err() != nil {
					return
				}
				c.logger.Error("failed to fetch message",
					zap.String("topic", topic),
					zap.Error(err),
				)
				continue
			}

			// Process the message
			if err := c.processMessage(ctx, reader, msg, handler); err != nil {
				c.logger.Error("failed to process message",
					zap.String("topic", topic),
					zap.Int64("offset", msg.Offset),
					zap.Error(err),
				)
				// Don't commit failed messages
				continue
			}

			// Commit the message
			if err := reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("failed to commit message",
					zap.String("topic", topic),
					zap.Int64("offset", msg.Offset),
					zap.Error(err),
				)
			}
		}
	}
}

// processMessage processes a single message with tracing.
func (c *KafkaConsumer) processMessage(ctx context.Context, reader *kafka.Reader, msg kafka.Message, handler MessageHandler) error {
	// Extract trace context from headers if tracing is enabled
	if c.cfg.GetEnableTracing() && c.tracer != nil {
		ctx = extractTraceContext(ctx, msg.Headers)

		var span trace.Span
		ctx, span = c.tracer.Start(ctx, "kafka.consume",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				attribute.String("messaging.system", "kafka"),
				attribute.String("messaging.destination", msg.Topic),
				attribute.Int("messaging.partition", msg.Partition),
				attribute.Int64("messaging.offset", msg.Offset),
			),
		)
		defer span.End()
	}

	// Convert headers
	headers := make(map[string]string)
	for _, h := range msg.Headers {
		headers[h.Key] = string(h.Value)
	}

	// Call the handler
	consumerMsg := ConsumerMessage{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Key:       msg.Key,
		Value:     msg.Value,
		Headers:   headers,
	}

	return handler(ctx, consumerMsg)
}

// Close closes all readers.
func (c *KafkaConsumer) Close() error {
	var errs []error
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close some readers: %v", errs)
	}
	return nil
}

// extractTraceContext extracts the trace context from Kafka headers.
func extractTraceContext(ctx context.Context, headers []kafka.Header) context.Context {
	carrier := &kafkaHeaderCarrier{headers: headers}
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

// Ensure KafkaConsumer implements Consumer.
var _ Consumer = (*KafkaConsumer)(nil)
