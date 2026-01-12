package gormfx_test

import (
	"database/sql"
	"testing"

	"github.com/quiqupltd/quiqupgo/gormfx"
	"github.com/quiqupltd/quiqupgo/gormfx/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // SQLite driver for testing
)

func TestStandardConfig(t *testing.T) {
	enableTracing := true
	cfg := &gormfx.StandardConfig{
		DB:            nil, // Would normally be a real *sql.DB
		MaxOpenConns:  25,
		MaxIdleConns:  5,
		EnableTracing: &enableTracing,
	}

	assert.Nil(t, cfg.GetDB())
	assert.Equal(t, 25, cfg.GetMaxOpenConns())
	assert.Equal(t, 5, cfg.GetMaxIdleConns())
	assert.True(t, cfg.GetEnableTracing())
}

func TestStandardConfig_Defaults(t *testing.T) {
	cfg := &gormfx.StandardConfig{}

	assert.Nil(t, cfg.GetDB())
	assert.Equal(t, 0, cfg.GetMaxOpenConns())
	assert.Equal(t, 0, cfg.GetMaxIdleConns())
	// EnableTracing defaults to true when nil
	assert.True(t, cfg.GetEnableTracing())
}

func TestStandardConfig_TracingDisabled(t *testing.T) {
	enableTracing := false
	cfg := &gormfx.StandardConfig{
		EnableTracing: &enableTracing,
	}

	assert.False(t, cfg.GetEnableTracing())
}

func TestNoopConfig(t *testing.T) {
	cfg := testutil.NewNoopConfig()

	assert.Nil(t, cfg.GetDB())
	assert.Equal(t, 0, cfg.GetMaxOpenConns())
	assert.Equal(t, 0, cfg.GetMaxIdleConns())
	assert.False(t, cfg.GetEnableTracing())
}

func TestNewTestDB(t *testing.T) {
	db, err := testutil.NewTestDB()
	require.NoError(t, err)
	require.NotNil(t, db)

	// Verify we can execute queries
	type TestModel struct {
		ID   uint
		Name string
	}

	err = db.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	// Create a record
	err = db.Create(&TestModel{Name: "test"}).Error
	require.NoError(t, err)

	// Read it back
	var result TestModel
	err = db.First(&result).Error
	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
}

func TestTestModule(t *testing.T) {
	var db *gorm.DB

	app := fx.New(
		fx.NopLogger,
		testutil.TestModule(),
		fx.Populate(&db),
	)

	require.NoError(t, app.Err())
	require.NotNil(t, db)

	// Verify we got a working database
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NotNil(t, sqlDB)

	err = sqlDB.Ping()
	require.NoError(t, err)
}

// Note: Integration tests for the actual GORM module with PostgreSQL would require
// a running database and are better suited for integration test suites.
// Example integration test structure:
//
// func TestModule_Integration(t *testing.T) {
//     if testing.Short() {
//         t.Skip("skipping integration test")
//     }
//
//     // Connect to a real PostgreSQL database
//     dsn := "host=localhost user=test password=test dbname=test port=5432 sslmode=disable"
//     sqlDB, err := sql.Open("postgres", dsn)
//     require.NoError(t, err)
//     defer sqlDB.Close()
//
//     var db *gorm.DB
//     app := fx.New(
//         fx.NopLogger,
//         tracing.NoopModule(),
//         fx.Provide(func() gormfx.Config {
//             return &gormfx.StandardConfig{
//                 DB:            sqlDB,
//                 MaxOpenConns:  25,
//                 MaxIdleConns:  5,
//                 EnableTracing: ptr(true),
//             }
//         }),
//         gormfx.Module(),
//         fx.Populate(&db),
//     )
//     // ... test with actual database
// }

// Ensure the config interface is satisfied
var _ gormfx.Config = (*gormfx.StandardConfig)(nil)
var _ gormfx.Config = (*testutil.NoopConfig)(nil)

// ptr is a helper for creating pointers to literals
func ptr[T any](v T) *T {
	return &v
}

func TestNewDB_NilDB(t *testing.T) {
	cfg := &gormfx.StandardConfig{
		DB: nil,
	}

	_, err := gormfx.NewDB(cfg, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sql.DB is required")
}

// TestNewDB_WithRealDB tests NewDB with a real SQLite connection
func TestNewDB_WithRealDB(t *testing.T) {
	// Open an in-memory SQLite database as sql.DB
	sqlDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	cfg := &gormfx.StandardConfig{
		DB:            sqlDB,
		MaxOpenConns:  10,
		MaxIdleConns:  2,
		EnableTracing: ptr(false), // Disable tracing for this test
	}

	db, err := gormfx.NewDB(cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Verify the connection works
	type TestModel struct {
		ID   uint
		Name string
	}

	err = db.AutoMigrate(&TestModel{})
	require.NoError(t, err)
}

func TestModule(t *testing.T) {
	// Open an in-memory SQLite database as sql.DB
	sqlDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	var db *gorm.DB

	app := fx.New(
		fx.NopLogger,
		testutil.NoopTracerProviderModule(),
		fx.Provide(func() gormfx.Config {
			return &gormfx.StandardConfig{
				DB:            sqlDB,
				MaxOpenConns:  10,
				MaxIdleConns:  2,
				EnableTracing: ptr(false),
			}
		}),
		gormfx.Module(),
		fx.Populate(&db),
	)

	require.NoError(t, app.Err())
	require.NotNil(t, db)

	// Start and stop to test lifecycle hooks
	ctx := t.Context()
	require.NoError(t, app.Start(ctx))
	require.NoError(t, app.Stop(ctx))
}

func TestModule_WithTracingEnabled(t *testing.T) {
	// Open an in-memory SQLite database as sql.DB
	sqlDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	var db *gorm.DB

	app := fx.New(
		fx.NopLogger,
		testutil.NoopTracerProviderModule(),
		fx.Provide(func() gormfx.Config {
			return &gormfx.StandardConfig{
				DB:            sqlDB,
				EnableTracing: ptr(true), // Tracing enabled
			}
		}),
		gormfx.Module(),
		fx.Populate(&db),
	)

	require.NoError(t, app.Err())
	require.NotNil(t, db)

	ctx := t.Context()
	require.NoError(t, app.Start(ctx))

	// Execute a query to verify tracing doesn't break things
	type TestModel struct {
		ID   uint
		Name string
	}
	err = db.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	require.NoError(t, app.Stop(ctx))
}
