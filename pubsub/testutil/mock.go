// Package testutil provides testing utilities for the pubsub module.
package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/quiqupltd/quiqupgo/pubsub"
	"go.uber.org/fx"
)

// NoopConfig is a test configuration for the pubsub module.
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
var _ pubsub.Config = (*NoopConfig)(nil)

// InMemoryPubSub is an in-memory implementation of Producer and Consumer for testing.
type InMemoryPubSub struct {
	mu          sync.RWMutex
	topics      map[string][]pubsub.Message
	subscribers map[string][]chan pubsub.ConsumerMessage
}

// NewInMemoryPubSub creates a new in-memory pubsub.
func NewInMemoryPubSub() *InMemoryPubSub {
	return &InMemoryPubSub{
		topics:      make(map[string][]pubsub.Message),
		subscribers: make(map[string][]chan pubsub.ConsumerMessage),
	}
}

// Publish sends a message to the in-memory topic.
func (p *InMemoryPubSub) Publish(ctx context.Context, topic string, key, value []byte) error {
	return p.PublishBatch(ctx, topic, []pubsub.Message{{Key: key, Value: value}})
}

// PublishBatch sends multiple messages to the in-memory topic.
func (p *InMemoryPubSub) PublishBatch(ctx context.Context, topic string, messages []pubsub.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Store messages
	p.topics[topic] = append(p.topics[topic], messages...)

	// Notify subscribers
	if subs, ok := p.subscribers[topic]; ok {
		for _, sub := range subs {
			for i, msg := range messages {
				select {
				case sub <- pubsub.ConsumerMessage{
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
func (p *InMemoryPubSub) Subscribe(ctx context.Context, topics []string, handler pubsub.MessageHandler) error {
	ch := make(chan pubsub.ConsumerMessage, 100)

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

// Close closes the in-memory pubsub.
func (p *InMemoryPubSub) Close() error {
	return nil
}

// GetMessages returns all messages for a topic.
func (p *InMemoryPubSub) GetMessages(topic string) []pubsub.Message {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]pubsub.Message(nil), p.topics[topic]...)
}

// Clear clears all messages.
func (p *InMemoryPubSub) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.topics = make(map[string][]pubsub.Message)
}

// Ensure InMemoryPubSub implements Producer and Consumer.
var _ pubsub.Producer = (*InMemoryPubSub)(nil)
var _ pubsub.Consumer = (*InMemoryPubSub)(nil)

// TestModule returns an fx.Option that provides an in-memory pubsub.
// Both Producer and Consumer are provided by the same InMemoryPubSub instance.
//
// Usage:
//
//	app := fx.New(
//	    pubsub_testutil.TestModule(),
//	    // ... other modules that depend on pubsub.Producer/Consumer
//	)
func TestModule() fx.Option {
	return fx.Module("pubsub-test",
		fx.Provide(func() *InMemoryPubSub {
			return NewInMemoryPubSub()
		}),
		fx.Provide(func(p *InMemoryPubSub) pubsub.Producer { return p }),
		fx.Provide(func(p *InMemoryPubSub) pubsub.Consumer { return p }),
	)
}
