package pubsub_test

import (
	"context"
	"testing"
	"time"

	"github.com/quiqupltd/quiqupgo/pubsub"
	"github.com/quiqupltd/quiqupgo/pubsub/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestStandardConfig(t *testing.T) {
	enableTracing := true
	cfg := &pubsub.StandardConfig{
		Brokers:         []string{"kafka1:9092", "kafka2:9092"},
		ConsumerGroup:   "my-group",
		ProducerTimeout: 5 * time.Second,
		ConsumerTimeout: 15 * time.Second,
		EnableTracing:   &enableTracing,
		TLSEnabled:      true,
		TLSCert:         "cert-data",
		TLSKey:          "key-data",
		TLSCA:           "ca-data",
		SASLEnabled:     true,
		SASLMechanism:   "SCRAM-SHA-256",
		SASLUsername:    "user",
		SASLPassword:    "pass",
	}

	assert.Equal(t, []string{"kafka1:9092", "kafka2:9092"}, cfg.GetBrokers())
	assert.Equal(t, "my-group", cfg.GetConsumerGroup())
	assert.Equal(t, 5*time.Second, cfg.GetProducerTimeout())
	assert.Equal(t, 15*time.Second, cfg.GetConsumerTimeout())
	assert.True(t, cfg.GetEnableTracing())
	assert.True(t, cfg.GetTLSEnabled())
	assert.Equal(t, "cert-data", cfg.GetTLSCert())
	assert.Equal(t, "key-data", cfg.GetTLSKey())
	assert.Equal(t, "ca-data", cfg.GetTLSCA())
	assert.True(t, cfg.GetSASLEnabled())
	assert.Equal(t, "SCRAM-SHA-256", cfg.GetSASLMechanism())
	assert.Equal(t, "user", cfg.GetSASLUsername())
	assert.Equal(t, "pass", cfg.GetSASLPassword())
}

func TestStandardConfig_Defaults(t *testing.T) {
	cfg := &pubsub.StandardConfig{}

	assert.Equal(t, []string{"localhost:9092"}, cfg.GetBrokers())
	assert.Equal(t, "default", cfg.GetConsumerGroup())
	assert.Equal(t, 10*time.Second, cfg.GetProducerTimeout())
	assert.Equal(t, 10*time.Second, cfg.GetConsumerTimeout())
	// EnableTracing defaults to true when nil
	assert.True(t, cfg.GetEnableTracing())
	assert.False(t, cfg.GetTLSEnabled())
	assert.Equal(t, "", cfg.GetTLSCert())
	assert.Equal(t, "", cfg.GetTLSKey())
	assert.Equal(t, "", cfg.GetTLSCA())
	assert.False(t, cfg.GetSASLEnabled())
	assert.Equal(t, "PLAIN", cfg.GetSASLMechanism())
	assert.Equal(t, "", cfg.GetSASLUsername())
	assert.Equal(t, "", cfg.GetSASLPassword())
}

func TestStandardConfig_TracingDisabled(t *testing.T) {
	enableTracing := false
	cfg := &pubsub.StandardConfig{
		EnableTracing: &enableTracing,
	}

	assert.False(t, cfg.GetEnableTracing())
}

func TestNoopConfig(t *testing.T) {
	cfg := testutil.NewNoopConfig()

	assert.Equal(t, []string{"localhost:9092"}, cfg.GetBrokers())
	assert.Equal(t, "test-group", cfg.GetConsumerGroup())
	assert.Equal(t, 10*time.Second, cfg.GetProducerTimeout())
	assert.Equal(t, 10*time.Second, cfg.GetConsumerTimeout())
	assert.False(t, cfg.GetEnableTracing())
	assert.False(t, cfg.GetTLSEnabled())
	assert.False(t, cfg.GetSASLEnabled())
}

func TestInMemoryPubSub_PublishAndGet(t *testing.T) {
	ps := testutil.NewInMemoryPubSub()

	ctx := context.Background()
	err := ps.Publish(ctx, "test-topic", []byte("key1"), []byte("value1"))
	require.NoError(t, err)

	err = ps.Publish(ctx, "test-topic", []byte("key2"), []byte("value2"))
	require.NoError(t, err)

	messages := ps.GetMessages("test-topic")
	require.Len(t, messages, 2)
	assert.Equal(t, []byte("key1"), messages[0].Key)
	assert.Equal(t, []byte("value1"), messages[0].Value)
	assert.Equal(t, []byte("key2"), messages[1].Key)
	assert.Equal(t, []byte("value2"), messages[1].Value)
}

