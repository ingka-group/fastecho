package config

import (
	"maps"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/ingka-group-digital/ocp-go-utils/api/core/env"
	"github.com/ingka-group-digital/ocp-go-utils/echozap"
)

type Options struct {
	SkipMetrics  bool
	SkipSwagger  bool
	SkipTracing  bool
	HealthChecks struct {
		Skip bool
		DB   *gorm.DB
	}
}

// ServiceConfig is the struct that contains the service configuration.
type ServiceConfig struct {
	Logger   *zap.Logger
	Database *gorm.DB
	Tracer   *trace.Tracer
	Env      env.EnvVars
}

// NewServiceConfig creates a new ServiceConfig.
func NewServiceConfig(extraEnvVars env.EnvVars, opts Options) (*ServiceConfig, *sdktrace.TracerProvider, error) {
	// Set environment variables MUST be the first step
	// merge default env vars with extra env vars
	var EnvVar = make(env.EnvVars)
	maps.Copy(EnvVar, env.DefaultEnvVars)
	maps.Copy(EnvVar, extraEnvVars)
	if !opts.SkipTracing {
		EnvVar[env.OtelTracing] = env.EnvVar{
			Value: "true",
		}
	}
	definedEnvs, err := env.SetEnv(EnvVar)
	if err != nil {
		return nil, nil, err
	}

	logger, err := echozap.New()
	if err != nil {
		return nil, nil, err
	}

	// only init tracing if it's not disabled
	var tracerProvider *sdktrace.TracerProvider
	var tracer *trace.Tracer
	// recheck because env var might have been overridden by env file
	if !opts.SkipTracing && definedEnvs[env.OtelTracing].Value == "true" {
		tracerProvider, tracer, err = newTracer(definedEnvs[env.ServiceName].Value)
		if err != nil {
			return nil, nil, err
		}
	}

	return &ServiceConfig{
		Logger: logger,
		Tracer: tracer,
		Env:    definedEnvs,
	}, tracerProvider, nil
}
