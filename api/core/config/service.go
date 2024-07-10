package config

import (
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/ingka-group-digital/ocp-go-utils/echozap"
)

// ServiceConfig is the struct that contains the service configuration.
type ServiceConfig struct {
	Logger   *zap.Logger
	Database *gorm.DB
	Tracer   *trace.Tracer
}

// NewServiceConfig creates a new ServiceConfig.
func NewServiceConfig(extraEnvVars *EnvVar, withPostgres bool) (*ServiceConfig, *sdktrace.TracerProvider, error) {
	// Set environment variables MUST be the first step
	err := setEnv(extraEnvVars, withPostgres)
	if err != nil {
		return nil, nil, err
	}

	logger, err := echozap.New()
	if err != nil {
		return nil, nil, err
	}

	var db *gorm.DB

	if withPostgres {
		DbConfig, err := NewDBConfig()
		if err != nil {
			return nil, nil, err
		}

		db, err := NewDB(DbConfig)
		if err != nil {
			return nil, nil, err
		}

		sqlDb, err := db.DB()
		if err != nil {
			return nil, nil, err
		}

		err = migrateDB(sqlDb, logger)
		if err != nil {
			return nil, nil, err
		}
	}

	// only init tracing if it's not disabled
	var tracerProvider *sdktrace.TracerProvider
	var tracer *trace.Tracer
	if Env[otelTracing].Value == "true" {
		tracerProvider, tracer, err = newTracer()
		if err != nil {
			return nil, nil, err
		}
	}

	return &ServiceConfig{
		Logger:   logger,
		Database: db,
		Tracer:   tracer,
	}, tracerProvider, nil
}
