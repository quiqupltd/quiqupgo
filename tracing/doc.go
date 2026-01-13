// Package tracing provides an uber/fx module for OpenTelemetry tracing and metrics.
//
// It exports TracerProvider, Tracer, MeterProvider, and Meter through dependency injection.
// Configure it by providing an implementation of the Config interface.
//
// # Basic Module Usage
//
// Example fx application setup:
//
//	fx.New(
//	    fx.Provide(func() tracing.Config {
//	        return &tracing.StandardConfig{
//	            ServiceName:     "my-service",
//	            EnvironmentName: "production",
//	            OTLPEndpoint:    "otel-collector:4318",
//	        }
//	    }),
//	    tracing.Module(),
//	)
//
// # BaseService for Service Tracing
//
// BaseService provides a reusable foundation for adding tracing to your service structs.
// It offers consistent span naming, automatic error recording, and a clean API.
//
// # Embedding BaseService
//
// Embed BaseService in your service structs to gain tracing capabilities:
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
// # Tracing Methods with Named Returns (Recommended)
//
// The Trace method returns a context and cleanup function. Use named returns to ensure
// errors are properly captured at function exit, not at defer registration time:
//
//	func (s *UserService) GetUser(ctx context.Context, id string) (user *User, err error) {
//	    ctx, end := s.Trace(ctx, "GetUser")
//	    defer end(&err)  // Pass pointer to error - captures final value
//
//	    user, err = s.db.QueryUser(ctx, id)
//	    if err != nil {
//	        return nil, fmt.Errorf("query user: %w", err)
//	    }
//	    return user, nil
//	}
//
// The span will be named "user.service.GetUser" and any error will be automatically
// recorded on the span with proper status codes.
//
// # Why Use &err Instead of err?
//
// Go evaluates defer arguments at registration time, not execution time:
//
//	// WRONG: err is nil when defer is registered
//	defer end(err)
//
//	// CORRECT: &err is dereferenced when defer executes
//	defer end(&err)
//
// # Callback Pattern with WithSpan
//
// For simpler cases or when named returns aren't convenient, use WithSpan:
//
//	func (s *UserService) DeleteUser(ctx context.Context, id string) error {
//	    return s.WithSpan(ctx, "DeleteUser", func(ctx context.Context) error {
//	        return s.db.DeleteUser(ctx, id)
//	    })
//	}
//
// # Generic Pattern with WithSpanResult
//
// For functions returning values, use the generic WithSpanResult function:
//
//	func (s *UserService) CountUsers(ctx context.Context) (int64, error) {
//	    return tracing.WithSpanResult(ctx, &s.BaseService, "CountUsers",
//	        func(ctx context.Context) (int64, error) {
//	            return s.db.CountUsers(ctx)
//	        })
//	}
//
// # Adding Span Attributes
//
// Pass span options to add attributes or links:
//
//	func (s *UserService) GetUser(ctx context.Context, id string) (user *User, err error) {
//	    ctx, end := s.Trace(ctx, "GetUser",
//	        trace.WithAttributes(attribute.String("user.id", id)),
//	    )
//	    defer end(&err)
//	    // ...
//	}
//
// # Accessing the Meter
//
// Use the Meter() method to create custom metrics:
//
//	func NewUserService(tracer trace.Tracer, meter metric.Meter, db *sql.DB) (*UserService, error) {
//	    svc := &UserService{
//	        BaseService: tracing.NewBaseService(tracer, meter, "user.service"),
//	        db:          db,
//	    }
//
//	    // Create custom counter
//	    counter, err := svc.Meter().Int64Counter("user.requests.total")
//	    if err != nil {
//	        return nil, err
//	    }
//	    svc.requestCounter = counter
//
//	    return svc, nil
//	}
//
// # Complete Example with fx
//
//	type GeocodingService struct {
//	    tracing.BaseService
//	    client       *http.Client
//	    cacheCounter metric.Int64Counter
//	}
//
//	func ProvideGeocodingService(tracer trace.Tracer, meter metric.Meter) (*GeocodingService, error) {
//	    svc := &GeocodingService{
//	        BaseService: tracing.NewBaseService(tracer, meter, "geocoding.service"),
//	        client:      &http.Client{Timeout: 10 * time.Second},
//	    }
//
//	    counter, err := meter.Int64Counter("geocoding.cache")
//	    if err != nil {
//	        return nil, err
//	    }
//	    svc.cacheCounter = counter
//
//	    return svc, nil
//	}
//
//	func (s *GeocodingService) Geocode(ctx context.Context, address string) (result *Location, err error) {
//	    ctx, end := s.Trace(ctx, "Geocode",
//	        trace.WithAttributes(attribute.String("address", address)),
//	    )
//	    defer end(&err)
//
//	    // Check cache
//	    if loc, ok := s.checkCache(ctx, address); ok {
//	        s.cacheCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("result", "hit")))
//	        return loc, nil
//	    }
//	    s.cacheCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("result", "miss")))
//
//	    // Call external API
//	    return s.callGeocodingAPI(ctx, address)
//	}
//
// # Testing with NoopModule
//
// Use testutil.NoopModule() for testing without actual tracing:
//
//	func TestMyService(t *testing.T) {
//	    app := fx.New(
//	        fx.NopLogger,
//	        testutil.NoopModule(),
//	        fx.Provide(NewMyService),
//	        fx.Invoke(func(svc *MyService) {
//	            // test your service
//	        }),
//	    )
//	    // ...
//	}
package tracing
