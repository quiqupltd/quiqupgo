package tracing

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
)

// TracingModule holds all the tracing components.
type TracingModule struct {
	TracerProvider *trace.TracerProvider
	Tracer         oteltrace.Tracer
	MeterProvider  *sdkmetric.MeterProvider
	Meter          metric.Meter
}

// moduleOptionSlice is a wrapper to allow fx.Supply of []ModuleOption.
type moduleOptionSlice []ModuleOption

// Module returns an fx.Option that provides OpenTelemetry tracing and metrics.
//
// It provides:
//   - trace.TracerProvider
//   - trace.Tracer
//   - metric.MeterProvider
//   - metric.Meter
//
// It requires:
//   - tracing.Config (must be provided by the application)
func Module(opts ...ModuleOption) fx.Option {
	return fx.Module("tracing",
		fx.Supply(moduleOptionSlice(opts)),
		fx.Provide(
			newTracingModule,
			provideTracerProvider,
			provideTracer,
			provideMeterProvider,
			provideMeter,
		),
		fx.Invoke(registerLifecycleHooks),
	)
}

// newTracingModule creates the TracingModule with all components.
func newTracingModule(lc fx.Lifecycle, cfg Config, opts moduleOptionSlice) (*TracingModule, error) {
	ctx := context.Background()

	// Create TracerProvider
	tp, err := GetTracerProvider(ctx, cfg, opts...)
	if err != nil {
		return nil, err
	}

	// Create MeterProvider
	mp, err := GetMeterProvider(ctx, cfg, opts...)
	if err != nil {
		return nil, err
	}

	// Get Tracer and Meter
	tracer := GetTracer(tp)
	meter := GetMeter(mp)

	return &TracingModule{
		TracerProvider: tp,
		Tracer:         tracer,
		MeterProvider:  mp,
		Meter:          meter,
	}, nil
}

// provideTracerProvider extracts TracerProvider as an interface.
func provideTracerProvider(tm *TracingModule) oteltrace.TracerProvider {
	if tm.TracerProvider == nil {
		// Return no-op provider if not configured
		return tracenoop.NewTracerProvider()
	}
	return tm.TracerProvider
}

// provideTracer extracts Tracer.
func provideTracer(tm *TracingModule) oteltrace.Tracer {
	return tm.Tracer
}

// provideMeterProvider extracts MeterProvider as an interface.
func provideMeterProvider(tm *TracingModule) metric.MeterProvider {
	if tm.MeterProvider == nil {
		// Return no-op provider if not configured
		return metricnoop.NewMeterProvider()
	}
	return tm.MeterProvider
}

// provideMeter extracts Meter.
func provideMeter(tm *TracingModule) metric.Meter {
	return tm.Meter
}

// registerLifecycleHooks registers shutdown hooks for graceful cleanup.
func registerLifecycleHooks(lc fx.Lifecycle, tm *TracingModule) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// Shutdown TracerProvider
			if tm.TracerProvider != nil {
				if err := ShutdownTracerProvider(ctx, tm.TracerProvider); err != nil {
					// Log error but don't fail shutdown
					_ = err
				}
			}

			// Shutdown MeterProvider
			if tm.MeterProvider != nil {
				if err := ShutdownMeterProvider(ctx, tm.MeterProvider); err != nil {
					// Log error but don't fail shutdown
					_ = err
				}
			}

			return nil
		},
	})
}