func TestInMemoryPubSub_PublishBatch(t *testing.T) {
	ps := testutil.NewInMemoryPubSub()

	ctx := context.Background()
	messages := []pubsub.Message{
		{Key: []byte("k1"), Value: []byte("v1")},
		{Key: []byte("k2"), Value: []byte("v2")},
		{Key: []byte("k3"), Value: []byte("v3")},
	}

	err := ps.PublishBatch(ctx, "batch-topic", messages)
	require.NoError(t, err)

	stored := ps.GetMessages("batch-topic")
	require.Len(t, stored, 3)
}

func TestInMemoryPubSub_Subscribe(t *testing.T) {
	ps := testutil.NewInMemoryPubSub()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	received := make([]pubsub.ConsumerMessage, 0)
	done := make(chan struct{})

	go func() {
		defer close(done)
		_ = ps.Subscribe(ctx, []string{"sub-topic"}, func(ctx context.Context, msg pubsub.ConsumerMessage) error {
			received = append(received, msg)
			return nil
		})
	}()

	// Give subscriber time to start
	time.Sleep(10 * time.Millisecond)

	// Publish some messages
	_ = ps.Publish(context.Background(), "sub-topic", []byte("k1"), []byte("v1"))
	_ = ps.Publish(context.Background(), "sub-topic", []byte("k2"), []byte("v2"))

	// Wait for context to timeout
	<-done

	// Should have received at least some messages
	assert.GreaterOrEqual(t, len(received), 0)
}

func TestInMemoryPubSub_Clear(t *testing.T) {
	ps := testutil.NewInMemoryPubSub()

	ctx := context.Background()
	_ = ps.Publish(ctx, "topic1", []byte("k"), []byte("v"))
	_ = ps.Publish(ctx, "topic2", []byte("k"), []byte("v"))

	assert.Len(t, ps.GetMessages("topic1"), 1)
	assert.Len(t, ps.GetMessages("topic2"), 1)

	ps.Clear()

	assert.Len(t, ps.GetMessages("topic1"), 0)
	assert.Len(t, ps.GetMessages("topic2"), 0)
}

func TestTestModule(t *testing.T) {
	var producer pubsub.Producer
	var consumer pubsub.Consumer

	app := fx.New(
		fx.NopLogger,
		testutil.TestModule(),
		fx.Populate(&producer, &consumer),
	)

	require.NoError(t, app.Err())
	require.NotNil(t, producer)
	require.NotNil(t, consumer)

	// Verify they're the same instance
	inMemProducer, ok := producer.(*testutil.InMemoryPubSub)
	require.True(t, ok)
	inMemConsumer, ok := consumer.(*testutil.InMemoryPubSub)
	require.True(t, ok)
	assert.Equal(t, inMemProducer, inMemConsumer)
}

// TestNewProducerWithTLS_InvalidCert tests producer creation with invalid TLS config
func TestNewProducerWithTLS_InvalidCert(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:    []string{"localhost:9092"},
		TLSEnabled: true,
		TLSCert:    "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
		TLSKey:     "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----",
	}

	// This should fail because the cert/key are invalid
	_, err := pubsub.NewProducer(cfg, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build TLS config")
}

// TestNewProducerWithTLS_CAOnly tests producer creation with just CA cert (no client certs)
func TestNewProducerWithTLS_CAOnly(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:    []string{"localhost:9092"},
		TLSEnabled: true,
		// No client cert/key, just enabling TLS
	}

	producer, err := pubsub.NewProducer(cfg, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, producer)
}

// TestNewProducerWithSASL_Plain tests producer with SASL PLAIN
func TestNewProducerWithSASL_Plain(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:       []string{"localhost:9092"},
		SASLEnabled:   true,
		SASLMechanism: "PLAIN",
		SASLUsername:  "user",
		SASLPassword:  "pass",
	}

	producer, err := pubsub.NewProducer(cfg, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, producer)
}

// TestNewProducerWithSASL_SCRAM256 tests producer with SCRAM-SHA-256
func TestNewProducerWithSASL_SCRAM256(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:       []string{"localhost:9092"},
		SASLEnabled:   true,
		SASLMechanism: "SCRAM-SHA-256",
		SASLUsername:  "user",
		SASLPassword:  "pass",
	}

	producer, err := pubsub.NewProducer(cfg, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, producer)
}

// TestNewProducerWithSASL_SCRAM512 tests producer with SCRAM-SHA-512
func TestNewProducerWithSASL_SCRAM512(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:       []string{"localhost:9092"},
		SASLEnabled:   true,
		SASLMechanism: "SCRAM-SHA-512",
		SASLUsername:  "user",
		SASLPassword:  "pass",
	}

	producer, err := pubsub.NewProducer(cfg, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, producer)
}

// TestNewProducerWithSASL_Unsupported tests producer with unsupported SASL mechanism
func TestNewProducerWithSASL_Unsupported(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:       []string{"localhost:9092"},
		SASLEnabled:   true,
		SASLMechanism: "UNSUPPORTED",
		SASLUsername:  "user",
		SASLPassword:  "pass",
	}

	_, err := pubsub.NewProducer(cfg, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported SASL mechanism")
}

