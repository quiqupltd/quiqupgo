package tracing

import (
	"time"

	"go.opentelemetry.io/otel/sdk/trace"
)

// moduleOptions holds the configurable options for the tracing module.
type moduleOptions struct {
	batchTimeout   time.Duration
	metricInterval time.Duration
	sampler        trace.Sampler
}

// defaultModuleOptions returns the default module options.
func defaultModuleOptions() *moduleOptions {
	return &moduleOptions{
		batchTimeout:   5 * time.Second,
		metricInterval: 10 * time.Second,
		sampler:        nil, // Use SDK default (ParentBased(AlwaysSample))
	}
}

// ModuleOption is a functional option for configuring the tracing module.
type ModuleOption func(*moduleOptions)

// WithBatchTimeout sets the batch timeout for the trace exporter.
// Default is 5 seconds.
func WithBatchTimeout(d time.Duration) ModuleOption {
	return func(o *moduleOptions) {
		o.batchTimeout = d
	}
}

// WithMetricInterval sets the interval for metric export.
// Default is 10 seconds.
func WithMetricInterval(d time.Duration) ModuleOption {
	return func(o *moduleOptions) {
		o.metricInterval = d
	}
}

// WithSampler sets a custom sampler for the TracerProvider.
// Default is nil (SDK default: ParentBased(AlwaysSample)).
func WithSampler(s trace.Sampler) ModuleOption {
	return func(o *moduleOptions) {
		o.sampler = s
	}
}

// WithAlwaysSample configures the tracer to always sample spans.
func WithAlwaysSample() ModuleOption {
	return func(o *moduleOptions) {
		o.sampler = trace.AlwaysSample()
	}
}

// WithNeverSample configures the tracer to never sample spans.
func WithNeverSample() ModuleOption {
	return func(o *moduleOptions) {
		o.sampler = trace.NeverSample()
	}
}

// WithTraceIDRatioBased configures the tracer to sample a fraction of traces.
// fraction should be between 0 and 1.
func WithTraceIDRatioBased(fraction float64) ModuleOption {
	return func(o *moduleOptions) {
		o.sampler = trace.TraceIDRatioBased(fraction)
	}
}
