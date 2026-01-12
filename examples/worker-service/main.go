// Package main demonstrates a worker service using all available modules.
//
// This example shows a Temporal worker service with:
//   - OpenTelemetry tracing
//   - Structured logging with zap
//   - Temporal workflow client
//   - Kafka messaging (producer and consumer)
//
// Note: This example requires running Temporal and Kafka servers.
//
// Usage:
//
//	go run ./examples/worker-service
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/quiqupltd/quiqupgo/logger"
	"github.com/quiqupltd/quiqupgo/kafka"
	"github.com/quiqupltd/quiqupgo/temporal"
	"github.com/quiqupltd/quiqupgo/tracing"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		// Provide configurations
		fx.Provide(
			newTracingConfig,
			newLoggerConfig,
			newTemporalConfig,
			newKafkaConfig,
		),

		// Include modules
		tracing.Module(),
		logger.Module(),
		temporal.Module(),
		kafka.Module(),

		// Start the worker
		fx.Invoke(registerWorker),
	).Run()
}

// newTracingConfig creates the tracing configuration.
func newTracingConfig() tracing.Config {
	return &tracing.StandardConfig{
		ServiceName:     "worker-service",
		EnvironmentName: "development",
		OTLPEndpoint:    "", // Empty = disabled for demo
	}
}

// newLoggerConfig creates the logger configuration.
func newLoggerConfig() logger.Config {
	return &logger.StandardConfig{
		ServiceName: "worker-service",
		Environment: "development",
	}
}

// newTemporalConfig creates the Temporal configuration.
// Note: Requires a running Temporal server at localhost:7233
func newTemporalConfig() temporal.Config {
	return &temporal.StandardConfig{
		HostPort:  "localhost:7233",
		Namespace: "default",
	}
}

// newKafkaConfig creates the Kafka configuration.
// Note: Requires a running Kafka server at localhost:9092
func newKafkaConfig() kafka.Config {
	enableTracing := false // Disable for demo
	return &kafka.StandardConfig{
		Brokers:       []string{"localhost:9092"},
		ConsumerGroup: "worker-service",
		EnableTracing: &enableTracing,
	}
}

// registerWorker sets up the Temporal worker and Kafka consumer.
func registerWorker(
	lc fx.Lifecycle,
	c client.Client,
	producer kafka.Producer,
	consumer kafka.Consumer,
	log *zap.Logger,
) {
	// Create Temporal worker
	w := worker.New(c, "worker-task-queue", worker.Options{})

	// Register workflows and activities
	w.RegisterWorkflow(ExampleWorkflow)
	w.RegisterActivity(PublishEventActivity(producer, log))

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("starting worker service")

			// Start Temporal worker
			if err := w.Start(); err != nil {
				return fmt.Errorf("failed to start Temporal worker: %w", err)
			}
			log.Info("Temporal worker started", zap.String("queue", "worker-task-queue"))

			// Start Kafka consumer in background
			go func() {
				log.Info("starting Kafka consumer")
				if err := consumer.Subscribe(ctx, []string{"events"}, handleMessage(log)); err != nil {
					log.Error("consumer error", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("stopping worker service")
			w.Stop()
			return consumer.Close()
		},
	})
}

// ExampleWorkflow is a simple workflow that processes data and publishes events.
func ExampleWorkflow(ctx workflow.Context, input string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ExampleWorkflow started", "input", input)

	// Execute activity
	var result string
	err := workflow.ExecuteActivity(ctx,
		PublishEventActivity(nil, nil), // Activity will be properly injected at runtime
		input,
	).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	logger.Info("ExampleWorkflow completed", "result", result)
	return result, nil
}

// PublishEventActivity publishes an event to Kafka.
func PublishEventActivity(producer kafka.Producer, log *zap.Logger) func(ctx context.Context, data string) (string, error) {
	return func(ctx context.Context, data string) (string, error) {
		if producer == nil {
			return fmt.Sprintf("processed: %s", data), nil
		}

		// Publish event to Kafka
		if err := producer.Publish(ctx, "events", []byte("event-key"), []byte(data)); err != nil {
			log.Error("failed to publish event", zap.Error(err))
			return "", err
		}

		log.Info("event published", zap.String("topic", "events"), zap.String("data", data))
		return fmt.Sprintf("published: %s", data), nil
	}
}

// handleMessage handles messages from Kafka.
func handleMessage(log *zap.Logger) kafka.MessageHandler {
	return func(ctx context.Context, msg kafka.ConsumerMessage) error {
		log.Info("received message",
			zap.String("topic", msg.Topic),
			zap.Int64("offset", msg.Offset),
			zap.String("value", string(msg.Value)),
		)

		// Simulate processing
		time.Sleep(100 * time.Millisecond)

		return nil
	}
}
