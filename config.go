package fastecho

import (
	"github.com/ingka-group/fastecho/env"
	"github.com/ingka-group/fastecho/router"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Config serves as input configuration for fastecho.
type Config struct {
	ExtraEnvs           env.Map
	ValidationRegistrar func(v *router.Validator) error
	Routes              func(e *echo.Echo, r *router.Router) error
	ContextProps        any
	Opts                Opts
	Plugins             []Plugin
}

// Opts define configuration options for fastecho.
type Opts struct {
	Metrics      MetricsOpts
	Tracing      TracingOpts
	HealthChecks HealthChecksOpts
}

// MetricsOpts define configuration options for metrics.
type MetricsOpts struct {
	Skip bool
}

// TracingOpts define configuration options for tracing.
type TracingOpts struct {
	Skip        bool
	ServiceName string
}

// HealthChecksOpts define configuration options for health checks.
type HealthChecksOpts struct {
	Skip bool
	DB   *gorm.DB
}

type Plugin struct {
	ValidationRegistrar func(v *router.Validator) error
	Routes              func(e *echo.Echo, r *router.Router) error
}

func (c *Config) Use(p Plugin) {
	c.Plugins = append(c.Plugins, p)
}
