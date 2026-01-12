//go:build integration

package kafka_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/quiqupltd/quiqupgo/fxutil"
	"github.com/quiqupltd/quiqupgo/logger/testutil"
	"github.com/quiqupltd/quiqupgo/kafka"
	tracingtest "github.com/quiqupltd/quiqupgo/tracing/testutil"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// getTestBroker returns the Kafka broker address from env or defaults to OrbStack URL.
func getTestBroker() string {
	if broker := os.Getenv("KAFKA_BROKERS"); broker != "" {
		return broker
	}
	if broker := os.Getenv("KAFKA_BROKER"); broker != "" {
		return broker
	}
	// Use external listener port for OrbStack access
	return "redpanda.quiqupgo.orb.local:19092"
}

// IntegrationTestConfig implements kafka.Config for integration tests.
type IntegrationTestConfig struct {
	brokers       []string
	consumerGroup string
	topic         string
}

func NewIntegrationTestConfig(topic string) *IntegrationTestConfig {
	return &IntegrationTestConfig{
		brokers:       []string{getTestBroker()},
		consumerGroup: fmt.Sprintf("test-group-%s", uuid.New().String()[:8]),
		topic:         topic,
	}
}

func (c *IntegrationTestConfig) GetBrokers() []string              { return c.brokers }
func (c *IntegrationTestConfig) GetConsumerGroup() string          { return c.consumerGroup }
func (c *IntegrationTestConfig) GetProducerTimeout() time.Duration { return 10 * time.Second }
func (c *IntegrationTestConfig) GetConsumerTimeout() time.Duration { return 10 * time.Second }
func (c *IntegrationTestConfig) GetEnableTracing() bool            { return false }
func (c *IntegrationTestConfig) GetTLSEnabled() bool               { return false }
func (c *IntegrationTestConfig) GetTLSCert() string                { return "" }
func (c *IntegrationTestConfig) GetTLSKey() string                 { return "" }
func (c *IntegrationTestConfig) GetTLSCA() string                  { return "" }
func (c *IntegrationTestConfig) GetSASLEnabled() bool              { return false }
func (c *IntegrationTestConfig) GetSASLMechanism() string          { return "" }
func (c *IntegrationTestConfig) GetSASLUsername() string           { return "" }
func (c *IntegrationTestConfig) GetSASLPassword() string           { return "" }

// IntegrationTestModule returns an fx.Option for integration testing with real Kafka.
func IntegrationTestModule(topic string) fx.Option {
	cfg := NewIntegrationTestConfig(topic)

	return fx.Module("kafka-integration-test",
		tracingtest.NoopModule(),
		testutil.NoopModule(),
		fx.Provide(func() kafka.Config {
			return cfg
		}),
		kafka.Module(),
	)
}

// KafkaIntegrationSuite tests the Kafka producer and consumer against Redpanda.
type KafkaIntegrationSuite struct {
	suite.Suite
	topic    string
	producer kafka.Producer
	consumer kafka.Consumer
	app      *fxtest.App
}

func TestKafkaIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(KafkaIntegrationSuite))
}

func (s *KafkaIntegrationSuite) SetupSuite() {
	// Use a unique topic for each test run
	// Redpanda auto-creates topics on first write
	s.topic = fmt.Sprintf("test-suite-%s", uuid.New().String()[:8])
}

func (s *KafkaIntegrationSuite) TearDownSuite() {
	// Topics auto-delete is handled by Redpanda retention policies
}

func (s *KafkaIntegrationSuite) SetupTest() {
	s.app = fxutil.TestApp(s.T(),
		IntegrationTestModule(s.topic),
		fx.Populate(&s.producer, &s.consumer),
	)
	s.app.RequireStart()
}

func (s *KafkaIntegrationSuite) TearDownTest() {
	s.app.RequireStop()
}

func (s *KafkaIntegrationSuite) TestPublishSingleMessage() {
	ctx := context.Background()

	err := s.producer.Publish(ctx, s.topic, []byte("test-key"), []byte("test-value"))
	s.Require().NoError(err)
}

func (s *KafkaIntegrationSuite) TestPublishBatchMessages() {
	ctx := context.Background()

	messages := []kafka.Message{
		{Key: []byte("batch-key-1"), Value: []byte("batch-value-1")},
		{Key: []byte("batch-key-2"), Value: []byte("batch-value-2")},
		{Key: []byte("batch-key-3"), Value: []byte("batch-value-3")},
	}

	err := s.producer.PublishBatch(ctx, s.topic, messages)
	s.Require().NoError(err)
}

func (s *KafkaIntegrationSuite) TestConsumeMessages() {
	ctx := context.Background()

	// Publish test messages
	testMessages := []string{"consume-msg-1", "consume-msg-2", "consume-msg-3"}
	for i, msg := range testMessages {
		err := s.producer.Publish(ctx, s.topic, []byte(fmt.Sprintf("consume-key-%d", i)), []byte(msg))
		s.Require().NoError(err)
	}

	// Give Redpanda a moment to commit
	time.Sleep(500 * time.Millisecond)

	// Consume messages
	consumeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	received := make([]string, 0, len(testMessages))
	done := make(chan struct{})

	handler := func(ctx context.Context, msg kafka.ConsumerMessage) error {
		received = append(received, string(msg.Value))
		if len(received) >= len(testMessages) {
			close(done)
		}
		return nil
	}

	go func() {
		_ = s.consumer.Subscribe(consumeCtx, []string{s.topic}, handler)
	}()

	select {
	case <-done:
		s.Len(received, len(testMessages))
	case <-consumeCtx.Done():
		s.Failf("timed out", "received %d/%d messages", len(received), len(testMessages))
	}
}

