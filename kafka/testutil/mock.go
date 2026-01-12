// Package testutil provides testing utilities for the kafka module.
package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/quiqupltd/quiqupgo/kafka"
	"go.uber.org/fx"
)

// NoopConfig is a test configuration for the kafka module.
type NoopConfig struct {
	brokers         []string
	consumerGroup   string
	producerTimeout time.Duration
	consumerTimeout time.Duration
	enableTracing   bool
}

// NewNoopConfig creates a NoopConfig with test defaults.
func NewNoopConfig() *NoopConfig {
	return &NoopConfig{
		brokers:         []string{"localhost:9092"},
		consumerGroup:   "test-group",
		producerTimeout: 10 * time.Second,
		consumerTimeout: 10 * time.Second,
		enableTracing:   false,
	}
}

func (c *NoopConfig) GetBrokers() []string              { return c.brokers }
func (c *NoopConfig) GetConsumerGroup() string          { return c.consumerGroup }
func (c *NoopConfig) GetProducerTimeout() time.Duration { return c.producerTimeout }
func (c *NoopConfig) GetConsumerTimeout() time.Duration { return c.consumerTimeout }
func (c *NoopConfig) GetEnableTracing() bool            { return c.enableTracing }
func (c *NoopConfig) GetTLSEnabled() bool               { return false }
func (c *NoopConfig) GetTLSCert() string                { return "" }
func (c *NoopConfig) GetTLSKey() string                 { return "" }
func (c *NoopConfig) GetTLSCA() string                  { return "" }
func (c *NoopConfig) GetSASLEnabled() bool              { return false }
func (c *NoopConfig) GetSASLMechanism() string          { return "PLAIN" }
func (c *NoopConfig) GetSASLUsername() string           { return "" }
func (c *NoopConfig) GetSASLPassword() string           { return "" }

// Ensure NoopConfig implements Config.
var _ kafka.Config = (*NoopConfig)(nil)

// InMemoryKafka is an in-memory implementation of Producer and Consumer for testing.
type InMemoryKafka struct {
	mu          sync.RWMutex
	topics      map[string][]kafka.Message
	subscribers map[string][]chan kafka.ConsumerMessage
}

// NewInMemoryKafka creates a new in-memory kafka.
func NewInMemoryKafka() *InMemoryKafka {
	return &InMemoryKafka{
		topics:      make(map[string][]kafka.Message),
		subscribers: make(map[string][]chan kafka.ConsumerMessage),
	}
}

// Publish sends a message to the in-memory topic.
func (p *InMemoryKafka) Publish(ctx context.Context, topic string, key, value []byte) error {
	return p.PublishBatch(ctx, topic, []kafka.Message{{Key: key, Value: value}})
}

// PublishBatch sends multiple messages to the in-memory topic.
func (p *InMemoryKafka) PublishBatch(ctx context.Context, topic string, messages []kafka.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Store messages
	p.topics[topic] = append(p.topics[topic], messages...)

	// Notify subscribers
	if subs, ok := p.subscribers[topic]; ok {
		for _, sub := range subs {
			for i, msg := range messages {
				select {
				case sub <- kafka.ConsumerMessage{
					Topic:   topic,
					Offset:  int64(len(p.topics[topic]) - len(messages) + i),
					Key:     msg.Key,
					Value:   msg.Value,
					Headers: msg.Headers,
				}:
				default:
					// Channel full, skip
				}
			}
		}
	}

	return nil
}

// Subscribe subscribes to the specified topics.
func (p *InMemoryKafka) Subscribe(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
	ch := make(chan kafka.ConsumerMessage, 100)

	p.mu.Lock()
	for _, topic := range topics {
		p.subscribers[topic] = append(p.subscribers[topic], ch)
	}
	p.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-ch:
			if err := handler(ctx, msg); err != nil {
				// Log error but continue
				continue
			}
		}
	}
}

// Close closes the in-memory kafka.
func (p *InMemoryKafka) Close() error {
	return nil
}

// GetMessages returns all messages for a topic.
func (p *InMemoryKafka) GetMessages(topic string) []kafka.Message {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]kafka.Message(nil), p.topics[topic]...)
}

// Clear clears all messages.
func (p *InMemoryKafka) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.topics = make(map[string][]kafka.Message)
}

// Ensure InMemoryKafka implements Producer and Consumer.
var _ kafka.Producer = (*InMemoryKafka)(nil)
var _ kafka.Consumer = (*InMemoryKafka)(nil)

// TestModule returns an fx.Option that provides an in-memory kafka.
// Both Producer and Consumer are provided by the same InMemoryKafka instance.
//
// Usage:
//
//	app := fx.New(
//	    kafka_testutil.TestModule(),
//	    // ... other modules that depend on kafka.Producer/Consumer
//	)
func TestModule() fx.Option {
	return fx.Module("kafka-test",
		fx.Provide(func() *InMemoryKafka {
			return NewInMemoryKafka()
		}),
		fx.Provide(func(p *InMemoryKafka) kafka.Producer { return p }),
		fx.Provide(func(p *InMemoryKafka) kafka.Consumer { return p }),
	)
}
