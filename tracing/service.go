package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// BaseService provides common tracing functionality that can be embedded in service structs.
// It offers a consistent way to create spans with automatic component name prefixing and
// proper error recording.
//
// The recommended pattern uses named return values with a deferred cleanup function,
// which ensures errors are properly captured at function exit time (not at defer registration).
//
// Basic usage with embedding:
//
//	type UserService struct {
//	    tracing.BaseService
//	    db *sql.DB
//	}
//
//	func NewUserService(tracer trace.Tracer, meter metric.Meter, db *sql.DB) *UserService {
//	    return &UserService{
//	        BaseService: tracing.NewBaseService(tracer, meter, "user.service"),
//	        db:          db,
//	    }
//	}
//
//	func (s *UserService) GetUser(ctx context.Context, id string) (user *User, err error) {
//	    ctx, end := s.Trace(ctx, "GetUser")
//	    defer end(&err)
//
//	    // Your logic here - errors are automatically recorded on the span
//	    return s.db.QueryUser(ctx, id)
//	}
//
// Using fx dependency injection:
//
//	func ProvideUserService(tracer trace.Tracer, meter metric.Meter, db *sql.DB) *UserService {
//	    return &UserService{
//	        BaseService: tracing.NewBaseService(tracer, meter, "user.service"),
//	        db:          db,
//	    }
//	}
//
// Alternative callback pattern with WithSpan (no named returns needed):
//
//	func (s *UserService) DeleteUser(ctx context.Context, id string) error {
//	    return s.WithSpan(ctx, "DeleteUser", func(ctx context.Context) error {
//	        return s.db.DeleteUser(ctx, id)
//	    })
//	}
type BaseService struct {
	tracer        trace.Tracer
	meter         metric.Meter
	componentName string
}

// NewBaseService creates a new BaseService with the given tracer, meter, and component name.
// The componentName is used as a prefix for all span names (e.g., "user.service.GetUser").
//
// Example:
//
//	base := tracing.NewBaseService(tracer, meter, "geocoder.domain")
func NewBaseService(tracer trace.Tracer, meter metric.Meter, componentName string) BaseService {
	return BaseService{
		tracer:        tracer,
		meter:         meter,
		componentName: componentName,
	}
}

// SpanEndFunc is a function that ends a span and records any error.
// It should be called with a pointer to the error return value.
type SpanEndFunc func(errPtr *error)

// Trace starts a new span with the component name prefixed to the operation name.
// It returns the context with the span and a cleanup function that should be deferred.
//
// IMPORTANT: Use named return values and pass a pointer to the error variable.
// This ensures the error is captured at function exit, not at defer registration.
//
// Correct usage:
//
//	func (s *Service) DoWork(ctx context.Context) (result string, err error) {
//	    ctx, end := s.Trace(ctx, "DoWork")
//	    defer end(&err)
//
//	    // errors assigned to 'err' will be recorded on the span
//	    result, err = s.actualWork(ctx)
//	    return result, err
//	}
//
// The cleanup function will:
//   - Record the error on the span (if non-nil)
//   - Set the span status to Error (if error is non-nil)
//   - End the span
func (s *BaseService) Trace(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, SpanEndFunc) {
	ctx, span := s.tracer.Start(ctx, fmt.Sprintf("%s.%s", s.componentName, name), opts...)

	return ctx, func(errPtr *error) {
		if errPtr != nil && *errPtr != nil {
			span.RecordError(*errPtr)
			span.SetStatus(codes.Error, (*errPtr).Error())
		}
		span.End()
	}
}

// WithSpan executes the given function within a new span.
// This is an alternative to the Trace pattern that doesn't require named returns.
//
// Example:
//
//	func (s *Service) ProcessItem(ctx context.Context, item Item) error {
//	    return s.WithSpan(ctx, "ProcessItem", func(ctx context.Context) error {
//	        // process the item
//	        return s.processor.Process(ctx, item)
//	    })
//	}
//
// For functions that return values, use [WithSpanResult] instead.
func (s *BaseService) WithSpan(ctx context.Context, name string, fn func(context.Context) error, opts ...trace.SpanStartOption) error {
	ctx, span := s.tracer.Start(ctx, fmt.Sprintf("%s.%s", s.componentName, name), opts...)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

// WithSpanResult executes the given function within a new span and returns its result.
// This is useful when you need to return a value along with an error.
//
// Example:
//
//	func (s *Service) FetchData(ctx context.Context, id string) (*Data, error) {
//	    return tracing.WithSpanResult(ctx, &s.BaseService, "FetchData",
//	        func(ctx context.Context) (*Data, error) {
//	            return s.repo.Get(ctx, id)
//	        })
//	}
func WithSpanResult[T any](ctx context.Context, s *BaseService, name string, fn func(context.Context) (T, error), opts ...trace.SpanStartOption) (T, error) {
	ctx, span := s.tracer.Start(ctx, fmt.Sprintf("%s.%s", s.componentName, name), opts...)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

// Tracer returns the underlying tracer for advanced use cases.
func (s *BaseService) Tracer() trace.Tracer {
	return s.tracer
}

// Meter returns the underlying meter for creating custom metrics.
//
// Example:
//
//	counter, _ := s.Meter().Int64Counter("requests.total")
//	counter.Add(ctx, 1)
func (s *BaseService) Meter() metric.Meter {
	return s.meter
}

// ComponentName returns the component name used for span prefixing.
func (s *BaseService) ComponentName() string {
	return s.componentName
}