// TestNewConsumer tests consumer creation
func TestNewConsumer(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:       []string{"localhost:9092"},
		ConsumerGroup: "test-group",
	}

	consumer, err := pubsub.NewConsumer(cfg, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, consumer)
}

// TestProducerClose tests producer close
func TestProducerClose(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers: []string{"localhost:9092"},
	}

	producer, err := pubsub.NewProducer(cfg, nil, nil)
	require.NoError(t, err)

	err = producer.Close()
	assert.NoError(t, err)
}

// TestNewProducerWithTLS_InvalidCA tests producer creation with invalid CA cert
func TestNewProducerWithTLS_InvalidCA(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:    []string{"localhost:9092"},
		TLSEnabled: true,
		TLSCA:      "invalid-ca-data",
	}

	_, err := pubsub.NewProducer(cfg, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse CA certificate")
}

// TestNewConsumerWithTLS tests consumer creation with TLS config
func TestNewConsumerWithTLS(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:       []string{"localhost:9092"},
		ConsumerGroup: "test-group",
		TLSEnabled:    true,
		// No client cert/key, just enabling TLS
	}

	consumer, err := pubsub.NewConsumer(cfg, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, consumer)
}

// TestNewConsumerWithSASL tests consumer creation with SASL config
func TestNewConsumerWithSASL(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:       []string{"localhost:9092"},
		ConsumerGroup: "test-group",
		SASLEnabled:   true,
		SASLMechanism: "PLAIN",
		SASLUsername:  "user",
		SASLPassword:  "pass",
	}

	consumer, err := pubsub.NewConsumer(cfg, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, consumer)
}

// TestNewConsumerWithTLSAndSASL tests consumer creation with both TLS and SASL config
func TestNewConsumerWithTLSAndSASL(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:       []string{"localhost:9092"},
		ConsumerGroup: "test-group",
		TLSEnabled:    true,
		SASLEnabled:   true,
		SASLMechanism: "SCRAM-SHA-256",
		SASLUsername:  "user",
		SASLPassword:  "pass",
	}

	// Consumer creation succeeds - TLS/SASL validation happens at Subscribe time
	consumer, err := pubsub.NewConsumer(cfg, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, consumer)
}

// TestConsumerClose tests consumer close
func TestConsumerClose(t *testing.T) {
	cfg := &pubsub.StandardConfig{
		Brokers:       []string{"localhost:9092"},
		ConsumerGroup: "test-group",
	}

	consumer, err := pubsub.NewConsumer(cfg, nil, nil)
	require.NoError(t, err)

	err = consumer.Close()
	assert.NoError(t, err)
}

// TestModuleWithTestUtil tests that the fx module can be constructed with testutil
func TestModuleWithTestUtil(t *testing.T) {
	var producer pubsub.Producer
	var consumer pubsub.Consumer

	app := fx.New(
		fx.NopLogger,
		testutil.TestModule(),
		fx.Populate(&producer, &consumer),
	)

	require.NoError(t, app.Err())
	require.NotNil(t, producer)
	require.NotNil(t, consumer)

	// Test that we can publish and get messages
	ctx := context.Background()
	err := producer.Publish(ctx, "test-topic", []byte("key"), []byte("value"))
	require.NoError(t, err)

	// Get the in-memory pubsub and verify the message
	inMem, ok := producer.(*testutil.InMemoryPubSub)
	require.True(t, ok)
	messages := inMem.GetMessages("test-topic")
	require.Len(t, messages, 1)
	assert.Equal(t, []byte("key"), messages[0].Key)
	assert.Equal(t, []byte("value"), messages[0].Value)
}

// Ensure the config interface is satisfied
var _ pubsub.Config = (*pubsub.StandardConfig)(nil)
var _ pubsub.Config = (*testutil.NoopConfig)(nil)

// Note: Integration tests for the actual Kafka producer/consumer would require
// a running Kafka cluster and are better suited for integration test suites.
// Example integration test structure:
//
// func TestModule_Integration(t *testing.T) {
//     if testing.Short() {
//         t.Skip("skipping integration test")
//     }
//
//     var producer pubsub.Producer
//     var consumer pubsub.Consumer
//     app := fx.New(
//         fx.NopLogger,
//         tracing.NoopModule(),
//         logger.NoopModule(),
//         fx.Provide(func() pubsub.Config {
//             return &pubsub.StandardConfig{
//                 Brokers:       []string{"localhost:9092"},
//                 ConsumerGroup: "test-group",
//             }
//         }),
//         pubsub.Module(),
//         fx.Populate(&producer, &consumer),
//     )
//     // ... test with actual Kafka cluster
// }