func (s *KafkaIntegrationSuite) TestMessageWithHeadersRoundTrip() {
	ctx := context.Background()

	// Use a unique topic for this test to avoid interference
	// Redpanda auto-creates topics on first write
	topic := fmt.Sprintf("roundtrip-%s", uuid.New().String()[:8])

	testKey := []byte("roundtrip-key")
	testValue := []byte(`{"event":"test","data":"hello"}`)
	correlationID := uuid.New().String()
	testHeaders := map[string]string{
		"content-type":   "application/json",
		"correlation-id": correlationID,
	}

	// Publish with headers
	err := s.producer.PublishBatch(ctx, topic, []kafka.Message{
		{
			Key:     testKey,
			Value:   testValue,
			Headers: testHeaders,
		},
	})
	s.Require().NoError(err)

	// Give Redpanda a moment
	time.Sleep(500 * time.Millisecond)

	// Consume
	consumeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	received := make(chan kafka.ConsumerMessage, 1)

	go func() {
		_ = s.consumer.Subscribe(consumeCtx, []string{topic}, func(ctx context.Context, msg kafka.ConsumerMessage) error {
			received <- msg
			return nil
		})
	}()

	select {
	case msg := <-received:
		s.Equal(testKey, msg.Key)
		s.Equal(testValue, msg.Value)
		s.Equal("application/json", msg.Headers["content-type"])
		s.Equal(correlationID, msg.Headers["correlation-id"])
	case <-consumeCtx.Done():
		s.Fail("timed out waiting for message")
	}
}

// TracingIntegrationSuite tests the Kafka producer and consumer with tracing enabled.
type TracingIntegrationSuite struct {
	suite.Suite
	topic    string
	producer kafka.Producer
	consumer kafka.Consumer
	app      *fxtest.App
}

func TestTracingIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TracingIntegrationSuite))
}

// TracingTestConfig implements kafka.Config with tracing enabled.
type TracingTestConfig struct {
	brokers       []string
	consumerGroup string
}

func (c *TracingTestConfig) GetBrokers() []string              { return c.brokers }
func (c *TracingTestConfig) GetConsumerGroup() string          { return c.consumerGroup }
func (c *TracingTestConfig) GetProducerTimeout() time.Duration { return 10 * time.Second }
func (c *TracingTestConfig) GetConsumerTimeout() time.Duration { return 10 * time.Second }
func (c *TracingTestConfig) GetEnableTracing() bool            { return true }
func (c *TracingTestConfig) GetTLSEnabled() bool               { return false }
func (c *TracingTestConfig) GetTLSCert() string                { return "" }
func (c *TracingTestConfig) GetTLSKey() string                 { return "" }
func (c *TracingTestConfig) GetTLSCA() string                  { return "" }
func (c *TracingTestConfig) GetSASLEnabled() bool              { return false }
func (c *TracingTestConfig) GetSASLMechanism() string          { return "" }
func (c *TracingTestConfig) GetSASLUsername() string           { return "" }
func (c *TracingTestConfig) GetSASLPassword() string           { return "" }

// TracingIntegrationTestModule returns an fx.Option for tracing integration testing.
func TracingIntegrationTestModule(topic string) fx.Option {
	cfg := &TracingTestConfig{
		brokers:       []string{getTestBroker()},
		consumerGroup: fmt.Sprintf("tracing-test-group-%s", uuid.New().String()[:8]),
	}

	return fx.Module("kafka-tracing-integration-test",
		tracingtest.NoopModule(),
		testutil.NoopModule(),
		fx.Provide(func() kafka.Config {
			return cfg
		}),
		kafka.Module(),
	)
}

func (s *TracingIntegrationSuite) SetupSuite() {
	s.topic = fmt.Sprintf("tracing-test-%s", uuid.New().String()[:8])
}

func (s *TracingIntegrationSuite) TearDownSuite() {}

func (s *TracingIntegrationSuite) SetupTest() {
	s.app = fxutil.TestApp(s.T(),
		TracingIntegrationTestModule(s.topic),
		fx.Populate(&s.producer, &s.consumer),
	)
	s.app.RequireStart()
}

func (s *TracingIntegrationSuite) TearDownTest() {
	s.app.RequireStop()
}

func (s *TracingIntegrationSuite) TestPublishWithTracing() {
	ctx := context.Background()

	// Publish a message - this exercises the tracing code path
	err := s.producer.Publish(ctx, s.topic, []byte("tracing-key"), []byte("tracing-value"))
	s.Require().NoError(err)
}

func (s *TracingIntegrationSuite) TestPublishBatchWithTracing() {
	ctx := context.Background()

	messages := []kafka.Message{
		{Key: []byte("batch-key-1"), Value: []byte("batch-value-1")},
		{Key: []byte("batch-key-2"), Value: []byte("batch-value-2")},
	}

	err := s.producer.PublishBatch(ctx, s.topic, messages)
	s.Require().NoError(err)
}

func (s *TracingIntegrationSuite) TestConsumeWithTracing() {
	ctx := context.Background()
	topic := fmt.Sprintf("tracing-consume-%s", uuid.New().String()[:8])

	// Publish test message
	err := s.producer.Publish(ctx, topic, []byte("trace-key"), []byte("trace-value"))
	s.Require().NoError(err)

	time.Sleep(500 * time.Millisecond)

	consumeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	received := make(chan kafka.ConsumerMessage, 1)

	go func() {
		_ = s.consumer.Subscribe(consumeCtx, []string{topic}, func(ctx context.Context, msg kafka.ConsumerMessage) error {
			received <- msg
			return nil
		})
	}()

	select {
	case msg := <-received:
		s.Equal([]byte("trace-key"), msg.Key)
		s.Equal([]byte("trace-value"), msg.Value)
	case <-consumeCtx.Done():
		s.Fail("timed out waiting for message")
	}
}

